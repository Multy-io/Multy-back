/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Multy-io/Multy-back/btc"
	"github.com/Multy-io/Multy-back/currencies"
	"github.com/Multy-io/Multy-back/eth"
	btcpb "github.com/Multy-io/Multy-back/node-streamer/btc"
	ethpb "github.com/Multy-io/Multy-back/node-streamer/eth"
	"github.com/Multy-io/Multy-back/store"

	"github.com/gin-gonic/gin"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	_ "github.com/jekabolt/slflog"
	"github.com/mitchellh/mapstructure"
)

const (
	socketIOOutMsg = "outcoming"
	socketIOInMsg  = "incoming"

	deviceTypeMac     = "mac"
	deviceTypeAndroid = "android"

	topicExchangeDay      = "exchangeDay"
	topicExchangeGdax     = "exchangeGdax"
	topicExchangePoloniex = "exchangePoloniex"
)

const (
	WirelessRoom = "wireless"

	ReceiverOn = "event:receiver:on"
	SenderOn   = "event:sender:on"

	SenderCheck = "event:sender:check"

	Filter = "event:filter"

	// wireless send
	NewReceiver     = "event:new:receiver"
	SendRaw         = "event:sendraw"
	PaymentSend     = "event:payment:send"
	PaymentReceived = "event:payment:received"

	stopReceive = "receiver:stop"
	stopSend    = "sender:stop"

	// multisig
	joinMultisig   = "join:multisig"
	leaveMultisig  = "leave:multisig"
	deleteMultisig = "delete:multisig"
	kickMultisig   = "kick:multisig"

	updateMultisig  = "update:multisig"
	deletedMultisig = "deleted:multisig"
	okMultisig      = "ok:multisig"
	errMultisig     = "err:multisig"

	msgSend    = "message:send"
	msgRecieve = "message:recieve"
)

func getHeaderDataSocketIO(headers http.Header) (*SocketIOUser, error) {
	userID := headers.Get("userID")
	if len(userID) == 0 {
		return nil, fmt.Errorf("wrong userID header")
	}

	deviceType := headers.Get("deviceType")
	if len(deviceType) == 0 {
		return nil, fmt.Errorf("wrong deviceType header")
	}

	jwtToken := headers.Get("jwtToken")
	if len(jwtToken) == 0 {
		return nil, fmt.Errorf("wrong jwtToken header")
	}

	return &SocketIOUser{
		userID:     userID,
		deviceType: deviceType,
		jwtToken:   jwtToken,
	}, nil
}

