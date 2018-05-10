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

	"github.com/Appscrunch/Multy-back/store"
	"github.com/gin-gonic/gin"
	"github.com/graarh/golang-socketio/transport"

	"github.com/graarh/golang-socketio"
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

	NewReceiver = "event:new:receiver"
	Pay         = "event:pay"
	PaymentSend = "event:payment:send"
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

func SetSocketIOHandlers(r *gin.RouterGroup, address, nsqAddr string, ratesDB store.UserStore) (*SocketIOConnectedPool, error) {
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

	//TODO: feature logic

	server.On(gosocketio.OnError, func(c *gosocketio.Channel) {
		pool.log.Errorf("Error occurs %s", c.Id())
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

	server.On(SenderOn, func(c *gosocketio.Channel, data store.SenderInData) string {
		pool.log.Infof("Sender become on: %v", c.Id())
		sender := store.Sender{
			ID:       data.UserID,
			UserCode: data.Code,
			Socket:   c,
		}
		pool.log.Infof("God data from sender: %v", sender)
		senderExist := false

		for _, cachedSender := range senders {
			if cachedSender.ID == sender.ID {
				senderExist = true
			}
		}
		if !senderExist {
			senders = append(senders, sender)
		}

		receiversM.Lock()
		receiver, ok := receivers[sender.UserCode]
		receiversM.Unlock()

		// Find receiver by the code
		if ok {
			sender.Socket.Emit(NewReceiver, receiver)
		} else {
			senderExist := false
			for _, cachedSender := range senders {
				if cachedSender.ID == sender.ID {
					senderExist = true
				}
			}
			if !senderExist {
				senders = append(senders, sender)
			}
		}
		return "ok"
	})

	server.On(ReceiverOn, func(c *gosocketio.Channel, data store.ReceiverInData) string {
		pool.log.Infof("Got messeage Receiver On:", data)
		c.Join(WirelessRoom)
		receiver := store.Receiver{
			ID:         data.ID,
			CurrencyID: data.CurrencyID,
			Amount:     data.Amount,
			UserCode:   data.UserCode,
			Socket:     c,
		}

		receiversM.Lock()
		_, ok := receivers[receiver.UserCode]

		if !ok {
			receivers[receiver.UserCode] = receiver
		}
		receiversM.Unlock()

		//Find sender
		for _, sender := range senders {
			if sender.UserCode == receiver.UserCode {
				sender.Socket.Emit(NewReceiver, receiver)
			}
		}
		return "ok"
	})

	server.On(SenderOn, func(c *gosocketio.Channel, data store.SenderInData) string {

		pool.log.Infof("Sender become on:", c.Id())

		sender := store.Sender{
			ID:       data.UserID,
			UserCode: data.Code,
			Socket:   c,
		}

		senderExist := false
		for _, cachedSender := range senders {
			if cachedSender.ID == sender.ID {
				senderExist = true
			}
		}

		if !senderExist {
			senders = append(senders, sender)
		}

		receiversM.Lock()
		receiver, ok := receivers[sender.UserCode]
		receiversM.Unlock()

		// Find Receiver by the code
		if ok {
			sender.Socket.Emit(NewReceiver, receiver)

		} else {
			senderExist := false
			for _, cachedSender := range senders {
				if cachedSender.ID == sender.ID {
					senderExist = true
				}
			}
			if !senderExist {
				senders = append(senders, sender)
			}
		}

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
