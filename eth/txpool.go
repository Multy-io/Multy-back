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

func getTrasaction(txHash string, rpc *ethrpc.EthRPC, c *Client) {
	// rawTX, err := rpc.EthGetTransactionByHash(txHash)
	rawTx, err := rpc.EthGetTransactionByHash(txHash)
	if err != nil {
		log.Printf("Get TX Err:", err)
	}
	c.parseETHTransaction(*rawTx, -1)

	// TODO:
	// add block transaction to database, or update existing if needed.
	// sent notification to user
	// parseRawTransaction(*rawTX, true)
	//
}