func SetSocketIOHandlers(restClient *RestClient, BTC *btc.BTCConn, ETH *eth.ETHConn, r *gin.RouterGroup, address, nsqAddr string, ratesDB store.UserStore) (*SocketIOConnectedPool, error) {
	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
	pool, err := InitConnectedPool(server, address, nsqAddr, ratesDB)
	if err != nil {
		return nil, fmt.Errorf("connection pool initialization: %s", err.Error())
	}

	chart, err := newExchangeChart(ratesDB)
	if err != nil {
		return nil, fmt.Errorf("exchange chart initialization: %s", err.Error())
	}
	pool.chart = chart

	receivers := make(map[string]store.Receiver)
	receiversM := sync.Mutex{}

	senders := []store.Sender{}

	server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		user, err := getHeaderDataSocketIO(c.RequestHeader())
		if err != nil {
			pool.log.Errorf("get socketio headers: %s", err.Error())
			return
		}
		user.pool = pool
		connectionID := c.Id()
		user.chart = pool.chart

		pool.m.Lock()
		defer pool.m.Unlock()
		userFromPool, ok := pool.users[user.userID]
		if !ok {
			pool.log.Debugf("new user")
			newSocketIOUser(connectionID, user, c, pool.log)
			pool.users[user.userID] = user
			userFromPool = user
		}

		userFromPool.conns[connectionID] = c
		pool.closeChByConnID[connectionID] = userFromPool.closeCh

		sendExchange(user, c)
		pool.log.Debugf("OnConnection done")
	})

	server.On(gosocketio.OnError, func(c *gosocketio.Channel) {
		pool.log.Errorf("Error occurs %s", c.Id())
	})

	//feature logic
	server.On(ReceiverOn, func(c *gosocketio.Channel, data store.Receiver) string {
		pool.log.Infof("Got messeage Receiver On:", data)
		c.Join(WirelessRoom)
		receiver := store.Receiver{
			ID:         data.ID,
			UserCode:   data.UserCode,
			CurrencyID: data.CurrencyID,
			NetworkID:  data.NetworkID,
			Amount:     data.Amount,
			Address:    data.Address,
			Socket:     c,
		}

		receiversM.Lock()
		_, ok := receivers[receiver.UserCode]
		if !ok {
			receivers[receiver.UserCode] = receiver
		}
		receiversM.Unlock()

		return "ok"
	})

	server.On(SenderCheck, func(c *gosocketio.Channel, nearIDs store.NearVisible) []store.Receiver {
		pool.log.Warnf("SenderCheck")
		nearReceivers := []store.Receiver{}
		receiversM.Lock()
		allR := receivers
		receiversM.Unlock()

		for _, id := range nearIDs.IDs {
			if res, ok := allR[id]; ok {
				nearReceivers = append(nearReceivers, res)
			}
		}
		c.Emit(SenderCheck, nearReceivers)
		return nearReceivers
	})

	server.On(SendRaw, func(c *gosocketio.Channel, raw store.RawHDTx) string {
		switch raw.CurrencyID {
		case currencies.Bitcoin:
			var resp *btcpb.ReplyInfo

			if raw.NetworkID == currencies.Test {
				resp, err = BTC.CliTest.EventSendRawTx(context.Background(), &btcpb.RawTx{
					Transaction: raw.Transaction,
				})
			}
			if raw.NetworkID == currencies.Main {
				resp, err = BTC.CliMain.EventSendRawTx(context.Background(), &btcpb.RawTx{
					Transaction: raw.Transaction,
				})
			}

			if err != nil {
				pool.log.Errorf("sendRawHDTransaction: restClient.BTC.CliMain.EventSendRawTx: %s", err.Error())
				c.Emit(SendRaw, err.Error())
				return err.Error()
			}

			if strings.Contains("err:", resp.GetMessage()) {
				pool.log.Errorf("sendRawHDTransaction: restClient.BTC.CliMain.EventSendRawTx:resp err %s", err.Error())
				c.Emit(SendRaw, resp.GetMessage())
				return err.Error()
			}

			if raw.IsHD && !strings.Contains("err:", resp.GetMessage()) {
				err = addAddressToWallet(raw.Address, raw.JWT, raw.CurrencyID, raw.NetworkID, raw.WalletIndex, raw.AddressIndex, restClient, nil)
				if err != nil {
					pool.log.Errorf("addAddressToWallet: %v", err.Error())
				}
				c.Emit(SendRaw, resp.GetMessage())
				receiversM.Lock()
				res := receivers[raw.UserCode]
				receiversM.Unlock()
				res.Socket.Emit(PaymentReceived, raw)
			}

			return "success:" + resp.GetMessage()

		case currencies.Ether:
			if raw.NetworkID == currencies.ETHMain {
				h, err := restClient.ETH.CliMain.EventSendRawTx(context.Background(), &ethpb.RawTx{
					Transaction: raw.Transaction,
				})
				if err != nil {
					pool.log.Errorf("sendRawHDTransaction:eth.SendRawTransaction %s", err.Error())
					return err.Error()
				}

				if strings.Contains("err:", h.GetMessage()) {
					pool.log.Errorf("sendRawHDTransaction: strings.Contains err: %s", err.Error())
					return err.Error()
				}
				return "success:" + h.GetMessage()
			}
			if raw.NetworkID == currencies.ETHTest {
				h, err := restClient.ETH.CliMain.EventSendRawTx(context.Background(), &ethpb.RawTx{
					Transaction: raw.Transaction,
				})
				if err != nil {
					pool.log.Errorf("sendRawHDTransaction:eth.SendRawTransaction %s", err.Error())
					return err.Error()
				}

				if strings.Contains("err:", h.GetMessage()) {
					pool.log.Errorf("sendRawHDTransaction: strings.Contains err: %s", err.Error())
					return err.Error()
				}
				return "success:" + h.GetMessage()
			}

		}
		return "err: no such curid or netid"
	})

	server.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		pool.log.Infof("Disconnected %s", c.Id())
		pool.removeUserConn(c.Id())
		for _, receiver := range receivers {
			if receiver.Socket.Id() == c.Id() {
				delete(receivers, receiver.UserCode)
				continue
			}
		}
		for i, sender := range senders {
			if sender.Socket.Id() == c.Id() {
				senders = append(senders[:i], senders[i+1:]...)
				continue
			}
		}
	})

	server.On(stopReceive, func(c *gosocketio.Channel) string {
		pool.log.Infof("Stop receive %s", c.Id())
		for _, receiver := range receivers {
			if receiver.Socket.Id() == c.Id() {
				delete(receivers, receiver.UserCode)
				continue
			}
		}

		receiversM.Lock()
		fmt.Println("stopReceive", receivers)
		receiversM.Unlock()
		return stopReceive + ":ok"
	})

	server.On(stopSend, func(c *gosocketio.Channel) string {
		pool.log.Infof("Stop send %s", c.Id())
		for i, sender := range senders {
			if sender.Socket.Id() == c.Id() {
				senders = append(senders[:i], senders[i+1:]...)
				continue
			}
		}
		return stopSend + ":ok"

	})

	server.On(msgSend, func(c *gosocketio.Channel, msg store.WsMessage) interface{} {
		switch msg.Type {
		case joinMultisig:
			msgMultisig := &store.MultisigMsg{}
			err := mapstructure.Decode(msg.Payload, msgMultisig)
			if err != nil {
				pool.log.Errorf("server.On:msgSend:joinMultisig:mapstructure.Decode %v", err.Error())
				return makeErr(msgMultisig.UserID, "can't join multisig: bad request: "+err.Error())
			}

			// if invite code exists
			if !ratesDB.CheckInviteCode(msgMultisig.InviteCode) {
				// check current multisig for able to joining
				multisig, msg, err := getMultisig(ratesDB, msgMultisig)
				if err != nil {
					pool.log.Errorf("server.On:msgSend:joinMultisig %v", err.Error())
					return msg
				}

				joined := len(multisig.Owners)

				if multisig.OwnersCount > joined {

					multisigToJoin := multisig
					owners := []store.AddressExtended{}
					for _, owner := range multisigToJoin.Owners {
						owners = append(owners, store.AddressExtended{
							Address: owner.Address,
							Creator: owner.Creator,
						})
					}

					owners = append(owners, store.AddressExtended{
						UserID:       msgMultisig.UserID,
						Address:      msgMultisig.Address,
						WalletIndex:  msgMultisig.WalletIndex,
						AddressIndex: 0,
						Associated:   true,
					})

					multisigToJoin.Owners = owners
					users := ratesDB.FindMultisigUsers(msgMultisig.InviteCode)
					err = ratesDB.JoinMultisig(msgMultisig.UserID, multisigToJoin)
					if err != nil {
						//db err
						pool.log.Errorf("server.On:MultisigMsgratesDB.MultisigMsg: %v", err.Error())
						return makeErr(msgMultisig.UserID, "can't join multisig: "+err.Error())
					}

					// send new multisig entitiy to all online owners by ws
					for _, user := range users {
						userMultisig, err := updateUserOwners(user, multisig, ratesDB)
						if err != nil {
							//db err
							pool.log.Errorf("server.On:MultisigMsgratesDB.MultisigMsg: %v", err.Error())
						}
						_, online := pool.users[user.UserID]
						if online {
							msg := store.WsMessage{
								Type:    updateMultisig,
								To:      user.UserID,
								Date:    time.Now().Unix(),
								Payload: userMultisig,
							}
							server.BroadcastTo("message", msgRecieve+":"+user.UserID, msg)
						}
					}

					return store.WsMessage{
						Type:    okMultisig,
						To:      msgMultisig.UserID,
						Date:    time.Now().Unix(),
						Payload: joinMultisig + ":ok",
					}
				}

			}

			return makeErr("", "wrong request payload")
		case leaveMultisig:
			msgMultisig := &store.MultisigMsg{}
			err := mapstructure.Decode(msg.Payload, msgMultisig)
			if err != nil {
				pool.log.Errorf("server.On:msgSend:leaveMultisig:mapstructure.Decode %v", err.Error())
				return makeErr(msgMultisig.UserID, "can't leave multisig: bad request: "+err.Error())
			}
			if !ratesDB.CheckInviteCode(msgMultisig.InviteCode) {
				multisig, msg, err := getMultisig(ratesDB, msgMultisig)
				if err != nil {
					pool.log.Errorf("server.On:msgSend:leaveMultisigч %v", err.Error())
					return msg
				}
				exists := false
				for _, owner := range multisig.Owners {
					if owner.Address == msgMultisig.Address {
						exists = true
						if owner.Creator {
							pool.log.Errorf("server.On:leaveMultisig: can't leave multisig if you are creator you need delete it")
							return makeErr(msgMultisig.UserID, "can't leave multisig if you are creator you need delete it ")
						}
					}
				}

				owners := []store.AddressExtended{}
				if exists {
					//delete multisig from user
					err := ratesDB.LeaveMultisig(msgMultisig.UserID, msgMultisig.InviteCode)
					if err != nil {
						pool.log.Errorf("server.On:leaveMultisig:ratesDB.LeaveMultisig : %v", err.Error())
						return makeErr(msgMultisig.UserID, "can't leave multisig: "+err.Error())
					}

					users := ratesDB.FindMultisigUsers(msgMultisig.InviteCode)

					//delete owner from owners list
					for _, owner := range multisig.Owners {
						if owner.Address != msgMultisig.Address {
							owners = append(owners, owner)
						}
					}
					multisig.Owners = owners

					for _, user := range users {
						userMultisig, err := updateUserOwners(user, multisig, ratesDB)
						if err != nil {
							//db err
							pool.log.Errorf("server.On:MultisigMsgratesDB.MultisigMsg: %v", err.Error())
						}

						_, online := pool.users[user.UserID]
						if online {
							msg := store.WsMessage{
								Type:    updateMultisig,
								To:      user.UserID,
								Date:    time.Now().Unix(),
								Payload: userMultisig,
							}
							server.BroadcastTo("message", msgRecieve+":"+user.UserID, msg)
						}
					}

					return store.WsMessage{
						Type:    okMultisig,
						To:      msgMultisig.UserID,
						Date:    time.Now().Unix(),
						Payload: leaveMultisig + ":ok",
					}

				}

				if !exists {
					pool.log.Errorf("server.On:leaveMultisig: can't leave multisig if you are not a owner")
					return makeErr(msgMultisig.UserID, "can't leave multisig if you are not a owner")
				}

			}

			return makeErr("", "wrong request payload: ")
		case kickMultisig:
			msgMultisig := &store.MultisigMsg{}
			err := mapstructure.Decode(msg.Payload, msgMultisig)
			if err != nil {
				pool.log.Errorf("server.On:msgSend:kickMultisig:mapstructure.Decode %v", err.Error())
				return makeErr(msgMultisig.UserID, "can't kik from multisig: bad request: "+err.Error())
			}
			if !ratesDB.CheckInviteCode(msgMultisig.InviteCode) {
				multisig, msg, err := getMultisig(ratesDB, msgMultisig)
				if err != nil {
					pool.log.Errorf("server.On:msgSend:joinMultisig %v", err.Error())
					return msg
				}

				admin := false
				for _, owner := range multisig.Owners {
					if owner.Address == msgMultisig.Address {
						if owner.Creator {
							admin = true
						}
					}
				}

				if !admin {
					pool.log.Errorf("server.On:kickMultisig: only admin can kik form ms: %v", err.Error())
					return makeErr(msgMultisig.UserID, "only admin can kik form ms")
				}

				owners := []store.AddressExtended{}
				if admin {
					//delete multisig from user
					err := ratesDB.KickMultisig(msgMultisig.AddressToKick, msgMultisig.InviteCode)
					if err != nil {
						pool.log.Errorf("server.On:kickMultisig:ratesDB.KickMultisig: %v", err.Error())
						return makeErr(msgMultisig.UserID, "can't kik from multisig: "+err.Error())
					}

					users := ratesDB.FindMultisigUsers(msgMultisig.InviteCode)

					//delete owner from owners list
					for _, owner := range multisig.Owners {
						if owner.Address != msgMultisig.AddressToKick {
							owners = append(owners, owner)
						}
					}
					multisig.Owners = owners

					for _, user := range users {
						userMultisig, err := updateUserOwners(user, multisig, ratesDB)
						if err != nil {
							//db err
							pool.log.Errorf("server.On:kickMultisig:updateUserOwners: %v", err.Error())
						}

						_, online := pool.users[user.UserID]
						if online {
							msg := store.WsMessage{
								Type:    updateMultisig,
								To:      user.UserID,
								Date:    time.Now().Unix(),
								Payload: userMultisig,
							}
							server.BroadcastTo("message", msgRecieve+":"+user.UserID, msg)
						}
					}

					return store.WsMessage{
						Type:    okMultisig,
						To:      msgMultisig.UserID,
						Date:    time.Now().Unix(),
						Payload: kickMultisig + ":ok",
					}

				}

			}

			return makeErr("", "wrong request payload")
		case deleteMultisig:
			msgMultisig := &store.MultisigMsg{}
			err := mapstructure.Decode(msg.Payload, msgMultisig)
			if err != nil {
				pool.log.Errorf("server.On:msgSend:deleteMultisig:mapstructure.Decode %v", err.Error())
				return makeErr(msgMultisig.UserID, "can't kik from multisig: bad request: "+err.Error())
			}
			if !ratesDB.CheckInviteCode(msgMultisig.InviteCode) {
				multisig, msg, err := getMultisig(ratesDB, msgMultisig)
				if err != nil {
					pool.log.Errorf("server.On:msgSend:joinMultisig %v", err.Error())
					return msg
				}

				admin := false
				for _, owner := range multisig.Owners {
					if owner.Address == msgMultisig.Address && owner.Creator {
						admin = true
					}
				}
				if !admin {
					pool.log.Errorf("server.On:deleteMultisig: only creator can kik form ms")
					return makeErr(msgMultisig.UserID, "only creator can kik form ms")
				}

				if admin {
					err := ratesDB.DeleteMultisig(msgMultisig.InviteCode)
					if err != nil {
						pool.log.Errorf("server.On:deleteMultisig:DeleteMultisig %v", err.Error())
						return makeErr(msgMultisig.UserID, "server.On:deleteMultisig:DeleteMultisig "+err.Error())
					}

					users := ratesDB.FindMultisigUsers(msgMultisig.InviteCode)

					for _, user := range users {
						_, online := pool.users[user.UserID]
						if online {
							msg := store.WsMessage{
								Type:    deletedMultisig,
								To:      user.UserID,
								Date:    time.Now().Unix(),
								Payload: msgMultisig.InviteCode,
							}
							server.BroadcastTo("message", msgRecieve+":"+user.UserID, msg)
						}
					}
					return store.WsMessage{
						Type:    okMultisig,
						To:      msgMultisig.UserID,
						Date:    time.Now().Unix(),
						Payload: deleteMultisig + ":ok",
					}
				}

				if !admin {
					pool.log.Errorf("server.On:deleteMultisig: can't delete multisig if you are not a creator")
					return makeErr(msgMultisig.UserID, "can't delete multisig if you are not a creator")
				}
			}

			return makeErr("", "wrong request payload: ")
		}
		return makeErr("", "wrong request message type: ")
	})

	server.On(msgRecieve, func(c *gosocketio.Channel, msg store.WsMessage) string {
		return ""
	})

	server.On("kek", func(c *gosocketio.Channel, msg string) string {
		fmt.Println("msg:", msg)
		return "ok:" + msg
	})

	server.On("kek1", func(c *gosocketio.Channel) string {
		return "ok"
	})

	serveMux := http.NewServeMux()
	serveMux.Handle("/socket.io/", server)

	pool.log.Infof("Starting socketIO server on %s address", address)
	go func() {
		pool.log.Panicf("%s", http.ListenAndServe(address, serveMux))
	}()
	return pool, nil
}

