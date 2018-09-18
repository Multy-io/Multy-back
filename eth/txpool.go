/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package eth

import (
	"strings"

	pb "github.com/Multy-io/Multy-ETH-node-service/node-streamer"
	"github.com/Multy-io/Multy-back/store"
)

func (c *Client) txpoolTransaction(txHash string) {
	// rawTX, err := rpc.EthGetTransactionByHash(txHash)
	rawTx, err := c.Rpc.EthGetTransactionByHash(txHash)
	if err != nil {
		log.Errorf("Get TX Err: %s", err.Error())
	}
	c.parseETHTransaction(*rawTx, -1, false)

	c.parseETHMultisig(*rawTx, -1, false)
	// log.Debugf("new txpool tx %v", rawTx.Hash)

	// add txpool record
	c.AddToMempool <- pb.MempoolRecord{
		Category: int32(rawTx.Gas),
		HashTX:   rawTx.Hash,
	}
	if strings.ToLower(rawTx.To) == strings.ToLower(c.Multisig.FactoryAddress) {

		go func() {
			fi, err := parseFactoryInput(rawTx.Input)
			if err != nil {
				log.Errorf("FactoryContract:parseInput: %s", err.Error())
			}
			fi.TxOfCreation = txHash
			fi.FactoryAddress = c.Multisig.FactoryAddress
			fi.DeployStatus = int64(store.MultisigStatusDeployPending)
			c.NewMultisig <- *fi
		}()
	}

}
