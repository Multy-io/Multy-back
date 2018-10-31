/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package node

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Multy-io/Multy-ETH-node-service/eth"
	pb "github.com/Multy-io/Multy-ETH-node-service/node-streamer"
	"github.com/Multy-io/Multy-ETH-node-service/streamer"
	"github.com/Multy-io/Multy-back/store"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jekabolt/slf"
	_ "github.com/jekabolt/slflog"
	"google.golang.org/grpc"
)

var log = slf.WithContext("NodeClient")

// NodeClient is a main struct of service
type NodeClient struct {
	Config      *Configuration
	Instance    *eth.Client
	GRPCserver  *streamer.Server
	Clients     *sync.Map // address to userid
	CliMultisig *eth.Multisig
}

// Init initializes Multy instance
func (nc *NodeClient) Init(conf *Configuration) (*NodeClient, error) {
	resyncUrl := FethResyncUrl(conf.NetworkID)
	conf.ResyncUrl = resyncUrl
	nc = &NodeClient{
		Config: conf,
	}

	var usersData sync.Map

	usersData.Store("address", store.AddressExtended{
		UserID:       "kek",
		WalletIndex:  1,
		AddressIndex: 2,
	})

	// initail initialization of clients data
	nc.Clients = &usersData

	//TODO: init contract clients
	multisig := eth.Multisig{
		FactoryAddress: conf.MultisigFactory,
		UsersContracts: sync.Map{},
	}

	multisig.UsersContracts.Store("0x7d2d50791f839aea9b3ebe2c1dfd4dea13bc12ca", "0x116FfA11DD8829524767f561da5d33D3D170E17d")

	// initail initialization of clients contracts data
	nc.CliMultisig = &multisig

	log.Infof("Users data initialization done")

	// init gRPC server
	lis, err := net.Listen("tcp", conf.GrpcPort)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err.Error())
	}

	// Creates a new gRPC server
	ethCli := eth.NewClient(&conf.EthConf, nc.Clients, nc.CliMultisig)
	if err != nil {
		return nil, fmt.Errorf("eth.NewClient initialization: %s", err.Error())
	}
	log.Infof("ETH client initialization done")

	nc.Instance = ethCli

	// Dial to abi client to reach smart contracts methods
	ABIconn, err := ethclient.Dial(conf.AbiClientUrl)
	if err != nil {
		log.Fatalf("Failed to connect to infura %v", err)
	}

	s := grpc.NewServer()
	srv := streamer.Server{
		UsersData:       nc.Clients,
		EthCli:          nc.Instance,
		Info:            &conf.ServiceInfo,
		Multisig:        nc.CliMultisig,
		NetworkID:       conf.NetworkID,
		ResyncUrl:       resyncUrl,
		EtherscanAPIKey: conf.EtherscanAPIKey,
		EtherscanAPIURL: conf.EtherscanAPIURL,
		ABIcli:          ABIconn,
		GRPCserver:      s,
		Listener:        lis,
		ReloadChan:      make(chan struct{}),
	}

	nc.GRPCserver = &srv

	pb.RegisterNodeCommuunicationsServer(s, &srv)
	go s.Serve(lis)

	go WathReload(srv.ReloadChan, nc)

	return nc, nil
}

func FethResyncUrl(networkid int) string {
	switch networkid {
	case 4:
		return "http://api-rinkeby.etherscan.io/api?sort=asc&endblock=99999999&startblock=0&address="
	case 3:
		return "http://api-ropsten.etherscan.io/api?sort=asc&endblock=99999999&startblock=0&address="
	case 1:
		return "http://api.etherscan.io/api?sort=asc&endblock=99999999&startblock=0&address="
	default:
		return "http://api.etherscan.io/api?sort=asc&endblock=99999999&startblock=0&address="
	}
}

func WathReload(reload chan struct{}, cli *NodeClient) {
	// func WathReload(reload chan struct{}, s *grpc.Server, srv *streamer.Server, lis net.Listener, conf *Configuration) {
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
