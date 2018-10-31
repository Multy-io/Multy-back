/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package eth

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"

	pb "github.com/Multy-io/Multy-ETH-node-service/node-streamer"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/jekabolt/slf"
	_ "github.com/jekabolt/slflog"
	"github.com/onrik/ethrpc"
)

var log = slf.WithContext("eth")

type Client struct {
	Rpc            *ethrpc.EthRPC
	Client         *rpc.Client
	config         *Conf
	TransactionsCh chan pb.ETHTransaction
	DeleteMempool  chan pb.MempoolToDelete
	AddToMempool   chan pb.MempoolRecord
	Block          chan pb.BlockHeight
	UsersData      *sync.Map
	NewMultisig    chan pb.Multisig
	Multisig       *Multisig
	AbiClient      *ethclient.Client
}

type Conf struct {
	Address string
	RpcPort string
	WsPort  string
}

func NewClient(conf *Conf, usersData *sync.Map, multisig *Multisig) *Client {
	c := &Client{
		config:         conf,
		TransactionsCh: make(chan pb.ETHTransaction),
		DeleteMempool:  make(chan pb.MempoolToDelete),
		AddToMempool:   make(chan pb.MempoolRecord),
		Block:          make(chan pb.BlockHeight),
		NewMultisig:    make(chan pb.Multisig),
		UsersData:      usersData,
		Multisig:       multisig,
	}

	go c.RunProcess()
	return c
}
func (c *Client) Shutdown() {
	c.Client.Close()
}

func (c *Client) RunProcess() error {
	log.Info("Run ETH Process")
	// c.Rpc = ethrpc.NewEthRPC("http://" + c.config.Address + c.config.RpcPort)
	c.Rpc = ethrpc.NewEthRPC("http" + c.config.Address + c.config.RpcPort)
	log.Infof("ETH RPC Connection %s", "http://"+c.config.Address+c.config.RpcPort)

	_, err := c.Rpc.EthNewPendingTransactionFilter()
	if err != nil {
		log.Errorf("NewClient:EthNewPendingTransactionFilter: %s", err.Error())
		return err
	}
	// client, err := rpc.Dial("ws://" + c.config.Address + c.config.WsPort)
	client, err := rpc.Dial("ws" + c.config.Address + c.config.WsPort)

	if err != nil {
		log.Errorf("Dial err: %s", err.Error())
		return err
	}
	c.Client = client
	log.Infof("ETH RPC Connection %s", "ws://"+c.config.Address+c.config.WsPort)

	ch := make(chan interface{})

	_, err = c.Client.Subscribe(context.Background(), "eth", ch, "newHeads")
	if err != nil {
		log.Errorf("Run: client.Subscribe: newHeads %s", err.Error())
		return err
	}

	_, err = c.Client.Subscribe(context.Background(), "eth", ch, "newPendingTransactions")
	if err != nil {
		log.Errorf("Run: client.Subscribe: newPendingTransactions %s", err.Error())
		return err
	}

	for {
		switch v := (<-ch).(type) {
		default:
			log.Errorf("Not found type: %v", v)
		case string:
			// tx pool transaction
			go c.txpoolTransaction(v)
			// fmt.Println(v)
		case map[string]interface{}:
			// tx block transactions
			// fmt.Println(v)
			go c.BlockTransaction(v["hash"].(string))
		}
	}

	defer client.Close()

	return nil
}
