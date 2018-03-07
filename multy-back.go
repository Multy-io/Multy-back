/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package multyback

import (
	"fmt"
	"io/ioutil"

	"github.com/Appscrunch/Multy-back/client"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
	"github.com/gin-gonic/gin"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
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
	log.Infof("UserStore initialization done on %s", conf.Database)

	// support bitcoin testnet
	wsBtcTest, err := InitWsNodeConn(conf.SupportedNodes[0], multy.userStore)
	if err != nil {
		return nil, fmt.Errorf("Init: InitWsNodeConn: %v on port %s", conf.SupportedNodes, err.Error())
	}
	multy.WsBtcTestnetCli = wsBtcTest

	// support bitcoin mainnet
	// wsBtcMain, err := InitWsNodeConn(conf.SupportedNodes[1], multy.userStore)
	// if err != nil {
	// 	return nil, fmt.Errorf("Init: InitWsNodeConn: %v on port %s", conf.SupportedNodes, err.Error())
	// }
	// multy.WsBtcMainnetCli = wsBtcTest

	if err = multy.initRoutes(conf); err != nil {
		return nil, fmt.Errorf("Router initialization: %s", err.Error())
	}

	return multy, nil
}

func InitWsNodeConn(ct CoinType, userStore store.UserStore) (*gosocketio.Client, error) {
	UsersData, err := userStore.FindUserDataChain(ct.СurrencyID, ct.NetworkID)
	if err != nil {
		return nil, fmt.Errorf("InitWsNodeConn: userStore.FindUserDataChain: curID :%d netID :%d err =%s", ct.СurrencyID, ct.NetworkID, err.Error())
	}
	if len(UsersData) == 0 {
		return nil, fmt.Errorf("InitWsNodeConn: empty UserData curID :%d netID :%d err =%s", ct.СurrencyID, ct.NetworkID, err.Error())
	}
	wsCli, err := gosocketio.Dial(
		gosocketio.GetUrl(ct.SocketURL, ct.SocketPort, false),
		transport.GetDefaultWebsocketTransport())
	if err != nil {
		return nil, fmt.Errorf("InitWsNodeConn: gosocketio.Dial: SocketURL :%s SocketPort :%d err =%s", ct.SocketURL, ct.SocketPort, err.Error())
	}

	err = wsCli.Emit(EventInitialAdd, UsersData)
	if err != nil {
		return nil, fmt.Errorf("InitWsNodeConn: wsBtcTest.Emit :%s SocketPort :%d err =%s", ct.SocketURL, ct.SocketPort, err.Error())
	}
	return wsCli, nil
}

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

func getCertificate(certFile string) string {
	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		log.Errorf("get certificate: %s", err.Error())
		return ""
	}
	if len(cert) > 1 {
		return string(cert[:len(cert)-1])
	}
	log.Errorf("get certificate: empty certificate")
	return ""
}
