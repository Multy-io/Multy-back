package client

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Appscrunch/Multy-back/btc"
	"github.com/Appscrunch/Multy-back/currencies"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/blockcypher/gobcy"
	"github.com/ventu-io/slf"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

const (
	msgErrHeaderError           = "wrong authorization headers"
	msgErrRequestBodyError      = "missing request body params"
	msgErrUserNotFound          = "user not found in db"
	msgErrRatesError            = "internal server error rates"
	msgErrDecodeWalletIndexErr  = "wrong wallet index"
	msgErrNoSpendableOuts       = "no spendable outputs"
	msgErrDecodeCurIndexErr     = "wrong currency index"
	msgErrAdressBalance         = "empty address or 3-rd party server error"
	msgErrChainIsNotImplemented = "current chain is not implemented"
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

func SetRestHandlers(userDB store.UserStore, btcConfTest, btcConfMain BTCApiConf, r *gin.Engine, clientRPC *rpcclient.Client) (*RestClient, error) {
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

	v1 := r.Group("/api/v1")
	v1.Use(restClient.middlewareJWT.MiddlewareFunc())
	{
		v1.POST("/wallet", restClient.addWallet())
		v1.POST("/address", restClient.addAddress())
		v1.GET("/wallets", restClient.getWallets())
		v1.GET("/transaction/feerate", restClient.getFeeRate())
		v1.GET("/outputs/spendable/:currencyid/:addr", restClient.getSpendableOutputs())
		v1.GET("/getexchangeprice/:from/:to", restClient.getExchangePrice())
		v1.POST("/transaction/send/:currencyid", restClient.sendRawTransaction())
		v1.GET("/address/balance/:currencyid/:address", restClient.getAdressBalance())
		v1.GET("/wallet/:walletindex/verbose/", restClient.getWalletVerbose())
		v1.GET("/wallets/verbose", restClient.getAllWalletsVerbose())
		v1.GET("/wallets/restore", restClient.restoreAllWallets())
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
				return store.User{}, false
			}
			return user, true
		},
		//Authorizator:  authorizator,
		//Unauthorized: unauthorized,
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

func (restClient *RestClient) addWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) < 2 {
			restClient.log.Errorf("addWallet: wrong Authorization header len\t[addr=%s]", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
			})
			return
		}

		token := authHeader[1]

		var (
			code    int
			message string
		)

		var wp WalletParams

		err := decodeBody(c, &wp)
		if err != nil {
			restClient.log.Errorf("addWallet: decodeBody: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		}

		wallet := createWallet(wp.CurrencyID, wp.Address, wp.AddressIndex, wp.WalletIndex, wp.WalletName)

		sel := bson.M{"devices.JWT": token}
		update := bson.M{"$push": bson.M{"wallets": wallet}}

		if err := restClient.userStore.Update(sel, update); err != nil {
			restClient.log.Errorf("addWallet: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			code = http.StatusBadRequest
			message = msgErrUserNotFound
		} else {
			code = http.StatusOK
			message = "wallet created"
		}

		c.JSON(http.StatusCreated, gin.H{
			"code":    code,
			"message": message,
		})
		return
	}
}

func (restClient *RestClient) addAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) < 2 {
			restClient.log.Errorf("addAddress: wrong Authorization header len\t[addr=%s]", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
			})
			return
		}

		token := authHeader[1]

		var sw SelectWallet
		err := decodeBody(c, &sw)
		if err != nil {
			restClient.log.Errorf("addAddress: decodeBody: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		}

		addr := store.Address{
			Address:      sw.Address,
			AddressIndex: sw.AddressIndex,
		}

		var (
			code    int
			message string
		)

		sel := bson.M{"wallets.walletIndex": sw.WalletIndex, "devices.JWT": token}
		update := bson.M{"$push": bson.M{"wallets.$.addresses": addr}}

		if err = restClient.userStore.Update(sel, update); err != nil {
			restClient.log.Errorf("addAddress: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			code = http.StatusBadRequest
			message = msgErrUserNotFound
		} else {
			code = http.StatusOK
			message = "address added"
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    code,
			"message": message,
		})
	}
}

