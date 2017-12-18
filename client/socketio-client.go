package client

import (
	"encoding/json"

	"sync"
	"time"

	"github.com/Appscrunch/Multy-back/btc"
	nsq "github.com/bitly/go-nsq"
	"github.com/graarh/golang-socketio"
	"github.com/ventu-io/slf"
)

const updateExchangeClient = time.Second * 5

type SocketIOConnectedPool struct {
	address string
	users   map[string]*SocketIOUser // socketio connections by client id
	m       *sync.RWMutex

	nsqConsumerExchange       *nsq.Consumer
	nsqConsumerBTCTransaction *nsq.Consumer

	chart  *exchangeChart
	server *gosocketio.Server
	log    slf.StructuredLogger
}

func InitConnectedPool(server *gosocketio.Server, address string) (*SocketIOConnectedPool, error) {
	pool := &SocketIOConnectedPool{
		m:       &sync.RWMutex{},
		users:   make(map[string]*SocketIOUser, 0),
		address: address,
		log:     slf.WithContext("connectedPool"),
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
		sConnPool.log.Infof("[%s]: %v", topicBTCTransactionUpdate, string(message.Body))

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
		sConnPool.log.Debugf("%s: id=%s\n", topicBTCTransactionUpdate, conn.Id())
		conn.Emit(topicBTCTransactionUpdate, newTransactionWithUserID)
	}
}

func (sConnPool *SocketIOConnectedPool) addUserConn(userID string, userObj *SocketIOUser) {
	sConnPool.log.Debugf("AddUserConn: ", userID)
	sConnPool.m.Lock()
	defer sConnPool.m.Unlock()

	(sConnPool.users[userID]) = userObj
}

func (sConnPool *SocketIOConnectedPool) removeUserConn(userID string) {
	sConnPool.log.Debugf("RemoveUserConn: %s", userID)
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

	log slf.StructuredLogger
}

func newSocketIOUser(id string, newUser *SocketIOUser, conn *gosocketio.Channel, log slf.StructuredLogger) *SocketIOUser {
	newUser.conns = make(map[string]*gosocketio.Channel, 0)
	newUser.conns[id] = conn
	newUser.log = log.WithField("userID", id)

	go newUser.runUpdateExchange()

	return newUser
}

func (sIOUser *SocketIOUser) runUpdateExchange() {
	sIOUser.log.Debugf("runUpdateExchange userID=%s", sIOUser.userID)
	tr := time.NewTicker(updateExchangeClient)

	for {
		select {
		case _ = <-tr.C:
			updateMsg := sIOUser.chart.getLast()
			for _, c := range sIOUser.conns {
				sIOUser.log.Debugf("runUpdateExchange: conn id=%s", c.Id())
				c.Emit(topicExchangeUpdate, updateMsg)
			}
		}
	}
}
