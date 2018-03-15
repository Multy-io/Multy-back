/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package multyback

import (
	"fmt"
	"time"

	"github.com/Appscrunch/Multy-back/btc"
	"github.com/Appscrunch/Multy-back/client"
	"github.com/Appscrunch/Multy-back/currencies"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
	"github.com/gin-gonic/gin"
	"github.com/graarh/golang-socketio"
)

var (
	pwdCurr = "multy-back"
	log     = slf.WithContext(pwdCurr)
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

	WsBtcTestnetCli *gosocketio.Client
	WsBtcMainnetCli *gosocketio.Client
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

	mainBtcCli, testBtcCli, err := btc.InitHandlers(&conf.Database, conf.SupportedNodes, conf.NSQAddress)
	if err != nil {
		return nil, fmt.Errorf("Init: btc.InitHandlers: %s", err.Error())
	}

	multy.WsBtcTestnetCli = testBtcCli
	multy.WsBtcMainnetCli = mainBtcCli
	log.Infof("WsBtcTestnetCli WsBtcMainnetCli initialization done √")

	go func() {
		for {
			time.Sleep(2 * time.Second)
			log.Infof("Is alife = %b \n", testBtcCli.IsAlive())
			testBtcCli.Emit("kek", "kek")
		}
	}()

	// initial add user data to node client test
	btcTestConf, err := fethCoinType(conf.SupportedNodes, currencies.Bitcoin, currencies.Test)
	if err != nil {
		return nil, fmt.Errorf("Init: InitWsNodeConn:  Test %v on port %s", conf.SupportedNodes, err.Error())
	}
	SetUserData(multy.WsBtcTestnetCli, btcTestConf, multy.userStore)
	log.Infof("WsBtcTestnetCli SetUserData Test initialization done √")

	// initial add user data to node client
	btcMainConf, err := fethCoinType(conf.SupportedNodes, currencies.Bitcoin, currencies.Main)
	if err != nil {
		return nil, fmt.Errorf("Init: InitWsNodeConn:  Main %v on port %s", conf.SupportedNodes, err.Error())
	}
	SetUserData(multy.WsBtcTestnetCli, btcMainConf, multy.userStore)
	log.Infof("WsBtcTestnetCli SetUserData Main initialization done √")

	if err = multy.initRoutes(conf); err != nil {
		return nil, fmt.Errorf("Router initialization: %s", err.Error())
	}

	return multy, nil
}

// SetUserData make initial userdata to node service
func SetUserData(wsCli *gosocketio.Client, ct *store.CoinType, userStore store.UserStore) error {

	// TODO: fix initial add
	UsersData, err := userStore.FindUserDataChain(ct.СurrencyID, ct.NetworkID)
	if err != nil {
		return fmt.Errorf("InitWsNodeConn: userStore.FindUserDataChain: curID :%d netID :%d err =%s", ct.СurrencyID, ct.NetworkID, err.Error())
	}
	fmt.Println(UsersData)

	// if len(UsersData) == 0 {
	// 	return nil, fmt.Errorf("InitWsNodeConn: empty UserData curID :%d netID :%d", ct.СurrencyID, ct.NetworkID)
	// }

	// TODO: fix initial add
	// err = wsCli.Emit(EventInitialAdd, UsersData)
	// if err != nil {
	// 	return nil, fmt.Errorf("InitWsNodeConn: wsBtcTest.Emit :%s SocketPort :%d err =%s", ct.SocketURL, ct.SocketPort, err.Error())
	// }

	return nil
}

// initRoutes initialize client communication services
// - http
// - socketio
// - firebase
func (multy *Multy) initRoutes(conf *Configuration) error {
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
		multy.WsBtcTestnetCli,
		multy.WsBtcMainnetCli,
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
