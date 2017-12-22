package client

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/graarh/golang-socketio/transport"

	"github.com/graarh/golang-socketio"
)

const (
	socketIOOutMsg = "outcoming"
	socketIOInMsg  = "incoming"

	deviceTypeMac     = "mac"
	deviceTypeAndroid = "android"
)

const (
	topicExchangeAll          = "exchangeAll"
	topicExchangeUpdate       = "exchangeUpdate"
	topicBTCTransactionUpdate = "btcTransaction"

	topicEthTransactionUpdate = "ethTransaction"

	EUR = "EUR"
	USD = "USD"
	ETH = "ETH"
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

func SetSocketIOHandlers(r *gin.RouterGroup, address string) (*SocketIOConnectedPool, error) {
	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())

	pool, err := InitConnectedPool(server, address)
	if err != nil {
		return nil, fmt.Errorf("connection pool initialization: %s", err.Error())
	}
	chart, err := initExchangeChart()
	if err != nil {
		return nil, fmt.Errorf("exchange chart initialization: %s", err.Error())
	}

	pool.chart = chart

	server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		pool.log.Debugf("connected: %s", c.Id())
		allRates := pool.chart.getAll()
		c.Emit(topicExchangeAll, allRates)

		user, err := getHeaderDataSocketIO(c.RequestHeader())
		if err != nil {
			pool.log.Errorf("get socketio headers: %s", err.Error())
			return
		}
		user.pool = pool
		connectionID := c.Id()
		user.chart = pool.chart

		pool.m.Lock()
		userFromPool, ok := pool.users[user.userID]
		if !ok {
			pool.log.Debugf("new user")
			newSocketIOUser(connectionID, user, c, pool.log)
			pool.users[user.userID] = user
			userFromPool = user
		}
		userFromPool.conns[connectionID] = c
		pool.closeChByConnID[connectionID] = userFromPool.closeCh
		pool.m.Unlock()
		pool.log.Debugf("OnConnection done")
	})

	server.On(gosocketio.OnError, func(c *gosocketio.Channel) {
		pool.log.Errorf("Error occurs %s", c.Id())
	})

	server.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		pool.log.Infof("Disconnected %s", c.Id())
		pool.removeUserConn(c.Id())
	})

	serveMux := http.NewServeMux()
	serveMux.Handle("/socket.io/", server)

	pool.log.Infof("Starting socketIO server on %s address", address)
	go func() {
		pool.log.Panicf("%s", http.ListenAndServe(address, serveMux))
	}()
	return pool, nil
}
