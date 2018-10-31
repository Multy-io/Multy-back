/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"net/http"
	_ "net/http/pprof"

	"github.com/Multy-io/Multy-ETH-node-service"
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
	globalOpt.ServiceInfo = store.ServiceInfo{
		Branch:    branch,
		Commit:    commit,
		Buildtime: buildtime,
	}

	flag.Parse()
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatalf("could not create memory profile: %v", err.Error())
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatalf("could not write memory profile: %v", err.Error())
		}
		f.Close()
	}

	nc := node.NodeClient{}
	node, err := nc.Init(&globalOpt)
	if err != nil {
		log.Fatalf("Server initialization: %s\n", err.Error())
	}
	fmt.Println(node)

	http.ListenAndServe(globalOpt.PprofPort, nil)

	block := make(chan bool)
	<-block
}
