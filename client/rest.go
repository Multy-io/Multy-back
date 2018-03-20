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
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Appscrunch/Multy-back/btc"
	"github.com/Appscrunch/Multy-back/currencies"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"

	pb "github.com/Appscrunch/Multy-back/node-streamer"

	"math/rand"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
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

	BTC *btc.BTCConn
}

type BTCApiConf struct {
	Token, Coin, Chain string
}

func SetRestHandlers(
	userDB store.UserStore,
	r *gin.Engine,
	donationAddresses []store.DonationInfo,
	btc *btc.BTCConn,
) (*RestClient, error) {
	restClient := &RestClient{
		userStore:         userDB,
		log:               slf.WithContext("rest-client"),
		donationAddresses: donationAddresses,
		BTC:               btc,
	}
	initMiddlewareJWT(restClient)

	r.POST("/auth", restClient.LoginHandler())
	r.GET("/server/config", restClient.getServerConfig())

	r.GET("/statuscheck", restClient.statusCheck())

	v1 := r.Group("/api/v1")
	v1.Use(restClient.middlewareJWT.MiddlewareFunc())
	{
		v1.POST("/wallet", restClient.addWallet())
		v1.DELETE("/wallet/:currencyid/:networkid/:walletindex", restClient.deleteWallet())         // add networkid
		v1.POST("/address", restClient.addAddress())                                                // add networkid
		v1.GET("/transaction/feerate/:currencyid/:networkid", restClient.getFeeRate())              // add networkid
		v1.GET("/outputs/spendable/:currencyid/:networkid/:addr", restClient.getSpendableOutputs()) // add networkid
		// v1.POST("/transaction/send/:currencyid", restClient.sendRawTransaction(btcNodeAddress)) 			// depricated
		v1.POST("/transaction/send", restClient.sendRawHDTransaction())                                                // add networkid
		v1.GET("/wallet/:walletindex/verbose/:currencyid/:networkid", restClient.getWalletVerbose())                   // add networkid
		v1.GET("/wallets/verbose", restClient.getAllWalletsVerbose())                                                  // add networkid
		v1.GET("/wallets/transactions/:currencyid/:networkid/:walletindex", restClient.getWalletTransactionsHistory()) // add networkid
		v1.POST("/wallet/name", restClient.changeWalletName())                                                         // add networkid
		v1.GET("/exchange/changelly/list", restClient.changellyListCurrencies())
		v1.GET("/drop", restClient.drop())
	}
	return restClient, nil
}

