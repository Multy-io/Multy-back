package client

import (
	"encoding/json"
	"time"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/gin-gonic/gin"
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
func createDevice(deviceid, ip, jwt, pushToken string, deviceType int) store.Device {
	return store.Device{
		DeviceID:       deviceid,
		PushToken:      pushToken,
		JWT:            jwt,
		LastActionIP:   ip,
		LastActionTime: time.Now().Unix(),
		DeviceType:     deviceType,
	}
}

func createWallet(currencyID int, address string, addressIndex int, walletIndex int, walletName string) store.Wallet {
	return store.Wallet{
		CurrencyID:     currencyID,
		WalletIndex:    walletIndex,
		WalletName:     walletName,
		LastActionTime: time.Now().Unix(),
		DateOfCreation: time.Now().Unix(),
		Status:         store.WalletStatusOK,
		Adresses: []store.Address{
			store.Address{
				Address:      address,
				AddressIndex: addressIndex,
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
