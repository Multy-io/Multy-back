/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package client

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Appscrunch/Multy-back/btc"
	"github.com/Appscrunch/Multy-back/currencies"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
	"github.com/blockcypher/gobcy"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
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
	msgErrNoSpendableOuts       = "no spendable outputs"
	msgErrDecodeCurIndexErr     = "wrong currency index"
	msgErrAdressBalance         = "empty address or 3-rd party server error"
	msgErrChainIsNotImplemented = "current chain is not implemented"
	msgErrUserHaveNoTxs         = "user have no transactions"
)

type RestClient struct {
	middlewareJWT *GinJWTMiddleware
	userStore     store.UserStore
	rpcClient     *rpcclient.Client
	// ballance api for test net
	apiBTCTest     gobcy.API
	btcConfTestnet BTCApiConf
	// ballance api for main net
	apiBTCMain     gobcy.API
	btcConfMainnet BTCApiConf

	log slf.StructuredLogger
}

type BTCApiConf struct {
	Token, Coin, Chain string
}

func SetRestHandlers(
	userDB store.UserStore,
	btcConfTest,
	btcConfMain BTCApiConf,
	r *gin.Engine,
	clientRPC *rpcclient.Client,
	btcNodeAddress string,
) (*RestClient, error) {
	restClient := &RestClient{
		userStore: userDB,
		rpcClient: clientRPC,

		btcConfTestnet: btcConfTest,
		btcConfMainnet: btcConfMain,

		apiBTCTest: gobcy.API{
			Token: btcConfTest.Token,
			Coin:  btcConfTest.Coin,
			Chain: btcConfTest.Chain,
		},
		apiBTCMain: gobcy.API{
			Token: btcConfMain.Token,
			Coin:  btcConfMain.Coin,
			Chain: btcConfMain.Chain,
		},
		log: slf.WithContext("rest-client"),
	}

	initMiddlewareJWT(restClient)

	r.POST("/auth", restClient.LoginHandler())
	r.GET("/server/config", restClient.getServerConfig())

	r.GET("/statuscheck", restClient.statusCheck())

	v1 := r.Group("/api/v1")
	v1.Use(restClient.middlewareJWT.MiddlewareFunc())
	{
		v1.POST("/wallet", restClient.addWallet())                                                          //nothing to change
		v1.DELETE("/wallet/:currencyid/:walletindex", restClient.deleteWallet())                            //todo add currency id √
		v1.POST("/address", restClient.addAddress())                                                        //todo add currency id √
		v1.GET("/transaction/feerate/:currencyid", restClient.getFeeRate())                                 //todo add currency id √
		v1.GET("/outputs/spendable/:currencyid/:addr", restClient.getSpendableOutputs())                    //nothing to change	√
		v1.POST("/transaction/send/:currencyid", restClient.sendRawTransaction(btcNodeAddress))             //todo add currency id √
		v1.GET("/wallet/:walletindex/verbose/:currencyid", restClient.getWalletVerbose())                   //todo add currency id √
		v1.GET("/wallets/verbose", restClient.getAllWalletsVerbose())                                       //nothing to change	√
		v1.GET("/wallets/transactions/:currencyid/:walletindex", restClient.getWalletTransactionsHistory()) //todo add currency id	√
		v1.POST("/wallet/name/:currencyid/:walletindex", restClient.changeWalletName())                     //todo add currency id √
		v1.GET("/exchange/changelly/list", restClient.changellyListCurrencies())
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
	Address      string `json:"address"`
	AddressIndex int    `json:"addressIndex"`
	WalletIndex  int    `json:"walletIndex"`
	WalletName   string `json:"walletName"`
}

type SelectWallet struct {
	CurrencyID   int    `json:"currencyID"`
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
	// we can't append wallet to user with already existed wallet id
	sel := bson.M{"devices.JWT": token, "wallets.walletIndex": wp.WalletIndex, "wallets.currencyID": wp.CurrencyID}
	err := restClient.userStore.FindUserErr(sel)

	if err == nil {
		// existed wallet
		return errors.New(msgErrWalletIndex)
	}
	if err != mgo.ErrNotFound {
		restClient.log.Errorf("addWallet: restClient.userStore.FindUser %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		return errors.New(http.StatusText(http.StatusInternalServerError))
	}

	wallet := createWallet(wp.CurrencyID, wp.Address, wp.AddressIndex, wp.WalletIndex, wp.WalletName)

	sel = bson.M{"devices.JWT": token}
	update := bson.M{"$push": bson.M{"wallets": wallet}}

	if err := restClient.userStore.Update(sel, update); err != nil {
		restClient.log.Errorf("addWallet: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		return errors.New(msgErrUserNotFound)
	} else {
		return nil
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
		}

		switch wp.CurrencyID {
		case currencies.Bitcoin:
			err := createCustomWallet(wp, token, restClient, c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusInternalServerError,
					"message": err.Error,
				})
				return
			}
			go resyncBTCAddress(wp.Address, c.Request.RemoteAddr, restClient)
		case currencies.Ether:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    code,
				"message": msgErrMethodNotImplennted,
			})
			return
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    code,
				"message": msgErrMethodNotImplennted,
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
	WalletName string `json:"walletname"`
	CurrencyID int    `json:"currencyID"`
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

		walletIndex, err := strconv.Atoi(c.Param("walletindex"))
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int wallet index:[%d] %s \t[addr=%s]", walletIndex, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeWalletIndexErr,
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

		switch cn.CurrencyID {
		case currencies.Bitcoin:
			sel := bson.M{"userID": user.UserID, "wallets.walletIndex": walletIndex}
			update := bson.M{
				"$set": bson.M{
					"wallets.$.walletName": cn.WalletName,
				},
			}
			err = restClient.userStore.Update(sel, update)
			if err != nil {
				restClient.log.Errorf("changeWalletName: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": msgErrNoWallet,
				})
				return
			}
		case currencies.Ether:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
			})
			return
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
			"donate": map[string]string{
				"BTC": "mzNZBhim9XGy66FkdzrehHwdWNgbiTYXCQ",
				"ETH": "0x54f46318d8f83c28b719ccf01ab4628e1e8f65fa",
			},
		}
		c.JSON(http.StatusOK, resp)
	}
}

