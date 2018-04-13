package eth

import (
	pb "github.com/Appscrunch/Multy-back/node-streamer/eth"
	"github.com/onrik/ethrpc"
)

func (c *Client) blockTransaction(hash string) {
	block, err := c.Rpc.EthGetBlockByHash(hash, true)
	if err != nil {
		log.Errorf("Get Block Err:%s", err.Error())
		return
	}

	txs := []ethrpc.Transaction{}
	if block.Transactions != nil {
		txs = block.Transactions
	} else {
		return
	}

	log.Debugf("New block -  lenght = %d", len(txs))

	for _, rawTx := range txs {
		c.parseETHTransaction(rawTx, int64(*rawTx.BlockNumber), false)
		c.DeleteMempool <- pb.MempoolToDelete{
			Hash: rawTx.Hash,
		}
	}
}
