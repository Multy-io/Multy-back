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

	"github.com/Appscrunch/Multy-back/btc"
	"github.com/Appscrunch/Multy-back/currencies"
	btcpb "github.com/Appscrunch/Multy-back/node-streamer/btc"
	"github.com/Appscrunch/Multy-back/store"

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
	PORT         = ":5555"
	WirelessRoom = "wireless"

	ReceiverOn = "event:receiver:on"
	SenderOn   = "event:sender:on"

	SenderCheck = "event:sender:check"

	Filter = "event:filter"

	NewReceiver     = "event:new:receiver"
	SendRaw         = "event:sendraw"
	PaymentSend     = "event:payment:send"
	PaymentReceived = "event:payment:received"
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

func SetSocketIOHandlers(restClient *RestClient, BTC *btc.BTCConn, r *gin.RouterGroup, address, nsqAddr string, ratesDB store.UserStore) (*SocketIOConnectedPool, error) {
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
		// pool.log.Debugf("connected: %s", c.Id())

		// moved to next release
		//ratesDay := pool.chart.getExchangeDay()
		//c.Emit(topicExchangeDay, ratesDay)

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

	server.On(ReceiverOn, func(c *gosocketio.Channel, data store.Receiver) string {
		pool.log.Infof("Got messeage Receiver On:", data)
		c.Join(WirelessRoom)
		receiver := store.Receiver{
			ID:         data.ID,
			UserCode:   data.UserCode,
			CurrencyID: data.CurrencyID,
			NetworkID:  data.NetworkID,
			Address:    data.Address,
			Amount:     data.Amount,
			Socket:     c,
		}

		receiversM.Lock()
		_, ok := receivers[receiver.UserCode]
		if !ok {
			receivers[receiver.UserCode] = receiver
		}
		receiversM.Unlock()

		//TODO:
		// wait for incoming tx

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

	server.On(SendRaw, func(c *gosocketio.Channel, raw store.RawHDTx) {

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
			return
		}

		if strings.Contains("err:", resp.GetMessage()) {
			pool.log.Errorf("sendRawHDTransaction: restClient.BTC.CliMain.EventSendRawTx:resp err %s", err.Error())
			c.Emit(SendRaw, err.Error())
			return
		}

		if raw.IsHD {
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

	serveMux := http.NewServeMux()
	serveMux.Handle("/socket.io/", server)

	pool.log.Infof("Starting socketIO server on %s address", address)
	go func() {
		pool.log.Panicf("%s", http.ListenAndServe(address, serveMux))
	}()
	return pool, nil
}