func checkBTCAddressbalance(address string, restClient *RestClient) int64 {
	var balance int64
	query := bson.M{"address": address}
	spOuts, err := restClient.userStore.GetAddressSpendableOutputs(query)
	if err != nil {
		return balance
	}

	for _, out := range spOuts {
		balance += out.TxOutAmount
	}
	return balance
}

func getBTCAddressSpendableOutputs(address string, restClient *RestClient) []store.SpendableOutputs {
	query := bson.M{"address": address}
	spOuts, err := restClient.userStore.GetAddressSpendableOutputs(query)
	if err != nil && err != mgo.ErrNotFound {
		restClient.log.Errorf("getkBTCAddressSpendableOutputs: rGetAddressSpendableOutputs: %s\t", err.Error())
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

		switch currencyId {
		case currencies.Bitcoin:
			var totalBalance int64
			for _, wallet := range user.Wallets {
				if wallet.WalletIndex == walletIndex {
					for _, address := range wallet.Adresses {
						totalBalance += checkBTCAddressbalance(address.Address, restClient)
					}
				}
			}

			if totalBalance != 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": msgErrWalletNonZeroBalance,
				})
				return
			}

			err := restClient.userStore.DeleteWallet(user.UserID, walletIndex)
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
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
			})
			return
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
		})

	}
}

