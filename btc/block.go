package btc

import (
	"log"

	"gopkg.in/mgo.v2/bson"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func parseNewBlock(hash *chainhash.Hash) {
	log.Printf("[DEBUG] New block connected %s", hash.String())

	// block Height
	// blockVerbose, err := rpcClient.GetBlockVerbose(hash)
	// blockHeight := blockVerbose.Height

	//parse all block transactions
	rawBlock, err := rpcClient.GetBlock(hash)
	allBlockTransactions, err := rawBlock.TxHashes()
	if err != nil {
		log.Printf("[ERR] parseNewBlock:rawBlock.TxHashes: %s", err.Error())
	}

	var user store.User

	// range over all block txID's and notify clients about including their transaction in block as input or output
	// delete by transaction hash record from mempool db to estimete tx speed
	for _, txHash := range allBlockTransactions {

		blockTxVerbose, err := rpcClient.GetRawTransactionVerbose(&txHash)
		if err != nil {
			log.Printf("[ERR] parseNewBlock:rpcClient.GetRawTransactionVerbose: %s ", err.Error())
			continue
		}

		// delete all block transations from memPoolDB
		query := bson.M{"hashtx": blockTxVerbose.Txid}
		err = mempoolRates.Remove(query)
		if err != nil {
			log.Printf("[ERR] parseNewBlock:mempoolRates.Remove: %s ", err.Error())
		} else {
			log.Printf("[DEBUG] Tx removed: %s ", blockTxVerbose.Txid)
		}

		// parse block tx outputs and notify
		for _, out := range blockTxVerbose.Vout {
			for _, address := range out.ScriptPubKey.Addresses {

				query := bson.M{"wallets.addresses.address": address}
				err := usersData.Find(query).One(&user)
				if err != nil {
					continue
				}
				log.Printf("[DEBUG] [IS OUR USER] parseNewBlock: usersData.Find = %s", address)

				chToClient <- BtcTransactionWithUserID{
					UserID: user.UserID,
					NotificationMsg: &BtcTransaction{
						TransactionType: txInBlock,
						Amount:          out.Value,
						TxID:            blockTxVerbose.Txid,
						Address:         address,
					},
				}

			}
		}

		// parse block tx inputs and notify
		for _, input := range blockTxVerbose.Vin {
			txHash, err := chainhash.NewHashFromStr(input.Txid)
			if err != nil {
				log.Printf("[ERR] parseNewBlock: chainhash.NewHashFromStr = %s", err)
			}
			previousTx, err := rpcClient.GetRawTransactionVerbose(txHash)
			if err != nil {
				log.Printf("[ERR] parseNewBlock:rpcClient.GetRawTransactionVerbose: %s ", err.Error())
				continue
			}

			for _, out := range previousTx.Vout {
				for _, address := range out.ScriptPubKey.Addresses {
					query := bson.M{"wallets.addresses.address": address}
					err := usersData.Find(query).One(&user)
					if err != nil {
						continue
					}
					log.Printf("[DEBUG] [IS OUR USER]-AS-OUT parseMempoolTransaction: usersData.Find = %s", address)

					chToClient <- BtcTransactionWithUserID{
						UserID: user.UserID,
						NotificationMsg: &BtcTransaction{
							TransactionType: txOutBlock,
							Amount:          out.Value,
							TxID:            blockTxVerbose.Txid,
							Address:         address,
						},
					}

				}
			}
		}

	}
}
