package client

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Appscrunch/Multy-back/btc"
	"github.com/Appscrunch/Multy-back/currencies"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
	"github.com/blockcypher/gobcy"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
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
		v1.DELETE("/wallet/:walletindex", restClient.deleteWallet())
		v1.POST("/address", restClient.addAddress())
		v1.GET("/transaction/feerate", restClient.getFeeRate())
		v1.GET("/outputs/spendable/:currencyid/:addr", restClient.getSpendableOutputs())
		v1.POST("/transaction/send/:currencyid", restClient.sendRawTransaction())
		v1.GET("/wallet/:walletindex/verbose", restClient.getWalletVerbose())
		v1.GET("/wallets/verbose", restClient.getAllWalletsVerbose())
		v1.GET("/wallets/transactions/:walletindex", restClient.getWalletTransactionsHistory())
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

		// we can't append wallet to user with already existed wallet id
		sel := bson.M{"wallets.walletIndex": wp.WalletIndex, "devices.JWT": token}
		err = restClient.userStore.FindUserErr(sel)

		if err == nil {
			// existed wallet
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrWalletIndex,
			})
			return
		}
		if err != mgo.ErrNotFound {
			restClient.log.Errorf("addWallet: restClient.userStore.FindUser %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": http.StatusText(http.StatusInternalServerError),
			})
			return
		}

		wallet := createWallet(wp.CurrencyID, wp.Address, wp.AddressIndex, wp.WalletIndex, wp.WalletName)

		sel = bson.M{"devices.JWT": token}
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

func (restClient *RestClient) deleteWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) < 2 {
			restClient.log.Errorf("getAllWalletsVerbose: wrong Authorization header len\t[addr=%s]", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
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

		// Check ballance on wallet
		userTxs := store.TxRecord{}
		query = bson.M{"userid": user.UserID}
		if err := restClient.userStore.FindUserTxs(query, &userTxs); err != nil {
			restClient.log.Errorf("getAllWalletsVerbose: restClient.userStore.FindUser: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)
		}

		var unspendTxs []store.MultyTX
		for _, tx := range userTxs.Transactions {
			if tx.TxStatus != "spend in mempool" && tx.TxStatus != "spend in block" {
				unspendTxs = append(unspendTxs, tx)
			}
		}

		var balance int
		var totalBalance int

		for _, wallet := range user.Wallets {
			if wallet.WalletIndex == walletIndex {
				for _, address := range wallet.Adresses {

					for _, tx := range unspendTxs {
						if tx.TxAddress == address.Address {
							balance += int(tx.TxOutAmount * float64(100000000))
						}
					}
					totalBalance += balance
					balance = 0
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

		sel := bson.M{"userID": user.UserID, "wallets.walletIndex": walletIndex}
		update := bson.M{
			"$set": bson.M{
				"wallets.$.status": store.WalletStatusDeleted,
			},
		}
		err = restClient.userStore.Update(sel, update)
		if err != nil {
			restClient.log.Errorf("deleteWallet: restClient.userStore.Update: %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrNoWallet,
			})
			return
		}

		/////////

		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})

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
			Address:      sw.Address,
			AddressIndex: sw.AddressIndex,
		}

		var (
			code    int
			message string
		)

		sel = bson.M{"devices.JWT": token, "wallets.walletIndex": sw.WalletIndex}
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
				"outs":    0,
			})
			return
		}
		token := authHeader[1]

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
		var spOuts []SpendableOutputs

		switch currencyID {
		case currencies.Testnet:
			userTxs := store.TxRecord{}
			query := bson.M{"userid": user.UserID, "transactions.txaddress": address}
			restClient.userStore.FindUserTxs(query, &userTxs)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": msgErrNoTransactionAddress,
					"outs":    0,
				})
				return
			}

			for _, tx := range userTxs.Transactions {
				if tx.TxAddress == address && tx.TxStatus == "incoming in block" {
					spOuts = append(spOuts, SpendableOutputs{
						TxID:        tx.TxID,
						TxOutID:     tx.TxOutID,
						TxOutAmount: int(tx.TxOutAmount * float64(100000000)),
						TxOutScript: tx.TxOutScript,
					})
				}
			}

		// case currencies.Bitcoin:
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

