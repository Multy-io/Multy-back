/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package client

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Multy-io/Multy-back/btc"
	"github.com/Multy-io/Multy-back/eth"
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
	joinMultisig       = 1
	leaveMultisig      = 2
	deleteMultisig     = 3
	kickMultisig       = 4
	checkMultisig      = 5
	viewTransaction    = 6
	declineTransaction = 7
	NotifyDeploy       = 8

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
	pool.Server = server

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
				return makeErr(msgMultisig.UserID, "can't join multisig: bad request: "+err.Error(), joinMultisig)
			}

			// if invite code exists
			if !ratesDB.CheckInviteCode(msgMultisig.InviteCode) {
				// check current multisig for able to joining
				users := ratesDB.FindMultisigUsers(msgMultisig.InviteCode)
				for _, user := range users {
					if user.UserID == msgMultisig.UserID {
						pool.log.Errorf("server.On:msgSend:joinMultisig we not support multiple join from same userid")
						return makeErr(msgMultisig.UserID, "we not support multiple join from same userid ", joinMultisig)
					}
				}

				if !ratesDB.IsRelatedAddress(msgMultisig.UserID, msgMultisig.Address) {
					pool.log.Errorf("server.On:msgSend:joinMultisig can't add addres with is not related userid ")
					return makeErr(msgMultisig.UserID, "can't add addres with is not related userid ", joinMultisig)
				}

				if !ratesDB.CheckMultisigCurrency(msgMultisig.InviteCode, msgMultisig.CurrencyID, msgMultisig.NetworkID) {
					pool.log.Errorf("server.On:msgSend:joinMultisig this invitecode accotiated with different currency/network  ")
					return makeErr(msgMultisig.UserID, "this invitecode accotiated with different currency/network", joinMultisig)
				}

				multisig, msg, err := getMultisig(ratesDB, msgMultisig, joinMultisig)
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
							Address:      owner.Address,
							Creator:      owner.Creator,
							WalletIndex:  owner.WalletIndex,
							AddressIndex: owner.AddressIndex,
						})
					}

					owners = append(owners, store.AddressExtended{
						UserID:       msgMultisig.UserID,
						Address:      msgMultisig.Address,
						WalletIndex:  msgMultisig.WalletIndex,
						AddressIndex: 0,
						Associated:   true,
					})

					deployStatus := store.MultisigStatusWaitingForJoin
					joined++
					if joined == multisig.OwnersCount {
						multisigToJoin.DeployStatus = store.MultisigStatusAllJoined
						deployStatus = store.MultisigStatusAllJoined
					}
					multisigToJoin.Owners = owners
					users := ratesDB.FindMultisigUsers(msgMultisig.InviteCode)
					err = ratesDB.JoinMultisig(msgMultisig.UserID, multisigToJoin)
					if err != nil {
						//db err
						pool.log.Errorf("server.On:MultisigMsgratesDB.MultisigMsg: %v", err.Error())
						return makeErr(msgMultisig.UserID, "can't join multisig: "+err.Error(), joinMultisig)
					}
					pool.log.Debugf("user %v joined %v multisig", msgMultisig.UserID, multisig.WalletName)
					// send new multisig entitiy to all online owners by ws
					for _, user := range users {
						userMultisig, err := updateUserOwners(user, multisig, ratesDB, deployStatus)
						if err != nil {
							//db err
							pool.log.Errorf("server.On:MultisigMsgratesDB.MultisigMsg: %v", err.Error())
						}
						_, online := pool.users[user.UserID]
						if online {
							msg := store.WsMessage{
								Type:    joinMultisig,
								To:      user.UserID,
								Date:    time.Now().Unix(),
								Payload: userMultisig,
							}
							server.BroadcastToAll(msgRecieve+":"+user.UserID, msg)
						}
					}

					return store.WsMessage{
						Type:    joinMultisig,
						To:      msgMultisig.UserID,
						Date:    time.Now().Unix(),
						Payload: "ok",
					}
				}

			}

			return makeErr("", "wrong request payload", joinMultisig)
		case leaveMultisig:
			msgMultisig := &store.MultisigMsg{}
			err := mapstructure.Decode(msg.Payload, msgMultisig)
			if err != nil {
				pool.log.Errorf("server.On:msgSend:leaveMultisig:mapstructure.Decode %v", err.Error())
				return makeErr(msgMultisig.UserID, "can't leave multisig: bad request: "+err.Error(), leaveMultisig)
			}
			if !ratesDB.CheckInviteCode(msgMultisig.InviteCode) {
				multisig, msg, err := getMultisig(ratesDB, msgMultisig, leaveMultisig)
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
							return makeErr(msgMultisig.UserID, "can't leave multisig if you are creator you need delete it ", leaveMultisig)
						}
					}
				}

				owners := []store.AddressExtended{}
				if exists {
					//delete multisig from user
					err := ratesDB.LeaveMultisig(msgMultisig.UserID, msgMultisig.InviteCode)
					if err != nil {
						pool.log.Errorf("server.On:leaveMultisig:ratesDB.LeaveMultisig : %v", err.Error())
						return makeErr(msgMultisig.UserID, "can't leave multisig: "+err.Error(), leaveMultisig)
					}
					pool.log.Debugf("user %v leave %v multisig", msgMultisig.UserID, multisig.WalletName)
					users := ratesDB.FindMultisigUsers(msgMultisig.InviteCode)

					//delete owner from owners list
					for _, owner := range multisig.Owners {
						if owner.Address != msgMultisig.Address {
							owners = append(owners, owner)
						}
					}
					multisig.Owners = owners
					for _, user := range users {
						userMultisig, err := updateUserOwners(user, multisig, ratesDB, store.MultisigStatusWaitingForJoin)
						if err != nil {
							//db err
							pool.log.Errorf("server.On:MultisigMsgratesDB.MultisigMsg: %v", err.Error())
						}

						_, online := pool.users[user.UserID]
						if online {
							msg := store.WsMessage{
								Type:    leaveMultisig,
								To:      user.UserID,
								Date:    time.Now().Unix(),
								Payload: userMultisig,
							}
							server.BroadcastToAll(msgRecieve+":"+user.UserID, msg)
						}
					}

					return store.WsMessage{
						Type:    leaveMultisig,
						To:      msgMultisig.UserID,
						Date:    time.Now().Unix(),
						Payload: "ok",
					}

				}

				if !exists {
					pool.log.Errorf("server.On:leaveMultisig: can't leave multisig if you are not a owner")
					return makeErr(msgMultisig.UserID, "can't leave multisig if you are not a owner", leaveMultisig)
				}

			}

			return makeErr("", "wrong request payload: ", leaveMultisig)
		case kickMultisig:
			msgMultisig := &store.MultisigMsg{}
			err := mapstructure.Decode(msg.Payload, msgMultisig)
			if err != nil {
				pool.log.Errorf("server.On:msgSend:kickMultisig:mapstructure.Decode %v", err.Error())
				return makeErr(msgMultisig.UserID, "can't kik from multisig: bad request: "+err.Error(), kickMultisig)
			}
			if !ratesDB.CheckInviteCode(msgMultisig.InviteCode) {
				multisig, msg, err := getMultisig(ratesDB, msgMultisig, kickMultisig)
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
					return makeErr(msgMultisig.UserID, "only admin can kik form ms", kickMultisig)
				}

				owners := []store.AddressExtended{}
				if admin {
					users := ratesDB.FindMultisigUsers(msgMultisig.InviteCode)
					//delete multisig from user
					err := ratesDB.KickMultisig(msgMultisig.AddressToKick, msgMultisig.InviteCode)
					if err != nil {
						pool.log.Errorf("server.On:kickMultisig:ratesDB.KickMultisig: %v", err.Error())
						return makeErr(msgMultisig.UserID, "can't kik from multisig: "+err.Error(), kickMultisig)
					}

					pool.log.Debugf("user %v kicked from %v multisig", msgMultisig.UserID, multisig.WalletName)

					//delete owner from owners list
					for _, owner := range multisig.Owners {
						if owner.Address != msgMultisig.AddressToKick {
							owners = append(owners, owner)
						}
					}
					multisig.Owners = owners

					for _, user := range users {
						userMultisig, err := updateUserOwners(user, multisig, ratesDB, store.MultisigStatusWaitingForJoin)
						if err != nil {
							//db err
							pool.log.Errorf("server.On:kickMultisig:updateUserOwners: %v", err.Error())
						}
						msUpd := store.MultisigExtended{
							Multisig:      *userMultisig,
							KickedAddress: msgMultisig.AddressToKick,
						}
						_, online := pool.users[user.UserID]
						if online {
							msg := store.WsMessage{
								Type:    kickMultisig,
								To:      user.UserID,
								Date:    time.Now().Unix(),
								Payload: msUpd,
							}
							server.BroadcastToAll(msgRecieve+":"+user.UserID, msg)
						}
					}

					return store.WsMessage{
						Type:    kickMultisig,
						To:      msgMultisig.UserID,
						Date:    time.Now().Unix(),
						Payload: "ok",
					}

				}

			}

			return makeErr("", "wrong request payload", kickMultisig)
		case deleteMultisig:
			msgMultisig := &store.MultisigMsg{}
			err := mapstructure.Decode(msg.Payload, msgMultisig)
			if err != nil {
				pool.log.Errorf("server.On:msgSend:deleteMultisig:mapstructure.Decode %v", err.Error())
				return makeErr(msgMultisig.UserID, "can't kik from multisig: bad request: "+err.Error(), deleteMultisig)
			}
			if !ratesDB.CheckInviteCode(msgMultisig.InviteCode) {
				multisig, msg, err := getMultisig(ratesDB, msgMultisig, deleteMultisig)
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
					pool.log.Errorf("server.On:deleteMultisig: only creator can delete ms")
					return makeErr(msgMultisig.UserID, "only creator can delete ms", deleteMultisig)
				}
				if admin {
					users := ratesDB.FindMultisigUsers(msgMultisig.InviteCode)
					err := ratesDB.DeleteMultisig(msgMultisig.InviteCode)
					if err != nil {
						pool.log.Errorf("server.On:deleteMultisig:DeleteMultisig %v", err.Error())
						return makeErr(msgMultisig.UserID, "server.On:deleteMultisig:DeleteMultisig "+err.Error(), deleteMultisig)
					}
					pool.log.Debugf("user %v delete %v multisig", msgMultisig.UserID, multisig.WalletName)
					for _, user := range users {
						_, online := pool.users[user.UserID]
						if online {
							msg := store.WsMessage{
								Type:    deleteMultisig,
								To:      user.UserID,
								Date:    time.Now().Unix(),
								Payload: msgMultisig.InviteCode,
							}
							server.BroadcastToAll(msgRecieve+":"+user.UserID, msg)
						}
					}
					return store.WsMessage{
						Type:    deleteMultisig,
						To:      msgMultisig.UserID,
						Date:    time.Now().Unix(),
						Payload: "ok",
					}
				}
			}

			return makeErr("", "wrong request payload: ", deleteMultisig)

		case checkMultisig:
			msgMultisig := &store.MultisigMsg{}
			err := mapstructure.Decode(msg.Payload, msgMultisig)
			if err != nil {
				pool.log.Errorf("server.On:msgSend:checkMultisig:mapstructure.Decode %v", err.Error())
				return makeErr(msgMultisig.UserID, "can't kik from multisig: bad request: "+err.Error(), checkMultisig)
			}
			icInfo := ratesDB.InviteCodeInfo(msgMultisig.InviteCode)
			pool.log.Debugf("user %v check multisig", msgMultisig.UserID)
			msg := store.WsMessage{
				Type:    checkMultisig,
				To:      msgMultisig.UserID,
				Date:    time.Now().Unix(),
				Payload: icInfo,
			}

			return msg
		case declineTransaction:
			msgMultisig := &store.MultisigMsg{}
			err := mapstructure.Decode(msg.Payload, msgMultisig)
			if err != nil {
				pool.log.Errorf("server.On:msgSend:DeclineTransaction:mapstructure.Decode %v", err.Error())
				return makeErr(msgMultisig.UserID, "can't kik from multisig: bad request: "+err.Error(), checkMultisig)
			}
			err = ratesDB.DeclineTransaction(msgMultisig.TxID, msgMultisig.Address, msgMultisig.CurrencyID, msgMultisig.NetworkID)
			if err != nil {
				pool.log.Errorf("server.On:msgSend:declineTransaction:DeclineTransaction %v", err.Error())
			}
			msg := store.WsMessage{
				Type:    declineTransaction,
				To:      msgMultisig.UserID,
				Date:    time.Now().Unix(),
				Payload: "declined",
			}

			return msg
		case viewTransaction:
			msgMultisig := &store.MultisigMsg{}
			err := mapstructure.Decode(msg.Payload, msgMultisig)
			if err != nil {
				pool.log.Errorf("server.On:msgSend:viewTransaction:mapstructure.Decode %v", err.Error())
				return makeErr(msgMultisig.UserID, "can't kik from multisig: bad request: "+err.Error(), checkMultisig)
			}
			err = ratesDB.ViewTransaction(msgMultisig.TxID, msgMultisig.Address, msgMultisig.CurrencyID, msgMultisig.NetworkID)
			if err != nil {
				pool.log.Errorf("server.On:msgSend:viewTransaction:ViewTransaction %v", err.Error())
			}
			msg := store.WsMessage{
				Type:    viewTransaction,
				To:      msgMultisig.UserID,
				Date:    time.Now().Unix(),
				Payload: "viewed",
			}
			return msg
		}

		return makeErr("", "wrong request message type: ", 0)
	})

	server.On(msgRecieve, func(c *gosocketio.Channel, msg store.WsMessage) string {
		return ""
	})

	serveMux := http.NewServeMux()
	serveMux.Handle("/socket.io/", server)

	pool.log.Infof("Starting socketIO server on %s address", address)
	go func() {
		pool.log.Panicf("%s", http.ListenAndServe(address, serveMux))
	}()
	return pool, nil
}

