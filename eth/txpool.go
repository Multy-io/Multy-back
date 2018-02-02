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

func getTrasaction(txHash string, rpc *ethrpc.EthRPC) {
	// rawTX, err := rpc.EthGetTransactionByHash(txHash)
	_, err := rpc.EthGetTransactionByHash(txHash)
	if err != nil {
		log.Printf("Get TX Err:", err)
	}
	// parseRawTransaction(*rawTX, true)
}