func (restClient *RestClient) getWallets() gin.HandlerFunc { //recieve rgb(255, 0, 0)
	return func(c *gin.Context) {
		var user store.User

		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) < 2 {
			restClient.log.Errorf("getWallets: wrong Authorization header len\t[addr=%s]", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":        http.StatusBadRequest,
				"message":     msgErrHeaderError,
				"userWallets": user,
			})
			return
		}

		token := authHeader[1]

		var (
			code    int
			message string
		)

		sel := bson.M{"devices.JWT": token}
		if err := restClient.userStore.FindUser(sel, &user); err != nil {
			restClient.log.Errorf("getWallets: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			code = http.StatusBadRequest
			message = msgErrUserNotFound
		} else {
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)
		}

		c.JSON(http.StatusOK, gin.H{
			"code":        code,
			"message":     message,
			"userWallets": user.Wallets,
		})

	}
}

func (restClient *RestClient) getFeeRate() gin.HandlerFunc {
	return func(c *gin.Context) {

		var rates []store.RatesRecord
		var sp EstimationSpeeds
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

	}
}

func (restClient *RestClient) getSpendableOutputs() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) < 2 {
			restClient.log.Errorf("getSpendableOutputs: wrong Authorization header len\t[addr=%s]", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
				"outs":    nil,
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
				"outs":    nil,
			})
			return
		}

		address := c.Param("addr")
		params := map[string]string{
			"unspentOnly": "true",
		}
		var (
			code     int
			message  string
			addrInfo gobcy.Addr
		)

		switch currencyID {
		case currencies.Testnet:
			addrInfo, err = restClient.apiBTCTest.GetAddrFull(address, params)
			if err != nil {
				restClient.log.Errorf("getSpendableOutputs: restClient.apiBTCMain.GetAddrFull : %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
				code = http.StatusInternalServerError
				message = http.StatusText(http.StatusInternalServerError)
			} else {
				code = http.StatusOK
				message = http.StatusText(http.StatusOK)
			}

		case currencies.Bitcoin:
			addrInfo, err = restClient.apiBTCMain.GetAddrFull(address, params)
			if err != nil {
				restClient.log.Errorf("getSpendableOutputs: restClient.apiBTCMain.GetAddrFull : %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
				code = http.StatusInternalServerError
				message = http.StatusText(http.StatusInternalServerError)
			} else {
				code = http.StatusOK
				message = http.StatusText(http.StatusOK)
			}
		}

		var spOuts []SpendableOutputs

		for _, v := range addrInfo.TXs {
			for key, output := range v.Outputs {
				for _, addr := range output.Addresses {
					if addr == address {
						spOuts = append(spOuts, SpendableOutputs{
							TxID:        v.Hash,
							TxOutID:     key,
							TxOutAmount: output.Value,
							TxOutScript: output.Script,
						})
					}
				}
			}
		}

		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
			"outs":    spOuts,
		})

	}
}

type SpendableOutputs struct {
	TxID        string `json:"txid"`
	TxOutID     int    `json:"txoutid"`
	TxOutAmount int    `json:"txoutamount"`
	TxOutScript string `json:"txoutscript"`
}

