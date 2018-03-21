/*
Copyright 2019 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"fmt"

	"google.golang.org/grpc"
	mgo "gopkg.in/mgo.v2"

	"github.com/Appscrunch/Multy-back/currencies"
	pb "github.com/Appscrunch/Multy-back/node-streamer"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
	nsq "github.com/bitly/go-nsq"
)

const (
	// TopicTransaction is a topic for sending notifications to clients
	TopicTransaction = "btcTransactionUpdate"
)

// BTCConn is a main struct of package
type BTCConn struct {
	NsqProducer      *nsq.Producer // a producer for sending data to clients
	CliTest          pb.NodeCommuunicationsClient
	CliMain          pb.NodeCommuunicationsClient
	WatchAddressTest chan pb.WatchAddress
	WatchAddressMain chan pb.WatchAddress
}

var log = slf.WithContext("btc")

//InitHandlers init nsq mongo and ws connection to node
// return main client , test client , err
func InitHandlers(dbConf *store.Conf, coinTypes []store.CoinType, nsqAddr string) (*BTCConn, error) {
	//declare pacakge struct
	cli := &BTCConn{}

	cli.WatchAddressMain = make(chan pb.WatchAddress)
	cli.WatchAddressTest = make(chan pb.WatchAddress)

	config := nsq.NewConfig()
	p, err := nsq.NewProducer(nsqAddr, config)
	if err != nil {
		return cli, fmt.Errorf("nsq producer: %s", err.Error())
	}

	cli.NsqProducer = p
	log.Infof("InitHandlers: nsq.NewProducer: √")

	db, err := mgo.Dial(dbConf.Address)
	if err != nil {
		log.Errorf("RunProcess: can't connect to DB: %s", err.Error())
		return cli, fmt.Errorf("mgo.Dial: %s", err.Error())
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

	// setup main net
	urlMain, err := fethCoinType(coinTypes, currencies.Bitcoin, currencies.Main)
	if err != nil {
		return cli, fmt.Errorf("fethCoinType: %s", err.Error())
	}

	cliMain, err := initGrpcClient(urlMain, mempoolRates)
	if err != nil {
		return cli, fmt.Errorf("initGrpcClient: %s", err.Error())
	}
	setGRPCHandlers(cliMain, cli.NsqProducer, currencies.Main, cli.WatchAddressMain)

	cli.CliMain = cliMain
	log.Infof("InitHandlers: initGrpcClient: Main: √")

	// setup testnet
	urlTest, err := fethCoinType(coinTypes, currencies.Bitcoin, currencies.Test)
	if err != nil {
		return cli, fmt.Errorf("fethCoinType: %s", err.Error())
	}
	cliTest, err := initGrpcClient(urlTest, mempoolRatesTest)
	if err != nil {
		return cli, fmt.Errorf("initGrpcClient: %s", err.Error())
	}
	setGRPCHandlers(cliTest, cli.NsqProducer, currencies.Test, cli.WatchAddressTest)

	cli.CliTest = cliTest
	log.Infof("InitHandlers: initGrpcClient: Test: √")

	return cli, nil
}

func initGrpcClient(url string, mpRates *mgo.Collection) (pb.NodeCommuunicationsClient, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		log.Errorf("initGrpcClient: grpc.Dial: %s", err.Error())
		return nil, err
	}
	// defer conn.Close()

	// Create a new  client
	client := pb.NewNodeCommuunicationsClient(conn)

	return client, nil
}

func fethCoinType(coinTypes []store.CoinType, currencyID, networkID int) (string, error) {
	for _, ct := range coinTypes {
		if ct.СurrencyID == currencyID && ct.NetworkID == networkID {
			return ct.GRPCUrl, nil
		}
	}
	return "", fmt.Errorf("fethCoinType: no such coin in config")
}

// BtcTransaction stuct for ws notifications
type BtcTransaction struct {
	TransactionType int    `json:"transactionType"`
	Amount          int64  `json:"amount"`
	TxID            string `json:"txid"`
	Address         string `json:"address"`
}

// BtcTransactionWithUserID sub-stuct for ws notifications
type BtcTransactionWithUserID struct {
	NotificationMsg *BtcTransaction
	UserID          string
}
