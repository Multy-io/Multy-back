/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"github.com/btcsuite/btcd/btcjson"
)

func mempoolTransaction(inTx *btcjson.TxRawResult) {
	log.Debugf("[MEMPOOL TX]")
	processTransaction(-1, inTx, false)

}
