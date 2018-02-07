/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package ethereum

import (
	"log"

	"github.com/onrik/ethrpc"
)

func getBlock(hash string, rpc *ethrpc.EthRPC, c *Client) {
	block, err := rpc.EthGetBlockByHash(hash, true)
	if err != nil {
		log.Printf("Get Block Err:%s", err.Error())
	}

	txs := block.Transactions
	log.Printf("New block -  lenght = %d ", len(txs))

	for _, rawTx := range txs {

		c.parseETHTransaction(rawTx, int64(*rawTx.BlockNumber))
		// TODO:
		// Send NOTIFY !!!
		// add block transaction to database, or update existing if needed.
		// sent notification to user
	}
}