func resyncBTCAddress(hash, RemoteAdd string, restClient *RestClient) {
	allResync := []resyncTx{}
	requestTimes := 0
	addrInfo, err := restClient.apiBTCTest.GetAddrFull(hash, map[string]string{"limit": "50"})
	if err != nil {
		restClient.log.Errorf("resyncAddress: restClient.apiBTCTest.GetAddrFull : %s \t[addr=%s]", err.Error(), RemoteAdd)
	}

	if addrInfo.FinalNumTX > 50 {
		requestTimes = int(float64(addrInfo.FinalNumTX) / 50.0)
	}

	for _, tx := range addrInfo.TXs {
		allResync = append(allResync, resyncTx{
			hash:        tx.Hash,
			blockHeight: tx.BlockHeight,
		})
	}

	for i := 0; i < requestTimes; i++ {
		addrInfo, err := restClient.apiBTCTest.GetAddrFull(hash, map[string]string{"limit": "50", "before": strconv.Itoa(allResync[len(allResync)-1].blockHeight)})
		if err != nil {
			restClient.log.Errorf("resyncAddress: restClient.apiBTCTest.GetAddrFull: %s \t[addr=%s]", err.Error(), RemoteAdd)
		}
		for _, tx := range addrInfo.TXs {
			allResync = append(allResync, resyncTx{
				hash:        tx.Hash,
				blockHeight: tx.BlockHeight,
			})
		}
	}

	reverseResyncTx(allResync)

	for _, reTx := range allResync {
		txHash, err := chainhash.NewHashFromStr(reTx.hash)
		if err != nil {
			restClient.log.Errorf("resyncAddress: chainhash.NewHashFromStr = %s\t[addr=%s]", err, RemoteAdd)
		}
		rawTx, err := btc.GetRawTransactionVerbose(txHash)
		if err != nil {
			restClient.log.Errorf("resyncAddress: rpcClient.GetRawTransactionVerbose = %s\t[addr=%s]", err, RemoteAdd)
		}
		btc.ProcessTransaction(int64(reTx.blockHeight), rawTx)
	}
}
func reverseResyncTx(ss []resyncTx) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

type resyncTx struct {
	hash        string
	blockHeight int
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

		var (
			code    int
			message string
		)

		switch sw.CurrencyID {
		case currencies.Bitcoin:
			// we can't append to wallet addresses with same indexes
			query := bson.M{"devices.JWT": token, "wallets.walletIndex": sw.WalletIndex, "wallets.addresses.addressIndex": sw.AddressIndex}
			sel := bson.M{
				"wallets": 1,
			}
			ws := store.WalletsSelect{}
			err = restClient.userStore.FindUserAddresses(query, sel, &ws)
			if err != nil {
				restClient.log.Errorf("addAddress: restClient.userStore.FindUserAddresses: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			}

			flag := false // add to db flag

			for _, wallet := range ws.Wallets {
				if wallet.WalletIndex == sw.WalletIndex {
					for _, addrVerbose := range wallet.Addresses {
						if addrVerbose.AddressIndex == sw.AddressIndex {
							flag = true
						}
					}
				}
			}

			if flag {
				// existed address
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": msgErrAddressIndex,
				})
				return
			}

			addr := store.Address{
				Address:        sw.Address,
				AddressIndex:   sw.AddressIndex,
				LastActionTime: time.Now().Unix(),
			}

			sel = bson.M{"devices.JWT": token, "wallets.walletIndex": sw.WalletIndex}
			update := bson.M{"$push": bson.M{"wallets.$.addresses": addr}}

			if err = restClient.userStore.Update(sel, update); err != nil {
				restClient.log.Errorf("addAddress: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
				code = http.StatusBadRequest
				message = msgErrUserNotFound
			} else {
				code = http.StatusOK
				message = "address added"
				go resyncBTCAddress(sw.Address, c.Request.RemoteAddr, restClient)
			}

		case currencies.Ether:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
			})
			return
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    code,
			"message": message,
		})
	}
}

