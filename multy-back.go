package multyback

import (
	"fmt"
	"log"

	"github.com/Appscrunch/Multy-back/btc"
	"github.com/Appscrunch/Multy-back/client"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"

	socketio "github.com/googollee/go-socket.io"
)

const (
	defaultServerAddress = "0.0.0.0:7778"
	version              = "v1"
)

// Multy is a main struct of service
type Multy struct {
	config     *Configuration
	clientPool *client.SocketIOConnectedPool

	userStore store.UserStore
	route     *gin.Engine

	btcClientCh chan btc.BtcTransactionWithUserID

	socketIO   *socketio.Server
	btcClient  *rpcclient.Client
	restClient *client.RestClient
}

// Init initializes Multy instance
func Init(conf *Configuration) (*Multy, error) {
	multy := &Multy{
		config: conf,
	}

	userStore, err := store.InitUserStore(conf.DataStoreAddress)
	if err != nil {
		return nil, err
	}
	multy.userStore = userStore

	// TODO: add channels for communitation
	log.Println("[DEBUG] InitHandlers")
	btcClient, btcClientCh, err := btc.InitHandlers()
	if err != nil {
		return nil, fmt.Errorf("blockchain api initialization: %s", err.Error())
	}
	log.Println("[INFO] btc handlers initialization done")
	multy.btcClient = btcClient
	multy.btcClientCh = btcClientCh

	multy.clientPool = client.InitConnectedPool(btcClientCh)

	if err = multy.initRoute(conf.Address); err != nil {
		return nil, fmt.Errorf("router initialization: %s", err.Error())
	}

	log.Println("[DEBUG] init done")
	return multy, nil
}

func (multy *Multy) initRoute(address string) error {
	router := gin.Default()

	gin.SetMode(gin.DebugMode)

	socketIORoute := router.Group("/socketio")
	socketIOServer, err := client.SetSocketIOHandlers(socketIORoute, multy.btcClientCh, multy.clientPool)
	if err != nil {
		return err
	}

	multy.route = router
	multy.socketIO = socketIOServer

	restClient, err := client.SetRestHandlers(multy.userStore, client.BTCApiConf{
		Token: "6b4e9ead6afe4803bd1e2d22b24b52ad",
		Coin:  "btc",
		Chain: "test3",
	},
		client.BTCApiConf{
			Token: "6b4e9ead6afe4803bd1e2d22b24b52ad",
			Coin:  "btc",
			Chain: "main",
		},
		router, multy.btcClient)
	if err != nil {
		return err
	}
	multy.restClient = restClient

	return nil
}

// Run runs service
func (m *Multy) Run() error {
	if m.config.Address == "" {
		log.Println("[INFO] listening on default addres: ", defaultServerAddress)
	}
	m.config.Address = defaultServerAddress

	log.Println("[DEBUG] running server")
	m.route.Run(m.config.Address)
	return nil
}
