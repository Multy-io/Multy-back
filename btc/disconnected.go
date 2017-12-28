package btc

import (
	"github.com/btcsuite/btcd/wire"
	"gopkg.in/mgo.v2/bson"
)

func blockDisconnected(header *wire.BlockHeader) {
	disconnectedHash := header.BlockHash()
	disconnectedBlock, err := rpcClient.GetBlock(&disconnectedHash)
	if err != nil {
		log.Errorf("blockDisconnected:rpcClient.GetBlock: %s", err.Error())
	}

	for _, tx := range disconnectedBlock.Transactions {
		query := bson.M{"transactions.txhash": tx.TxHash()}
		update := bson.M{
			"$set": bson.M{
				"transactions.$.txstatus": TxStatusRejectedFromBlock,
			},
		}
		err = txsData.Update(query, update)
		if err != nil {
			log.Errorf("txsData.Update add new tx to user: %s", err.Error())
		}

	}

}
