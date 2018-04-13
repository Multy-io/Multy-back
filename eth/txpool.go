/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package eth

import (
	pb "github.com/Appscrunch/Multy-back/node-streamer/eth"
)

func (c *Client) txpoolTransaction(txHash string) {
	// rawTX, err := rpc.EthGetTransactionByHash(txHash)
	rawTx, err := c.Rpc.EthGetTransactionByHash(txHash)
	if err != nil {
		log.Errorf("Get TX Err: %s", err.Error())
	}
	c.parseETHTransaction(*rawTx, -1, false)
	log.Debugf("new txpool tx %v", rawTx.Hash)

	// add txpool record
	c.AddToMempool <- pb.MempoolRecord{
		Category: int32(rawTx.Gas),
		HashTX:   rawTx.Hash,
	}
}
