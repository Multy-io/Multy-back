/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package nseth

import (
	"github.com/Multy-io/Multy-back/store"
)

// Configuration is a struct with all service options
type Configuration struct {
	CanaryTest      bool
	Name            string
	GrpcPort        string
	MultisigFactory string
	EthConf         Conf
	ServiceInfo     store.ServiceInfo
	NetworkID       int
	ResyncUrl       string
	AbiClientUrl    string
	EtherscanAPIURL string
	EtherscanAPIKey string
	PprofPort       string
}
