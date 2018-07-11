/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package client

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Multy-io/Multy-back/btc"
	"github.com/Multy-io/Multy-back/currencies"
	"github.com/Multy-io/Multy-back/eth"
	"github.com/Multy-io/Multy-back/store"
	"github.com/jekabolt/slf"

	btcpb "github.com/Multy-io/Multy-back/node-streamer/btc"
	ethpb "github.com/Multy-io/Multy-back/node-streamer/eth"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin" // gin-swagger middleware
	// swagger embed files
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	msgErrMissingRequestParams  = "missing request parametrs"
	msgErrServerError           = "internal server error"
	msgErrNoWallet              = "no such wallet"
	msgErrWalletNonZeroBalance  = "can't delete non zero balance wallet"
	msgErrWalletIndex           = "already existing wallet index"
	msgErrTxHistory             = "not found any transaction history"
	msgErrAddressIndex          = "already existing address index"
	msgErrMethodNotImplennted   = "method is not implemented"
	msgErrHeaderError           = "wrong authorization headers"
	msgErrRequestBodyError      = "missing request body params"
	msgErrUserNotFound          = "user not found in db"
	msgErrNoTransactionAddress  = "zero balance"
	msgErrNoSpendableOutputs    = "no spendable outputs"
	msgErrRatesError            = "internal server error rates"
	msgErrDecodeWalletIndexErr  = "wrong wallet index"
	msgErrDecodeNetworkIDErr    = "wrong network id"
	msgErrNoSpendableOuts       = "no spendable outputs"
	msgErrDecodeCurIndexErr     = "wrong currency index"
	msgErrDecodenetworkidErr    = "wrong networkid index"
	msgErrAdressBalance         = "empty address or 3-rd party server error"
	msgErrChainIsNotImplemented = "current chain is not implemented"
	msgErrUserHaveNoTxs         = "user have no transactions"
)

type RestClient struct {
	middlewareJWT *GinJWTMiddleware
	userStore     store.UserStore
	rpcClient     *rpcclient.Client
	// // ballance api for test net
	// apiBTCTest     gobcy.API
	btcConfTestnet BTCApiConf
	// // ballance api for main net
	// apiBTCMain     gobcy.API
	btcConfMainnet BTCApiConf

	log slf.StructuredLogger

	donationAddresses []store.DonationInfo

	BTC          *btc.BTCConn
	ETH          *eth.ETHConn
	MultyVerison store.ServerConfig
	Secretkey    string
}

type BTCApiConf struct {
	Token, Coin, Chain string
}

func SetRestHandlers(
	userDB store.UserStore,
	r *gin.Engine,
	donationAddresses []store.DonationInfo,
	btc *btc.BTCConn,
	eth *eth.ETHConn,
	mv store.ServerConfig,
	secretkey string,
) (*RestClient, error) {
	restClient := &RestClient{
		userStore:         userDB,
		log:               slf.WithContext("rest-client"),
		donationAddresses: donationAddresses,
		BTC:               btc,
		ETH:               eth,
		MultyVerison:      mv,
		Secretkey:         secretkey,
	}
	initMiddlewareJWT(restClient)

	r.POST("/auth", restClient.LoginHandler())
	r.GET("/server/config", restClient.getServerConfig())

	r.GET("/donations", restClient.donations())

	v1 := r.Group("/api/v1")
	v1.Use(restClient.middlewareJWT.MiddlewareFunc())
	{
		v1.POST("/wallet", restClient.addWallet())
		v1.DELETE("/wallet/:currencyid/:networkid/:walletindex", restClient.deleteWallet())
		v1.POST("/address", restClient.addAddress())
		v1.GET("/transaction/feerate/:currencyid/:networkid", restClient.getFeeRate())
		v1.GET("/outputs/spendable/:currencyid/:networkid/:addr", restClient.getSpendableOutputs())
		v1.POST("/transaction/send", restClient.sendRawHDTransaction())
		v1.GET("/wallet/:walletindex/verbose/:currencyid/:networkid", restClient.getWalletVerbose())
		v1.GET("/wallets/verbose", restClient.getAllWalletsVerbose())
		v1.GET("/wallets/transactions/:currencyid/:networkid/:walletindex", restClient.getWalletTransactionsHistory())
		v1.POST("/wallet/name", restClient.changeWalletName())
		v1.POST("/resync/wallet/:currencyid/:networkid/:walletindex", restClient.resyncWallet())
		v1.GET("/exchange/changelly/list", restClient.changellyListCurrencies())
	}
	return restClient, nil
}

func initMiddlewareJWT(restClient *RestClient) {
	restClient.middlewareJWT = &GinJWTMiddleware{
		Realm:      "test zone",
		Key:        []byte(restClient.Secretkey), // config
		Timeout:    time.Hour,
		MaxRefresh: time.Hour,
		Authenticator: func(userId, deviceId, pushToken string, deviceType int, c *gin.Context) (store.User, bool) {
			query := bson.M{"userID": userId}

			user := store.User{}

			err := restClient.userStore.FindUser(query, &user)

			if err != nil || len(user.UserID) == 0 {
				return user, false
			}
			return user, true
		},
		Unauthorized: nil,
		TokenLookup:  "header:Authorization",

		TokenHeadName: "Bearer",

		TimeFunc: time.Now,
	}
}

type WalletParams struct {
	CurrencyID   int    `json:"currencyID"`
	NetworkID    int    `json:"networkID"`
	Address      string `json:"address"`
	AddressIndex int    `json:"addressIndex"`
	WalletIndex  int    `json:"walletIndex"`
	WalletName   string `json:"walletName"`
}

type SelectWallet struct {
	CurrencyID   int    `json:"currencyID"`
	NetworkID    int    `json:"networkID"`
	WalletIndex  int    `json:"walletIndex"`
	Address      string `json:"address"`
	AddressIndex int    `json:"addressIndex"`
}

type EstimationSpeeds struct {
	VerySlow int
	Slow     int
	Medium   int
	Fast     int
	VeryFast int
}

type EstimationSpeedsETH struct {
	VerySlow string
	Slow     string
	Medium   string
	Fast     string
	VeryFast string
}

type Tx struct {
	Transaction   string `json:"transaction"`
	AllowHighFees bool   `json:"allowHighFees"`
}

type DisplayWallet struct {
	Chain    string          `json:"chain"`
	Adresses []store.Address `json:"addresses"`
}

type WalletExtended struct {
	CuurencyID  int         `bson:"chain"`       //cuurencyID
	WalletIndex int         `bson:"walletIndex"` //walletIndex
	Addresses   []AddressEx `bson:"addresses"`
}

type AddressEx struct { // extended
	AddressID int    `bson:"addressID"` //addressIndex
	Address   string `bson:"address"`
	Amount    int    `bson:"amount"` //float64
}

func getToken(c *gin.Context) (string, error) {
	authHeader := strings.Split(c.GetHeader("Authorization"), " ")
	if len(authHeader) < 2 {
		return "", errors.New(msgErrHeaderError)
	}
	return authHeader[1], nil
}

