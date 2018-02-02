/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package ethereum

import (
	"context"
	"log"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/onrik/ethrpc"
)

type Client struct {
	rpc    *ethrpc.EthRPC
	client *rpc.Client
	config *Conf
	db     store.UserStore
	log    slf.StructuredLogger
}

type Conf struct {
	Address string
	RpcPort string
	WsPort  string
	//...
}

func NewClient(conf *Conf, db store.UserStore) *Client {
	c := &Client{
		config: conf,
		db:     db,
		log:    slf.WithContext("ethereum-client"),
	}
	go c.Run()
	return c
}

func (c *Client) Run() error {
	c.log.Info("Run ETH Process")

	// c.rpc = ethrpc.NewEthRPC("http://" + c.config.Address + c.config.RpcPort)
	c.rpc = ethrpc.NewEthRPC("http://" + c.config.Address + c.config.RpcPort)
	// http://88.198.47.112:18545
	// ws://88.198.47.112:8545
	c.log.Debugf("ETH RPC Connectedon %s", "http://"+c.config.Address+c.config.RpcPort)

	_, err := c.rpc.EthNewPendingTransactionFilter()
	if err != nil {
		c.log.Errorf("NewClient:EthNewPendingTransactionFilter: %s", err.Error())
		return err
	}

	// client, err := rpc.Dial("ws://" + c.config.Address + c.config.WsPort)
	client, err := rpc.Dial("ws://" + c.config.Address + c.config.WsPort)

	if err != nil {
		c.log.Errorf("Dial err: %s", err.Error())
		return err
	}
	c.client = client
	c.log.Debugf("ETH RPC Connectedon %s", "ws://"+c.config.Address+c.config.WsPort)

	ch := make(chan interface{})

	_, err = c.client.Subscribe(context.Background(), "eth", ch, "newHeads")
	if err != nil {
		c.log.Errorf("Run: client.Subscribe: newHeads %s", err.Error())
		return err
	}

	_, err = c.client.Subscribe(context.Background(), "eth", ch, "newPendingTransactions")
	if err != nil {
		c.log.Errorf("Run: client.Subscribe: newPendingTransactions %s", err.Error())
		return err
	}

	for {
		//v:= <-c
		switch v := (<-ch).(type) {
		default:
			log.Printf("Not found type:", v)
		case string:
			// tx pool transaction
			go getTrasaction(v, c.rpc)
		case map[string]interface{}:
			// tx block transactions
			go getBlock(v["hash"].(string), c.rpc)
		}
	}

	defer client.Close()

	return nil
}