func (restClient *RestClient) sendRawTransaction() gin.HandlerFunc {
	return func(c *gin.Context) {

		restClient.log.Infof("btc.Cert=%s\n", btc.Cert)

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

func (restClient *RestClient) getWalletVerboseOld() gin.HandlerFunc {
	return func(c *gin.Context) {
		var av []AddressVerbose
		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) < 2 {
			restClient.log.Errorf("getWalletVerbose: wrong Authorization header len\t[addr=%s]", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
				"wallet":  av,
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
				"wallet":  av,
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

func (restClient *RestClient) getWalletVerbose() gin.HandlerFunc {
	return func(c *gin.Context) {
		var wv []WalletVerbose
		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) < 2 {
			restClient.log.Errorf("getAllWalletsVerbose: wrong Authorization header len\t[addr=%s]", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
				"wallet":  wv,
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
			if tx.TxStatus != "spend in mempool" && tx.TxStatus != "spend in block" {
				unspendTxs = append(unspendTxs, tx)
			}
		}

		var av []AddressVerbose
		var spOuts []SpendableOutputs
		var balance int

		for _, wallet := range user.Wallets {
			if wallet.WalletIndex == walletIndex {
				for _, address := range wallet.Adresses {

					for _, tx := range unspendTxs {
						if tx.TxAddress == address.Address {
							balance += int(tx.TxOutAmount * float64(100000000))

							spOuts = append(spOuts, SpendableOutputs{
								TxID:              tx.TxID,
								TxOutID:           tx.TxOutID,
								TxOutAmount:       int(tx.TxOutAmount * float64(100000000)),
								TxOutScript:       tx.TxOutScript,
								TxStatus:          tx.TxStatus,
								AddressIndex:      address.AddressIndex,
								StockExchangeRate: []StockExchangeRate{}, // from db
							})
						}
					}

					av = append(av, AddressVerbose{
						Address:       address.Address,
						AddressIndex:  address.AddressIndex,
						Amount:        balance,
						SpendableOuts: spOuts,
					})
					spOuts = []SpendableOutputs{}
					balance = 0
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
	Address       string             `json:"address"`
	AddressIndex  int                `json:"addressindex"`
	Amount        int                `json:"amount"`
	SpendableOuts []SpendableOutputs `json:"spendableoutputs"`
}
type SpendableOutputs struct {
	TxID              string              `json:"txid"`
	TxOutID           int                 `json:"txoutid"`
	TxOutAmount       int                 `json:"txoutamount"`
	TxOutScript       string              `json:"txoutscript"`
	AddressIndex      int                 `json:"addressindex"`
	TxStatus          string              `json:"txstatus"`
	StockExchangeRate []StockExchangeRate `json:"stockexchangerate"`
}
type StockExchangeRate struct {
	ExchangeName   string `json:"exchangename"`
	FiatEquivalent int    `json:"fiatequivalent"`
	TotalAmount    int    `json:"totalamount"`
}

func (restClient *RestClient) getAllWalletsVerbose() gin.HandlerFunc {
	return func(c *gin.Context) {
		var wv []WalletVerbose
		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) < 2 {
			restClient.log.Errorf("getAllWalletsVerbose: wrong Authorization header len\t[addr=%s]", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
				"wallets": wv,
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
			c.JSON(code, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrUserNotFound,
				"wallets": wv,
			})
			return
		} else {
			code = http.StatusOK
			message = http.StatusText(http.StatusOK)
		}

		userTxs := store.TxRecord{}
		query = bson.M{"userid": user.UserID}
		if err := restClient.userStore.FindUserTxs(query, &userTxs); err != nil {
			restClient.log.Errorf("getAllWalletsVerbose: restClient.userStore.FindUser: user %s\t[addr=%s]", err.Error(), c.Request.RemoteAddr)
		}

		var unspendTxs []store.MultyTX
		for _, tx := range userTxs.Transactions {
			if tx.TxStatus != "spend in mempool" && tx.TxStatus != "spend in block" {
				unspendTxs = append(unspendTxs, tx)
			}
		}

		var av []AddressVerbose
		var spOuts []SpendableOutputs
		var balance int

		for _, wallet := range user.Wallets {
			if wallet.Status == store.WalletStatusOK {

				for _, address := range wallet.Adresses {

					for _, tx := range unspendTxs {
						if tx.TxAddress == address.Address {
							balance += int(tx.TxOutAmount * float64(100000000))

							spOuts = append(spOuts, SpendableOutputs{
								TxID:              tx.TxID,
								TxOutID:           tx.TxOutID,
								TxOutAmount:       int(tx.TxOutAmount * float64(100000000)),
								TxOutScript:       tx.TxOutScript,
								TxStatus:          tx.TxStatus,
								AddressIndex:      address.AddressIndex,
								StockExchangeRate: []StockExchangeRate{}, // from db
							})
						}
					}

					av = append(av, AddressVerbose{
						Address:       address.Address,
						AddressIndex:  address.AddressIndex,
						Amount:        balance,
						SpendableOuts: spOuts,
					})
					spOuts = []SpendableOutputs{}
					balance = 0
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
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
			"wallets": wv,
		})
	}
}

func (restClient *RestClient) getWalletTransactionsHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		var walletTxs []store.MultyTX
		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) < 2 {
			restClient.log.Errorf("getAllWalletsVerbose: wrong Authorization header len\t[addr=%s]", c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": msgErrHeaderError,
				"history": walletTxs,
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
				"message": msgErrDecodeWalletIndexErr,
				"history": walletTxs,
			})
			return
		}

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
			if tx.WalletIndex == walletIndex {
				walletTxs = append(walletTxs, tx)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": http.StatusText(http.StatusOK),
			"history": walletTxs,
		})
	}
}
