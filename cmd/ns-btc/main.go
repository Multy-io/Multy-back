/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package main

import (
	"fmt"

	nsbtc "github.com/Multy-io/Multy-back/ns-btc"
	"github.com/Multy-io/Multy-back/store"
	"github.com/jekabolt/config"
	"github.com/jekabolt/slf"
	_ "github.com/jekabolt/slflog"
)

var (
	log = slf.WithContext("ns-btc").WithCaller(slf.CallerShort)

	// Set externaly during build
	branch    string
	commit    string
	lasttag   string
	buildtime string
)

// TODO: add all default params
var globalOpt = nsbtc.Configuration{
	CanaryTest: false,
	Name:       "my-test-back",
}

func main() {
	log.Info("============================================================")
	log.Info("Node service BTC starting")
	log.Infof("branch: %s", branch)
	log.Infof("commit: %s", commit)
	log.Infof("build time: %s", buildtime)

	log.Info("Reading configuration...")

	config.ReadGlobalConfig(&globalOpt, "multy configuration")
	log.Infof("CONFIGURATION=%+v", globalOpt)

	if globalOpt.CanaryTest == true {
		log.Info("This is a CanaryTest run, quitting immediatelly...")
		return
	}

	globalOpt.ServiceInfo = store.ServiceInfo{
		Branch:    branch,
		Commit:    commit,
		Buildtime: buildtime,
	}

	nc := nsbtc.NodeClient{}
	node, err := nc.Init(&globalOpt)
	if err != nil {
		log.Fatalf("Server initialization: %s\n", err.Error())
	}
	fmt.Println(node)

	block := make(chan struct{})
	<-block
}
