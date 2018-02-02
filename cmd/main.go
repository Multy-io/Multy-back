/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package main

import (
	"github.com/KristinaEtc/config"
	_ "github.com/KristinaEtc/slflog"

	multy "github.com/Appscrunch/Multy-back"
	"github.com/Appscrunch/Multy-back/client"
	"github.com/Appscrunch/Multy-back/eth"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
)

var (
	log = slf.WithContext("main")

	branch    string
	commit    string
	buildtime string
)

// TODO: add all default params
var globalOpt = multy.Configuration{
	Name: "my-test-back",
	Database: store.Conf{
		Address:             "localhost:27017",
		DBUsers:             "userDB-test",
		DBFeeRates:          "BTCMempool-test",
		DBTx:                "DBTx-test",
		DBStockExchangeRate: "dev-DBStockExchangeRate",
	},
	RestAddress:  "localhost:7778",
	SocketioAddr: "localhost:7780",
	BTCAPITest: client.BTCApiConf{
		Token: "btc-test-token",
		Coin:  "btc",
		Chain: "test3",
	},
	BTCAPIMain: client.BTCApiConf{
		Token: "btc-main-token",
		Coin:  "btc",
		Chain: "main",
	},
	NSQAddress:     "nsq:4150",
	BTCNodeAddress: "localhost:18334",
	Etherium: ethereum.Conf{
		Address: "88.198.47.112",
		RpcPort: ":18545",
		WsPort:  ":8545",
	},
}

func main() {
	config.ReadGlobalConfig(&globalOpt, "multy configuration")

	log.Error("--------------------------------new multy back server session")
	log.Infof("CONFIGURATION=%+v", globalOpt)

	log.Infof("branch: %s", branch)
	log.Infof("commit: %s", commit)
	log.Infof("build time: %s", buildtime)

	mu, err := multy.Init(&globalOpt)
	if err != nil {
		log.Fatalf("Server initialization: %s\n", err.Error())
	}

	if err = mu.Run(); err != nil {
		log.Fatalf("Server running: %s\n", err.Error())
	}
}