func (restClient *RestClient) sendRawTransaction() gin.HandlerFunc {
	return func(c *gin.Context) {

		connCfg := &rpcclient.ConnConfig{
			Host:         "localhost:18334",
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
			// TODO: remove fatal in package
			restClient.log.Fatal(err.Error())
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
	}
}

type RawTx struct { // remane RawClientTransaction
	Transaction string `json:"transaction"` //HexTransaction
}

func (restClient *RestClient) getAdressBalance() gin.HandlerFunc { //recieve rgb(255, 0, 0)
	return func(c *gin.Context) {
		address := c.Param("address")
		currencyID, err := strconv.Atoi(c.Param("currencyid"))
		restClient.log.Debugf("getAdressBalance [%d] \t[addr=%s]", currencyID, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getAdressBalance: non int currencyID:[%d] %s \t[addr=%s]", currencyID, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":     http.StatusBadRequest,
				"message":  msgErrDecodeCurIndexErr,
				"ballance": nil,
			})
			return
		}
		var (
			code     int
			message  string
			ballance int
		)

		switch currencyID {
		case currencies.Testnet:
			addr, err := restClient.apiBTCTest.GetAddr(address, nil)
			if err != nil {
				restClient.log.Errorf("getAdressBalance: restClient.apiBTCTest.GetAddr : %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
				code = http.StatusInternalServerError
				message = http.StatusText(http.StatusInternalServerError)
				ballance = 0
			} else {
				code = http.StatusOK
				message = http.StatusText(http.StatusOK)
				ballance = addr.Balance
			}

			c.JSON(http.StatusOK, gin.H{
				"code":     code,
				"message":  message,
				"ballance": ballance,
			})

		case currencies.Bitcoin:
			addr, err := restClient.apiBTCMain.GetAddr(address, nil)
			if err != nil {
				restClient.log.Errorf("getAdressBalance: restClient.apiBTCMain.GetAddr : %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
				code = http.StatusInternalServerError
				message = http.StatusText(http.StatusInternalServerError)
				ballance = 0
				return
			} else {
				code = http.StatusOK
				message = http.StatusText(http.StatusOK)
				ballance = addr.Balance
			}
			c.JSON(http.StatusOK, gin.H{
				"code":     code,
				"message":  message,
				"ballance": ballance,
			})
		default:
			c.JSON(http.StatusOK, gin.H{
				"code":     http.StatusBadRequest,
				"message":  msgErrChainIsNotImplemented,
				"ballance": 0,
			})
		}
	}
}

func (restClient *RestClient) blank() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func (restClient *RestClient) getWalletVerbose() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) < 2 {
			restClient.log.Errorf("getWalletVerbose: wrong Authorization header len\t[addr=%s]", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
				"wallet":  nil,
			})
			return
		}
		token := authHeader[1]

		walletIndex, err := strconv.Atoi(c.Param("walletindex"))
		restClient.log.Debugf("getWalletVerbose [%d] \t[walletindexr=%s]", walletIndex, c.Request.RemoteAddr)
		if err != nil {
			restClient.log.Errorf("getWalletVerbose: non int wallet index:[%d] %s \t[addr=%s]", walletIndex, err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrDecodeWalletIndexErr,
				"wallet":  nil,
			})
			return
		}

		query := bson.M{"devices.JWT": token}

		user := store.User{}
		var (
			code    int
			message string
		)

		if err := restClient.userStore.FindUser(query, &user); err != nil {
			restClient.log.Errorf("getWalletVerbose: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			code = http.StatusBadRequest
			message = msgErrUserNotFound
		} else {
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)
		}

		var av []AddressVerbose
		var spOuts []SpendableOutputs

		params := map[string]string{
			"unspentOnly": "true",
		}

		for _, wallet := range user.Wallets {
			if wallet.WalletIndex == walletIndex {
				for _, address := range wallet.Adresses {
					addrInfo, err := restClient.apiBTCTest.GetAddrFull(address.Address, params)
					if err != nil {
						restClient.log.Errorf("getWalletVerbose: restClient.apiBTCTest.GetAddrFull : %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
						code = http.StatusInternalServerError
						message = http.StatusText(http.StatusInternalServerError) // fix naming
						continue
					}

					for _, v := range addrInfo.TXs {
						for key, output := range v.Outputs {
							for _, addr := range output.Addresses {
								if addr == address.Address {
									spOuts = append(spOuts, SpendableOutputs{
										TxID:        v.Hash,
										TxOutID:     key,
										TxOutAmount: output.Value,
										TxOutScript: output.Script,
									})
								}
							}
						}
					}

					av = append(av, AddressVerbose{
						Address:       address.Address,
						AddressIndex:  address.AddressIndex,
						Amount:        addrInfo.Balance,
						SpendableOuts: spOuts,
					})
				}
				spOuts = []SpendableOutputs{}
			}
		}

		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
			"wallet":  av,
		})

	}
}

