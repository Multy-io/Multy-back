/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package main

import (
	"fmt"

	"github.com/KristinaEtc/config"
	"github.com/KristinaEtc/slf"
	_ "github.com/KristinaEtc/slflog"
	"github.com/Multy-io/Multy-BTC-node-service"
	"github.com/Multy-io/Multy-back/store"
)

var (
	log = slf.WithContext("main")

	branch    string
	commit    string
	buildtime string
)

// TODO: add all default params
var globalOpt = node.Configuration{
	Name: "my-test-back",
}

func main() {
	config.ReadGlobalConfig(&globalOpt, "multy configuration")
	log.Infof("CONFIGURATION=%+v", globalOpt)
	log.Infof("branch: %s", branch)
	log.Infof("commit: %s", commit)
	log.Infof("build time: %s", buildtime)
	globalOpt.ServiceInfo = store.ServiceInfo{
		Branch:    branch,
		Commit:    commit,
		Buildtime: buildtime,
	}

	node, err := node.Init(&globalOpt)
	if err != nil {
		log.Fatalf("Server initialization: %s\n", err.Error())
	}
	fmt.Println(node)

	block := make(chan bool)
	<-block
}
