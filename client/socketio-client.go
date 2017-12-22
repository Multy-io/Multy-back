package client

import (
	"encoding/json"
	"log"

	"sync"
	"time"

	"github.com/Appscrunch/Multy-back/btc"
	"github.com/KristinaEtc/slf"
	nsq "github.com/bitly/go-nsq"
	"github.com/graarh/golang-socketio"
)

const updateExchangeClient = time.Second * 5

type SocketIOConnectedPool struct {
	address         string
	users           map[string]*SocketIOUser // socketio connections by client id
	closeChByConnID map[string]chan string   // when connection was finished, send close signal to his goroutine
	m               *sync.RWMutex

	nsqConsumerExchange       *nsq.Consumer
	nsqConsumerBTCTransaction *nsq.Consumer

	chart  *exchangeChart
	server *gosocketio.Server
	log    slf.StructuredLogger
}

func InitConnectedPool(server *gosocketio.Server, address string) (*SocketIOConnectedPool, error) {
	pool := &SocketIOConnectedPool{
		m:               &sync.RWMutex{},
		users:           make(map[string]*SocketIOUser, 0),
		address:         address,
		log:             slf.WithContext("connectedPool"),
		closeChByConnID: make(map[string]chan string, 0),
	}
	pool.log.Info("InitConnectedPool")

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
		var newTransactionWithUserID = btc.BtcTransactionWithUserID{}
		if err := json.Unmarshal(message.Body, &newTransactionWithUserID); err != nil {
			sConnPool.log.Errorf("topic btc transaction update: %s", err.Error())
			return err
		}
		go sConnPool.sendTransactionNotify(newTransactionWithUserID)
		return nil
	}))

	err = consumer.ConnectToNSQD("127.0.0.1:4150")
	if err != nil {
		sConnPool.log.Errorf("nsq exchange: %s", err.Error())
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
		conn.Emit(topicBTCTransactionUpdate, newTransactionWithUserID)
	}
}

func (sConnPool *SocketIOConnectedPool) removeUserConn(connID string) {
	sConnPool.log.Debugf("RemoveUserConn by conn ID: %s", connID)
	sConnPool.m.Lock()
	defer sConnPool.m.Unlock()

	if closeCh, ok := sConnPool.closeChByConnID[connID]; !ok {
		sConnPool.log.Errorf("trying to disconnect user, which didn't connected")
	} else {
		sConnPool.log.Debugf("sending to close chan id=%s", connID)
		delete(sConnPool.closeChByConnID, connID)
		closeCh <- connID
	}
}

func (sConnPool *SocketIOConnectedPool) removeUserFromPool(userID string) {
	sConnPool.log.Debugf("removeUserFromPool")
	sConnPool.m.Lock()
	defer sConnPool.m.Unlock()

	delete(sConnPool.users, userID)
}

type SocketIOUser struct {
	userID     string
	deviceType string
	jwtToken   string

	pool *SocketIOConnectedPool

	chart *exchangeChart

	nsqExchangeUpdateConsumer *nsq.Consumer
	nsqBTCTRxConsumer         *nsq.Consumer
	nsqConfig                 *nsq.Config

	conns map[string]*gosocketio.Channel

	closeCh            chan string
	tickerLastExchange *time.Ticker

	log slf.StructuredLogger
}

func newSocketIOUser(id string, newUser *SocketIOUser, conn *gosocketio.Channel, log slf.StructuredLogger) *SocketIOUser {
	newUser.conns = make(map[string]*gosocketio.Channel, 0)
	newUser.conns[id] = conn
	newUser.log = log.WithField("userID", newUser.userID)
	newUser.closeCh = make(chan string, 0)

	go newUser.runUpdateExchange()

	return newUser
}

func (sIOUser *SocketIOUser) runUpdateExchange() {
	sIOUser.log.Debugf("runUpdateExchange")
	sIOUser.tickerLastExchange = time.NewTicker(updateExchangeClient)

	for {
		select {
		case _ = <-sIOUser.tickerLastExchange.C:
			updateMsg := sIOUser.chart.getLast()
			for _, c := range sIOUser.conns {
				sIOUser.log.Debugf("get last: conn id=%s", c.Id())
				c.Emit(topicExchangeUpdate, updateMsg)
			}
		case connID := <-sIOUser.closeCh:
			log.Println("disconnecting conn id=", connID)
			if conn, ok := sIOUser.conns[connID]; !ok {
				sIOUser.log.Warnf("trying to close conn which doesnt' exists: %s", connID)
			} else {
				conn.Close()
				delete(sIOUser.conns, connID)
				if len(sIOUser.conns) == 0 {
					sIOUser.log.Infof("no connections for user %s", sIOUser.userID)
					sIOUser.tickerLastExchange.Stop()
					sIOUser.pool.removeUserFromPool(sIOUser.userID)
					return
				}
			}
		}
	}
}
