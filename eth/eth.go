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
	Rpc                 *ethrpc.EthRPC
	Client              *rpc.Client
	config              *Conf
	TransactionsStream  chan pb.ETHTransaction
	DeleteMempoolStream chan pb.MempoolToDelete
	AddToMempoolStream  chan pb.MempoolRecord
	BlockStream         chan pb.BlockHeight
	NewMultisigStream   chan pb.Multisig
	RPCStream           chan interface{}
	Done                <-chan interface{}
	Stop                chan struct{}
	UsersData           *sync.Map
	Multisig            *Multisig
	AbiClient           *ethclient.Client
}

type Conf struct {
	Address string
	RpcPort string
	WsPort  string
}

func NewClient(conf *Conf, usersData *sync.Map, multisig *Multisig) *Client {
	c := &Client{
		config:              conf,
		TransactionsStream:  make(chan pb.ETHTransaction),
		DeleteMempoolStream: make(chan pb.MempoolToDelete),
		AddToMempoolStream:  make(chan pb.MempoolRecord),
		BlockStream:         make(chan pb.BlockHeight),
		NewMultisigStream:   make(chan pb.Multisig),
		Done:                make(chan interface{}),
		Stop:                make(chan struct{}),
		UsersData:           usersData,
		Multisig:            multisig,
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

	c.RPCStream = ch

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

	// done := or(c.fanIn(ch)...)

	// c.Done = done

	for {
		switch v := (<-ch).(type) {
		default:
			log.Errorf("Not found type: %v", v)
		case string:
			go c.txpoolTransaction(v)
		case map[string]interface{}:
			go c.BlockTransaction(v["hash"].(string))
		case nil:
			defer func() {
				c.Stop <- struct{}{}
			}()
			defer client.Close()
			log.Debugf("RPC stream closed")
			return nil
		}
	}

}
