/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package node

import "github.com/Appscrunch/Multy-ETH-node-service/eth"

// Configuration is a struct with all service options
type Configuration struct {
	Name     string
	GrpcPort string
	EthConf  eth.Conf
}

// // BTCApiConf provide blockcypher api
// type BTCApiConf struct {
// 	Token, Coin, Chain string
// }