func getMultisig(uStore store.UserStore, msgMultisig *store.MultisigMsg, method int) (*store.Multisig, store.WsMessage, error) {
	msg := store.WsMessage{}
	multisig, err := uStore.FindMultisig(msgMultisig.UserID, msgMultisig.InviteCode)
	if err != nil {
		msg = store.WsMessage{
			Type:    method,
			To:      msgMultisig.UserID,
			Date:    time.Now().Unix(),
			Payload: "can't join multisig: " + err.Error(),
		}
	}
	return multisig, msg, err
}

func makeErr(userid, errorStr string, method int) store.WsMessage {
	return store.WsMessage{
		Type:    method,
		To:      userid,
		Date:    time.Now().Unix(),
		Payload: errorStr,
	}
}

func updateUserOwners(user store.User, multisig *store.Multisig, uStore store.UserStore, deployStatus int) (*store.Multisig, error) {
	// Clean addttion tags
	owners := map[string]store.AddressExtended{}
	for _, owner := range multisig.Owners {
		owners[owner.Address] = store.AddressExtended{
			Address:      owner.Address,
			Creator:      owner.Creator,
			WalletIndex:  owner.WalletIndex,
			AddressIndex: owner.AddressIndex,
		}
	}

	for _, wallet := range user.Wallets {
		for _, address := range wallet.Adresses {
			owner, ok := owners[address.Address]
			if ok {
				owners[address.Address] = store.AddressExtended{
					Address:      address.Address,
					AddressIndex: address.AddressIndex,
					WalletIndex:  wallet.WalletIndex,
					UserID:       user.UserID,
					Creator:      owner.Creator,
					Associated:   true,
				}
			}
			break
		}
	}

	fo := []store.AddressExtended{}
	for _, owner := range owners {
		fo = append(fo, owner)
	}

	err := uStore.UpdateMultisigOwners(user.UserID, multisig.InviteCode, fo, deployStatus)
	return multisig, err

}
