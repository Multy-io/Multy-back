/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package nsbtc

import (
	"sync"
	"time"

	pb "github.com/Multy-io/Multy-back/ns-btc-protobuf"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	_ "github.com/jekabolt/slflog"
)

type Client struct {
	RPCClient      *rpcclient.Client
	ResyncCh       chan pb.Resync
	TransactionsCh chan pb.BTCTransaction
	AddSpOut       chan pb.AddSpOut
	DelSpOut       chan pb.ReqDeleteSpOut
	DeleteMempool  chan pb.MempoolToDelete
	AddToMempool   chan pb.MempoolRecord
	Block          chan pb.BlockHeight
	UsersData      *sync.Map
	rpcConf        *rpcclient.ConnConfig
}

// var log = slf.WithContext("btc")

func NewClient(certFromConf []byte, btcNodeAddress string, usersData *sync.Map) (*Client, error) {

	cli := &Client{
		ResyncCh:       make(chan pb.Resync),
		TransactionsCh: make(chan pb.BTCTransaction),
		AddSpOut:       make(chan pb.AddSpOut),
		DelSpOut:       make(chan pb.ReqDeleteSpOut),
		DeleteMempool:  make(chan pb.MempoolToDelete),
		AddToMempool:   make(chan pb.MempoolRecord),
		Block:          make(chan pb.BlockHeight),
		rpcConf: &rpcclient.ConnConfig{
			Host:         btcNodeAddress,
			User:         "multy",
			Pass:         "multy",
			Endpoint:     "ws",
			Certificates: certFromConf,
			HTTPPostMode: false, // Bitcoin core only supports HTTP POST mode
			DisableTLS:   false, // Bitcoin core does not provide TLS by default
		},
		UsersData: usersData,
	}

	log.Infof("cert= %d bytes\n", len(certFromConf))

	go cli.RunProcess(btcNodeAddress)
	return cli, nil
}

func (c *Client) RunProcess(btcNodeAddress string) error {
	log.Info("Run Process")

	ntfnHandlers := rpcclient.NotificationHandlers{
		OnBlockConnected: func(hash *chainhash.Hash, height int32, t time.Time) {
			log.Debugf("OnBlockConnected: %v (%d) %v", hash, height, t)
			if hash != nil {
				go c.BlockTransactions(hash)
			}
			if hash == nil {
				log.Errorf("OnBlockConnected:hash is nil")
			}
			c.Block <- pb.BlockHeight{Height: int64(height)}
		},
		OnTxAcceptedVerbose: func(txDetails *btcjson.TxRawResult) {
			if txDetails != nil {
				go c.mempoolTransaction(txDetails)
			}
			if txDetails == nil {
				log.Errorf("OnTxAcceptedVerbose:txDetails is nil")
			}
		},
		OnFilteredBlockDisconnected: func(height int32, header *wire.BlockHeader) {

		},
	}

	RPCClient, err := rpcclient.New(c.rpcConf, &ntfnHandlers)
	if err != nil {
		log.Errorf("RunProcess(): RPCclient.New %s\n", err.Error())
		return err
	}

	// Register for block connect and disconnect notifications.
	if err = RPCClient.NotifyBlocks(); err != nil {
		return err
	}
	log.Info("NotifyBlocks: Registration Complete")

	// Register for new transaction in mempool notifications.
	if err = RPCClient.NotifyNewTransactions(true); err != nil {
		return err
	}
	log.Info("NotifyNewTransactions: Registration Complete")

	c.RPCClient = RPCClient

	c.RPCClient.WaitForShutdown()
	return nil
}
