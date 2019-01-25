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

	"github.com/Multy-io/Multy-back/exchanger"

	// exchanger "github.com/Multy-io/Multy-back-exchange-service"
	"github.com/Multy-io/Multy-back/btc"
	"github.com/Multy-io/Multy-back/client"
	"github.com/Multy-io/Multy-back/currencies"
	"github.com/Multy-io/Multy-back/eth"
	btcpb "github.com/Multy-io/Multy-back/ns-btc-protobuf"
	ethpb "github.com/Multy-io/Multy-back/ns-eth-protobuf"
	"github.com/Multy-io/Multy-back/store"
	"github.com/gin-gonic/gin"
	"github.com/jekabolt/slf"
)

var (
	log = slf.WithContext("multy-back").WithCaller(slf.CallerShort)
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

	ExchangerFactory *exchanger.FactoryExchanger
}

// Init initializes Multy instance
func Init(conf *Configuration) (*Multy, error) {
	multy := &Multy{
		config:           conf,
		ExchangerFactory: &exchanger.FactoryExchanger{},
	}
	// DB initialization
	userStore, err := store.InitUserStore(conf.Database)
	if err != nil {
		return nil, fmt.Errorf("DB initialization: %s on port %s", err.Error(), conf.Database.Address)
	}
	multy.userStore = userStore
	log.Infof("UserStore initialization done on %s √", conf.Database)

	//BTC
	btcCli, err := btc.InitHandlers(&conf.Database, conf.SupportedNodes, conf.NSQAddress)
	if err != nil {
		return nil, fmt.Errorf("Init: btc.InitHandlers: %s", err.Error())
	}
	btcVer, err := btcCli.CliMain.ServiceInfo(context.Background(), &btcpb.Empty{})
	multy.BTC = btcCli
	log.Infof(" BTC initialization done on %v √", btcVer)

	// ETH
	ethCli, err := eth.InitHandlers(&conf.Database, conf.SupportedNodes, conf.NSQAddress)
	if err != nil {
		return nil, fmt.Errorf("Init: btc.InitHandlers: %s", err.Error())
	}
	ethVer, err := ethCli.CliMain.ServiceInfo(context.Background(), &ethpb.Empty{})
	multy.ETH = ethCli
	log.Infof(" ETH initialization done on %v √", ethVer)

	//users data set
	sv, err := multy.SetUserData(multy.userStore, conf.SupportedNodes)
	if err != nil {
		return nil, fmt.Errorf("Init: multy.SetUserData: %s", err.Error())
	}
	log.Infof("Users data  initialization done √")

	log.Debugf("Server versions %v", sv)

	// REST handlers
	if err = multy.initHttpRoutes(conf); err != nil {
		return nil, fmt.Errorf("Router initialization: %s", err.Error())
	}

	err = multy.ConfigureExchangers()
	if err != nil {
		return nil, fmt.Errorf("Failed to configure exchangers, [%s]", err.Error())
	}

	return multy, nil
}

func (m *Multy) ConfigureExchangers() error {
	m.ExchangerFactory.SetExchangersConfig(m.config.Exchangers)

	for _, exchangerConfig := range m.config.Exchangers {
		if exchangerConfig.IsActive {
			// exchanger warm-up
			_, err := m.ExchangerFactory.GetExchanger(exchangerConfig.Name)
			if err != nil {
				log.Errorf("Failed to initialize [%s] exchanger, [%s]", exchangerConfig.Name, err.Error())
			}

			log.Infof("Exchanger: name [%s] init completed", exchangerConfig.Name)
		}
	}

	return nil
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
		if len(usersContracts) == 0 {
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
func (multy *Multy) initHttpRoutes(conf *Configuration) error {
	router := gin.Default()
	multy.route = router
	gin.SetMode(gin.DebugMode)

	f, err := os.OpenFile("../currencies/erc20tokens.json", os.O_RDONLY, os.FileMode(0644))
	// f, err := os.OpenFile("/currencies/erc20tokens.json")
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
		multy.userStore,
		router,
		conf.DonationAddresses,
		multy.BTC,
		multy.ETH,
		conf.MultyVerison,
		conf.Secretkey,
		conf.MobileVersions,
		tokenList,
		conf.BrowserDefault,
		multy.ExchangerFactory,
	)
	if err != nil {
		return err
	}
	multy.restClient = restClient

	// socketIO server initialization. server -> mobile client
	socketIORoute := router.Group("/socketio")
	socketIOPool, err := client.SetSocketIOHandlers(multy.restClient, multy.BTC, multy.ETH, socketIORoute, conf.SocketioAddr, conf.NSQAddress, multy.userStore)
	if err != nil {
		return err
	}
	multy.clientPool = socketIOPool
	multy.ETH.WsServer = multy.clientPool.Server
	multy.BTC.WsServer = multy.clientPool.Server

	firebaseClient, err := client.InitFirebaseConn(&conf.Firebase, multy.route, conf.NSQAddress)
	if err != nil {
		return err
	}
	multy.firebaseClient = firebaseClient

	return nil
}

// Run runs service
func (multy *Multy) Run() error {
	log.Debugf("Listening Rest address: %d", multy.config.RestAddress)
	log.Debugf("Listening Socketio address %v", multy.config.SocketioAddr)
	log.Info("Running server")
	multy.route.Run(multy.config.RestAddress)
	return nil
}

func fetchCoinType(coinTypes []store.CoinType, currencyID, networkID int) (*store.CoinType, error) {
	for _, ct := range coinTypes {
		if ct.СurrencyID == currencyID && ct.NetworkID == networkID {
			return &ct, nil
		}
	}
	return nil, fmt.Errorf("fetchCoinType: no such coin in config")
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