func createCustomWallet(wp WalletParams, token string, restClient *RestClient, c *gin.Context) error {
	user := store.User{}
	query := bson.M{"devices.JWT": token}

	err := restClient.userStore.FindUser(query, &user)
	if err != nil {
		restClient.log.Errorf("createCustomWallet: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		err = errors.New(msgErrUserNotFound)
		return err
	}

	for _, wallet := range user.Wallets {
		if wallet.CurrencyID == wp.CurrencyID && wallet.NetworkID == wp.NetworkID && wallet.WalletIndex == wp.WalletIndex {
			err = errors.New(msgErrWalletIndex)
			return err
		}
	}

	sel := bson.M{"devices.JWT": token}
	wallet := createWallet(wp.CurrencyID, wp.NetworkID, wp.Address, wp.AddressIndex, wp.WalletIndex, wp.WalletName)
	update := bson.M{"$push": bson.M{"wallets": wallet}}

	err = restClient.userStore.Update(sel, update)
	if err != nil {
		restClient.log.Errorf("addWallet: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		err := errors.New(msgErrServerError)
		return err
	}

	err = AddWatchAndResync(wp.CurrencyID, wp.NetworkID, wp.WalletIndex, wp.AddressIndex, user.UserID, wp.Address, restClient)
	if err != nil {
		restClient.log.Errorf("createCustomWallet: AddWatchAndResync: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		err := errors.New(msgErrServerError)
		return err
	}

	return nil
}

func changeName(cn ChangeName, token string, restClient *RestClient, c *gin.Context) error {
	user := store.User{}
	query := bson.M{"devices.JWT": token}

	if err := restClient.userStore.FindUser(query, &user); err != nil {
		restClient.log.Errorf("deleteWallet: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		err := errors.New(msgErrUserNotFound)
		return err
	}
	var position int

	// change multisig name
	if cn.Address != "" {
		query := bson.M{"userID": user.UserID, "multisig.contractaddress": cn.Address}
		update := bson.M{
			"$set": bson.M{
				"multisig.$.walletname": cn.WalletName,
			},
		}

		err := restClient.userStore.Update(query, update)
		return err
	}

	// change wallet name
	if cn.Address == "" {

		for i, wallet := range user.Wallets {
			if wallet.NetworkID == cn.NetworkID && wallet.WalletIndex == cn.WalletIndex && wallet.CurrencyID == cn.CurrencyID {
				position = i
				break
			}
		}
		sel := bson.M{"userID": user.UserID, "wallets.walletIndex": cn.WalletIndex, "wallets.networkID": cn.NetworkID}
		update := bson.M{
			"$set": bson.M{
				"wallets." + strconv.Itoa(position) + ".walletName": cn.WalletName,
			},
		}
		return restClient.userStore.Update(sel, update)
	}

	return errors.New(msgErrNoWallet)

}

func addAddressToWallet(address, token string, currencyID, networkid, walletIndex, addressIndex int, restClient *RestClient, c *gin.Context) error {
	user := store.User{}
	query := bson.M{"devices.JWT": token}

	if err := restClient.userStore.FindUser(query, &user); err != nil {
		// restClient.log.Errorf("deleteWallet: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		return errors.New(msgErrUserNotFound)
	}

	var position int
	for i, wallet := range user.Wallets {
		if wallet.NetworkID == networkid && wallet.CurrencyID == currencyID && wallet.WalletIndex == walletIndex {
			position = i
			for _, walletAddress := range wallet.Adresses {
				if walletAddress.AddressIndex == addressIndex {
					return errors.New(msgErrAddressIndex)
				}
			}
		}
	}

	addr := store.Address{
		Address:        address,
		AddressIndex:   addressIndex,
		LastActionTime: time.Now().Unix(),
	}

	//TODO: make no possibility to add eth address
	sel := bson.M{"devices.JWT": token, "wallets.currencyID": currencyID, "wallets.networkID": networkid, "wallets.walletIndex": walletIndex}
	update := bson.M{"$push": bson.M{"wallets." + strconv.Itoa(position) + ".addresses": addr}}
	if err := restClient.userStore.Update(sel, update); err != nil {
		// restClient.log.Errorf("addAddressToWallet: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		return errors.New(msgErrServerError)
	}

	return nil

	// return AddWatchAndResync(currencyID, networkid, walletIndex, addressIndex, user.UserID, address, restClient)

}

func AddWatchAndResync(currencyID, networkid, walletIndex, addressIndex int, userid, address string, restClient *RestClient) error {

	err := NewAddressNode(address, userid, currencyID, networkid, walletIndex, addressIndex, restClient)
	if err != nil {
		restClient.log.Errorf("AddWatchAndResync: NewAddressWs: currencies.Main: WsBtcMainnetCli.Emit:resync %s\t", err.Error())
		return err
	}

	return nil
}

func NewAddressNode(address, userid string, currencyID, networkID, walletIndex, addressIndex int, restClient *RestClient) error {

	//add new re-sync to map
	restClient.BTC.Resync.Store(address, true)

	switch currencyID {
	case currencies.Bitcoin:
		if networkID == currencies.Main {
			restClient.BTC.WatchAddressMain <- btcpb.WatchAddress{
				Address:      address,
				UserID:       userid,
				WalletIndex:  int32(walletIndex),
				AddressIndex: int32(addressIndex),
			}
		}

		if networkID == currencies.Test {
			restClient.BTC.WatchAddressTest <- btcpb.WatchAddress{
				Address:      address,
				UserID:       userid,
				WalletIndex:  int32(walletIndex),
				AddressIndex: int32(addressIndex),
			}
		}
	case currencies.Ether:
		if networkID == currencies.ETHMain {
			restClient.ETH.WatchAddressMain <- ethpb.WatchAddress{
				Address:      address,
				UserID:       userid,
				WalletIndex:  int32(walletIndex),
				AddressIndex: int32(addressIndex),
			}
		}

		if networkID == currencies.ETHTest {
			restClient.ETH.WatchAddressTest <- ethpb.WatchAddress{
				Address:      address,
				UserID:       userid,
				WalletIndex:  int32(walletIndex),
				AddressIndex: int32(addressIndex),
			}
		}
	}
	return nil
}

func (restClient *RestClient) addWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getToken(c)
		if err != nil {
			restClient.log.Errorf("addWallet: getToken: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
			})
			return
		}

		var (
			code    = http.StatusOK
			message = http.StatusText(http.StatusOK)
		)

		var wp WalletParams

		err = decodeBody(c, &wp)
		if err != nil {
			restClient.log.Errorf("addWallet: decodeBody: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrRequestBodyError,
			})
			return
		}

		err = createCustomWallet(wp, token, restClient, c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"code":    code,
			"tine":    time.Now().Unix(),
			"message": message,
		})
		return
	}
}

type ChangeName struct {
	WalletName  string `json:"walletname"`
	CurrencyID  int    `json:"currencyID"`
	WalletIndex int    `json:"walletIndex"`
	NetworkID   int    `json:"networkId"`
	Address     string `json:"address"`
}

func (restClient *RestClient) changeWalletName() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
			})
			return
		}

		var cn ChangeName
		err = decodeBody(c, &cn)
		if err != nil {
			restClient.log.Errorf("changeWalletName: decodeBody: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrRequestBodyError,
			})
			return
		}
		err = changeName(cn, token, restClient, c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": http.StatusText(http.StatusOK),
		})

	}
}

func (restClient *RestClient) donations() gin.HandlerFunc {
	return func(c *gin.Context) {
		donationInfo := []store.Donation{}
		for _, da := range restClient.donationAddresses {
			b := checkBTCAddressbalance(da.DonationAddress, currencies.Bitcoin, currencies.Main, restClient)
			donationInfo = append(donationInfo, store.Donation{
				FeatureID: da.FeatureCode,
				Address:   da.DonationAddress,
				Amount:    b,
				Status:    1,
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"code":      http.StatusOK,
			"message":   http.StatusText(http.StatusOK),
			"donations": donationInfo,
		})
	}
}

func (restClient *RestClient) getServerConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		resp := map[string]interface{}{
			"stockexchanges": map[string][]string{
				"poloniex": []string{"usd_btc", "eth_btc", "eth_usd", "btc_usd"},
				"gdax":     []string{"eur_btc", "usd_btc", "eth_btc", "eth_usd", "eth_eur", "btc_usd"},
			},
			"servertime": time.Now().UTC().Unix(),
			"api":        "0.01",
			"android": map[string]int{
				"soft": 7,
				"hard": 7,
			},
			"version": restClient.MultyVerison,
			"ios": map[string]int{
				"soft": 49,
				"hard": 49,
			},
			"donate": restClient.donationAddresses,
		}
		c.JSON(http.StatusOK, resp)
	}
}

func checkBTCAddressbalance(address string, currencyID, networkid int, restClient *RestClient) int64 {
	var balance int64
	spOuts, err := restClient.userStore.GetAddressSpendableOutputs(address, currencyID, networkid)
	if err != nil {
		return balance
	}

	for _, out := range spOuts {
		balance += out.TxOutAmount
	}
	return balance
}

func getBTCAddressSpendableOutputs(address string, currencyID, networkID int, restClient *RestClient) []store.SpendableOutputs {
	spOuts, err := restClient.userStore.GetAddressSpendableOutputs(address, currencyID, networkID)
	if err != nil && err != mgo.ErrNotFound {
		restClient.log.Errorf("getBTCAddressSpendableOutputs: GetAddressSpendableOutputs: %s\t", err.Error())
	}
	return spOuts
}

