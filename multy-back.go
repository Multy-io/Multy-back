/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package multyback

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	// exchanger "github.com/Multy-io/Multy-back-exchange-service"
	btcpb "github.com/Multy-io/Multy-BTC-node-service/node-streamer"
	ethpb "github.com/Multy-io/Multy-ETH-node-service/node-streamer"
	"github.com/Multy-io/Multy-back/btc"
	"github.com/Multy-io/Multy-back/client"
	"github.com/Multy-io/Multy-back/currencies"
	"github.com/Multy-io/Multy-back/eth"
	"github.com/Multy-io/Multy-back/store"
	"github.com/gin-gonic/gin"
	"github.com/jekabolt/slf"
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
	ETH *eth.ETHConn
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

	multy.setupConfiguredNodes()

	//users data set
	sv, err := multy.SetUserData(multy.userStore, multy.config.SupportedNodes)
	if err != nil {
		return nil, fmt.Errorf("Init: multy.SetUserData: %s ", err.Error())
	}
	log.Infof("Users data initialization done √")

	log.Debugf("Server versions %v", sv)

	// REST handlers
	if err = multy.initHttpRoutes(); err != nil {
		return nil, fmt.Errorf("Router initialization: %s ", err.Error())
	}

	if err = multy.registerWebSocketEvents(); err != nil {
		return nil, fmt.Errorf("Failed to register websocket events: [%s] ", err.Error())
	}

	return multy, nil
}


func (m *Multy) registerWebSocketEvents() error {
	// socketIO server initialization. server -> mobile client
	socketIORoute := m.route.Group("/socketio")
	socketIOPool, err := client.SetSocketIOHandlers(m.restClient, m.BTC, m.ETH, socketIORoute, m.config.SocketioAddr,
		m.config.NSQAddress, m.userStore)
	if err != nil {
		return err
	}
	m.clientPool = socketIOPool

	for _, nodeConfig := range m.config.SupportedNodes {
		switch nodeConfig.СurrencyID {
		case currencies.Bitcoin:
			m.BTC.WsServer = m.clientPool.Server
			break
		case currencies.Ether:
			m.ETH.WsServer = m.clientPool.Server
			break
		}
	}

	return nil
}


func (m *Multy) setupConfiguredNodes() {
	for _, nodeConfig := range m.config.SupportedNodes {
		switch nodeConfig.СurrencyID {
		case currencies.Bitcoin:
			//BTC
			btcCli, err := btc.InitHandlers(&m.config.Database, m.config.SupportedNodes, m.config.NSQAddress)
			if err != nil {
				log.Errorf("Init: btc.InitHandlers: %s", err.Error())
				break
			}
			btcVer, err := btcCli.CliMain.ServiceInfo(context.Background(), &btcpb.Empty{})
			m.BTC = btcCli
			log.Infof(" BTC initialization done on %v √", btcVer)
			break
		case currencies.Ether:
			// ETH
			ethCli, err := eth.InitHandlers(&m.config.Database, m.config.SupportedNodes, m.config.NSQAddress)
			if err != nil {
				log.Errorf("Init: btc.InitHandlers: [%s] ", err.Error())
				break
			}
			ethVer, err := ethCli.CliMain.ServiceInfo(context.Background(), &ethpb.Empty{})
			m.ETH = ethCli
			log.Infof(" ETH initialization done on %v √", ethVer)
			break
		}
	}
}


