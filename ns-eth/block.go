package nseth

import (
	"strings"

	pb "github.com/Multy-io/Multy-back/ns-eth-protobuf"

	"github.com/onrik/ethrpc"
)

func (c *Client) BlockTransaction(hash string) {
	block, err := c.Rpc.EthGetBlockByHash(hash, true)
	if err != nil {
		log.Errorf("Get Block Err:%s", err.Error())
		return
	}

	go func(blockNum int64) {
		log.Debugf("new block number = %v", blockNum)
		c.BlockStream <- pb.BlockHeight{
			Height: blockNum,
		}
	}(int64(block.Number))

	txs := []ethrpc.Transaction{}
	if block.Transactions != nil {
		txs = block.Transactions
	} else {
		return
	}

	log.Debugf("New block -  lenght = %d", len(txs))

	for _, rawTx := range txs {
		c.parseETHMultisig(rawTx, int64(*rawTx.BlockNumber), false)
		c.parseETHTransaction(rawTx, int64(*rawTx.BlockNumber), false)
		go func(hash string) {
			c.DeleteMempoolStream <- pb.MempoolToDelete{
				Hash: hash,
			}
		}(rawTx.Hash)

		if strings.ToLower(rawTx.To) == strings.ToLower(c.Multisig.FactoryAddress) {
			log.Debugf("%v %s %v", strings.ToLower(rawTx.To), ":", strings.ToLower(c.Multisig.FactoryAddress))
			go func(hash string) {
				go c.FactoryContract(hash)
			}(rawTx.Hash)
		}
	}
}
