/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package multyback

import (
	"fmt"
	"io/ioutil"

	"github.com/Appscrunch/Multy-back/btc"
	"github.com/Appscrunch/Multy-back/client"
	"github.com/Appscrunch/Multy-back/eth"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
)

var (
	pwdCurr = "multy-back"
	log     = slf.WithContext(pwdCurr)
)

const (
	defaultServerAddress = "0.0.0.0:6678"
	version              = "v1"
)

// Multy is a main struct of service
type Multy struct {
	config     *Configuration
	clientPool *client.SocketIOConnectedPool
	route      *gin.Engine

	userStore store.UserStore

	btcClient      *rpcclient.Client
	restClient     *client.RestClient
	firebaseClient *client.FirebaseClient

	ethClient *ethereum.Client
}

// Init initializes Multy instance
func Init(conf *Configuration) (*Multy, error) {
	multy := &Multy{
		config: conf,
	}

	userStore, err := store.InitUserStore(conf.Database)
	if err != nil {
		return nil, fmt.Errorf("DB initialization: %s", err.Error())
	}
	multy.userStore = userStore

	btcClient, err := btc.InitHandlers(getCertificate(conf.BTCSertificate), &conf.Database, conf.NSQAddress, conf.BTCNodeAddress)
	if err != nil {
		return nil, fmt.Errorf("Blockchain api initialization: %s", err.Error())
	}
	log.Debug("BTC client initialization done")
	multy.btcClient = btcClient

	multy.ethClient = ethereum.NewClient(&conf.Etherium, userStore)
	log.Debug("ETH client initialization done")

	if err = multy.initRoutes(conf); err != nil {
		return nil, fmt.Errorf("Router initialization: %s", err.Error())
	}

	return multy, nil
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

func (multy *Multy) initRoutes(conf *Configuration) error {
	router := gin.Default()
	multy.route = router

	gin.SetMode(gin.DebugMode)

	socketIORoute := router.Group("/socketio")
	socketIOPool, err := client.SetSocketIOHandlers(socketIORoute, conf.SocketioAddr, conf.NSQAddress, multy.userStore)
	if err != nil {
		return err
	}
	multy.clientPool = socketIOPool

	restClient, err := client.SetRestHandlers(
		multy.ethClient,
		multy.userStore,
		conf.BTCAPITest,
		conf.BTCAPIMain,
		router,
		multy.btcClient,
		conf.BTCNodeAddress)
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
