/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package multyback

import (
	"github.com/Multy-io/Multy-back/client"
	exchangerCommon "github.com/Multy-io/Multy-back/exchanger/common"
	"github.com/Multy-io/Multy-back/store"
)

// Configuration is a struct with all service options
type Configuration struct {
	Name              string
	Database          store.Conf
	SocketioAddr      string
	RestAddress       string
	Firebase          client.FirebaseConf
	NSQAddress        string
	BTCNodeAddress    string
	DonationAddresses []store.DonationInfo
	MultyVerison      store.ServerConfig
	ServicesInfo []store.ServiceInfo
	Secretkey    string
	store.MobileVersions
	BrowserDefault store.BrowserDefault

	SupportedNodes []store.CoinType
	Exchangers	[]exchangerCommon.BasicExchangeConfiguration
}
