package nsbtc

import (
	pb "github.com/Multy-io/Multy-back/ns-btc-protobuf"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// ProcessTransaction from block
func (c *Client) BlockTransactions(hash *chainhash.Hash) {
	log.Debugf("New block connected %s", hash.String())
	// block Height

	blockVerbose, err := c.RPCClient.GetBlockVerbose(hash)
	if err != nil {
		log.Errorf("parseNewBlock:GetBlockVerbose: %s", err.Error())
		return
	}
	blockHeight := blockVerbose.Height

	//parse all block transactions
	rawBlock, err := c.RPCClient.GetBlock(hash)
	allBlockTransactions, err := rawBlock.TxHashes()
	if err != nil {
		log.Errorf("parseNewBlock:rawBlock.TxHashes: %s", err.Error())
	}

	// Broadcast to client to delete mempool
	for _, hash := range allBlockTransactions {
		c.DeleteMempool <- pb.MempoolToDelete{
			Hash: hash.String(),
		}
	}

	log.Debug("Start process TXs ")
	for _, txHash := range allBlockTransactions {

		blockTxVerbose, err := c.RPCClient.GetRawTransactionVerbose(&txHash)
		if err != nil {
			log.Errorf("parseNewBlock:RPCClient.GetRawTransactionVerbose: %s", err.Error())
			continue
		}

		go c.ProcessTransaction(blockHeight, blockTxVerbose, false)
	}
	log.Debug("END")
}

func (c *Client) ResyncBlock(blockVerbose *btcjson.GetBlockVerboseResult) {
	blockHeight := blockVerbose.Height
	log.Debugf("ResyncBlock on height %v", blockVerbose.Height)

	//parse all block transactions
	hash, _ := chainhash.NewHashFromStr(blockVerbose.Hash)
	rawBlock, err := c.RPCClient.GetBlock(hash)
	allBlockTransactions, err := rawBlock.TxHashes()
	if err != nil {
		log.Errorf("parseNewBlock:rawBlock.TxHashes: %s", err.Error())
	}

	for _, txHash := range allBlockTransactions {
		blockTxVerbose, err := c.RPCClient.GetRawTransactionVerbose(&txHash)
		if err != nil {
			log.Errorf("parseNewBlock:RPCClient.GetRawTransactionVerbose: %s", err.Error())
			continue
		}
		c.ProcessTransaction(blockHeight, blockTxVerbose, false)
	}
}
