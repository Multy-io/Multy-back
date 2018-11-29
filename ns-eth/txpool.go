/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package nseth

import (
	"strings"

	pb "github.com/Multy-io/Multy-back/ns-eth-protobuf"
	"github.com/Multy-io/Multy-back/store"
)

func (c *Client) txpoolTransaction(txHash string) {
	// rawTX, err := rpc.EthGetTransactionByHash(txHash)
	rawTx, err := c.Rpc.EthGetTransactionByHash(txHash)
	if err != nil {
		log.Errorf("c.Rpc.EthGetTransactionByHash:Get TX Err: %s", err.Error())
	}
	c.parseETHTransaction(*rawTx, -1, false)

	c.parseETHMultisig(*rawTx, -1, false)
	// log.Debugf("new txpool tx %v", rawTx.Hash)

	// add txpool record
	c.AddToMempoolStream <- pb.MempoolRecord{
		Category: int32(rawTx.Gas),
		HashTX:   rawTx.Hash,
	}
	if strings.ToLower(rawTx.To) == strings.ToLower(c.Multisig.FactoryAddress) {

		// go func() {
		fi, err := parseFactoryInput(rawTx.Input)
		if err != nil {
			log.Errorf("txpoolTransaction:parseFactoryInput: %s", err.Error())
		}
		fi.TxOfCreation = txHash
		fi.FactoryAddress = c.Multisig.FactoryAddress
		fi.DeployStatus = int64(store.MultisigStatusDeployPending)
		c.NewMultisigStream <- *fi
		// }()
	}

}