// SetUserData make initial userdata to node service
func (m *Multy) SetUserData(userStore store.UserStore, ct []store.CoinType) ([]store.ServiceInfo, error) {
	servicesInfo := []store.ServiceInfo{}
	for _, conCred := range ct {
		usersData, err := userStore.FindUserDataChain(conCred.СurrencyID, conCred.NetworkID)
		if err != nil {
			return servicesInfo, fmt.Errorf("SetUserData: userStore.FindUserDataChain: curID :%d netID :%d err =%s", conCred.СurrencyID, conCred.NetworkID, err.Error())
		}
		if len(usersData) == 0 {
			log.Infof("Empty userdata")
		}

		usersContracts, err := userStore.FindUsersContractsChain(conCred.СurrencyID, conCred.NetworkID)
		if err != nil {
			return servicesInfo, fmt.Errorf("SetUserData: userStore.FindUsersContractsChain: curID :%d netID :%d err =%s", conCred.СurrencyID, conCred.NetworkID, err.Error())
		}
		if len(usersData) == 0 {
			log.Infof("Empty userscontracts")
		}

		switch conCred.СurrencyID {
		case currencies.Bitcoin:
			var cli btcpb.NodeCommunicationsClient
			switch conCred.NetworkID {
			case currencies.Main:
				cli = m.BTC.CliMain
			case currencies.Test:
				cli = m.BTC.CliTest
			default:
				log.Errorf("setGRPCHandlers: wrong networkID:")
			}

			//TODO: Re State
			go m.restoreState(conCred, cli)

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
			resp, err := cli.EventInitialAdd(context.Background(), &genUd)
			if err != nil {
				return servicesInfo, fmt.Errorf("SetUserData:  btcCli.CliMain.EventInitialAdd: curID :%d netID :%d err =%s", conCred.СurrencyID, conCred.NetworkID, err.Error())
			}
			log.Debugf("Btc EventInitialAdd: resp: %s", resp.Message)

			sv, err := cli.ServiceInfo(context.Background(), &btcpb.Empty{})
			if err != nil {
				return servicesInfo, fmt.Errorf("SetUserData:  cli.ServiceInfo: curID :%d netID :%d err =%s", conCred.СurrencyID, conCred.NetworkID, err.Error())
			}
			servicesInfo = append(servicesInfo, store.ServiceInfo{
				Branch:    sv.Branch,
				Commit:    sv.Commit,
				Buildtime: sv.Buildtime,
				Lasttag:   sv.Lasttag,
			})

		case currencies.Ether:
			var cli ethpb.NodeCommunicationsClient
			switch conCred.NetworkID {
			case currencies.ETHMain:
				cli = m.ETH.CliMain
			case currencies.ETHTest:
				cli = m.ETH.CliTest
			default:
				log.Errorf("setGRPCHandlers: wrong networkID:")
			}

			//TODO: Restore state
			go m.restoreState(conCred, cli)

			genUd := ethpb.UsersData{
				Map:            map[string]*ethpb.AddressExtended{},
				UsersContracts: usersContracts,
			}

			for address, ex := range usersData {
				if ex.WalletIndex == -1 {
					genUd.Map[address] = &ethpb.AddressExtended{
						UserID:       "imported",
						WalletIndex:  int32(ex.WalletIndex),
						AddressIndex: int32(ex.AddressIndex),
					}
				} else {
					genUd.Map[address] = &ethpb.AddressExtended{
						UserID:       ex.UserID,
						WalletIndex:  int32(ex.WalletIndex),
						AddressIndex: int32(ex.AddressIndex),
					}
				}
			}
			resp, err := cli.EventInitialAdd(context.Background(), &genUd)
			if err != nil {
				return servicesInfo, fmt.Errorf("SetUserData: Ether.EventInitialAdd: curID :%d netID :%d err =%s", conCred.СurrencyID, conCred.NetworkID, err.Error())
			}
			log.Debugf("Ether cli.EventInitialAdd: resp: %s", resp.Message)

			sv, err := cli.ServiceInfo(context.Background(), &ethpb.Empty{})
			if err != nil {
				return servicesInfo, fmt.Errorf("SetUserData:  cli.ServiceInfo: curID :%d netID :%d err =%s", conCred.СurrencyID, conCred.NetworkID, err.Error())
			}
			servicesInfo = append(servicesInfo, store.ServiceInfo{
				Branch:    sv.Branch,
				Commit:    sv.Commit,
				Buildtime: sv.Buildtime,
				Lasttag:   sv.Lasttag,
			})
		}
	}

	return nil, nil
}