func getMultisig(uStore store.UserStore, msgMultisig *store.MultisigMsg) (*store.Multisig, store.WsMessage, error) {
	msg := store.WsMessage{}
	multisig, err := uStore.FindMultisig(msgMultisig.UserID, msgMultisig.InviteCode)
	if err != nil {
		msg = store.WsMessage{
			Type:    errMultisig,
			To:      msgMultisig.UserID,
			Date:    time.Now().Unix(),
			Payload: "can't join multisig: " + err.Error(),
		}
	}
	return multisig, msg, err
}

func makeErr(userid, errorStr string) store.WsMessage {
	return store.WsMessage{
		Type:    errMultisig,
		To:      userid,
		Date:    time.Now().Unix(),
		Payload: errorStr,
	}
}

func updateUserOwners(user store.User, multisig *store.Multisig, uStore store.UserStore) (*store.Multisig, error) {
	// Clean addttion tags
	owners := []store.AddressExtended{}
	for _, owner := range multisig.Owners {
		owners = append(owners, store.AddressExtended{
			Address: owner.Address,
			Creator: owner.Creator,
		})
	}
	fetchedOwners := []store.AddressExtended{}
	for _, wallet := range user.Wallets {
		for _, address := range wallet.Adresses {
			for _, owner := range owners {
				if owner.Address == address.Address {
					fetchedOwners = append(fetchedOwners, store.AddressExtended{
						Address:      address.Address,
						AddressIndex: address.AddressIndex,
						WalletIndex:  wallet.WalletIndex,
						UserID:       user.UserID,
						Creator:      owner.Creator,
						Associated:   true,
					})
				} else {
					fetchedOwners = append(fetchedOwners, owner)
				}
			}
		}
	}
	err := uStore.UpdateMultisigOwners(user.UserID, multisig.InviteCode, fetchedOwners)
	return multisig, err

}