func (restClient *RestClient) deleteWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
			})
			return
		}

		var multisigAddress string
		walletIndex, err := strconv.Atoi(c.Param("walletindex"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[walletindexr=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			multisigAddress = c.Param("walletindex")
		}

		currencyId, err := strconv.Atoi(c.Param("currencyid"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[currencyId=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int currency id:[%d] %s \t[addr=%s]", currencyId, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeCurIndexErr,
			})
			return
		}

		networkid, err := strconv.Atoi(c.Param("networkid"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[networkid=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int networkid index:[%d] %s \t[addr=%s]", networkid, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodenetworkidErr,
			})
			return
		}

		var (
			code    int
			message string
		)

		user := store.User{}
		query := bson.M{"devices.JWT": token}
		if err := restClient.userStore.FindUser(query, &user); err != nil {
			restClient.log.Errorf("deleteWallet: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrUserNotFound,
			})
			return
		}
		code = http.StatusOK
		message = http.StatusText(http.StatusOK)

		var totalBalance int64

		switch currencyId {
		case currencies.Bitcoin:

			if networkid == currencies.Main {
				for _, wallet := range user.Wallets {
					if wallet.WalletIndex == walletIndex {
						for _, address := range wallet.Adresses {
							totalBalance += checkBTCAddressbalance(address.Address, currencyId, networkid, restClient)
						}
					}
				}
			}
			if networkid == currencies.Test {
				for _, wallet := range user.Wallets {
					if wallet.WalletIndex == walletIndex {
						for _, address := range wallet.Adresses {
							totalBalance += checkBTCAddressbalance(address.Address, currencyId, networkid, restClient)
						}
					}
				}
			}

			if totalBalance == 0 {
				err := restClient.userStore.DeleteWallet(user.UserID, "", walletIndex, currencyId, networkid)
				if err != nil {
					restClient.log.Errorf("deleteWallet: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
					c.JSON(http.StatusBadRequest, gin.H{
						"code":    http.StatusInternalServerError,
						"message": msgErrNoWallet,
					})
					return
				}
			}

			if totalBalance != 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": msgErrWalletNonZeroBalance,
				})
				return
			}

			code = http.StatusOK
			message = http.StatusText(http.StatusOK)

		case currencies.Ether:

			var address string
			// delete multisig
			if multisigAddress != "" {
				address = multisigAddress
			}
			// delete wallwt
			if multisigAddress != "" {

				for _, wallet := range user.Wallets {
					if wallet.WalletIndex == walletIndex {
						if len(wallet.Adresses) > 0 {
							address = wallet.Adresses[0].Address
						}
					}
				}
			}

			balance := &ethpb.Balance{}
			if networkid == currencies.ETHMain {
				balance, err = restClient.ETH.CliMain.EventGetAdressBalance(context.Background(), &ethpb.AddressToResync{
					Address: address,
				})
			}
			if networkid == currencies.ETHTest {
				restClient.ETH.CliTest.EventGetAdressBalance(context.Background(), &ethpb.AddressToResync{
					Address: address,
				})
			}

			if balance.Balance == "0" || balance.Balance == "" {
				err := restClient.userStore.DeleteWallet(user.UserID, address, walletIndex, currencyId, networkid)
				if err != nil {
					restClient.log.Errorf("deleteWallet: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
					c.JSON(http.StatusBadRequest, gin.H{
						"code":    http.StatusInternalServerError,
						"message": msgErrNoWallet,
					})
					return
				}
			}

			if balance.Balance != "0" && balance.Balance != "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": msgErrWalletNonZeroBalance,
				})
				return
			}

			code = http.StatusOK
			message = http.StatusText(http.StatusOK)
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
			})
			return
		}

		if totalBalance != 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrWalletNonZeroBalance,
			})
			return
		}
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	}
}

func (restClient *RestClient) addAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
			})
			return
		}
		var sw SelectWallet
		err = decodeBody(c, &sw)
		if err != nil {
			restClient.log.Errorf("addAddress: decodeBody: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		}

		err = addAddressToWallet(sw.Address, token, sw.CurrencyID, sw.NetworkID, sw.WalletIndex, sw.AddressIndex, restClient, c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusText(http.StatusBadRequest),
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"code":    http.StatusText(http.StatusCreated),
			"message": "wallet created",
		})
	}
}

func (restClient *RestClient) getFeeRate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var sp EstimationSpeeds
		currencyID, err := strconv.Atoi(c.Param("currencyid"))
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int currency id: %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"speeds":  sp,
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeCurIndexErr,
			})
			return
		}

		networkid, err := strconv.Atoi(c.Param("networkid"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[networkid=%s]", networkid, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int networkid:[%d] %s \t[addr=%s]", networkid, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodenetworkidErr,
			})
			return
		}

		switch currencyID {
		case currencies.Bitcoin:
			// restClient.BTC.M.Lock()
			// mempool := *restClient.BTC.BtcMempool
			// restClient.BTC.M.Unlock()

			type kv struct {
				Key   string
				Value int
			}

			var mp []kv
			restClient.BTC.BtcMempool.Range(func(k, v interface{}) bool {
				mp = append(mp, kv{k.(string), v.(int)})
				return true
			})

			// var mp []kv
			// for k, v := range mempool {
			// 	mp = append(mp, kv{k, v})
			// }

			sort.Slice(mp, func(i, j int) bool {
				return mp[i].Value > mp[j].Value
			})

			// for _, kv := range ss {
			// 	fmt.Printf("%s, %d\n", kv.Key, kv.Value)
			// }

			// var speeds []string{}
			var slowestValue, slowValue, mediumValue, fastValue, fastestValue int

			memPoolSize := len(mp)

			if memPoolSize <= 2000 && memPoolSize > 0 {
				//low rates logic

				fastestPosition := int(memPoolSize / 100 * 5)
				fastPosition := int(memPoolSize / 100 * 30)
				mediumPosition := int(memPoolSize / 100 * 50)
				slowPosition := int(memPoolSize / 100 * 80)
				//slowestPosition := int(memPoolSize)

				slowestValue = 2

				slowValue = mp[slowPosition].Value

				if slowValue < 2 {
					slowValue = 2
				}

				mediumValue = mp[mediumPosition].Value
				fastValue = mp[fastPosition].Value
				fastestValue = mp[fastestPosition].Value

			} else if memPoolSize == 0 {
				slowestValue = 2
				slowValue = 2
				mediumValue = 3
				fastValue = 5
				fastestValue = 10
			} else {
				//high rates logic
				fastestPosition := 100
				fastPosition := 500
				mediumPosition := 2000
				slowPosition := int(memPoolSize / 100 * 70)
				slowestPosition := int(memPoolSize / 100 * 90)

				slowestValue = mp[slowestPosition].Value

				if slowestValue < 2 {
					slowestValue = 2
				}

				slowValue = mp[slowPosition].Value

				if slowValue < 2 {
					slowValue = 2
				}

				mediumValue = mp[mediumPosition].Value
				fastValue = mp[fastPosition].Value
				fastestValue = mp[fastestPosition].Value

			}

			if fastValue > fastestValue {
				fastestValue = fastValue
			}
			if mediumValue > fastValue {
				fastValue = mediumValue
			}
			if slowValue > mediumValue {
				mediumValue = slowValue
			}
			if slowestValue > slowValue {
				slowValue = slowestValue
			}

			sp = EstimationSpeeds{
				VerySlow: slowestValue,
				Slow:     slowValue,
				Medium:   mediumValue,
				Fast:     fastValue,
				VeryFast: fastestValue,
			}

			restClient.log.Debugf("FeeRates for Bitcoin network id %d is: %v :\n memPoolSize is: %v ", networkid, sp, memPoolSize)

			c.JSON(http.StatusOK, gin.H{
				"speeds":  sp,
				"code":    http.StatusOK,
				"message": http.StatusText(http.StatusOK),
			})
		case currencies.Ether:
			//TODO: make eth feerate
			//var rate *ethpb.GasPrice
			var err error
			// switch networkid {
			// case currencies.ETHMain:
			// 	rate, err = restClient.ETH.CliMain.EventGetGasPrice(context.Background(), &ethpb.Empty{})
			// case currencies.ETHTest:
			// 	rate, err = restClient.ETH.CliTest.EventGetGasPrice(context.Background(), &ethpb.Empty{})
			// default:
			// 	restClient.log.Errorf("getFeeRate:currencies.Ether: no such networkid")
			// }

			if err != nil {
				restClient.log.Errorf("getFeeRate:currencies.Ether:restClient.ETH.Cli: %v ", err.Error())
			}
			//speed, _ := strconv.Atoi(rate.GetGas())
			switch networkid {
			case currencies.ETHMain:
				// c.JSON(http.StatusOK, gin.H{
				// 	"speeds": EstimationSpeeds{
				// 		VerySlow: speed / 2,
				// 		Slow:     speed,
				// 		Medium:   speed * 15 / 10,
				// 		Fast:     speed * 2,
				// 		VeryFast: speed * 25 / 10,
				// 	},
				// 	"code":    http.StatusOK,
				// 	"message": http.StatusText(http.StatusOK),
				// })

				c.JSON(http.StatusOK, gin.H{
					"speeds": EstimationSpeeds{
						VerySlow: 9 * 1000000000,
						Slow:     10 * 1000000000,
						Medium:   14 * 1000000000,
						Fast:     20 * 1000000000,
						VeryFast: 25 * 1000000000,
					},
					"code":    http.StatusOK,
					"message": http.StatusText(http.StatusOK),
				})
			case 3:
				c.JSON(http.StatusOK, gin.H{
					"speeds": EstimationSpeeds{
						VerySlow: 1000000000,
						Slow:     2000000000,
						Medium:   3000000000,
						Fast:     4000000000,
						VeryFast: 5000000000,
					},
					"code":    http.StatusOK,
					"message": http.StatusText(http.StatusOK),
				})

			case 4:
				c.JSON(http.StatusOK, gin.H{
					"speeds": EstimationSpeeds{
						VerySlow: 1000000000,
						Slow:     2000000000,
						Medium:   3000000000,
						Fast:     4000000000,
						VeryFast: 5000000000,
					},
					"code":    http.StatusOK,
					"message": http.StatusText(http.StatusOK),
				})
			}

		default:

		}

	}
}

