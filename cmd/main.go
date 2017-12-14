package main

import (
	"fmt"
	"log"

	mylty "github.com/Appscrunch/Multy-back"
)

const defaultConfigFile = "config.yml"

var (
	branch    string
	commit    string
	buildtime string
)

func main() {
	serviceInfo := fmt.Sprintf("[INFO] multy back-end  service\n"+
		"\tbranch: \t%s\n"+
		"\tcommit: \t%s\n"+
		"\tbuild time: \t%s\n", branch, commit, buildtime)
	log.Println(serviceInfo)

	conf, err := mylty.GetConfig(defaultConfigFile)
	if err != nil {
		log.Fatal("[ERR] configuration init: ", err.Error())
	}
	log.Printf("[DEBUG] configuration: %+v\n", conf)

	mu, err := mylty.Init(conf)
	if err != nil {
		log.Fatal("[ERR] Server initialization: ", err.Error())
	}

	if err = mu.Run(); err != nil {
		log.Fatal("[ERR] Server running: ", err.Error())
	}
}
