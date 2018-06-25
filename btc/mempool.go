package btc

import (
	pb "github.com/Multy-io/Multy-back/node-streamer/btc"
	"github.com/btcsuite/btcd/btcjson"
)

// ProcessTransaction from mempool
func (c *Client) mempoolTransaction(inTx *btcjson.TxRawResult) {
	// Brodcast new mempool transaction to mempool event
	rec := c.rawTxToMempoolRec(inTx)
	c.AddToMempool <- pb.MempoolRecord{
		Category: int32(rec.Category),
		HashTX:   rec.HashTX,
	}

	// Process tx for tx history and spendable outs
	c.ProcessTransaction(-1, inTx, false)
}