// initRoutes initialize client communication services
// - http
// - socketio
// - firebase
func (m *Multy) initHttpRoutes() error {
	router := gin.Default()
	m.route = router
	gin.SetMode(gin.DebugMode)

	f, err := os.OpenFile("../currencies/erc20tokens.json", os.O_RDONLY, os.FileMode(0644))
	if err != nil {
		return err
	}

	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	tokenList := store.VerifiedTokenList{}
	_ = json.Unmarshal(bs, &tokenList)

	restClient, err := client.SetRestHandlers(
		m.userStore,
		router,
		m.config.DonationAddresses,
		m.BTC,
		m.ETH,
		m.config.MultyVerison,
		m.config.Secretkey,
		m.config.MobileVersions,
		tokenList,
		m.config.BrowserDefault,
	)
	if err != nil {
		return err
	}
	m.restClient = restClient
	firebaseClient, err := client.InitFirebaseConn(&m.config.Firebase, m.route, m.config.NSQAddress)
	if err != nil {
		return err
	}
	m.firebaseClient = firebaseClient

	return nil
}


// Run runs service
func (m *Multy) Run() {
	log.Info("Running server")
	m.route.Run(m.config.RestAddress)
}


func (m *Multy) restoreState(coinType store.CoinType, ncClient interface{}) {
	if coinType.AccuracyRange > 0 {
		var height int64
		height, err := m.userStore.FetchLastSyncBlockState(coinType.NetworkID, coinType.СurrencyID)
		if err != nil {
			log.Warnf("SetUserData:  btcCli.CliMain.cli.FetchLastSyncBlockState: curID :%d netID :%d err =%s", coinType.СurrencyID, coinType.NetworkID, err.Error())
		}
		log.Debugf("Last height recorded %v last trusted block %v netid:%v curid", height, height-int64(coinType.AccuracyRange), coinType.NetworkID, coinType.СurrencyID)
		height = height - int64(coinType.AccuracyRange)

		if height > 0 {
			switch coinType.СurrencyID {
			case currencies.Ether:
				var cli ethpb.NodeCommunicationsClient
				cli = ncClient.(ethpb.NodeCommunicationsClient)
				rp, err := cli.SyncState(context.Background(), &ethpb.BlockHeight{Height: height})
				if err != nil {
					log.Errorf("SetUserData:  restoreState:cli.SyncState: curID :%d netID :%d err =%s", coinType.СurrencyID, coinType.NetworkID, err.Error())
				}
				if strings.Contains("err:", rp.GetMessage()) {
					log.Errorf("SetUserData:  Contains err : curID :%d netID :%d err =%s", coinType.СurrencyID, coinType.NetworkID, err.Error())
				}
				log.Debugf("Restored state processing on curid =%v netid =%v", coinType.СurrencyID, coinType.NetworkID)
			case currencies.Bitcoin:
				var cli btcpb.NodeCommunicationsClient
				cli = ncClient.(btcpb.NodeCommunicationsClient)
				rp, err := cli.SyncState(context.Background(), &btcpb.BlockHeight{Height: height})
				if err != nil {
					log.Errorf("SetUserData:  btcCli.CliMain.cli.SyncState: curID :%d netID :%d err =%s", coinType.СurrencyID, coinType.NetworkID, err.Error())
				}
				if strings.Contains("err:", rp.GetMessage()) {
					log.Errorf("SetUserData:  Contains err : curID :%d netID :%d err =%s", coinType.СurrencyID, coinType.NetworkID, err.Error())
				}
				log.Debugf("Restored state processing on curid =%v netid =%v", coinType.СurrencyID, coinType.NetworkID)
			default:
				log.Errorf("setGRPCHandlers: wrong СurrencyID")
			}
		} else {
			log.Errorf("FetchLastSyncBlockState : h > 0")
		}
	} else {
		log.Warnf("Restore last state is disabled for c = %v n = %v ", coinType.СurrencyID, coinType.NetworkID)
	}

}
