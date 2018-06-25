/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package node

import (
	"github.com/Multy-io/Multy-ETH-node-service/eth"
	"github.com/Multy-io/Multy-back/store"
)

// Configuration is a struct with all service options
type Configuration struct {
	Name            string
	GrpcPort        string
	MultisigFactory string
	EthConf         eth.Conf
	ServiceInfo     store.ServiceInfo
}
