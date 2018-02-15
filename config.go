/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package multyback

import (
	"github.com/Appscrunch/Multy-back/client"
	"github.com/Appscrunch/Multy-back/eth"
	"github.com/Appscrunch/Multy-back/store"
)

// Configuration is a struct with all service options
type Configuration struct {
	Name              string
	Database          store.Conf
	SocketioAddr      string
	RestAddress       string
	BTCAPIMain        client.BTCApiConf
	BTCAPITest        client.BTCApiConf
	Firebase          client.FirebaseConf
	Etherium          ethereum.Conf
	BTCSertificate    string
	NSQAddress        string
	BTCNodeAddress    string
	DonationAddresses map[string]string
}
