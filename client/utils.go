package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/gin-gonic/gin"
)

func responseErr(c *gin.Context, err error, code int) {
	if err != nil {
		c.JSON(code, gin.H{
			"code":    code,
			"message": http.StatusText(code),
		})
	}

	return
}

func decodeBody(c *gin.Context, to interface{}) error {
	body := json.NewDecoder(c.Request.Body)
	err := body.Decode(to)
	defer c.Request.Body.Close()
	return err
}

func makeRequest(c *gin.Context, url string, to interface{}) {
	response, err := http.Get(url)
	responseErr(c, err, http.StatusServiceUnavailable) // 503

	data, err := ioutil.ReadAll(response.Body)
	responseErr(c, err, http.StatusInternalServerError) // 500

	err = json.Unmarshal(data, to)
	responseErr(c, err, http.StatusInternalServerError) // 500
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
		LastActionTime: time.Now(),
		DeviceType:     deviceType,
	}
}

func createWallet(currencyID int, address string, addressIndex int, walletIndex int, walletName string) store.Wallet {
	return store.Wallet{
		CurrencyID:     currencyID,
		WalletIndex:    walletIndex,
		WalletName:     walletName,
		LastActionTime: time.Now(),
		DateOfCreation: time.Now(),
		Adresses: []store.Address{
			store.Address{
				Address:      address,
				AddressIndex: addressIndex,
			},
		},
	}
}
