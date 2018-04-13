/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package main

import (
	"fmt"

	"github.com/Appscrunch/Multy-ETH-node-service"
	"github.com/KristinaEtc/config"
	"github.com/KristinaEtc/slf"
	_ "github.com/KristinaEtc/slflog"
)

var (
	log = slf.WithContext("main")

	branch    string
	commit    string
	buildtime string
)

// TODO: add all default params
var globalOpt = node.Configuration{
	Name: "eth-node-service",
}

func main() {
	config.ReadGlobalConfig(&globalOpt, "multy configuration")
	log.Error("--------------------------------new multy back server session")
	log.Infof("CONFIGURATION=%+v", globalOpt)

	log.Infof("branch: %s", branch)
	log.Infof("commit: %s", commit)
	log.Infof("build time: %s", buildtime)

	node, err := node.Init(&globalOpt)
	if err != nil {
		log.Fatalf("Server initialization: %s\n", err.Error())
	}
	fmt.Println(node)

	block := make(chan bool)
	<-block
}
