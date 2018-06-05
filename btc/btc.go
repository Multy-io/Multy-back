/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"sync"
	"time"

	pb "github.com/Appscrunch/Multy-back/node-streamer/btc"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
	_ "github.com/KristinaEtc/slflog"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

type Client struct {
	RpcClient      *rpcclient.Client
	ResyncCh       chan pb.Resync
	TransactionsCh chan pb.BTCTransaction
	AddSpOut       chan pb.AddSpOut
	DelSpOut       chan pb.ReqDeleteSpOut
	DeleteMempool  chan pb.MempoolToDelete
	AddToMempool   chan pb.MempoolRecord
	UsersData      *map[string]store.AddressExtended
	rpcConf        *rpcclient.ConnConfig
	UserDataM      *sync.Mutex
	RpcClientM     *sync.Mutex
}

var log = slf.WithContext("btc")

func NewClient(certFromConf []byte, btcNodeAddress string, usersData *map[string]store.AddressExtended, udm, rpcm *sync.Mutex) (*Client, error) {

	cli := &Client{
		ResyncCh:       make(chan pb.Resync),
		TransactionsCh: make(chan pb.BTCTransaction),
		AddSpOut:       make(chan pb.AddSpOut),
		DelSpOut:       make(chan pb.ReqDeleteSpOut),
		DeleteMempool:  make(chan pb.MempoolToDelete),
		AddToMempool:   make(chan pb.MempoolRecord),
		rpcConf: &rpcclient.ConnConfig{
			Host:         btcNodeAddress,
			User:         "multy",
			Pass:         "multy",
			Endpoint:     "ws",
			Certificates: certFromConf,
			HTTPPostMode: false, // Bitcoin core only supports HTTP POST mode
			DisableTLS:   false, // Bitcoin core does not provide TLS by default
		},
		UsersData:  usersData,
		UserDataM:  udm,
		RpcClientM: rpcm,
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
			go c.BlockTransactions(hash)
		},
		OnTxAcceptedVerbose: func(txDetails *btcjson.TxRawResult) {
			// log.Debugf("OnTxAcceptedVerbose: new transaction id = %v \n ud = %v lock = %v", txDetails.Txid, c.UsersData, c.UserDataM)
			go c.mempoolTransaction(txDetails)
		},
		OnFilteredBlockDisconnected: func(height int32, header *wire.BlockHeader) {

		},
	}

	rpcClient, err := rpcclient.New(c.rpcConf, &ntfnHandlers)
	if err != nil {
		log.Errorf("RunProcess(): rpcclient.New %s\n", err.Error())
		return err
	}

	// Register for block connect and disconnect notifications.
	if err = rpcClient.NotifyBlocks(); err != nil {
		return err
	}
	log.Info("NotifyBlocks: Registration Complete")

	// Register for new transaction in mempool notifications.
	if err = rpcClient.NotifyNewTransactions(true); err != nil {
		return err
	}
	log.Info("NotifyNewTransactions: Registration Complete")

	c.RpcClientM.Lock()
	c.RpcClient = rpcClient
	c.RpcClientM.Unlock()

	c.RpcClient.WaitForShutdown()
	return nil
}
