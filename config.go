/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package multyback

import (
	"github.com/Multy-io/Multy-back/client"
	"github.com/Multy-io/Multy-back/exchanger"
	"github.com/Multy-io/Multy-back/store"
	"github.com/Multy-io/Multy-back/types"
)

// Configuration is a struct with all service options
type Configuration struct {
	CanaryTest        bool
	Name              string
	Database          store.Conf
	SocketioAddr      string
	RestAddress       string
	Firebase          client.FirebaseConf
	NSQAddress        string
	BTCNodeAddress    string
	DonationAddresses []store.DonationInfo
	MultyVerison      store.ServerConfig
	ServicesInfo      []store.ServiceInfo
	Secretkey         string
	store.MobileVersions
	BrowserDefault store.BrowserDefault

	ETHDefaultGasPrice types.TransactionFeeRateEstimation

	SupportedNodes []store.CoinType
	Exchangers     []exchanger.BasicExchangeConfiguration
}