func avg(arr []store.MempoolRecord) int {
	total := 0
	for _, value := range arr {
		total += value.Category
	}
	if total == 0 {
		return 0
	}
	return total / len(arr)
}

func (restClient *RestClient) getSpendableOutputs() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
			})
			return
		}

		currencyID, err := strconv.Atoi(c.Param("currencyid"))
		restClient.log.Errorf("getSpendableOutputs [%d] \t[addr=%s]", currencyID, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getSpendableOutputs: non int currencyID:[%d] %s \t[addr=%s]", currencyID, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeCurIndexErr,
				"outs":    0,
			})
			return
		}

		networkid, err := strconv.Atoi(c.Param("networkid"))
		restClient.log.Debugf("getSpendableOutputs \t[networkid=%s]", c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getSpendableOutputs: non int networkid : %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodenetworkidErr,
			})
			return
		}

		address := c.Param("addr")

		var (
			code    int
			message string
		)

		user := store.User{}
		query := bson.M{"devices.JWT": token}
		if err := restClient.userStore.FindUser(query, &user); err != nil {
			restClient.log.Errorf("getAllWalletsVerbose: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrUserNotFound,
				"outs":    0,
			})
			return
		} else {
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)
		}

		spOuts, err := restClient.userStore.GetAddressSpendableOutputs(address, currencyID, networkid)
		if err != nil {
			restClient.log.Errorf("getSpendableOutputs: GetAddressSpendableOutputs:[%d] %s \t[addr=%s]", currencyID, err.Error(), c.Request.RemoteAddr)
		}

		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
			"outs":    spOuts,
		})
	}
}

type RawHDTx struct {
	CurrencyID int `json:"currencyid"`
	NetworkID  int `json:"networkID"`
	Payload    `json:"payload"`
}

type Payload struct {
	Address             string   `json:"address"`
	AddressIndex        int      `json:"addressindex"`
	WalletIndex         int      `json:"walletindex"`
	Transaction         string   `json:"transaction"`
	IsHD                bool     `json:"ishd"`
	MultisigFactory     bool     `json:"multisigfactory"`
	WalletName          string   `json:"walletname"`
	Owners              []string `json:"owners"`
	ConfirmationsNeeded int      `json:"confirmationsneeded"`
}

func (restClient *RestClient) sendRawHDTransaction() gin.HandlerFunc {
	return func(c *gin.Context) {

		var rawTx RawHDTx
		if err := decodeBody(c, &rawTx); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrRequestBodyError,
			})
		}

		token, err := getToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
			})
			return
		}

		user := store.User{}
		query := bson.M{"devices.JWT": token}
		if err := restClient.userStore.FindUser(query, &user); err != nil {
			restClient.log.Errorf("sendRawHDTransaction: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)

			return
		}
		code := http.StatusOK
		switch rawTx.CurrencyID {
		case currencies.Bitcoin:
			if rawTx.NetworkID == currencies.Main {
				err := NewAddressNode(rawTx.Address, user.UserID, rawTx.CurrencyID, rawTx.NetworkID, rawTx.WalletIndex, rawTx.AddressIndex, restClient)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    http.StatusInternalServerError,
						"message": "err: " + err.Error(),
					})
					return
				}

				resp, err := restClient.BTC.CliMain.EventSendRawTx(context.Background(), &btcpb.RawTx{
					Transaction: rawTx.Transaction,
				})

				if err != nil {
					restClient.log.Errorf("sendRawHDTransaction: restClient.BTC.CliMain.EventSendRawTx: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
					code = http.StatusBadRequest
					c.JSON(code, gin.H{
						"code":    code,
						"message": err.Error(),
					})
					return
				}

				if strings.Contains("err:", resp.GetMessage()) {
					restClient.log.Errorf("sendRawHDTransaction: restClient.BTC.CliMain.EventSendRawTx:resp err %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
					code = http.StatusBadRequest
					c.JSON(code, gin.H{
						"code":    code,
						"message": resp.GetMessage(),
					})
					return
				}

				if rawTx.IsHD {
					err = addAddressToWallet(rawTx.Address, token, rawTx.CurrencyID, rawTx.NetworkID, rawTx.WalletIndex, rawTx.AddressIndex, restClient, c)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{
							"code":    http.StatusBadRequest,
							"message": err.Error(),
						})
						return
					}
				}

				c.JSON(code, gin.H{
					"code":    code,
					"message": resp.Message,
				})
				return

			}
			if rawTx.NetworkID == currencies.Test {

				err := NewAddressNode(rawTx.Address, user.UserID, rawTx.CurrencyID, rawTx.NetworkID, rawTx.WalletIndex, rawTx.AddressIndex, restClient)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    http.StatusInternalServerError,
						"message": "err: " + err.Error(),
					})
					return
				}

				resp, err := restClient.BTC.CliTest.EventSendRawTx(context.Background(), &btcpb.RawTx{
					Transaction: rawTx.Transaction,
				})
				if err != nil {
					restClient.log.Errorf("sendRawHDTransaction: restClient.BTC.CliMain.EventSendRawTx: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
					code = http.StatusBadRequest
					c.JSON(code, gin.H{
						"code":    code,
						"message": err.Error(),
					})
					return
				}

				if strings.Contains("err:", resp.GetMessage()) {
					restClient.log.Errorf("sendRawHDTransaction: restClient.BTC.CliMain.EventSendRawTx:resp err %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
					code = http.StatusBadRequest
					c.JSON(code, gin.H{
						"code":    code,
						"message": resp.GetMessage(),
					})
					return
				}

				if rawTx.IsHD {
					err = addAddressToWallet(rawTx.Address, token, rawTx.CurrencyID, rawTx.NetworkID, rawTx.WalletIndex, rawTx.AddressIndex, restClient, c)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{
							"code":    http.StatusBadRequest,
							"message": err.Error(),
						})
						return
					}
				}
				flag := 4
				for {
					ex := restClient.userStore.CheckTx(resp.GetMessage())
					if ex {
						break
					}
					flag++
					if flag == 4 {
						break
					}
				}

				c.JSON(code, gin.H{
					"code":    code,
					"message": resp.Message,
				})
				return
			}
		case currencies.Ether:
			if rawTx.NetworkID == currencies.ETHMain {
				hash, err := restClient.ETH.CliMain.EventSendRawTx(context.Background(), &ethpb.RawTx{
					Transaction: rawTx.Transaction,
				})
				if err != nil {
					restClient.log.Errorf("sendRawHDTransaction:eth.SendRawTransaction %s", err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    http.StatusInternalServerError,
						"message": err.Error(),
					})
					return
				}
				// TODO: Make a wallet

				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": hash,
				})
				return
			}
			if rawTx.NetworkID == currencies.ETHTest {
				hash, err := restClient.ETH.CliTest.EventSendRawTx(context.Background(), &ethpb.RawTx{
					Transaction: rawTx.Transaction,
				})
				if err != nil {
					restClient.log.Errorf("sendRawHDTransaction:eth.SendRawTransaction %s", err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    http.StatusInternalServerError,
						"message": err.Error(),
					})
					return
				}

				// TODO: Make a wallet

				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": hash,
				})
				return
			}

		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
			})
		}
	}
}