func (restClient *RestClient) getFeeRate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var sp EstimationSpeeds
		currencyId, err := strconv.Atoi(c.Param("currencyid"))
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int currency id: %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"speeds":  sp,
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeCurIndexErr,
			})
			return
		}

		switch currencyId {
		case currencies.Bitcoin:
			var rates []store.RatesRecord
			speeds := []int{
				1, 2, 3, 4, 5,
			}
			if err := restClient.userStore.GetAllRates("category", &rates); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"speeds":  sp,
					"code":    http.StatusInternalServerError,
					"message": msgErrRatesError,
				})
			}

			sp = EstimationSpeeds{
				VerySlow: rates[speeds[0]].Category,
				Slow:     rates[speeds[1]].Category,
				Medium:   rates[speeds[2]].Category,
				Fast:     rates[speeds[3]].Category,
				VeryFast: rates[speeds[4]].Category,
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
		var spOuts []store.SpendableOutputs

		switch currencyID {
		case currencies.Bitcoin:

			//todo remove query
			query := bson.M{"userid": user.UserID, "transactions.txaddress": address}
			spOuts, err = restClient.userStore.GetAddressSpendableOutputs(query)
			if err != nil {
				restClient.log.Errorf("getSpendableOutputs: GetAddressSpendableOutputs:[%d] %s \t[addr=%s]", currencyID, err.Error(), c.Request.RemoteAddr)
			}
			userTxs := store.TxRecord{}

			restClient.userStore.FindUserTxs(query, &userTxs)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": msgErrNoTransactionAddress,
					"outs":    0,
				})
				return
			}

		default:
			code = http.StatusBadRequest
			message = msgErrMethodNotImplennted
		}

		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
			"outs":    spOuts,
		})
	}
}

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
			restClient.log.Infof("btc.Cert=%s\n", btc.Cert)

			connCfg := &rpcclient.ConnConfig{
				Host:         btcNodeAddress,
				User:         "multy",
				Pass:         "multy",
				HTTPPostMode: true,  // Bitcoin core only supports HTTP POST mode
				DisableTLS:   false, // Bitcoin core does not provide TLS by default
				Certificates: []byte(btc.Cert),
			}
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

type RawTx struct { // remane RawClientTransaction
	Transaction string `json:"transaction"` //HexTransaction
}

func (restClient *RestClient) getWalletVerbose() gin.HandlerFunc {
	return func(c *gin.Context) {
		var wv []WalletVerbose
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

		var (
			code    int
			message string
		)
		user := store.User{}
		query := bson.M{"devices.JWT": token, "wallets.walletIndex": walletIndex}

		if err := restClient.userStore.FindUser(query, &user); err != nil {
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

		switch currencyId {
		case currencies.Bitcoin:
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)

			userTxs := store.TxRecord{}
			query = bson.M{"userid": user.UserID}
			if err := restClient.userStore.FindUserTxs(query, &userTxs); err != nil {
				restClient.log.Errorf("getAllWalletsVerbose: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": msgErrNoSpendableOutputs,
					"wallet":  wv,
				})
				return
			}

			var unspendTxs []store.MultyTX
			for _, tx := range userTxs.Transactions {
				if tx.TxStatus == store.TxStatusAppearedInMempoolIncoming || tx.TxStatus == store.TxStatusAppearedInBlockIncoming || tx.TxStatus == store.TxStatusInBlockConfirmedIncoming { // pending and actual ballance
					unspendTxs = append(unspendTxs, tx)
				}
			}

			var av []AddressVerbose

			for _, wallet := range user.Wallets {
				if wallet.WalletIndex == walletIndex { // specify wallet index

					for _, address := range wallet.Adresses {
						av = append(av, AddressVerbose{
							LastActionTime: address.LastActionTime,
							Address:        address.Address,
							AddressIndex:   address.AddressIndex,
							Amount:         int(checkBTCAddressbalance(address.Address, restClient)),
							SpendableOuts:  getBTCAddressSpendableOutputs(address.Address, restClient),
						})
					}
					wv = append(wv, WalletVerbose{
						WalletIndex:    wallet.WalletIndex,
						CurrencyID:     wallet.CurrencyID,
						WalletName:     wallet.WalletName,
						LastActionTime: wallet.LastActionTime,
						DateOfCreation: wallet.DateOfCreation,
						VerboseAddress: av,
					})
					av = []AddressVerbose{}
				}

			}
		case currencies.Ether:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
			})
			return
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
	WalletIndex    int              `json:"walletindex"`
	WalletName     string           `json:"walletname"`
	LastActionTime int64            `json:"lastactiontime"`
	DateOfCreation int64            `json:"dateofcreation"`
	VerboseAddress []AddressVerbose `json:"addresses"`
}
type AddressVerbose struct {
	LastActionTime int64                    `json:"lastActionTime"`
	Address        string                   `json:"address"`
	AddressIndex   int                      `json:"addressindex"`
	Amount         int                      `json:"amount"`
	SpendableOuts  []store.SpendableOutputs `json:"spendableoutputs"`
}

