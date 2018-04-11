/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package multyback

import (
	"context"
	"fmt"

	"github.com/Appscrunch/Multy-back/btc"
	"github.com/Appscrunch/Multy-back/client"
	"github.com/Appscrunch/Multy-back/currencies"
	btcpb "github.com/Appscrunch/Multy-back/node-streamer/btc"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
	"github.com/gin-gonic/gin"
)

var (
	log = slf.WithContext("multy-back")
)

const (
	defaultServerAddress = "0.0.0.0:6678"
	version              = "v1"
)

const (
	EventConnection    = "connection"
	EventInitialAdd    = "allUsers"
	EventResyncAddress = "resync"
	EventSendRawTx     = "sendRaw"
	EventAddNewAddress = "newUser"
	Room               = "node"
)

// Multy is a main struct of service
type Multy struct {
	config     *Configuration
	clientPool *client.SocketIOConnectedPool
	route      *gin.Engine

	userStore store.UserStore

	restClient     *client.RestClient
	firebaseClient *client.FirebaseClient

	BTC *btc.BTCConn
}

// Init initializes Multy instance
func Init(conf *Configuration) (*Multy, error) {
	multy := &Multy{
		config: conf,
	}

	// DB initialization
	userStore, err := store.InitUserStore(conf.Database)
	if err != nil {
		return nil, fmt.Errorf("DB initialization: %s on port %s", err.Error(), conf.Database.Address)
	}
	multy.userStore = userStore
	log.Infof("UserStore initialization done on %s √", conf.Database)

	//TODO: Mempool Data delete
	// multy.userStore.DeleteMempool()
	log.Infof("Mempool Data delete √")

	btcCli, err := btc.InitHandlers(&conf.Database, conf.SupportedNodes, conf.NSQAddress)
	if err != nil {
		return nil, fmt.Errorf("Init: btc.InitHandlers: %s", err.Error())
	}

	multy.BTC = btcCli

	log.Infof(" BTC initialization done √")

	err = SetUserData(multy.BTC, multy.userStore, conf.SupportedNodes)

	log.Infof("BTC Users data  initialization done √")

	if err = multy.initHttpRoutes(conf); err != nil {
		return nil, fmt.Errorf("Router initialization: %s", err.Error())
	}
	return multy, nil
}

// SetUserData make initial userdata to node service
func SetUserData(btcCli *btc.BTCConn, userStore store.UserStore, ct []store.CoinType) error {
	for _, conCred := range ct {
		switch conCred.СurrencyID {
		case currencies.Bitcoin:
			usersData, err := userStore.FindUserDataChain(conCred.СurrencyID, conCred.NetworkID)
			if err != nil {
				return fmt.Errorf("SetUserData: userStore.FindUserDataChain: curID :%d netID :%d err =%s", conCred.СurrencyID, conCred.NetworkID, err.Error())
			}
			if len(usersData) == 0 {
				log.Infof("Empty userdata")
			}
			log.Errorf("\n\n\n usersData cur%d subnet%d \n %v \n\n", conCred.СurrencyID, conCred.NetworkID, usersData)

			switch conCred.NetworkID {
			case currencies.Main:
				genUd := btcpb.UsersData{
					Map: map[string]*btcpb.AddressExtended{},
				}
				for address, ex := range usersData {
					genUd.Map[address] = &btcpb.AddressExtended{
						UserID:       ex.UserID,
						WalletIndex:  int32(ex.WalletIndex),
						AddressIndex: int32(ex.AddressIndex),
					}
				}
				resp, err := btcCli.CliMain.EventInitialAdd(context.Background(), &genUd)
				if err != nil {
					return fmt.Errorf("SetUserData:  btcCli.CliMain.EventInitialAdd: curID :%d netID :%d err =%s", conCred.СurrencyID, conCred.NetworkID, err.Error())
				}
				log.Debugf("btcCli.CliMain.EventInitialAdd: resp: %s", resp.Message)
			case currencies.Test:
				genUd := btcpb.UsersData{
					Map: map[string]*btcpb.AddressExtended{},
				}
				for address, ex := range usersData {
					genUd.Map[address] = &btcpb.AddressExtended{
						UserID:       ex.UserID,
						WalletIndex:  int32(ex.WalletIndex),
						AddressIndex: int32(ex.AddressIndex),
					}
				}
				resp, err := btcCli.CliTest.EventInitialAdd(context.Background(), &genUd)
				if err != nil {
					return fmt.Errorf("SetUserData: btcCli.CliTest.EventInitialAdd: curID :%d netID :%d err =%s", conCred.СurrencyID, conCred.NetworkID, err.Error())
				}
				log.Debugf("btcCli.CliMain.EventInitialAdd: resp: %s", resp.Message)
			}

		case currencies.Ether:
			//soon
		}
	}

	if btcCli.CliTest != nil {
		log.Infof("SetUserData: users data initialization done Test √")
	} else {
		log.Infof("SetUserData: users data initialization Test FAILED")
	}

	if btcCli.CliMain == nil {
		log.Infof("SetUserData users data initialization done Main √")
	} else {
		log.Infof("SetUserData: users data initialization Main FAILED")
	}

	return nil
}

// initRoutes initialize client communication services
// - http
// - socketio
// - firebase
func (multy *Multy) initHttpRoutes(conf *Configuration) error {
	router := gin.Default()
	multy.route = router

	gin.SetMode(gin.DebugMode)

	// socketIO server initialization. server -> mobile client
	socketIORoute := router.Group("/socketio")
	socketIOPool, err := client.SetSocketIOHandlers(socketIORoute, conf.SocketioAddr, conf.NSQAddress, multy.userStore)
	if err != nil {
		return err
	}
	multy.clientPool = socketIOPool

	restClient, err := client.SetRestHandlers(
		multy.userStore,
		router,
		conf.DonationAddresses,
		multy.BTC,
		store.ServerConfig,
		conf.MultyVerison,
	)
	if err != nil {
		return err
	}
	multy.restClient = restClient

	firebaseClient, err := client.InitFirebaseConn(&conf.Firebase, multy.route, conf.NSQAddress)
	if err != nil {
		return err
	}
	multy.firebaseClient = firebaseClient

	return nil
}

// Run runs service
func (multy *Multy) Run() error {
	log.Info("Running server")
	multy.route.Run(multy.config.RestAddress)
	return nil
}

func fethCoinType(coinTypes []store.CoinType, currencyID, networkID int) (*store.CoinType, error) {
	for _, ct := range coinTypes {
		if ct.СurrencyID == currencyID && ct.NetworkID == networkID {
			return &ct, nil
		}
	}
	return nil, fmt.Errorf("fethCoinType: no such coin in config")
}
