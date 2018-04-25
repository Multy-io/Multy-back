/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func decodeBody(c *gin.Context, to interface{}) error {
	body := json.NewDecoder(c.Request.Body)
	err := body.Decode(to)
	defer c.Request.Body.Close()
	return err
}

func createUser(userid string, device []store.Device, wallets []store.Wallet) store.User {
	return store.User{
		UserID:  userid,
		Devices: device,
		Wallets: wallets,
	}
}
func createDevice(deviceid, ip, jwt, pushToken, appVersion string, deviceType int) store.Device {
	return store.Device{
		DeviceID:       deviceid,
		PushToken:      pushToken,
		JWT:            jwt,
		LastActionIP:   ip,
		LastActionTime: time.Now().Unix(),
		AppVersion:     appVersion,
		DeviceType:     deviceType,
	}
}

func createWallet(currencyID, networkID int, address string, addressIndex int, walletIndex int, walletName string) store.Wallet {

	return store.Wallet{
		CurrencyID:     currencyID,
		NetworkID:      networkID,
		WalletIndex:    walletIndex,
		WalletName:     walletName,
		LastActionTime: time.Now().Unix(),
		DateOfCreation: time.Now().Unix(),
		Status:         store.WalletStatusOK,
		Adresses: []store.Address{
			store.Address{
				Address:        address,
				AddressIndex:   addressIndex,
				LastActionTime: time.Now().Unix(),
			},
		},
	}

}

func newEmptyTx(userID string) store.TxRecord {
	return store.TxRecord{
		UserID:       userID,
		Transactions: []store.MultyTX{},
	}
}

func newWebSocketConn(addr string) (*websocket.Conn, error) {
	fmt.Printf("addr=%s", addr)
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func reconnectWebSocketConn(addr string, log slf.StructuredLogger) (*websocket.Conn, error) {
	var (
		c   *websocket.Conn
		err error

		secToRecon = time.Duration(time.Second * 2) // start time for reconnect function
		numOfRecon = 0
	)

	for {
		c, err = newWebSocketConn(addr)
		if err == nil {
			return c, nil
		}

		log.Errorf("reconnecting: %s", err)
		log.Errorf("secToRecon=%f/numOfRecon=%d", secToRecon.Seconds(), numOfRecon)
		ticker := time.NewTicker(secToRecon)
		select {
		case _ = <-ticker.C:
			if secToRecon < backOffLimit {
				randomAdd := secToRecon / 100 * (20 + time.Duration(r1.Int31n(10)))
				secToRecon = secToRecon*2 + time.Duration(randomAdd)
				numOfRecon++
			} else {
				// back off limit was reached
				return nil, err
			}
		}
	}
}
