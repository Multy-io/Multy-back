/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"github.com/Appscrunch/Multy-back/store"
	"github.com/btcsuite/btcd/wire"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func blockDisconnected(header *wire.BlockHeader) {
	disconnectedHash := header.BlockHash()
	disconnectedBlock, err := rpcClient.GetBlock(&disconnectedHash)
	if err != nil {
		log.Errorf("blockDisconnected:rpcClient.GetBlock: %s", err.Error())
	}

	rejectedTx := store.MultyTX{}
	for _, tx := range disconnectedBlock.Transactions {
		query := bson.M{"transactions.txhash": tx.TxHash()}
		txsData.Find(query).One(&rejectedTx)
		if err != nil {
			if err == mgo.ErrNotFound {
				continue
			}
			log.Errorf("blockDisconnected:txsData.Find: %s", err.Error())
			continue
		}

		query = bson.M{"transactions.txhash": tx.TxHash()}
		update := bson.M{
			"$set": bson.M{
				"transactions.$.txstatus": rejectedTx.TxStatus * -1,
			},
		}
		err = txsData.Update(query, update)
		if err != nil {
			log.Errorf("txsData.Update add new tx to user: %s", err.Error())
		}

	}

}
