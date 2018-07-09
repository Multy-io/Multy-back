/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package node

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"sync"

	"github.com/KristinaEtc/slf"
	"github.com/Multy-io/Multy-BTC-node-service/btc"
	"github.com/Multy-io/Multy-BTC-node-service/streamer"
	pb "github.com/Multy-io/Multy-back/node-streamer/btc"
	"github.com/Multy-io/Multy-back/store"
	"github.com/blockcypher/gobcy"
	"google.golang.org/grpc"
)

var (
	log = slf.WithContext("NodeClient")
)

// Multy is a main struct of service

// NodeClient is a main struct of service
type NodeClient struct {
	Config     *Configuration
	Instance   *btc.Client
	GRPCserver *streamer.Server
	Clients    *sync.Map // address to userid
	// Clients sync.Map // address to userid
	BtcApi *gobcy.API
}

// Init initializes Multy instance
func Init(conf *Configuration) (*NodeClient, error) {
	cli := &NodeClient{
		Config: conf,
	}

	usersData := sync.Map{}

	usersData.Store("2MvPhdUf3cwaadRKsSgbQ2SXc83CPcBJezT", store.AddressExtended{
		UserID:       "kek",
		WalletIndex:  1,
		AddressIndex: 2,
	})

	api := gobcy.API{
		Token: conf.BTCAPI.Token,
		Coin:  conf.BTCAPI.Coin,
		Chain: conf.BTCAPI.Chain,
	}
	cli.BtcApi = &api
	log.Debug("btc api initialization done √")

	// initail initialization of clients data
	cli.Clients = &usersData
	log.Debug("Users data initialization done √")

	// init gRPC server
	lis, err := net.Listen("tcp", conf.GrpcPort)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err.Error())
	}

	btcClient, err := btc.NewClient(getCertificate(conf.BTCSertificate), conf.BTCNodeAddress, cli.Clients)
	if err != nil {
		return nil, fmt.Errorf("Blockchain api initialization: %s", err.Error())
	}
	log.Debug("BTC client initialization done √")
	cli.Instance = btcClient

	// Creates a new gRPC server
	s := grpc.NewServer()
	srv := streamer.Server{
		UsersData: cli.Clients,
		BtcAPI:    cli.BtcApi,
		M:         &sync.Mutex{},
		BtcCli:    btcClient,
		Info:      &conf.ServiceInfo,
	}

	pb.RegisterNodeCommuunicationsServer(s, &srv)
	go s.Serve(lis)
	log.Debug("NodeCommuunications Server initialization done √")

	return cli, nil
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
