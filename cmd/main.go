/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package main

import (
	"github.com/jekabolt/config"
	_ "github.com/jekabolt/slflog"

	multy "github.com/Multy-io/Multy-back"
	"github.com/Multy-io/Multy-back/store"
	"github.com/jekabolt/slf"
	"fmt"
)

var (
	log = slf.WithContext("main")

	branch    string
	commit    string
	buildtime string
	lasttag   string
	// TODO: add all default params
	globalOpt = multy.Configuration{
		Name: "my-test-back",
		Database: store.Conf{
			Address:             "localhost:27017",
			DBUsers:             "userDB-test",
			DBFeeRates:          "BTCMempool-test",
			DBTx:                "DBTx-test",
			DBStockExchangeRate: "dev-DBStockExchangeRate",
			Username:            "Username",
			Password:            "Password",
		},
		RestAddress:    "localhost:7778",
		SocketioAddr:   "localhost:7780",
		NSQAddress:     "nsq:4150",
		BTCNodeAddress: "localhost:18334",
	}
)

func main() {
	config.ReadGlobalConfig(&globalOpt, "multy configuration")
<<<<<<< HEAD

	log.Error("--------------------------------new multy back server session")

	log.Infof("CONFIGURATION=%+v", globalOpt)
=======
	log.Infof("CONFIGURATION=%+v", globalOpt.SupportedNodes)
>>>>>>> release_1.3

	log.Infof("branch: %s", branch)
	log.Infof("commit: %s", commit)
	log.Infof("build time: %s", buildtime)
	log.Infof("tag: %s", lasttag)

<<<<<<< HEAD
	gracefulStop := make(chan os.Signal)

	signal.Notify(gracefulStop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-gracefulStop
		fmt.Println("")
		log.Infof("Got shutting down signal")
		log.Infof("Shutting down")
		os.Exit(1)
	}()

=======
>>>>>>> release_1.3
	sc := store.ServerConfig{
		BranchName: branch,
		CommitHash: commit,
		Build:      buildtime,
		Tag:        lasttag,
	}

	globalOpt.MultyVerison = sc

	mu, err := multy.Init(&globalOpt)
	if err != nil {
		log.Fatalf("Server initialization: %s\n", err.Error())
	}

	if err = mu.Run(); err != nil {
		log.Fatalf("Server running: %s\n", err.Error())
	}

}
