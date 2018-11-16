/*
Copyright 2019 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"fmt"
	"sync"

	"google.golang.org/grpc"
	mgo "gopkg.in/mgo.v2"

	pb "github.com/Multy-io/Multy-BTC-node-service/node-streamer"
	"github.com/Multy-io/Multy-back/currencies"
	"github.com/Multy-io/Multy-back/store"
	nsq "github.com/bitly/go-nsq"
	"github.com/graarh/golang-socketio"
	"github.com/jekabolt/slf"
)

// BTCConn is a main struct of package
type BTCConn struct {
	NsqProducer      *nsq.Producer // a producer for sending data to clients
	CliTest          pb.NodeCommunicationsClient
	CliMain          pb.NodeCommunicationsClient
	WatchAddressTest chan pb.WatchAddress
	WatchAddressMain chan pb.WatchAddress

	BtcMempool     sync.Map
	BtcMempoolTest sync.Map

	Resync *sync.Map

	WsServer *gosocketio.Server
}

var log = slf.WithContext("btc")

//InitHandlers init nsq mongo and ws connection to node
// return main client , test client , err
func InitHandlers(dbConf *store.Conf, coinTypes []store.CoinType, nsqAddr string) (*BTCConn, error) {
	//declare pacakge struct
	cli := &BTCConn{
		BtcMempool:     sync.Map{},
		BtcMempoolTest: sync.Map{},
		Resync:         &sync.Map{},
	}

	cli.WatchAddressMain = make(chan pb.WatchAddress)
	cli.WatchAddressTest = make(chan pb.WatchAddress)

	config := nsq.NewConfig()
	p, err := nsq.NewProducer(nsqAddr, config)
	if err != nil {
		return cli, fmt.Errorf("nsq producer: %s", err.Error())
	}

	cli.NsqProducer = p
	log.Infof("InitHandlers: nsq.NewProducer: √")

	addr := []string{dbConf.Address}

	mongoDBDial := &mgo.DialInfo{
		Addrs:    addr,
		Username: dbConf.Username,
		Password: dbConf.Password,
	}

	db, err := mgo.DialWithInfo(mongoDBDial)
	if err != nil {
		return nil, err
	}

	// HACK: this made to acknowledge that queried data has already inserted to db
	db.SetSafe(&mgo.Safe{
		W:        1,
		WTimeout: 100,
		J:        true,
	})
	if err != nil {
		log.Errorf("RunProcess: can't connect to DB: %s", err.Error())
		return cli, fmt.Errorf("mgo.Dial: %s", err.Error())
	}
	log.Infof("InitHandlers: mgo.Dial: √")

	usersData = db.DB(dbConf.DBUsers).C(store.TableUsers) // all db tables
	exRate = db.DB(dbConf.DBStockExchangeRate).C("TableStockExchangeRate")

	// main
	txsData = db.DB(dbConf.DBTx).C(dbConf.TableTxsDataBTCMain)
	spendableOutputs = db.DB(dbConf.DBTx).C(dbConf.TableSpendableOutputsBTCMain)
	spentOutputs = db.DB(dbConf.DBTx).C(dbConf.TableSpentOutputsBTCMain)

	// test
	txsDataTest = db.DB(dbConf.DBTx).C(dbConf.TableTxsDataBTCTest)
	spendableOutputsTest = db.DB(dbConf.DBTx).C(dbConf.TableSpendableOutputsBTCTest)
	spentOutputsTest = db.DB(dbConf.DBTx).C(dbConf.TableSpentOutputsBTCTest)

	restoreState = db.DB(dbConf.DBRestoreState).C(dbConf.TableState)

	// setup main net
	coinTypeMain, err := store.FetchCoinType(coinTypes, currencies.Bitcoin, currencies.Main)
	if err != nil {
		return cli, fmt.Errorf("fetchCoinType: %s", err.Error())
	}

	cliMain, err := initGrpcClient(coinTypeMain.GRPCUrl)
	if err != nil {
		return cli, fmt.Errorf("initGrpcClient: %s", err.Error())
	}
	cli.CliMain = cliMain

	// setup testnet
	coinTypeTest, err := store.FetchCoinType(coinTypes, currencies.Bitcoin, currencies.Test)
	if err != nil {
		return cli, fmt.Errorf("fetchCoinType: %s", err.Error())
	}
	cliTest, err := initGrpcClient(coinTypeTest.GRPCUrl)
	if err != nil {
		return cli, fmt.Errorf("initGrpcClient: %s", err.Error())
	}

	cli.CliTest = cliTest

	cli.setGRPCHandlers(currencies.Test, coinTypeTest.AccuracyRange)
	log.Infof("InitHandlers: initGrpcClient: Test: √")

	cli.setGRPCHandlers(currencies.Main, coinTypeMain.AccuracyRange)
	log.Infof("InitHandlers: initGrpcClient: Main: √")

	return cli, nil
}

func initGrpcClient(url string) (pb.NodeCommunicationsClient, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		log.Errorf("initGrpcClient: grpc.Dial: %s", err.Error())
		return nil, err
	}

	// Create a new  client
	client := pb.NewNodeCommunicationsClient(conn)
	return client, nil
}