type AddressVerbose struct {
	Address       string             `json:"address"`
	AddressIndex  int                `json:"addressindex"`
	Amount        int                `json:"amount"`
	SpendableOuts []SpendableOutputs `json:"spendableoutputs"`
}

func (restClient *RestClient) getAllWalletsVerbose() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) < 2 {
			restClient.log.Errorf("getAllWalletsVerbose: wrong Authorization header len\t[addr=%s]", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
				"wallets": nil,
			})
			return
		}
		token := authHeader[1]

		var (
			code    int
			message string
		)
		user := store.User{}
		query := bson.M{"devices.JWT": token}

		if err := restClient.userStore.FindUser(query, &user); err != nil {
			restClient.log.Errorf("getAllWalletsVerbose: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			code = http.StatusBadRequest
			message = msgErrUserNotFound
		} else {
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)
		}

		var wv []WalletVerbose
		var av []AddressVerbose
		var spOuts []SpendableOutputs
		params := map[string]string{
			"unspentOnly": "true",
		}

		for _, wallet := range user.Wallets {

			for _, address := range wallet.Adresses {
				addrInfo, err := restClient.apiBTCTest.GetAddrFull(address.Address, params)
				if err != nil {
					restClient.log.Errorf("getAllWalletsVerbose: restClient.apiBTCTest.GetAddrFull : %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
					code = http.StatusInternalServerError
					message = http.StatusText(http.StatusInternalServerError)
					continue
				}

				for _, v := range addrInfo.TXs {
					for key, output := range v.Outputs {
						for _, addr := range output.Addresses {
							if addr == address.Address {
								spOuts = append(spOuts, SpendableOutputs{
									TxID:        v.Hash,
									TxOutID:     key,
									TxOutAmount: output.Value,
									TxOutScript: output.Script,
								})
							}
						}
					}
				}

				av = append(av, AddressVerbose{
					Address:       address.Address,
					AddressIndex:  address.AddressIndex,
					Amount:        addrInfo.Balance,
					SpendableOuts: spOuts,
				})
				spOuts = []SpendableOutputs{}
			}

			wv = append(wv, WalletVerbose{
				WalletIndex:    wallet.WalletIndex,
				VerboseAddress: av,
			})
			av = []AddressVerbose{}
		}
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
			"wallets": wv,
		})
	}
}

type WalletVerbose struct {
	WalletIndex    int              `json:"walletindex"`
	VerboseAddress []AddressVerbose `json:"addresses"`
}

