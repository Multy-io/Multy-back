/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package node

// Configuration is a struct with all service options
type Configuration struct {
	Name           string
	GrpcPort       string
	BTCSertificate string
	BTCNodeAddress string
	BTCAPI         BTCApiConf
}

// BTCApiConf provide blockcypher api
type BTCApiConf struct {
	Token, Coin, Chain string
}
