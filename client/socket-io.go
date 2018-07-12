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

	"github.com/Multy-io/Multy-back/btc"
	"github.com/Multy-io/Multy-back/currencies"
	"github.com/Multy-io/Multy-back/eth"
	btcpb "github.com/Multy-io/Multy-back/node-streamer/btc"
	ethpb "github.com/Multy-io/Multy-back/node-streamer/eth"
	"github.com/Multy-io/Multy-back/store"

	"github.com/gin-gonic/gin"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
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

	// off chain multisig logic
	server.On(joinMultisig, func(c *gosocketio.Channel, jm store.JoinMultisig) string {
		//TODO: check current multisig for able to joining
		multisig, err := ratesDB.FindMultisig(jm)
		if err != nil {
			pool.log.Errorf("server.On:joinMultisig : %v", err.Error())
			return "can't join multisig: " + err.Error()
		}

		joined := len(multisig.Owners)
		if multisig.OwnersCount < joined {
			err = ratesDB.JoinMultisig(jm, multisig)
			if err != nil {
				//db err
				pool.log.Errorf("server.On:joinMultisigratesDB.JoinMultisig: %v", err.Error())
				return "can't join multisig: " + err.Error()
			}
			return joinMultisig + ":ok"
		}

		if multisig.OwnersCount == joined {
			pool.log.Errorf("server.On:joinMultisig: can't join multisig: sufficient number of owners")
			return "can't join multisig: sufficient number of owners"
		}
		return ""
	})

	server.On(leaveMultisig, func(c *gosocketio.Channel, jm store.JoinMultisig) string {
		//TODO: check current multisig for able to joining
		multisig, err := ratesDB.FindMultisig(jm)
		if err != nil {
			pool.log.Errorf("server.On:joinMultisig : %v", err.Error())
			return "can't join multisig: " + err.Error()
		}
		exists := false
		for _, owner := range multisig.Owners {
			if owner.Address == jm.Address {
				exists = true
				if owner.Creator {
					pool.log.Errorf("server.On:leaveMultisig: can't leave multisig if you are creator you need delete it")
					return "can't leave multisig if you are creator you need delete it "
				}
			}
		}

		owners := []store.AddressExtended{}
		if exists {
			//delete multisig from user
			err := ratesDB.LeaveMultisig(jm)
			if err != nil {
				pool.log.Errorf("server.On:leaveMultisig:ratesDB.LeaveMultisig : %v", err.Error())
				return "can't leave multisig: " + err.Error()
			}

			//delete owner from owners list
			for _, owner := range multisig.Owners {
				if owner.Address != jm.Address {
					owners = append(owners, owner)
				}
			}
			ratesDB.CleanOwnerList(jm, multisig, owners)

		}

		if !exists {
			pool.log.Errorf("server.On:leaveMultisig: can't leave multisig if you are not a owner")
			return "can't leave multisig if you are not a owner"
		}
		return ""
	})

	//TODO:
	//TODO:
	//TODO:
	server.On(deleteMultisig, func(c *gosocketio.Channel, jm store.JoinMultisig) string {
		//TODO: check current multisig for able to joining
		multisig, err := ratesDB.FindMultisig(jm)
		if err != nil {
			pool.log.Errorf("server.On:joinMultisig : %v", err.Error())
			return "can't join multisig: " + err.Error()
		}
		owner := false
		for _, owner := range multisig.Owners {
			if owner.Address == jm.Address {
				if owner.Creator {

					//delete
				}
			}
		}

		if !owner {
			pool.log.Errorf("server.On:leaveMultisig: can't delete multisig if you are not a creator")
			return "can't delete multisig if you are not a creator"
		}
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
