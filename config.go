/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package node

import "github.com/Appscrunch/Multy-back/store"

// Configuration is a struct with all service options
type Configuration struct {
	Name           string
	GrpcPort       string
	BTCSertificate string
	BTCNodeAddress string
	BTCAPI         BTCApiConf
	ServiceInfo    store.ServiceInfo
}

// BTCApiConf provide blockcypher api
type BTCApiConf struct {
	Token, Coin, Chain string
}
