/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
	nsq "github.com/bitly/go-nsq"
	"github.com/btcsuite/btcd/rpcclient"
)

const (
	// TopicTransaction is a topic for sending notifies to clients
	TopicTransaction = "btcTransactionUpdate"
)

// Dirty hack - this will be wrapped to a struct
var (
	rpcClient   = &rpcclient.Client{}
	nsqProducer *nsq.Producer // a producer for sending data to clients
	rpcConf     *rpcclient.ConnConfig
)

var log = slf.WithContext("btc")

func InitHandlers(certFromConf string, dbConf *store.Conf, nsqAddr, btcNodeAddress string) (*rpcclient.Client, error) {
	config := nsq.NewConfig()
	p, err := nsq.NewProducer(nsqAddr, config)
	if err != nil {
		return nil, fmt.Errorf("nsq producer: %s", err.Error())
	}
	nsqProducer = p

	Cert = certFromConf
	connCfg.Certificates = []byte(Cert)
	log.Infof("cert=%s\n", Cert)

	db, err := mgo.Dial(dbConf.Address)
	if err != nil {
		log.Errorf("RunProcess: Cand connect to DB: %s", err.Error())
		return nil, err
	}

	usersData = db.DB(dbConf.DBUsers).C(store.TableUsers) // all db tables
	mempoolRates = db.DB(dbConf.DBFeeRates).C(store.TableFeeRates)
	txsData = db.DB(dbConf.DBTx).C(store.TableBTC)
	exRate = db.DB("dev-DBStockExchangeRate").C("TableStockExchangeRate")

	go RunProcess(btcNodeAddress)
	return rpcClient, nil
}

type BtcTransaction struct {
	TransactionType int     `json:"transactionType"`
	Amount          float64 `json:"amount"`
	TxID            string  `json:"txid"`
	Address         string  `json:"address"`
}

type BtcTransactionWithUserID struct {
	NotificationMsg *BtcTransaction
	UserID          string
}
