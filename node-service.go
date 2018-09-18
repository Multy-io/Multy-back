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
func Init(conf *Configuration) (*NodeClient, error) {
	resyncUrl := FethResyncUrl(conf.NetworkID)
	conf.ResyncUrl = resyncUrl
	cli := &NodeClient{
		Config: conf,
	}

	var usersData sync.Map

	usersData.Store("address", store.AddressExtended{
		UserID:       "kek",
		WalletIndex:  1,
		AddressIndex: 2,
	})

	// initail initialization of clients data
	cli.Clients = &usersData

	//TODO: init contract clients
	multisig := eth.Multisig{
		FactoryAddress: conf.MultisigFactory,
		UsersContracts: sync.Map{},
	}

	multisig.UsersContracts.Store("0x7d2d50791f839aea9b3ebe2c1dfd4dea13bc12ca", "0x116FfA11DD8829524767f561da5d33D3D170E17d")

	// initail initialization of clients contracts data
	cli.CliMultisig = &multisig

	log.Infof("Users data initialization done")

	// init gRPC server
	lis, err := net.Listen("tcp", conf.GrpcPort)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err.Error())
	}
	// Creates a new gRPC server

	ethCli := eth.NewClient(&conf.EthConf, cli.Clients, cli.CliMultisig)
	if err != nil {
		return nil, fmt.Errorf("eth.NewClient initialization: %s", err.Error())
	}
	log.Infof("ETH client initialization done")

	cli.Instance = ethCli

	ABIconn, err := ethclient.Dial(conf.AbiClientUrl)
	if err != nil {
		log.Fatalf("Failed to connect to infura %v", err)
	}

	s := grpc.NewServer()
	srv := streamer.Server{
		UsersData: cli.Clients,
		EthCli:    cli.Instance,
		Info:      &conf.ServiceInfo,
		Multisig:  cli.CliMultisig,
		NetworkID: conf.NetworkID,
		ResyncUrl: resyncUrl,
		ABIcli:    ABIconn,
	}

	pb.RegisterNodeCommuunicationsServer(s, &srv)
	go s.Serve(lis)
	return cli, nil
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