func errHandler(resp string) bool {
	return strings.Contains(resp, "err:")
}

type RawTx struct { // remane RawClientTransaction
	Transaction string `json:"transaction"` //HexTransaction
}

func (restClient *RestClient) getWalletVerbose() gin.HandlerFunc {
	return func(c *gin.Context) {
		var wv []interface{}
		//var wv []WalletVerbose
		token, err := getToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
			})
			return
		}

		var multisigAddress string
		walletIndex, err := strconv.Atoi(c.Param("walletindex"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[walletindexr=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			multisigAddress = c.Param("walletindex")
		}

		currencyId, err := strconv.Atoi(c.Param("currencyid"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[currencyId=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int currency id:[%d] %s \t[addr=%s]", currencyId, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeCurIndexErr,
			})
			return
		}

		networkId, err := strconv.Atoi(c.Param("networkid"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[networkID=%s]", networkId, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int networkid:[%d] %s \t[addr=%s]", networkId, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeNetworkIDErr,
				"wallet":  wv,
			})
			return
		}

		var (
			code    int
			message string
		)

		user := store.User{}

		switch currencyId {
		case currencies.Bitcoin:
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)
			var av []AddressVerbose

			query := bson.M{"devices.JWT": token}

			if err := restClient.userStore.FindUser(query, &user); err != nil {
				restClient.log.Errorf("getAllWalletsVerbose: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			}

			// fetch wallet with concrete networkid currencyid and wallet index
			wallet := store.Wallet{}
			for _, w := range user.Wallets {
				if w.NetworkID == networkId && w.CurrencyID == currencyId && w.WalletIndex == walletIndex {
					wallet = w
					break
				}
			}

			if len(wallet.Adresses) == 0 {
				restClient.log.Errorf("getAllWalletsVerbose: restClient.userStore.FindUser:\t[addr=%s]", c.Request.RemoteAddr)
				c.JSON(code, gin.H{
					"code":    http.StatusBadRequest,
					"message": msgErrUserNotFound,
					"wallet":  wv,
				})
				return
			}
			var pending bool
			for _, address := range wallet.Adresses {
				spOuts := getBTCAddressSpendableOutputs(address.Address, currencyId, networkId, restClient)
				for _, spOut := range spOuts {
					if spOut.TxStatus == store.TxStatusAppearedInMempoolIncoming {
						pending = true
					}
				}
				// TODO:
				_, sync := restClient.BTC.Resync.Load(address.Address)

				av = append(av, AddressVerbose{
					LastActionTime: address.LastActionTime,
					Address:        address.Address,
					AddressIndex:   address.AddressIndex,
					Amount:         int64(checkBTCAddressbalance(address.Address, currencyId, networkId, restClient)),
					SpendableOuts:  spOuts,
					IsSyncing:      sync,
				})
			}
			wv = append(wv, WalletVerbose{
				WalletIndex:    wallet.WalletIndex,
				CurrencyID:     wallet.CurrencyID,
				NetworkID:      wallet.NetworkID,
				WalletName:     wallet.WalletName,
				LastActionTime: wallet.LastActionTime,
				DateOfCreation: wallet.DateOfCreation,
				VerboseAddress: av,
				Pending:        pending,
			})
			av = []AddressVerbose{}

		case currencies.Ether:
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)

			var av []ETHAddressVerbose

			query := bson.M{"devices.JWT": token}
			if err := restClient.userStore.FindUser(query, &user); err != nil {
				restClient.log.Errorf("getAllWalletsVerbose: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			}

			// fetch wallet with concrete networkid currencyid and wallet index
			var pending bool
			var totalBalance string
			var pendingBalance string
			var waletNonce int64
			wallet := store.Wallet{}
			multisig := store.Multisig{}

			// multisig verbose
			if multisigAddress != "" {
				for _, m := range user.Multisigs {
					if m.NetworkID == networkId && m.CurrencyID == currencyId && m.ContractAddress == multisigAddress {
						multisig = m
						break
					}
				}
				if len(multisig.ContractAddress) == 0 {
					restClient.log.Errorf("getAllWalletsVerbose: len(multisig.ContractAddress) == 0:\t[addr=%s]", c.Request.RemoteAddr)
					c.JSON(code, gin.H{
						"code":    http.StatusBadRequest,
						"message": msgErrNoWallet,
						"wallet":  wv,
					})
					return
				}

				amount := &ethpb.Balance{}
				nonce := &ethpb.Nonce{}

				var err error
				adr := ethpb.AddressToResync{
					Address: multisig.ContractAddress,
				}

				switch networkId {
				case currencies.ETHTest:
					nonce, err = restClient.ETH.CliTest.EventGetAdressNonce(context.Background(), &adr)
					amount, err = restClient.ETH.CliTest.EventGetAdressBalance(context.Background(), &adr)
				case currencies.ETHMain:
					nonce, err = restClient.ETH.CliMain.EventGetAdressNonce(context.Background(), &adr)
					amount, err = restClient.ETH.CliMain.EventGetAdressBalance(context.Background(), &adr)
				default:
					c.JSON(code, gin.H{
						"code":    http.StatusBadRequest,
						"message": msgErrMethodNotImplennted,
						"wallet":  wv,
					})
					return
				}

				if err != nil {
					restClient.log.Errorf("EventGetAdressNonce || EventGetAdressBalance: %v", err.Error())
				}

				totalBalance = amount.GetBalance()
				pendingBalance = amount.GetPendingBalance()

				p, _ := strconv.Atoi(amount.GetPendingBalance())
				b, _ := strconv.Atoi(amount.GetBalance())

				if p != b {
					pending = true
					multisig.LastActionTime = time.Now().Unix()
				}

				if p == b {
					pendingBalance = "0"
				}

				waletNonce = nonce.GetNonce()

				av = append(av, ETHAddressVerbose{
					LastActionTime: multisig.LastActionTime,
					Address:        multisig.ContractAddress,
					Amount:         totalBalance,
					Nonce:          nonce.Nonce,
				})
				wv = append(wv, WalletVerboseETH{
					CurrencyID:     multisig.CurrencyID,
					NetworkID:      multisig.NetworkID,
					WalletName:     multisig.WalletName,
					LastActionTime: multisig.LastActionTime,
					DateOfCreation: multisig.DateOfCreation,
					Nonce:          waletNonce,
					Balance:        totalBalance,
					PendingBalance: pendingBalance,
					VerboseAddress: av,
					Pending:        pending,
					Multisig: MultisigVerbose{
						Owners:         multisig.Owners,
						Confirmations:  multisig.Confirmations,
						IsDeployed:     multisig.DeployStatus,
						FactoryAddress: multisig.FactoryAddress,
						TxOfCreation:   multisig.TxOfCreation,
					},
				})

			}

			// wallet verbose
			if multisigAddress == "" {
				for _, w := range user.Wallets {
					if w.NetworkID == networkId && w.CurrencyID == currencyId && w.WalletIndex == walletIndex {
						wallet = w
						break
					}
				}

				if len(wallet.Adresses) == 0 {
					restClient.log.Errorf("getAllWalletsVerbose: len(wallet.Adresses) == 0:\t[addr=%s]", c.Request.RemoteAddr)
					c.JSON(code, gin.H{
						"code":    http.StatusBadRequest,
						"message": msgErrNoWallet,
						"wallet":  wv,
					})
					return
				}

				for _, address := range wallet.Adresses {
					amount := &ethpb.Balance{}
					nonce := &ethpb.Nonce{}

					var err error
					adr := ethpb.AddressToResync{
						Address: address.Address,
					}

					switch networkId {
					case currencies.ETHTest:
						nonce, err = restClient.ETH.CliTest.EventGetAdressNonce(context.Background(), &adr)
						amount, err = restClient.ETH.CliTest.EventGetAdressBalance(context.Background(), &adr)
					case currencies.ETHMain:
						nonce, err = restClient.ETH.CliMain.EventGetAdressNonce(context.Background(), &adr)
						amount, err = restClient.ETH.CliMain.EventGetAdressBalance(context.Background(), &adr)
					default:
						c.JSON(code, gin.H{
							"code":    http.StatusBadRequest,
							"message": msgErrMethodNotImplennted,
							"wallet":  wv,
						})
						return
					}

					if err != nil {
						restClient.log.Errorf("EventGetAdressNonce || EventGetAdressBalance: %v", err.Error())
					}

					totalBalance = amount.GetBalance()
					pendingBalance = amount.GetPendingBalance()

					p, _ := strconv.Atoi(amount.GetPendingBalance())
					b, _ := strconv.Atoi(amount.GetBalance())
					// pendingBalance = strconv.Itoa(p - b)
					// pendingAmount = strconv.Itoa(p - b)

					if p != b {
						pending = true
						address.LastActionTime = time.Now().Unix()
					}

					if p == b {
						pendingBalance = "0"
					}

					waletNonce = nonce.GetNonce()

					av = append(av, ETHAddressVerbose{
						LastActionTime: address.LastActionTime,
						Address:        address.Address,
						AddressIndex:   address.AddressIndex,
						Amount:         totalBalance,
						Nonce:          nonce.Nonce,
					})

				}

				wv = append(wv, WalletVerboseETH{
					WalletIndex:    wallet.WalletIndex,
					CurrencyID:     wallet.CurrencyID,
					NetworkID:      wallet.NetworkID,
					WalletName:     wallet.WalletName,
					LastActionTime: wallet.LastActionTime,
					DateOfCreation: wallet.DateOfCreation,
					Nonce:          waletNonce,
					Balance:        totalBalance,
					PendingBalance: pendingBalance,
					VerboseAddress: av,
					Pending:        pending,
				})
				av = []ETHAddressVerbose{}

			}

		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
			})
			return
		}

		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
			"wallet":  wv,
		})
	}
}

