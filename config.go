/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package multyback

import (
	"github.com/Appscrunch/Multy-back/client"
	"github.com/Appscrunch/Multy-back/store"
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
	DonationAddresses map[string]string

	SupportedNodes []CoinType
}

type CoinType struct {
	СurrencyID int
	NetworkID  int
	SocketPort int
	SocketURL  string
}
