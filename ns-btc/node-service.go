/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package nsbtc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"sync"
	"time"

	pb "github.com/Multy-io/Multy-back/ns-btc-protobuf"
	"github.com/Multy-io/Multy-back/store"
	"github.com/blockcypher/gobcy"
	"github.com/jekabolt/slf"
	_ "github.com/jekabolt/slflog"
	"google.golang.org/grpc"
)

var (
	log = slf.WithContext("NodeClient")
)

// Multy is a main struct of service

// NodeClient is a main struct of service
type NodeClient struct {
	Config     *Configuration
	Instance   *Client
	GRPCserver *Server
	Clients    *sync.Map // address to userid
	BtcApi     *gobcy.API
}

// Init initializes Multy instance
func (nc *NodeClient) Init(conf *Configuration) (*NodeClient, error) {
	nc = &NodeClient{
		Config: conf,
	}

	usersData := sync.Map{}

	usersData.Store("address", store.AddressExtended{
		UserID:       "kek",
		WalletIndex:  1,
		AddressIndex: 2,
	})

	api := gobcy.API{
		Token: conf.BTCAPI.Token,
		Coin:  conf.BTCAPI.Coin,
		Chain: conf.BTCAPI.Chain,
	}
	nc.BtcApi = &api
	log.Debug("btc api initialization done √")

	// initail initialization of clients data
	nc.Clients = &usersData
	log.Debug("Users data initialization done √")

	// init gRPC server
	lis, err := net.Listen("tcp", conf.GrpcPort)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err.Error())
	}

	btcClient, err := NewClient(getCertificate(conf.BTCSertificate), conf.BTCNodeAddress, nc.Clients)
	if err != nil {
		return nil, fmt.Errorf("Blockchain api initialization: %s", err.Error())
	}
	log.Debug("BTC client initialization done √")
	nc.Instance = btcClient

	// Creates a new gRPC server
	s := grpc.NewServer()
	srv := Server{
		UsersData:  nc.Clients,
		BtcAPI:     nc.BtcApi,
		M:          &sync.Mutex{},
		BtcCli:     btcClient,
		Info:       &conf.ServiceInfo,
		GRPCserver: s,
		Listener:   lis,
		ReloadChan: make(chan struct{}),
	}

	nc.GRPCserver = &srv

	pb.RegisterNodeCommunicationsServer(s, &srv)

	go s.Serve(lis)

	go WathReload(srv.ReloadChan, nc)

	// go ContinuousResync(nc)

	go log.Debug("NodeCommuunications Server initialization done √")

	return nc, nil
}

func getCertificate(certFile string) []byte {
	cert, err := ioutil.ReadFile(certFile)
	cert = bytes.Trim(cert, "\x00")

	if err != nil {
		log.Errorf("get certificate: %s", err.Error())
		return []byte{}
	}
	if len(cert) > 1 {
		return cert
	}
	log.Errorf("get certificate: empty certificate")
	return []byte{}
}

func WathReload(reload chan struct{}, cli *NodeClient) {
	for {
		select {
		case _ = <-reload:
			ticker := time.NewTicker(1 * time.Second)

			err := cli.GRPCserver.Listener.Close()
			if err != nil {
				log.Errorf("WathReload:lis.Close %v", err.Error())
			}
			cli.GRPCserver.GRPCserver.Stop()
			log.Warnf("WathReload:Successfully stopped")
			for _ = range ticker.C {
				_, err := cli.Init(cli.Config)
				if err != nil {
					log.Errorf("WathReload:Init %v ", err)
					continue
				}

				log.Warnf("WathReload:Successfully reloaded")
				return
			}
		}
	}
}

// func ContinuousResync(cli *NodeClient) {
// 	log.Warnf("ContinuousResync")
// 	time.Sleep(time.Second * 30)
// 	var once sync.Once
// 	once.Do(func() {
// 		// ticker := time.NewTicker(30 * time.Second)
// 		resyncRange := cli.Config.ContinuousResyncCap
// 		resyncBlocks := []*chainhash.Hash{}
// 		// disable continuous resync
// 		if resyncRange == 0 {
// 			return
// 		}
// 		// for range ticker.C {
// 		// for range cli.Instance.Block {
// 		blockHash, _, err := cli.Instance.RPCClient.GetBestBlock()
// 		if err != nil {
// 			log.Errorf("ContinuousResync:GetBestBlock( %v", err.Error())
// 			// continue
// 		}

// 		resyncBlocks = append(resyncBlocks, blockHash)

// 		for index := 0; index < resyncRange; index++ {
// 			log.Warnf("ContinuousResync process %v ", index)
// 			block, err := cli.Instance.RPCClient.GetBlock(resyncBlocks[index])
// 			if err != nil {
// 				log.Errorf("ContinuousResync:GetBlock %v", err.Error())
// 				continue
// 			}
// 			prevBlock := block.Header.PrevBlock
// 			resyncBlocks = append(resyncBlocks, &prevBlock)

// 		}

// 		// cut block that we alreaby synced
// 		// resyncBlocks = resyncBlocks[1:]

// 		for i := len(resyncBlocks) - 1; i >= 0; i-- {
// 			cli.Instance.BlockTransactions(resyncBlocks[i])
// 		}
// 		log.Warnf("ContinuousResync:DONE")

// 		// }
// 	})

// }
