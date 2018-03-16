/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"

	"github.com/Appscrunch/Multy-back/currencies"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
	nsq "github.com/bitly/go-nsq"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

const (
	// TopicTransaction is a topic for sending notifies to clients
	TopicTransaction = "btcTransactionUpdate"
)

// Dirty hack - this will be wrapped to a struct
var (
	nsqProducer *nsq.Producer // a producer for sending data to clients
	WsCliTest   *gosocketio.Client
	WsCliMain   *gosocketio.Client
)

var log = slf.WithContext("btc")

//InitHandlers init nsq mongo and ws connection to node
// return main client , test client , err
func InitHandlers(dbConf *store.Conf, coinTypes []store.CoinType, nsqAddr string) (*gosocketio.Client, *gosocketio.Client, error) {
	config := nsq.NewConfig()
	p, err := nsq.NewProducer(nsqAddr, config)
	if err != nil {
		return WsCliMain, WsCliTest, fmt.Errorf("nsq producer: %s", err.Error())
	}
	nsqProducer = p
	log.Infof("InitHandlers: nsq.NewProducer: √")

	db, err := mgo.Dial(dbConf.Address)
	if err != nil {
		log.Errorf("RunProcess: can't connect to DB: %s", err.Error())
		return WsCliMain, WsCliTest, fmt.Errorf("mgo.Dial: %s", err.Error())
	}
	log.Infof("InitHandlers: mgo.Dial: √")

	usersData = db.DB(dbConf.DBUsers).C(store.TableUsers) // all db tables
	exRate = db.DB("DBStockExchangeRate").C("TableStockExchangeRate")

	//TODO: set table names from conf

	// main
	mempoolRates = db.DB(dbConf.DBFeeRates).C(store.TableFeeRates)
	txsData = db.DB(dbConf.DBTx).C("BTCMainTxData")
	spendableOutputs = db.DB(dbConf.DBTx).C("BTCMainspendableOutputs")

	// test
	mempoolRatesTest = db.DB(dbConf.DBFeeRates).C("Rates-Test")
	txsDataTest = db.DB(dbConf.DBTx).C("BTCTestTxData")
	spendableOutputsTest = db.DB(dbConf.DBTx).C("BTCTestspendableOutputs")

	// TODO: uncomment to support mainnet
	/*
		// setup mainnet
		url, port, err := fethCoinType(coinTypes, currencies.Bitcoin, currencies.Main)
		if err != nil {
			return WsCliMain, WsCliTest,fmt.Errorf("fethCoinType: %s", err.Error())
		}
		mainnetCli, err := gosocketio.Dial(
			gosocketio.GetUrl(url, port, false),
			transport.GetDefaultWebsocketTransport())
		if err != nil {
			return WsCliMain, WsCliTest,fmt.Errorf("gosocketio.Dial: %s", err.Error())
		}
		WsCliMain = mainnetCli
		log.Infof("InitHandlers: gosocketio.Dial Main: √")
		go SetWsHandlers(WsCliMain, currencies.Main)
		log.Infof("InitHandlers: SetWsHandlers Main: √")
	*/

	// setup testnet
	url, port, err := fethCoinType(coinTypes, currencies.Bitcoin, currencies.Test)
	if err != nil {
		return WsCliMain, WsCliTest, fmt.Errorf("fethCoinType: %s", err.Error())
	}

	t := transport.GetDefaultWebsocketTransport()
	// t.PingInterval = time.Second * 5
	// t.PingTimeout = time.Second * 5
	// t.ReceiveTimeout = time.Second * 5
	// t.SendTimeout = time.Second * 5
	// t.BufferSize = 1000000000

	fmt.Println(transport.GetDefaultWebsocketTransport().ReceiveTimeout)
	testnetCli, err := gosocketio.Dial(
		gosocketio.GetUrl(url, port, false),
		t)
	if err != nil {
		return WsCliMain, WsCliTest, fmt.Errorf("gosocketio.Dial: %s", err.Error())
	}

	WsCliTest = testnetCli
	log.Infof("InitHandlers: gosocketio.Dial Test: √")
	SetWsHandlers(WsCliTest, currencies.Test)
	log.Infof("InitHandlers: SetWsHandlers Test: √")

	//request all mempool
	err = WsCliTest.Emit("getAllMempool", nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	return WsCliMain, WsCliTest, nil
}

func fethCoinType(coinTypes []store.CoinType, currencyID, networkID int) (string, int, error) {
	for _, ct := range coinTypes {
		if ct.СurrencyID == currencyID && ct.NetworkID == networkID {
			return ct.SocketURL, ct.SocketPort, nil
		}
	}
	return "", 0, fmt.Errorf("fethCoinType: no such coin in config")
}

type BtcTransaction struct {
	TransactionType int    `json:"transactionType"`
	Amount          int64  `json:"amount"`
	TxID            string `json:"txid"`
	Address         string `json:"address"`
}

type BtcTransactionWithUserID struct {
	NotificationMsg *BtcTransaction
	UserID          string
}