type WalletVerbose struct {
	CurrencyID     int              `json:"currencyid"`
	NetworkID      int              `json:"networkid"`
	WalletIndex    int              `json:"walletindex"`
	WalletName     string           `json:"walletname"`
	LastActionTime int64            `json:"lastactiontime"`
	DateOfCreation int64            `json:"dateofcreation"`
	VerboseAddress []AddressVerbose `json:"addresses"`
	Pending        bool             `json:"pending"`
}

type WalletVerboseETH struct {
	CurrencyID     int                 `json:"currencyid"`
	NetworkID      int                 `json:"networkid"`
	WalletIndex    int                 `json:"walletindex"`
	WalletName     string              `json:"walletname"`
	LastActionTime int64               `json:"lastactiontime"`
	DateOfCreation int64               `json:"dateofcreation"`
	Nonce          int64               `json:"nonce"`
	PendingBalance string              `json:"pendingbalance"`
	Balance        string              `json:"balance"`
	VerboseAddress []ETHAddressVerbose `json:"addresses"`
	Pending        bool                `json:"pending"`
	Multisig       MultisigVerbose     `json:"multisig,omitempty"`
}

type AddressVerbose struct {
	LastActionTime int64                    `json:"lastactiontime"`
	Address        string                   `json:"address"`
	AddressIndex   int                      `json:"addressindex"`
	Amount         int64                    `json:"amount"`
	SpendableOuts  []store.SpendableOutputs `json:"spendableoutputs,omitempty"`
	Nonce          int64                    `json:"nonce,omitempty"`
	IsSyncing      bool                     `json:"issyncing"`
}

type ETHAddressVerbose struct {
	LastActionTime int64  `json:"lastactiontime"`
	Address        string `json:"address"`
	AddressIndex   int    `json:"addressindex"`
	Amount         string `json:"amount"`
	Nonce          int64  `json:"nonce,omitempty"`
}

type MultisigVerbose struct {
	Owners         []store.AddressExtended `json:"owners"`
	Confirmations  int                     `json:"Confirmations"`
	IsDeployed     bool                    `json:"isdeployed"`
	FactoryAddress string                  `json:"factoryaddress"`
	TxOfCreation   string                  `json:"txofcreation"`
}

type StockExchangeRate struct {
	ExchangeName   string `json:"exchangename"`
	FiatEquivalent int    `json:"fiatequivalent"`
	TotalAmount    int    `json:"totalamount"`
}

type TopIndex struct {
	CurrencyID int `json:"currencyid"`
	NetworkID  int `json:"networkid"`
	TopIndex   int `json:"topindex"`
}

func findTopIndexes(walletsBTC []store.Wallet) []TopIndex {
	top := map[TopIndex]int{} // currency id -> topindex
	topIndex := []TopIndex{}
	for _, wallet := range walletsBTC {
		top[TopIndex{wallet.CurrencyID, wallet.NetworkID, 0}]++
	}
	for value, maxIndex := range top {
		topIndex = append(topIndex, TopIndex{
			CurrencyID: value.CurrencyID,
			NetworkID:  value.NetworkID,
			TopIndex:   maxIndex,
		})
	}
	return topIndex
}

func fetchUndeletedWallets(walletsBTC []store.Wallet) []store.Wallet {
	//func fetchUndeletedWallets(wallets []store.Wallet) []store.Wallet {
	okWalletsBTC := []store.Wallet{}

	for _, wallet := range walletsBTC {
		if wallet.Status == store.WalletStatusOK {
			okWalletsBTC = append(okWalletsBTC, wallet)
		}
	}

	return okWalletsBTC
}

