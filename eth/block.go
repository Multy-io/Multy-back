package eth

import (
	"strings"

	pb "github.com/Multy-io/Multy-ETH-node-service/node-streamer"

	"github.com/onrik/ethrpc"
)

func (c *Client) BlockTransaction(hash string) {
	block, err := c.Rpc.EthGetBlockByHash(hash, true)
	if err != nil {
		log.Errorf("Get Block Err:%s", err.Error())
		return
	}

	go func() {
		log.Debugf("new block number = %v", block.Number)
		c.Block <- pb.BlockHeight{
			Height: int64(block.Number),
		}
	}()

	txs := []ethrpc.Transaction{}
	if block.Transactions != nil {
		txs = block.Transactions
	} else {

		return
	}

	// log.Errorf("block.Transactions %v", len(block.Transactions))

	for _, rawTx := range txs {
		c.parseETHMultisig(rawTx, int64(*rawTx.BlockNumber), false)
		c.parseETHTransaction(rawTx, int64(*rawTx.BlockNumber), false)
		go func() {
			c.DeleteMempool <- pb.MempoolToDelete{
				Hash: rawTx.Hash,
			}
		}()

		if strings.ToLower(rawTx.To) == strings.ToLower(c.Multisig.FactoryAddress) {
			log.Debugf("%v %s %v", strings.ToLower(rawTx.To), ":", strings.ToLower(c.Multisig.FactoryAddress))
			go c.FactoryContract(rawTx.Hash)
		}
	}

}
