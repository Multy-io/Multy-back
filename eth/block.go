package eth

import (
	pb "github.com/Multy-io/Multy-back/node-streamer/eth"
	"github.com/onrik/ethrpc"
)

func (c *Client) BlockTransaction(hash string) {
	block, err := c.Rpc.EthGetBlockByHash(hash, true)
	if err != nil {
		log.Errorf("Get Block Err:%s", err.Error())
		return
	}
	log.Debugf("new block number = %v", block.Number)
	c.Block <- pb.BlockHeight{
		Height: int64(block.Number),
	}

	txs := []ethrpc.Transaction{}
	if block.Transactions != nil {
		txs = block.Transactions
	} else {
		return
	}

	for _, rawTx := range txs {
		c.parseETHTransaction(rawTx, int64(*rawTx.BlockNumber), false)
		c.DeleteMempool <- pb.MempoolToDelete{
			Hash: rawTx.Hash,
		}
		// if strings.ToLower(rawTx.To) == strings.ToLower("0x116FfA11DD8829524767f561da5d33D3D170E17d") {
		// 	// fmt.Println("\n\ntx.to \n\n", tx.To)
		// 	c.FactoryContract(rawTx.Hash)
		// }
	}
}
