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

func getBlock(hash string, rpc *ethrpc.EthRPC) {
	block, err := rpc.EthGetBlockByHash(hash, true)
	if err != nil {
		log.Printf("Get Block Err:%s", err.Error())
	}

	txs := block.Transactions
	log.Printf("New block -  lenght = %d ", len(txs))
	// for _, rawTx := range txs {
	// 	parseRawTransaction(rawTx, false)
	// }
}