func (restClient *RestClient) getAllWalletsVerbose() gin.HandlerFunc {
	return func(c *gin.Context) {
		var wv []interface{}
		//var wv []WalletVerbose
		token, err := getToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
			})
			return
		}
		var (
			code    int
			message string
		)
		user := store.User{}
		query := bson.M{"devices.JWT": token}

		if err := restClient.userStore.FindUser(query, &user); err != nil {
			restClient.log.Errorf("getAllWalletsVerbose: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(code, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrUserNotFound,
				"wallets": wv,
			})
			return
		}

		topIndexes := findTopIndexes(user.Wallets)

		code = http.StatusOK
		message = http.StatusText(http.StatusOK)

		okWallets := fetchUndeletedWallets(user.Wallets)

		userTxs := []store.MultyTX{}

		for _, wallet := range okWallets {
			switch wallet.CurrencyID {
			case currencies.Bitcoin:
				var av []AddressVerbose
				var pending bool
				for _, address := range wallet.Adresses {
					spOuts := getBTCAddressSpendableOutputs(address.Address, wallet.CurrencyID, wallet.NetworkID, restClient)

					//all user txs
					err = restClient.userStore.GetAllWalletTransactions(user.UserID, wallet.CurrencyID, wallet.NetworkID, &userTxs)
					if err != nil {
						//empty history
					}

					for _, tx := range userTxs {
						if len(tx.TxAddress) > 0 {
							if tx.TxAddress[0] == address.Address {
								if tx.TxStatus == store.TxStatusAppearedInMempoolIncoming || tx.TxStatus == store.TxStatusAppearedInMempoolOutcoming {
									pending = true
								}
							}
						}
					}

					// restClient.BTC.ResyncM.Lock()
					// re := *restClient.BTC.Resync
					// restClient.BTC.ResyncM.Unlock()
					_, sync := restClient.BTC.Resync.Load(address.Address)

					// sync := re[address.Address]

					av = append(av, AddressVerbose{
						LastActionTime: address.LastActionTime,
						Address:        address.Address,
						AddressIndex:   address.AddressIndex,
						Amount:         int64(checkBTCAddressbalance(address.Address, wallet.CurrencyID, wallet.NetworkID, restClient)),
						SpendableOuts:  spOuts,
						IsSyncing:      sync,
					})

				}

				wv = append(wv, WalletVerbose{
					WalletIndex:    wallet.WalletIndex,
					CurrencyID:     wallet.CurrencyID,
					NetworkID:      wallet.NetworkID,
					WalletName:     wallet.WalletName,
					LastActionTime: wallet.LastActionTime,
					DateOfCreation: wallet.DateOfCreation,
					VerboseAddress: av,
					Pending:        pending,
				})
				av = []AddressVerbose{}
				userTxs = []store.MultyTX{}
			case currencies.Ether:
				var av []ETHAddressVerbose
				var pending bool
				var walletNonce int64

				var totalBalance string
				var pendingBalance string
				for _, address := range wallet.Adresses {
					amount := &ethpb.Balance{}
					nonce := &ethpb.Nonce{}

					var err error
					adr := ethpb.AddressToResync{
						Address: address.Address,
					}

					switch wallet.NetworkID {
					case currencies.ETHTest:
						nonce, err = restClient.ETH.CliTest.EventGetAdressNonce(context.Background(), &adr)
						amount, err = restClient.ETH.CliTest.EventGetAdressBalance(context.Background(), &adr)
					case currencies.ETHMain:
						nonce, err = restClient.ETH.CliMain.EventGetAdressNonce(context.Background(), &adr)
						amount, err = restClient.ETH.CliMain.EventGetAdressBalance(context.Background(), &adr)
					default:
						c.JSON(code, gin.H{
							"code":    http.StatusBadRequest,
							"message": msgErrMethodNotImplennted,
							"wallet":  wv,
						})
						return
					}

					if err != nil {
						restClient.log.Errorf("EventGetAdressNonce || EventGetAdressBalance: %v", err.Error())
					}

					totalBalance = amount.GetBalance()
					pendingBalance = amount.GetPendingBalance()

					p, _ := strconv.Atoi(amount.GetPendingBalance())
					b, _ := strconv.Atoi(amount.GetBalance())

					if p != b {
						pending = true
					}

					if p == b {
						pendingBalance = "0"
					}
					walletNonce = nonce.GetNonce()

					av = append(av, ETHAddressVerbose{
						LastActionTime: address.LastActionTime,
						Address:        address.Address,
						AddressIndex:   address.AddressIndex,
						Amount:         totalBalance,
						Nonce:          nonce.Nonce,
					})

				}

				wv = append(wv, WalletVerboseETH{
					WalletIndex:    wallet.WalletIndex,
					CurrencyID:     wallet.CurrencyID,
					NetworkID:      wallet.NetworkID,
					Balance:        totalBalance,
					PendingBalance: pendingBalance,
					Nonce:          walletNonce,
					WalletName:     wallet.WalletName,
					LastActionTime: wallet.LastActionTime,
					DateOfCreation: wallet.DateOfCreation,
					VerboseAddress: av,
					Pending:        pending,
				})
				av = []ETHAddressVerbose{}
			default:

			}

		}
		for _, multisig := range user.Multisigs {
			var av []ETHAddressVerbose
			var pending bool

			var totalBalance string
			var pendingBalance string

			amount := &ethpb.Balance{}
			nonce := &ethpb.Nonce{}

			adr := ethpb.AddressToResync{
				Address: multisig.ContractAddress,
			}

			switch multisig.NetworkID {
			case currencies.ETHTest:
				nonce, err = restClient.ETH.CliTest.EventGetAdressNonce(context.Background(), &adr)
				amount, err = restClient.ETH.CliTest.EventGetAdressBalance(context.Background(), &adr)
			case currencies.ETHMain:
				nonce, err = restClient.ETH.CliMain.EventGetAdressNonce(context.Background(), &adr)
				amount, err = restClient.ETH.CliMain.EventGetAdressBalance(context.Background(), &adr)
			default:
				c.JSON(code, gin.H{
					"code":    http.StatusBadRequest,
					"message": msgErrMethodNotImplennted,
					"wallet":  wv,
				})
				return
			}

			totalBalance = amount.GetBalance()
			pendingBalance = amount.GetPendingBalance()

			p, _ := strconv.Atoi(amount.GetPendingBalance())
			b, _ := strconv.Atoi(amount.GetBalance())

			if p != b {
				pending = true
				multisig.LastActionTime = time.Now().Unix()
			}

			if p == b {
				pendingBalance = "0"
			}

			waletNonce := nonce.GetNonce()

			av = append(av, ETHAddressVerbose{
				LastActionTime: multisig.LastActionTime,
				Address:        multisig.ContractAddress,
				Amount:         totalBalance,
				Nonce:          nonce.Nonce,
			})

			wv = append(wv, WalletVerboseETH{
				CurrencyID:     multisig.CurrencyID,
				NetworkID:      multisig.NetworkID,
				WalletName:     multisig.WalletName,
				LastActionTime: multisig.LastActionTime,
				DateOfCreation: multisig.DateOfCreation,
				Nonce:          waletNonce,
				Balance:        totalBalance,
				PendingBalance: pendingBalance,
				VerboseAddress: av,
				Pending:        pending,
				Multisig: MultisigVerbose{
					Owners:         multisig.Owners,
					Confirmations:  multisig.Confirmations,
					IsDeployed:     multisig.DeployStatus,
					FactoryAddress: multisig.FactoryAddress,
					TxOfCreation:   multisig.TxOfCreation,
				},
			})

		}

		c.JSON(code, gin.H{
			"code":       code,
			"message":    message,
			"wallets":    wv,
			"topindexes": topIndexes,
		})

	}
}

