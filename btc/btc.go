/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"time"

	pb "github.com/Appscrunch/Multy-back/node-streamer/btc"
	"github.com/KristinaEtc/slf"
	_ "github.com/KristinaEtc/slflog"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

// Dirty hack - this will be wrapped to a struct
var (
	RpcClient      = &rpcclient.Client{}
	TransactionsCh = make(chan pb.BTCTransaction)
	AddSpOut       = make(chan pb.AddSpOut)
	DelSpOut       = make(chan pb.ReqDeleteSpOut)
	DeleteMempool  = make(chan pb.MempoolToDelete)
	AddToMempool   = make(chan pb.MempoolRecord)
	rpcConf        *rpcclient.ConnConfig
)

var connCfg = &rpcclient.ConnConfig{
	Host:     "192.168.0.121:18334",
	User:     "multy",
	Pass:     "multy",
	Endpoint: "ws",
	//Certificates: []byte(Cert), // add it in InitHandlers function

	HTTPPostMode: false, // Bitcoin core only supports HTTP POST mode
	DisableTLS:   false, // Bitcoin core does not provide TLS by default

}

var log = slf.WithContext("btc")

func InitHandlers(certFromConf []byte, btcNodeAddress string, usersData *map[string]string) (*rpcclient.Client, error) {
	connCfg.Certificates = certFromConf
	log.Infof("cert= %d bytes\n", len(certFromConf))

	go RunProcess(btcNodeAddress, usersData)
	return RpcClient, nil
}

func RunProcess(btcNodeAddress string, usersData *map[string]string) error {
	log.Info("Run Process")

	ntfnHandlers := rpcclient.NotificationHandlers{
		OnBlockConnected: func(hash *chainhash.Hash, height int32, t time.Time) {
			log.Debugf("OnBlockConnected: %v (%d) %v", hash, height, t)
			go blockTransactions(hash, usersData)
		},
		OnTxAcceptedVerbose: func(txDetails *btcjson.TxRawResult) {
			// log.Debugf("OnTxAcceptedVerbose: new transaction id = %v", txDetails.Txid)
			go mempoolTransaction(txDetails, usersData)
		},
		OnFilteredBlockDisconnected: func(height int32, header *wire.BlockHeader) {

		},
	}

	//overwrite btc node address
	connCfg.Host = btcNodeAddress

	var err error
	RpcClient, err = rpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		log.Errorf("RunProcess(): rpcclient.New %s\n", err.Error())
		return err
	}

	// Register for block connect and disconnect notifications.
	if err = RpcClient.NotifyBlocks(); err != nil {
		return err
	}
	log.Info("NotifyBlocks: Registration Complete")

	// Register for new transaction in mempool notifications.
	if err = RpcClient.NotifyNewTransactions(true); err != nil {
		return err
	}
	log.Info("NotifyNewTransactions: Registration Complete")
	RpcClient.WaitForShutdown()
	return nil
}
