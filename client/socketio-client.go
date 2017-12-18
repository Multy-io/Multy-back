package client

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/Appscrunch/Multy-back/btc"
	nsq "github.com/bitly/go-nsq"
	"github.com/graarh/golang-socketio"
)

const updateExchangeClient = time.Second * 5

type SocketIOConnectedPool struct {
	users map[string]*SocketIOUser // socketio connections by client id
	m     *sync.RWMutex

	nsqConsumerExchange       *nsq.Consumer
	nsqConsumerBTCTransaction *nsq.Consumer

	chart  *exchangeChart
	server *gosocketio.Server
}

func InitConnectedPool(server *gosocketio.Server) (*SocketIOConnectedPool, error) {
	log.Println("[DEBUG] InitConnectedPool")
	pool := &SocketIOConnectedPool{
		m:     &sync.RWMutex{},
		users: make(map[string]*SocketIOUser, 0),
	}
	nsqConsumerBTCTransaction, err := pool.newConsumerBTCTransaction()
	if err != nil {
		return nil, err
	}
	pool.nsqConsumerBTCTransaction = nsqConsumerBTCTransaction

	return pool, nil
}

func (sConnPool *SocketIOConnectedPool) newConsumerBTCTransaction() (*nsq.Consumer, error) {
	consumer, err := nsq.NewConsumer(topicBTCTransactionUpdate, "all", nsq.NewConfig())
	if err != nil {
		return nil, err
	}

	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		log.Printf("[%s]: %v", topicBTCTransactionUpdate, string(message.Body))

		var newTransactionWithUserID = btc.BtcTransactionWithUserID{}
		if err := json.Unmarshal(message.Body, &newTransactionWithUserID); err != nil {
			log.Println("[ERR] topic btc transaction update: ", err.Error())
			return err
		}
		go sConnPool.sendTransactionNotify(newTransactionWithUserID)
		return nil
	}))

	err = consumer.ConnectToNSQD("127.0.0.1:4150")
	if err != nil {
		log.Printf("[ERR] nsq exchange: %s\n", err.Error())
	}

	return consumer, nil
}

func (sConnPool *SocketIOConnectedPool) sendTransactionNotify(newTransactionWithUserID btc.BtcTransactionWithUserID) {
	sConnPool.m.Lock()
	defer sConnPool.m.Unlock()

	if _, ok := sConnPool.users[newTransactionWithUserID.UserID]; !ok {
		return
	}
	userID := newTransactionWithUserID.UserID
	userConns := sConnPool.users[userID].conns

	for _, conn := range userConns {
		log.Printf("[DEBUG] %s: id=%s\n", topicBTCTransactionUpdate, conn.Id())
		conn.Emit(topicBTCTransactionUpdate, newTransactionWithUserID)
	}
}

func (sConnPool *SocketIOConnectedPool) addUserConn(userID string, userObj *SocketIOUser) {
	log.Println("DEBUG AddUserConn: ", userID)
	sConnPool.m.Lock()
	defer sConnPool.m.Unlock()

	(sConnPool.users[userID]) = userObj
}

func (sConnPool *SocketIOConnectedPool) removeUserConn(userID string) {
	log.Println("DEBUG RemoveUserConn: ", userID)
	sConnPool.m.Lock()
	defer sConnPool.m.Unlock()

	delete(sConnPool.users, userID)
}

type SocketIOUser struct {
	userID     string
	deviceType string
	jwtToken   string

	chart *exchangeChart

	nsqExchangeUpdateConsumer *nsq.Consumer
	nsqBTCTRxConsumer         *nsq.Consumer
	nsqConfig                 *nsq.Config

	conns map[string]*gosocketio.Channel
}

func newSocketIOUser(id string, connectedUser *SocketIOUser, conn *gosocketio.Channel) *SocketIOUser {
	connectedUser.conns = make(map[string]*gosocketio.Channel, 0)
	connectedUser.conns[id] = conn

	go connectedUser.runUpdateExchange()

	return connectedUser
}

func (sIOUser *SocketIOUser) runUpdateExchange() {
	log.Println("[DEBUG] runUpdateExchange userID=", sIOUser.userID)
	tr := time.NewTicker(updateExchangeClient)

	for {
		select {
		case _ = <-tr.C:
			updateMsg := sIOUser.chart.getLast()
			for _, c := range sIOUser.conns {
				log.Printf("conn %v\n", c)
				c.Emit(topicExchangeUpdate, updateMsg)
			}
		}
	}
}
