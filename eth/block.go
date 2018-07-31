package eth

import (
	_ "github.com/KristinaEtc/slflog"
	pb "github.com/Multy-io/Multy-back/node-streamer/eth"
	"github.com/onrik/ethrpc"
)

func (c *Client) BlockTransaction(hash string) {
	block, err := c.Rpc.EthGetBlockByHash(hash, true)
	if err != nil {
		log.Errorf("Get Block Err:%s", err.Error())
		return
	}
	c.Block <- pb.BlockHeight{
		Height: int64(block.Number),
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
