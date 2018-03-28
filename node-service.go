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
	"time"

	"github.com/Appscrunch/Multy-BTC-node-service/btc"
	"github.com/Appscrunch/Multy-BTC-node-service/streamer"
	pb "github.com/Appscrunch/Multy-back/node-streamer/btc"
	"github.com/KristinaEtc/slf"
	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcd/rpcclient"
	"google.golang.org/grpc"
)

var (
	log = slf.WithContext("NodeClient")
)

// Multy is a main struct of service

// NodeClient is a main struct of service
type NodeClient struct {
	Config     *Configuration
	Instance   *rpcclient.Client
	GRPCserver *streamer.Server
	Clients    *map[string]string // address to userid
	BtcApi     *gobcy.API
}

// Init initializes Multy instance
func Init(conf *Configuration) (*NodeClient, error) {
	cli := &NodeClient{
		Config: conf,
	}

	var usersData = map[string]string{
		"2MvPhdUf3cwaadRKsSgbQ2SXc83CPcBJezT": "baka",
	}

	// wait for initial users data
	for {
		fmt.Println("No users data", usersData)
		if len(usersData) == 0 {
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}

	api := gobcy.API{
		Token: conf.BTCAPI.Token,
		Coin:  conf.BTCAPI.Coin,
		Chain: conf.BTCAPI.Chain,
	}
	cli.BtcApi = &api
	log.Debug("btc api initialization done")

	// initail initialization of clients data
	cli.Clients = &usersData
	log.Debug("Users data initialization done")

	// init gRPC server
	lis, err := net.Listen("tcp", conf.GrpcPort)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err.Error())
	}
	// Creates a new gRPC server
	s := grpc.NewServer()
	srv := streamer.Server{
		UsersData: cli.Clients,
		BtcAPI:    cli.BtcApi,
		M:         &sync.Mutex{},
	}
	pb.RegisterNodeCommuunicationsServer(s, &srv)
	go s.Serve(lis)

	// cli.GRPCserver = srv.

	_, err = btc.InitHandlers(getCertificate(conf.BTCSertificate), conf.BTCNodeAddress, cli.Clients)
	if err != nil {
		return nil, fmt.Errorf("Blockchain api initialization: %s", err.Error())
	}
	log.Debug("BTC client initialization done")
	cli.Instance = btc.RpcClient

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