func initMiddlewareJWT(restClient *RestClient) {
	restClient.middlewareJWT = &GinJWTMiddleware{
		Realm:      "test zone",
		Key:        []byte("secret key"), // config
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

	for _, walletBTC := range user.Wallets {
		if walletBTC.CurrencyID == wp.CurrencyID && walletBTC.NetworkID == wp.NetworkID && walletBTC.WalletIndex == wp.WalletIndex {
			err = errors.New(msgErrWalletIndex)
			return err
		}
	}

	for _, walletETH := range user.WalletsETH {
		if walletETH.CurrencyID == wp.CurrencyID && walletETH.NetworkID == wp.NetworkID && walletETH.WalletIndex == wp.WalletIndex {
			err = errors.New(msgErrWalletIndex)
			return err
		}
	}

	sel := bson.M{"devices.JWT": token}

	switch wp.CurrencyID {
	case currencies.Bitcoin:
		walletBTC := createWallet(wp.CurrencyID, wp.NetworkID, wp.Address, wp.AddressIndex, wp.WalletIndex, wp.WalletName)
		update := bson.M{"$push": bson.M{"wallets": walletBTC}}

		err = restClient.userStore.Update(sel, update)
		if err != nil {
			restClient.log.Errorf("addWallet: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			err := errors.New(msgErrServerError)
			return err
		}

		//TODO: uncomment
		err = AddWatchAndResync(wp.CurrencyID, wp.NetworkID, user.UserID, wp.Address, restClient)
		if err != nil {
			restClient.log.Errorf("createCustomWallet: AddWatchAndResync: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			err := errors.New(msgErrServerError)
			return err
		}
	case currencies.Ether:
		walletETH := createWallet(wp.CurrencyID, wp.NetworkID, wp.Address, wp.AddressIndex, wp.WalletIndex, wp.WalletName)
		update := bson.M{"$push": bson.M{"walletsEth": walletETH}}

		err = restClient.userStore.Update(sel, update)
		if err != nil {
			restClient.log.Errorf("addWallet: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			err := errors.New(msgErrServerError)
			return err
		}

		err = AddWatchAndResync(wp.CurrencyID, wp.NetworkID, user.UserID, wp.Address, restClient)
		if err != nil {
			restClient.log.Errorf("createCustomWallet: AddWatchAndResync: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			err := errors.New(msgErrServerError)
			return err
		}

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

	switch cn.CurrencyID {
	case currencies.Bitcoin:
		for _, wallet := range user.Wallets {
			if wallet.CurrencyID == cn.CurrencyID && wallet.WalletIndex == cn.WalletIndex && wallet.NetworkID == cn.NetworkID {
				sel := bson.M{"userID": user.UserID, "wallets.walletIndex": cn.WalletIndex, "wallets.networkID": cn.NetworkID}
				update := bson.M{
					"$set": bson.M{
						"wallets.$.walletName": cn.WalletName,
					},
				}
				err := restClient.userStore.Update(sel, update)
				if err != nil {
					err := errors.New(msgErrServerError)
					return err
				}
				return nil
			}
		}
	case currencies.Ether:
		for _, walletETH := range user.WalletsETH {
			if walletETH.CurrencyID == cn.CurrencyID && walletETH.WalletIndex == cn.WalletIndex && walletETH.NetworkID == cn.NetworkID {
				sel := bson.M{"userID": user.UserID, "walletsEth.walletIndex": cn.WalletIndex, "walletsEth.networkID": cn.NetworkID}
				update := bson.M{
					"$set": bson.M{
						"walletsEth.$.walletName": cn.WalletName,
					},
				}
				err := restClient.userStore.Update(sel, update)
				if err != nil {
					err := errors.New(msgErrServerError)
					return err
				}
				return nil
			}
		}
	}

	err := errors.New(msgErrNoWallet)
	return err

}

func addAddressToWallet(address, token string, currencyID, networkid, walletIndex, addressIndex int, restClient *RestClient, c *gin.Context) error {
	user := store.User{}
	query := bson.M{"devices.JWT": token}

	if err := restClient.userStore.FindUser(query, &user); err != nil {
		restClient.log.Errorf("deleteWallet: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		err := errors.New(msgErrUserNotFound)
		return err
	}

	for _, wallet := range user.Wallets {
		if wallet.NetworkID == networkid && wallet.CurrencyID == currencyID && wallet.WalletIndex == walletIndex {
			for _, walletAddress := range wallet.Adresses {
				if walletAddress.AddressIndex == addressIndex {
					err := errors.New(msgErrAddressIndex)
					return err
				}
			}
		}
	}

	addr := store.Address{
		Address:        address,
		AddressIndex:   addressIndex,
		LastActionTime: time.Now().Unix(),
	}
	sel := bson.M{"devices.JWT": token, "wallets.walletIndex": walletIndex}
	update := bson.M{"$push": bson.M{"wallets.$.addresses": addr}}
	if err := restClient.userStore.Update(sel, update); err != nil {
		restClient.log.Errorf("addAddressToWallet: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		err := errors.New(msgErrServerError)
		return err
	}

	AddWatchAndResync(currencyID, networkid, user.UserID, address, restClient)
	// TODO: implement re-sync
	// TODO: add to watch address

	return nil

}

func AddWatchAndResync(currencyID, networkid int, userid, address string, restClient *RestClient) error {

	err := NewAddressNode(address, userid, currencyID, networkid, restClient)
	if err != nil {
		restClient.log.Errorf("AddWatchAndResync: NewAddressWs: currencies.Main: WsBtcMainnetCli.Emit:resync %s\t", err.Error())
		return err
	}

	return nil
}

func NewAddressNode(address, userid string, currencyID, networkID int, restClient *RestClient) error {
	switch currencyID {
	case currencies.Bitcoin:
		if networkID == currencies.Main {
			restClient.BTC.WatchAddressMain <- pb.WatchAddress{
				Address: address,
				UserID:  userid,
			}
		}

		if networkID == currencies.Test {
			restClient.BTC.WatchAddressTest <- pb.WatchAddress{
				Address: address,
				UserID:  userid,
			}
		}
	case currencies.Ether:
		//soon
	}
	return nil
}

func (restClient *RestClient) drop() gin.HandlerFunc {
	return func(c *gin.Context) {
		// restClient.userStore.DropTest()
	}
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

func (restClient *RestClient) statusCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, `{"Status":"ok"}`)
	}
}

func (restClient *RestClient) getServerConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		resp := map[string]interface{}{
			"stockexchanges": map[string][]string{
				"poloniex": []string{"usd_btc", "eth_btc", "eth_usd", "btc_usd"},
				"gdax":     []string{"eur_btc", "usd_btc", "eth_btc", "eth_usd", "eth_eur", "btc_usd"},
			},
			"servertime": time.Now().Unix(),
			"api":        "0.01",
			"android": map[string]int{
				"soft": 1,
				"hard": 1,
			},
			"ios": map[string]int{
				"soft": 18,
				"hard": 1,
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

		walletIndex, err := strconv.Atoi(c.Param("walletindex"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[walletindexr=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int wallet index:[%d] %s \t[addr=%s]", walletIndex, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeWalletIndexErr,
			})
			return
		}

		currencyId, err := strconv.Atoi(c.Param("currencyid"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[currencyId=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int wallet index:[%d] %s \t[addr=%s]", walletIndex, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeCurIndexErr,
			})
			return
		}

		networkid, err := strconv.Atoi(c.Param("networkid"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[networkid=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int networkid index:[%d] %s \t[addr=%s]", walletIndex, err.Error(), c.Request.RemoteAddr)
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

			err := restClient.userStore.DeleteWallet(user.UserID, walletIndex, currencyId, networkid)
			if err != nil {
				restClient.log.Errorf("deleteWallet: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": msgErrNoWallet,
				})
				return
			}
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)

		case currencies.Ether:
			// var totalBalance int64
			// for _, wallet := range user.Wallets {
			// 	if wallet.WalletIndex == walletIndex {
			// 		for _, address := range wallet.Adresses {
			// 			balance, err := restClient.eth.GetAddressBalance(address.Address)
			// 			if err != nil {
			// 				restClient.log.Errorf("deleteWallet: eth.GetAddressBalance: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			// 			}
			// 			totalBalance += balance.Int64()
			// 		}
			// 	}
			// }
			// if totalBalance != 0 {
			// 	c.JSON(http.StatusBadRequest, gin.H{
			// 		"code":    http.StatusBadRequest,
			// 		"message": msgErrWalletNonZeroBalance,
			// 	})
			// 	return
			// }
			// err := restClient.userStore.DeleteWallet(user.UserID, walletIndex)
			// if err != nil {
			// 	restClient.log.Errorf("deleteWallet: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			// 	c.JSON(http.StatusBadRequest, gin.H{
			// 		"code":    http.StatusBadRequest,
			// 		"message": msgErrNoWallet,
			// 	})
			// 	return
			// }
			// code = http.StatusOK
			// message = http.StatusText(http.StatusOK)
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
			restClient.log.Errorf("getWalletVerbose: non int networkid index:[%d] %s \t[addr=%s]", networkid, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodenetworkidErr,
			})
			return
		}

		switch currencyID {
		case currencies.Bitcoin:
			var rates []store.RatesRecord
			speeds := []int{
				1, 2, 3, 4, 5,
			}
			if err := restClient.userStore.GetAllRates(currencyID, networkid, "category", &rates); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"speeds":  sp,
					"code":    http.StatusInternalServerError,
					"message": msgErrRatesError,
				})
			}
			value := len(rates) / 4

			for i := 0; i < len(speeds); i++ {
				speeds[i] = value * i
			}
			fmt.Println(speeds)
			// for i := 0; i < len(speeds); i++ {
			// 	arr := rates[speeds[i]:speeds[i+1]]
			// 	fmt.Println("len ", len(arr))
			// }

			sp = EstimationSpeeds{
				VerySlow: avg(rates[0:speeds[1]]),
				Slow:     avg(rates[speeds[1]:speeds[2]]),
				Medium:   avg(rates[speeds[2]:speeds[3]]),
				Fast:     avg(rates[speeds[3]:speeds[4]]),
				VeryFast: avg(rates[speeds[4]-1 : len(rates)-1]),
			}
			c.JSON(http.StatusOK, gin.H{
				"speeds":  sp,
				"code":    http.StatusOK,
				"message": http.StatusText(http.StatusOK),
			})
		case currencies.Ether:

		default:

		}

	}
}

func avg(arr []store.RatesRecord) int {
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
			restClient.log.Errorf("getSpendableOutputs: non int networkid index: %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
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
	Address      string `json:"address"`
	AddressIndex int    `json:"addressindex"`
	WalletIndex  int    `json:"walletindex"`
	Transaction  string `json:"transaction"`
	IsHD         bool   `json:"ishd"`
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

				resp, err := restClient.BTC.CliMain.EventSendRawTx(context.Background(), &pb.RawTx{
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
				c.JSON(code, gin.H{
					"code":    code,
					"message": resp.Message,
				})
				return

			}
			if rawTx.NetworkID == currencies.Test {

				resp, err := restClient.BTC.CliTest.EventSendRawTx(context.Background(), &pb.RawTx{
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
				c.JSON(code, gin.H{
					"code":    code,
					"message": resp.Message,
				})
				return
			}

		case currencies.Ether:
			// hash, err := restClient.eth.SendRawTransaction(rawTx.Transaction)
			// if err != nil {
			// 	restClient.log.Errorf("sendRawHDTransaction:eth.SendRawTransaction %s", err.Error())
			// 	c.JSON(http.StatusInternalServerError, gin.H{
			// 		"code":    http.StatusInternalServerError,
			// 		"message": err.Error(),
			// 	})
			// 	return
			// }

			// c.JSON(http.StatusOK, gin.H{
			// 	"code":    http.StatusOK,
			// 	"message": hash,
			// })
			// return
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

/*
func (restClient *RestClient) sendRawTransaction(btcNodeAddress string) gin.HandlerFunc {
	return func(c *gin.Context) {

		currencyID, err := strconv.Atoi(c.Param("currencyid"))
		if err != nil {
			restClient.log.Errorf("getSpendableOutputs: non int currencyID:[%d] %s \t[addr=%s]", currencyID, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeCurIndexErr,
				"outs":    0,
			})
			return
		}

		switch currencyID {
		case currencies.Bitcoin:
			// Notice the notification parameter is nil since notifications are
			// not supported in HTTP POST mode.
			client, err := rpcclient.New(connCfg, nil)
			if err != nil {
				restClient.log.Errorf("sendRawTransaction: rpcclient.New  \t[addr=%s]", err, c.Request.RemoteAddr)
			}
			defer client.Shutdown()

			var rawTx RawTx

			decodeBody(c, &rawTx)
			txid, err := client.SendCyberRawTransaction(rawTx.Transaction, true)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": err.Error(),
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"TransactionHash": txid,
				})
			}
		case currencies.Ether:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
			})
		}
	}
}
*/

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

		walletIndex, err := strconv.Atoi(c.Param("walletindex"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[walletindexr=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int wallet index:[%d] %s \t[addr=%s]", walletIndex, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeWalletIndexErr,
				"wallet":  wv,
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
		//query := bson.M{"devices.JWT": token, "wallets.walletIndex": walletIndex}
		queryBTC := bson.M{"devices.JWT": token, "wallets.walletIndex": walletIndex}

		if err := restClient.userStore.FindUser(queryBTC, &user); err != nil {
			restClient.log.Errorf("getAllWalletsVerbose: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(code, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrUserNotFound,
				"wallet":  wv,
			})
			return
		}

		queryETH := bson.M{"devices.JWT": token, "walletsEth.walletIndex": walletIndex}

		if err := restClient.userStore.FindUser(queryETH, &user); err != nil {
			restClient.log.Errorf("getAllWalletsVerbose: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(code, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrUserNotFound,
				"wallet":  wv,
			})
			return
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

		//networkid, err := strconv.Atoi(c.Param("networkid"))
		//restClient.log.Debugf("getWalletVerbose [%d] \t[networkid=%s]", walletIndex, c.Request.RemoteAddr)
		//if err != nil {
		//	restClient.log.Errorf("getWalletVerbose: non int networkid index:[%d] %s \t[addr=%s]", walletIndex, err.Error(), c.Request.RemoteAddr)
		//	c.JSON(http.StatusBadRequest, gin.H{
		//		"code":    http.StatusBadRequest,
		//		"message": msgErrDecodenetworkidErr,
		//	})
		//	return
		//}

		switch currencyId {
		case currencies.Bitcoin:
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)
			var av []AddressVerbose

			for _, wallet := range user.Wallets {
				if wallet.WalletIndex == walletIndex { // specify wallet index

					var pending bool
					for _, address := range wallet.Adresses {
						spOuts := getBTCAddressSpendableOutputs(address.Address, currencyId, networkId, restClient)

						for _, spOut := range spOuts {
							if spOut.TxStatus == store.TxStatusAppearedInMempoolIncoming {
								pending = true
							}
						}

						av = append(av, AddressVerbose{
							LastActionTime: address.LastActionTime,
							Address:        address.Address,
							AddressIndex:   address.AddressIndex,
							Amount:         int64(checkBTCAddressbalance(address.Address, currencyId, networkId, restClient)),
							SpendableOuts:  spOuts,
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
				}
			}
		case currencies.Ether:
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)
			//TODO maybe change it!
			var av []AddressVerbose

			for _, walletETH := range user.WalletsETH {
				if walletETH.WalletIndex == walletIndex { // specify wallet index

					var pending bool

					//TODO remove this hardcode!
					pend := rand.Int31n(2)
					if pend == 0 {
						pending = true
					} else {
						pending = false
					}

					for _, address := range walletETH.Adresses {

						av = append(av, AddressVerbose{
							LastActionTime: address.LastActionTime,
							Address:        address.Address,
							AddressIndex:   address.AddressIndex,
							Amount:         walletETH.Balance,
						})
					}
					wv = append(wv, WalletVerboseETH{
						WalletIndex:    walletETH.WalletIndex,
						CurrencyID:     walletETH.CurrencyID,
						NetworkID:      walletETH.NetworkID,
						WalletName:     walletETH.WalletName,
						LastActionTime: walletETH.LastActionTime,
						DateOfCreation: walletETH.DateOfCreation,
						Nonce:          walletETH.Nonce,
						Balance:        walletETH.Balance,
						VerboseAddress: av,
						Pending:        pending,
					})
					av = []AddressVerbose{}
				}
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
	CurrencyID     int              `json:"currencyid"`
	NetworkID      int              `json:"networkid"`
	WalletIndex    int              `json:"walletindex"`
	WalletName     string           `json:"walletname"`
	LastActionTime int64            `json:"lastactiontime"`
	DateOfCreation int64            `json:"dateofcreation"`
	Nonce          int64            `json:"nonce"`
	Balance        int64            `json:"balance"`
	VerboseAddress []AddressVerbose `json:"addresses"`
	Pending        bool             `json:"pending"`
}

type AddressVerbose struct {
	LastActionTime int64                    `json:"lastActionTime"`
	Address        string                   `json:"address"`
	AddressIndex   int                      `json:"addressindex"`
	Amount         int64                    `json:"amount"`
	SpendableOuts  []store.SpendableOutputs `json:"spendableoutputs"`
}

type ETHAddressVerbose struct {
	LastActionTime int64  `json:"lastActionTime"`
	Address        string `json:"address"`
	AddressIndex   int    `json:"addressindex"`
	Amount         int64  `json:"amount"`
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

func findTopIndexes(walletsBTC []store.Wallet, walletsETH []store.WalletETH) []TopIndex {
	top := map[TopIndex]int{} // currency id -> topindex
	topIndex := []TopIndex{}
	for _, wallet := range walletsBTC {
		top[TopIndex{wallet.CurrencyID, wallet.NetworkID, 0}]++
	}

	for _, wallet := range walletsETH {
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

func fetchUndeletedWallets(walletsBTC []store.Wallet, walletsETH []store.WalletETH) ([]store.Wallet, []store.WalletETH) {
	//func fetchUndeletedWallets(wallets []store.Wallet) []store.Wallet {
	okWalletsBTC := []store.Wallet{}
	okWalletsETH := []store.WalletETH{}

	for _, wallet := range walletsBTC {
		if wallet.Status == store.WalletStatusOK {
			okWalletsBTC = append(okWalletsBTC, wallet)
		}
	}

	for _, wallet := range walletsETH {
		if wallet.Status == store.WalletStatusOK {
			okWalletsETH = append(okWalletsETH, wallet)
		}
	}

	return okWalletsBTC, okWalletsETH
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

		topIndexes := findTopIndexes(user.Wallets, user.WalletsETH)

		code = http.StatusOK
		message = http.StatusText(http.StatusOK)

		var av []AddressVerbose

		okWalletsBTC, okWalletsETH := fetchUndeletedWallets(user.Wallets, user.WalletsETH)

		//
		//walletBTC, ok := walletRaw.(store.Wallet)
		//if ok {
		//	//this is BTC wallet
		//	if walletBTC.Status == store.WalletStatusOK {
		//		okWallets = append(okWallets, walletBTC)
		//	}
		//}

		for _, wallet := range okWalletsBTC {
			var pending bool

			for _, address := range wallet.Adresses {

				spOuts := getBTCAddressSpendableOutputs(address.Address, wallet.CurrencyID, wallet.NetworkID, restClient)
				for _, spOut := range spOuts {
					if spOut.TxStatus == store.TxStatusAppearedInMempoolIncoming {
						pending = true
					}
				}

				av = append(av, AddressVerbose{
					LastActionTime: address.LastActionTime,
					Address:        address.Address,
					AddressIndex:   address.AddressIndex,
					Amount:         int64(checkBTCAddressbalance(address.Address, wallet.CurrencyID, wallet.NetworkID, restClient)),
					SpendableOuts:  spOuts,
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

		}

		for _, walletETH := range okWalletsETH {
			var pending bool

			//TODO remove this hardcode!
			pend := rand.Int31n(2)
			if pend == 0 {
				pending = true
			} else {
				pending = false
			}

			for _, address := range walletETH.Adresses {

				av = append(av, AddressVerbose{
					LastActionTime: address.LastActionTime,
					Address:        address.Address,
					AddressIndex:   address.AddressIndex,
					Amount:         walletETH.Balance,
				})
			}
			wv = append(wv, WalletVerboseETH{
				WalletIndex:    walletETH.WalletIndex,
				CurrencyID:     walletETH.CurrencyID,
				NetworkID:      walletETH.NetworkID,
				WalletName:     walletETH.WalletName,
				LastActionTime: walletETH.LastActionTime,
				DateOfCreation: walletETH.DateOfCreation,
				Nonce:          walletETH.Nonce,
				Balance:        walletETH.Balance,
				VerboseAddress: av,
				Pending:        pending,
			})
			av = []AddressVerbose{}
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

		walletIndex, err := strconv.Atoi(c.Param("walletindex"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[walletindexr=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int wallet index:[%d] %s \t[addr=%s]", walletIndex, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeWalletIndexErr,
				"history": walletTxs,
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
		restClient.log.Debugf("getWalletVerbose [%d] \t[networkid=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int networkid index:[%d] %s \t[addr=%s]", walletIndex, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodenetworkidErr,
			})
			return
		}

		switch currencyId {
		case currencies.Bitcoin:

			var blockHeight int64
			switch networkid {
			case currencies.Test:
				resp, err := restClient.BTC.CliTest.EventGetBlockHeight(context.Background(), &pb.Empty{})
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
				resp, err := restClient.BTC.CliMain.EventGetBlockHeight(context.Background(), &pb.Empty{})
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

			for _, tx := range userTxs {
				//New Logic
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

			for i := 0; i < len(walletTxs); i++ {
				walletTxs[i].Confirmations = int(blockHeight-walletTxs[i].BlockHeight) + 1
			}

		case currencies.Ether:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
				"history": walletTxs,
			})
			return
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
				"history": walletTxs,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": http.StatusText(http.StatusOK),
			"history": walletTxs,
		})
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
