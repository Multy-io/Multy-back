/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package main

import (
	"flag"
	"fmt"

	"net/http"
	_ "net/http/pprof"

	ns "github.com/Multy-io/Multy-back/ns-eth"
	"github.com/Multy-io/Multy-back/store"
	"github.com/jekabolt/config"
	"github.com/jekabolt/slf"
	_ "github.com/jekabolt/slflog"
)

var (
	log       = slf.WithContext("main")
	branch    string
	commit    string
	buildtime string

	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

var globalOpt = ns.Configuration{
	Name: "eth-node-service",
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

	nc := ns.NodeClient{}
	node, err := nc.Init(&globalOpt)
	if err != nil {
		log.Fatalf("Server initialization: %s\n", err.Error())
	}
	fmt.Println(node)

	http.ListenAndServe(globalOpt.PprofPort, nil)

	block := make(chan bool)
	<-block
}
