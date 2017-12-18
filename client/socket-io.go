package client

import (
	"fmt"
	"log"
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

func SetSocketIOHandlers(r *gin.RouterGroup) (*SocketIOConnectedPool, error) {
	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())

	connectedPool, err := InitConnectedPool(server)
	if err != nil {
		return nil, fmt.Errorf("connection pool initialization: %s", err.Error())
	}
	chart, err := initExchangeChart()
	if err != nil {
		return nil, fmt.Errorf("exchange chart initialization: %s", err.Error())
	}

	connectedPool.chart = chart

	server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		fmt.Println("[DEBUG] connected:", c.Id())
		c.Emit("exchangeAll", connectedPool.chart.getAll())

		user, err := getHeaderDataSocketIO(c.RequestHeader())
		if err != nil {
			log.Printf("[ERR] get socketio headers: %s\n", err.Error())
			return
		}
		connectionID := c.Id()
		user.chart = connectedPool.chart

		newConn := newSocketIOUser(connectionID, user, c)
		connectedPool.addUserConn(user.userID, newConn)

		log.Println("[DEBUG] OnConnection done")
	})

	/*server.On("getExchangeReq", func(c *gosocketio.Channel, req EventGetExchangeReq) EventGetExchangeResp {
		log.Printf("[DEBUG] getExchange: user=%s, req=%+v\n", c.Id(), req)
		resp := processGetExchangeEvent(req)
		log.Printf("[DEBUG] getExchange: user=%s, resp=%+v\n", c.Id(), resp)
		return resp
	})*/

	server.On(gosocketio.OnError, func(c *gosocketio.Channel) {
		log.Println("Error occurs")
	})

	server.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		log.Println("Disconnected", c.Id())
	})

	serveMux := http.NewServeMux()
	serveMux.Handle("/socket.io/", server)

	log.Println("Starting server...")
	go func() {
		log.Panic(http.ListenAndServe("0.0.0.0:6680", serveMux))
	}()
	return connectedPool, nil
}