func (restClient *RestClient) restoreAllWallets() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) < 2 {
			restClient.log.Errorf("restoreAllWallets: wrong Authorization header len\t[addr=%s]", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
				"wallets": nil,
			})
			return
		}
		token := authHeader[1]

		var (
			code    int
			message string
		)
		user := store.User{}
		query := bson.M{"devices.JWT": token}

		if err := restClient.userStore.FindUser(query, &user); err != nil {
			restClient.log.Errorf("restoreAllWallets: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			code = http.StatusBadRequest
			message = msgErrUserNotFound
		} else {
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)
		}

		var rw []RestoreWallet
		var av []AddressVerbose
		var spOuts []SpendableOutputs
		params := map[string]string{
			"unspentOnly": "true",
		}

		for _, wallet := range user.Wallets {
			for _, address := range wallet.Adresses {
				addrInfo, err := restClient.apiBTCTest.GetAddrFull(address.Address, params)
				if err != nil {
					restClient.log.Errorf("restoreAllWallets: restClient.apiBTCTest.GetAddrFull : %s \t[addr=%s]", err.Error(), c.Request.RemoteAddr)
					code = http.StatusConflict
					message = "server error or no outputs on some address" // fix status code

					continue
				}

				for _, v := range addrInfo.TXs {
					for key, output := range v.Outputs {
						for _, addr := range output.Addresses {
							if addr == address.Address {
								spOuts = append(spOuts, SpendableOutputs{
									TxID:        v.Hash,
									TxOutID:     key,
									TxOutAmount: output.Value,
									TxOutScript: output.Script,
								})
							}
						}
					}
				}

				av = append(av, AddressVerbose{
					Address:       address.Address,
					AddressIndex:  address.AddressIndex,
					Amount:        addrInfo.Balance,
					SpendableOuts: spOuts,
				})
				spOuts = []SpendableOutputs{}
			}

			rw = append(rw, RestoreWallet{
				CurrencyID:     wallet.CurrencyID,
				WalletIndex:    wallet.WalletIndex,
				WalletName:     wallet.WalletName,
				LastActionTime: wallet.LastActionTime,
				DateOfCreation: wallet.DateOfCreation,
				VerboseAddress: av,
			})
			av = []AddressVerbose{}
		}
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
			"wallets": rw,
		})
	}
}

type RestoreWallet struct {
	CurrencyID     int              `json:"currencyID"`
	WalletIndex    int              `json:"walletIndex"`
	WalletName     string           `json:"walletName"`
	LastActionTime time.Time        `json:"lastActionTime"`
	DateOfCreation time.Time        `json:"dateOfCreation"`
	VerboseAddress []AddressVerbose `json:"addresses"`
}

func (restClient *RestClient) getBlock() gin.HandlerFunc {
	return func(c *gin.Context) {
		height := c.Param("height")

		url := "https://bitaps.com/api/block/" + height

		if height == "last" {
			url = "https://bitaps.com/api/block/latest"
		}

		response, err := http.Get(url)
		responseErr(c, err, http.StatusServiceUnavailable) // 503

		data, err := ioutil.ReadAll(response.Body)
		responseErr(c, err, http.StatusInternalServerError) // 500

		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Write(data)
	}
}

func (restClient *RestClient) getTickets() gin.HandlerFunc { // shapeshift
	return func(c *gin.Context) {
		pair := c.Param("pair")
		url := "https://shapeshift.io/marketinfo/" + pair

		var to map[string]interface{}

		makeRequest(c, url, &to)
		// fv := v.Convert(floatType).Float()

		c.JSON(http.StatusOK, gin.H{
			"pair":     to["pair"],
			"rate":     to["rate"],
			"limit":    to["limit"],
			"minimum":  to["minimum"],
			"minerFee": to["minerFee"],
		})
	}
}

func (restClient *RestClient) getExchangePrice() gin.HandlerFunc {
	return func(c *gin.Context) {
		from := strings.ToUpper(c.Param("from"))
		to := strings.ToUpper(c.Param("to"))

		url := "https://min-api.cryptocompare.com/data/price?fsym=" + from + "&tsyms=" + to
		var er map[string]interface{}
		makeRequest(c, url, &er)

		c.JSON(http.StatusOK, gin.H{
			to: er[to],
		})
	}
}

func (restClient *RestClient) getTransactionInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		txid := c.Param("txid")
		url := "https://bitaps.com/api/transaction/" + txid
		response, err := http.Get(url)
		responseErr(c, err, http.StatusServiceUnavailable) // 503

		data, err := ioutil.ReadAll(response.Body)
		responseErr(c, err, http.StatusInternalServerError) // 500

		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Write(data)
	}
}