func (restClient *RestClient) getWalletTransactionsHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		var walletTxs []store.MultyTX
		token, err := getToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
			})
			return
		}

		var multisig string
		walletIndex, err := strconv.Atoi(c.Param("walletindex"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[walletindexr=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			multisig = c.Param("walletindex")
		}

		currencyId, err := strconv.Atoi(c.Param("currencyid"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[currencyId=%s]", currencyId, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int currency id:[%d] %s \t[addr=%s]", currencyId, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeCurIndexErr,
			})
			return
		}

		networkid, err := strconv.Atoi(c.Param("networkid"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[networkid=%s]", networkid, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int networkid index:[%d] %s \t[addr=%s]", networkid, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodenetworkidErr,
			})
			return
		}

		user := store.User{}
		sel := bson.M{"devices.JWT": token}
		err = restClient.userStore.FindUser(sel, &user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrUserNotFound,
				"history": walletTxs,
			})
			return
		}

		switch currencyId {
		case currencies.Bitcoin:

			var blockHeight int64
			switch networkid {
			case currencies.Test:
				resp, err := restClient.BTC.CliTest.EventGetBlockHeight(context.Background(), &btcpb.Empty{})
				if err != nil {
					restClient.log.Errorf("getWalletTransactionsHistory: restClient.BTC.CliTest.EventGetBlockHeight %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    http.StatusInternalServerError,
						"message": http.StatusText(http.StatusInternalServerError),
					})
					return
				}
				blockHeight = resp.Height
			case currencies.Main:
				resp, err := restClient.BTC.CliMain.EventGetBlockHeight(context.Background(), &btcpb.Empty{})
				if err != nil {
					restClient.log.Errorf("getWalletTransactionsHistory: restClient.BTC.CliMain.EventGetBlockHeight %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    http.StatusInternalServerError,
						"message": http.StatusText(http.StatusInternalServerError),
					})
					return
				}
				blockHeight = resp.Height
			}

			userTxs := []store.MultyTX{}
			err = restClient.userStore.GetAllWalletTransactions(user.UserID, currencyId, networkid, &userTxs)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": msgErrTxHistory,
					"history": walletTxs,
				})
				return
			}

			walletAddresses := []string{}
			for _, wallet := range user.Wallets {
				if wallet.WalletIndex == walletIndex {
					for _, addresses := range wallet.Adresses {
						walletAddresses = append(walletAddresses, addresses.Address)
					}
				}
			}

			// TODO: fix this logic same wallet to samewallet tx

			for _, address := range walletAddresses {
				for _, tx := range userTxs {
					if len(tx.TxAddress) > 0 {
						if tx.TxAddress[0] == address {
							var isTheSameWallet = false
							for _, input := range tx.WalletsInput {
								if walletIndex == input.WalletIndex {
									isTheSameWallet = true
								}
							}

							for _, output := range tx.WalletsOutput {
								if walletIndex == output.WalletIndex {
									isTheSameWallet = true
								}
							}

							if isTheSameWallet {
								walletTxs = append(walletTxs, tx)
							}
						}
					}
				}
			}
			for i := 0; i < len(walletTxs); i++ {
				if walletTxs[i].BlockHeight == -1 {
					walletTxs[i].Confirmations = 0
				} else {
					walletTxs[i].Confirmations = int(blockHeight-walletTxs[i].BlockHeight) + 1
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"message": http.StatusText(http.StatusOK),
				"history": walletTxs,
			})
			return

		case currencies.Ether:
			var blockHeight int64

			switch networkid {
			case currencies.ETHTest:
				resp, err := restClient.ETH.CliTest.EventGetBlockHeight(context.Background(), &ethpb.Empty{})
				if err != nil {
					restClient.log.Errorf("getWalletTransactionsHistory: restClient.BTC.CliTest.EventGetBlockHeight %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    http.StatusInternalServerError,
						"message": http.StatusText(http.StatusInternalServerError),
					})
					return
				}
				blockHeight = resp.Height
			case currencies.ETHMain:
				resp, err := restClient.ETH.CliMain.EventGetBlockHeight(context.Background(), &ethpb.Empty{})
				if err != nil {
					restClient.log.Errorf("getWalletTransactionsHistory: restClient.BTC.CliTest.EventGetBlockHeight %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    http.StatusInternalServerError,
						"message": http.StatusText(http.StatusInternalServerError),
					})
					return
				}
				blockHeight = resp.Height
			}

			//history for ether wallet
			userTxs := []store.TransactionETH{}
			if multisig == "" {
				err = restClient.userStore.GetAllWalletEthTransactions(user.UserID, currencyId, networkid, &userTxs)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"code":    http.StatusBadRequest,
						"message": msgErrTxHistory,
						"history": walletTxs,
					})
					return
				}

				for i := 0; i < len(userTxs); i++ {
					if userTxs[i].BlockHeight == -1 {
						userTxs[i].Confirmations = 0
					} else {
						userTxs[i].Confirmations = int(blockHeight-userTxs[i].BlockHeight) + 1
					}
				}

				history := []store.TransactionETH{}
				for _, tx := range userTxs {
					if tx.WalletIndex == walletIndex {
						history = append(history, tx)
					}
				}

				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": http.StatusText(http.StatusOK),
					"history": history,
				})
				return
			}

			//history for ether multisig
			if multisig != "" {
				err = restClient.userStore.GetAllMultisigEthTransactions(multisig, currencyId, networkid, &userTxs)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"code":    http.StatusBadRequest,
						"message": msgErrTxHistory,
						"history": walletTxs,
					})
					return
				}

				for i := 0; i < len(userTxs); i++ {
					if userTxs[i].BlockHeight == -1 {
						userTxs[i].Confirmations = 0
					} else {
						userTxs[i].Confirmations = int(blockHeight-userTxs[i].BlockHeight) + 1
					}
				}

				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": http.StatusText(http.StatusOK),
					"history": userTxs,
				})
				return

			}

		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
				"history": walletTxs,
			})
			return
		}

	}
}

type TxHistory struct {
	TxID        string               `json:"txid"`
	TxHash      string               `json:"txhash"`
	TxOutScript string               `json:"txoutscript"`
	TxAddress   string               `json:"address"`
	TxStatus    int                  `json:"txstatus"`
	TxOutAmount int64                `json:"txoutamount"`
	TxOutID     int                  `json:"txoutid"`
	WalletIndex int                  `json:"walletindex"`
	BlockTime   int64                `json:"blocktime"`
	BlockHeight int64                `json:"blockheight"`
	TxFee       int64                `json:"txfee"`
	MempoolTime int64                `json:"mempooltime"`
	BtcToUsd    float64              `json:"btctousd"`
	TxInputs    []store.AddresAmount `json:"txinputs"`
	TxOutputs   []store.AddresAmount `json:"txoutputs"`
}

func (restClient *RestClient) resyncWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
			})
			return
		}

		walletIndex, err := strconv.Atoi(c.Param("walletindex"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[walletindexr=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("resyncWallet: non int wallet index:[%d] %s \t[addr=%s]", walletIndex, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeWalletIndexErr,
			})
			return
		}

		currencyID, err := strconv.Atoi(c.Param("currencyid"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[currencyId=%s]", currencyID, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("resyncWallet: non int currency id:[%d] %s \t[addr=%s]", currencyID, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeCurIndexErr,
			})
			return
		}

		networkID, err := strconv.Atoi(c.Param("networkid"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[networkid=%s]", networkID, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("resyncWallet: non int networkid index:[%d] %s \t[addr=%s]", networkID, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodenetworkidErr,
			})
			return
		}

		user := store.User{}
		sel := bson.M{"devices.JWT": token}
		err = restClient.userStore.FindUser(sel, &user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrUserNotFound,
			})
			return
		}

		walletToResync := store.Wallet{}
		for _, wallet := range user.Wallets {
			if wallet.CurrencyID == currencyID && wallet.NetworkID == networkID && wallet.WalletIndex == walletIndex {
				walletToResync = wallet
			}
		}

		if len(walletToResync.Adresses) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrUserHaveNoTxs,
			})
			return
		}

		switch currencyID {
		case currencies.Bitcoin:
			if networkID == currencies.Main {
				for _, address := range walletToResync.Adresses {
					restClient.BTC.CliMain.EventResyncAddress(context.Background(), &btcpb.AddressToResync{
						Address:      address.Address,
						UserID:       user.UserID,
						WalletIndex:  int32(walletIndex),
						AddressIndex: int32(address.AddressIndex),
					})
					err := restClient.userStore.DeleteHistory(currencyID, networkID, address.Address)
					if err != nil {
						restClient.log.Errorf("resyncWallet case currencies.Bitcoin:Main: %v", err.Error())
					}
				}
			}

			if networkID == currencies.Test {
				for _, address := range walletToResync.Adresses {
					restClient.BTC.CliTest.EventResyncAddress(context.Background(), &btcpb.AddressToResync{
						Address:      address.Address,
						UserID:       user.UserID,
						WalletIndex:  int32(walletIndex),
						AddressIndex: int32(address.AddressIndex),
					})
					err := restClient.userStore.DeleteHistory(currencyID, networkID, address.Address)
					if err != nil {
						restClient.log.Errorf("resyncWallet case currencies.Bitcoin:Test: %v", err.Error())
					}
				}

			}
		case currencies.Ether:
			if networkID == currencies.ETHMain {

			}

			if networkID == currencies.ETHTest {

			}
		}

	}
}

func (restClient *RestClient) changellyListCurrencies() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiUrl := "https://api.changelly.com"
		apiKey := "8015e09ba78243ad889db470ec48fed4"
		apiSecret := "712bfcf899dd235b0af1d66922d5962e8c85a909635f838688a38b5f12c4d03a"
		cr := ChangellyReqest{
			JsonRpc: "2.0",
			ID:      1,
			Method:  "getCurrencies",
			Params:  []string{},
		}
		bs, err := json.Marshal(cr)
		if err != nil {
			restClient.log.Errorf("changellyListCurrencies: json.Marshal: %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			//
			return
		}

		sign := ComputeHmac512(bs, apiSecret)
		req, err := http.NewRequest("GET", apiUrl, nil)
		if err != nil {
			restClient.log.Errorf("changellyListCurrencies: http.NewRequest: %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			//
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api-key", apiKey)
		req.Header.Set("sign", sign)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			restClient.log.Errorf("changellyListCurrencies: http.Client.Do: %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			//
			return
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		c.JSON(http.StatusOK, gin.H{
			"code":    resp.StatusCode,
			"message": string(body),
		})
	}
}

func ComputeHmac512(message []byte, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha512.New, key)
	h.Write(message)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

type ChangellyReqest struct {
	JsonRpc string   `json:"jsonrpc"`
	ID      int      `json:"id"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
}