type StockExchangeRate struct {
	ExchangeName   string `json:"exchangename"`
	FiatEquivalent int    `json:"fiatequivalent"`
	TotalAmount    int    `json:"totalamount"`
}

func (restClient *RestClient) getAllWalletsVerbose() gin.HandlerFunc {
	return func(c *gin.Context) {
		var wv []WalletVerbose
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
				"wallet":  wv,
			})
			return
		}

		currencyId, err := strconv.Atoi(c.Param("currencyid"))
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int currency id:[%d] %s \t[addr=%s]", currencyId, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeCurIndexErr,
			})
			return
		}

		switch currencyId {
		case currencies.Bitcoin:
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)

			userTxs := store.TxRecord{}
			query = bson.M{"userid": user.UserID}
			if err := restClient.userStore.FindUserTxs(query, &userTxs); err != nil {
				restClient.log.Errorf("getAllWalletsVerbose: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": msgErrNoSpendableOutputs,
					"wallet":  wv,
				})
				return
			}

			var unspendTxs []store.MultyTX
			for _, tx := range userTxs.Transactions {
				if tx.TxStatus == store.TxStatusAppearedInMempoolIncoming || tx.TxStatus == store.TxStatusAppearedInBlockIncoming || tx.TxStatus == store.TxStatusInBlockConfirmedIncoming { // pending and actual ballance
					unspendTxs = append(unspendTxs, tx)
				}
			}

			var av []AddressVerbose


			for _, wallet := range user.Wallets {

				for _, address := range wallet.Adresses {
					av = append(av, AddressVerbose{
						LastActionTime: address.LastActionTime,
						Address:        address.Address,
						AddressIndex:   address.AddressIndex,
						Amount:         int(checkBTCAddressbalance(address.Address, restClient)),
						SpendableOuts:  getBTCAddressSpendableOutputs(address.Address, restClient),
					})
				}
				wv = append(wv, WalletVerbose{
					WalletIndex:    wallet.WalletIndex,
					CurrencyID:     wallet.CurrencyID,
					WalletName:     wallet.WalletName,
					LastActionTime: wallet.LastActionTime,
					DateOfCreation: wallet.DateOfCreation,
					VerboseAddress: av,
				})
				av = []AddressVerbose{}

			}
		case currencies.Ether:
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrChainIsNotImplemented,
			})
			return
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

		// txHistory := []TxHistory{}
		switch currencyId {
		case currencies.Bitcoin:
			query := bson.M{"userid": user.UserID}
			userTxs := store.TxRecord{}
			err = restClient.userStore.GetAllWalletTransactions(query, &userTxs)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    http.StatusBadRequest,
					"message": msgErrTxHistory,
					"history": walletTxs,
				})
				return
			}

			for _, tx := range userTxs.Transactions {
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

				//OLD logic
				//for _, output := range tx.WalletsOutput {
				//	if output.WalletIndex == walletIndex {
				//		walletTxs = append(walletTxs, tx)
				//	}
				//}
				//for _, input := range tx.WalletsInput {
				//	if input.WalletIndex == walletIndex {
				//		walletTxs = append(walletTxs, tx)
				//	}
				//}
			}

			/*
				for _, walletTx := range walletTxs {
					txHistory = append(txHistory, TxHistory{
						TxID:        walletTx.TxID,
						TxHash:      walletTx.TxHash,
						TxOutScript: walletTx.TxOutScript,
						TxAddress:   walletTx.TxAddress,
						TxStatus:    walletTx.TxStatus,
						TxOutAmount: walletTx.TxOutAmount,
						TxOutID:     walletTx.TxOutID,
						WalletIndex: walletTx.WalletIndex,
						BlockTime:   walletTx.BlockTime,
						BlockHeight: walletTx.BlockHeight,
						TxFee:       walletTx.TxFee,
						BtcToUsd:    walletTx.StockExchangeRate[0].Exchanges.BTCtoUSD,
						TxInputs:    walletTx.TxInputs,
						TxOutputs:   walletTx.TxOutputs,
						MempoolTime: walletTx.MempoolTime,
					})
				}
			*/
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
