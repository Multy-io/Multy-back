package btc

import (
	pb "github.com/Appscrunch/Multy-back/node-streamer/btc"
	"github.com/btcsuite/btcd/btcjson"
)

// ProcessTransaction from mempool
func mempoolTransaction(inTx *btcjson.TxRawResult, usersData *map[string]string) {
	log.Debugf("[MEMPOOL TX]")
	// fmt.Println(*usersData)
	// Brodcast new mempool transaction to mempool event

	rec := rawTxToMempoolRec(inTx)
	AddToMempool <- pb.MempoolRecord{
		Category: int32(rec.Category),
		HashTX:   rec.HashTX,
	}

	// Process tx for tx history and spendable outs
	processTransaction(-1, inTx, false, usersData)
}
