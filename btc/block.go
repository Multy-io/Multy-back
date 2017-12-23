package btc

import (
	"encoding/json"
	"fmt"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func notifyNewBlockTx(hash *chainhash.Hash) {
	log.Debugf("New block connected %s", hash.String())

	// block Height
	// blockVerbose, err := rpcClient.GetBlockVerbose(hash)
	// blockHeight := blockVerbose.Height

	//parse all block transactions
	rawBlock, err := rpcClient.GetBlock(hash)
	allBlockTransactions, err := rawBlock.TxHashes()
	if err != nil {
		log.Errorf("parseNewBlock:rawBlock.TxHashes: %s", err.Error())
	}

	var user store.User

	// range over all block txID's and notify clients about including their transaction in block as input or output
	// delete by transaction hash record from mempool db to estimete tx speed
	for _, txHash := range allBlockTransactions {

		blockTxVerbose, err := rpcClient.GetRawTransactionVerbose(&txHash)
		if err != nil {
			log.Errorf("parseNewBlock:rpcClient.GetRawTransactionVerbose: %s", err.Error())
			continue
		}

		// delete all block transations from memPoolDB
		query := bson.M{"hashtx": blockTxVerbose.Txid}
		err = mempoolRates.Remove(query)
		if err != nil {
			log.Errorf("parseNewBlock:mempoolRates.Remove: %s", err.Error())
		} else {
			log.Debugf("Tx removed: %s", blockTxVerbose.Txid)
		}

		// parse block tx outputs and notify
		for _, out := range blockTxVerbose.Vout {
			for _, address := range out.ScriptPubKey.Addresses {

				query := bson.M{"wallets.addresses.address": address}
				err := usersData.Find(query).One(&user)
				if err != nil {
					continue
				}
				log.Debugf("[IS OUR USER] parseNewBlock: usersData.Find = %s", address)

				txMsq := BtcTransactionWithUserID{
					UserID: user.UserID,
					NotificationMsg: &BtcTransaction{
						TransactionType: txInBlock,
						Amount:          out.Value,
						TxID:            blockTxVerbose.Txid,
						Address:         address,
					},
				}
				sendNotifyToClients(&txMsq)

			}
		}

		// parse block tx inputs and notify
		for _, input := range blockTxVerbose.Vin {
			txHash, err := chainhash.NewHashFromStr(input.Txid)
			if err != nil {
				log.Errorf("parseNewBlock: chainhash.NewHashFromStr = %s", err)
			}
			previousTx, err := rpcClient.GetRawTransactionVerbose(txHash)
			if err != nil {
				log.Errorf("parseNewBlock:rpcClient.GetRawTransactionVerbose: %s ", err.Error())
				continue
			}

			for _, out := range previousTx.Vout {
				for _, address := range out.ScriptPubKey.Addresses {
					query := bson.M{"wallets.addresses.address": address}
					err := usersData.Find(query).One(&user)
					if err != nil {
						continue
					}
					log.Debugf("[IS OUR USER]-AS-OUT parseMempoolTransaction: usersData.Find = %s", address)

					txMsq := BtcTransactionWithUserID{
						UserID: user.UserID,
						NotificationMsg: &BtcTransaction{
							TransactionType: txOutBlock,
							Amount:          out.Value,
							TxID:            blockTxVerbose.Txid,
							Address:         address,
						},
					}
					sendNotifyToClients(&txMsq)
				}
			}
		}

	}
}

func sendNotifyToClients(txMsq *BtcTransactionWithUserID) {
	newTxJSON, err := json.Marshal(txMsq)
	if err != nil {
		log.Errorf("sendNotifyToClients: [%+v] %s\n", txMsq, err.Error())
		return
	}

	err = nsqProducer.Publish(TopicTransaction, newTxJSON)
	if err != nil {
		log.Errorf("nsq publish new transaction: [%+v] %s\n", txMsq, err.Error())
		return
	}
	return
}

func blockTransactions(hash *chainhash.Hash) {
	log.Debugf("New block connected %s", hash.String())

	// block Height
	blockVerbose, err := rpcClient.GetBlockVerbose(hash)
	blockHeight := blockVerbose.Height

	//parse all block transactions
	rawBlock, err := rpcClient.GetBlock(hash)
	allBlockTransactions, err := rawBlock.TxHashes()
	if err != nil {
		log.Errorf("parseNewBlock:rawBlock.TxHashes: %s", err.Error())
	}

	var user store.User

	for _, txHash := range allBlockTransactions {

		blockTxVerbose, err := rpcClient.GetRawTransactionVerbose(&txHash)
		if err != nil {
			log.Errorf("parseNewBlock:rpcClient.GetRawTransactionVerbose: %s", err.Error())
			continue
		}

		// apear as output
		for _, output := range blockTxVerbose.Vout {
			for _, address := range output.ScriptPubKey.Addresses {
				query := bson.M{"wallets.addresses.address": address}
				err := usersData.Find(query).One(&user)
				if err != nil {
					continue
					// is not our user
				} else {
					fmt.Println("[ITS OUR USER] ", user.UserID)
				}

				err = txsData.Find(bson.M{"userid": user.UserID}).One(nil)
				if err == mgo.ErrNotFound {
					newRec := newTxRecord(user.UserID, blockTxVerbose.Txid, blockTxVerbose.Hash, output.ScriptPubKey.Hex, address, TxStatusAppearedInBlockIncoming, int(output.N), blockHeight, output.Value)
					err = txsData.Insert(newRec)
					if err != nil {
						log.Errorf("[txsData.Insert: %s", err.Error())
					}
					continue
				} else if err != nil && err != mgo.ErrNotFound {
					log.Errorf("txsData.Find: %s", err.Error())
				}

				sel := bson.M{"userid": user.UserID, "transactions.txid": blockTxVerbose.Txid, "transactions.txaddress": address}
				err = txsData.Find(sel).One(nil)
				if err == mgo.ErrNotFound {
					newTx := newMultyTX(blockTxVerbose.Txid, blockTxVerbose.Hash, output.ScriptPubKey.Hex, address, TxStatusAppearedInBlockIncoming, int(output.N), blockHeight, output.Value)
					sel = bson.M{"userid": user.UserID}
					update := bson.M{"$push": bson.M{"transactions": newTx}}
					err = txsData.Update(sel, update)
					if err != nil {
						log.Errorf("txsData.Update add new tx to user: %s", err.Error())
					}
					continue
				} else if err != nil && err != mgo.ErrNotFound {
					log.Errorf("[ERR]txsData.Find: %s", err.Error())
				}

				sel = bson.M{"userid": user.UserID, "transactions.txid": blockTxVerbose.Txid, "transactions.txaddress": address}
				update := bson.M{
					"$set": bson.M{
						"transactions.$.txstatus":      TxStatusAppearedInBlockIncoming,
						"transactions.$.txblockheight": blockHeight,
					},
				}
				err = txsData.Update(sel, update)
				if err != nil {
					log.Errorf("parseNewBlock:outputsData.Insert case nil: %s", err.Error())
				}

			}
		}

		// apear as input
		for _, input := range blockTxVerbose.Vin {
			hash, err := chainhash.NewHashFromStr(input.Txid)
			if err != nil {
				log.Errorf("parseNewBlock:outputsData.chainhash.NewHashFromStr: %s", err.Error())
			}
			previousTxVerbose, err := rpcClient.GetRawTransactionVerbose(hash)
			if err != nil {
				log.Errorf("parseNewBlock:rpcClient.GetRawTransactionVerbose: %s", err.Error())
				continue
			}

			for _, address := range previousTxVerbose.Vout[input.Vout].ScriptPubKey.Addresses {
				query := bson.M{"wallets.addresses.address": address}
				// Is it's our user transaction
				err := usersData.Find(query).One(&user)
				if err != nil {
					continue
					// Is not our user
				} else {
					log.Debugf("[ITS OUR USER] %s", user.UserID)
				}

				// Is our user already have transactions.
				err = txsData.Find(bson.M{"userid": user.UserID}).One(nil)
				if err == mgo.ErrNotFound {
					// Users first transaction.
					newRec := newTxRecord(user.UserID, blockTxVerbose.Txid, blockTxVerbose.Hash, previousTxVerbose.Vout[input.Vout].ScriptPubKey.Hex, address, TxStatusAppearedInBlockOutcoming, int(previousTxVerbose.Vout[input.Vout].N), blockHeight, previousTxVerbose.Vout[input.Vout].Value)
					err = txsData.Insert(newRec)
					if err != nil {
						log.Errorf("txsData.Insert: %s", err.Error())
					}
					continue
				} else if err != nil && err != mgo.ErrNotFound {
					log.Errorf("txsData.Find: %s", err.Error())
				}

				// Is our user already have this transactions.
				sel := bson.M{"userid": user.UserID, "transactions.txid": blockTxVerbose.Txid, "transactions.txaddress": address}
				err = txsData.Find(sel).One(nil)
				if err == mgo.ErrNotFound {
					// User have no transaction like this. Add to DB.
					newTx := newMultyTX(blockTxVerbose.Txid, blockTxVerbose.Hash, previousTxVerbose.Vout[input.Vout].ScriptPubKey.Hex, address, TxStatusAppearedInBlockOutcoming, int(previousTxVerbose.Vout[input.Vout].N), blockHeight, previousTxVerbose.Vout[input.Vout].Value)
					sel = bson.M{"userid": user.UserID}
					update := bson.M{"$push": bson.M{"transactions": newTx}}
					err = txsData.Update(sel, update)
					if err != nil {
						log.Errorf("txsData.Update add new tx to user: %s", err.Error())
					}
					continue
				} else if err != nil && err != mgo.ErrNotFound {
					log.Errorf("[ERR]txsData.Find: %s", err.Error())
				}

				// User have this transaction but with another status.
				// Update statsus and block height.
				sel = bson.M{"userid": user.UserID, "transactions.txid": blockTxVerbose.Txid, "transactions.txaddress": address}
				update := bson.M{
					"$set": bson.M{
						"transactions.$.txstatus":      TxStatusAppearedInBlockOutcoming,
						"transactions.$.txblockheight": blockHeight,
					},
				}
				err = txsData.Update(sel, update)
				if err != nil {
					log.Errorf("parseNewBlock:outputsData.Insert case nil: %s", err.Error())
				}

			}

		}
	}

}

type MultyTX struct {
	TxID          string  `json:"txid"`
	TxHash        string  `json:"txhash"`
	TxOutID       int     `json:"txoutid"`
	TxOutAmount   float64 `json:"txoutamount"`
	TxOutScript   string  `json:"txoutscript"`
	TxAddress     string  `json:"address"`
	TxBlockHeight int64   `json:"blockheight"`
	TxStatus      string  `json:"txstatus"`
}

type TxRecord struct {
	UserID       string    `json:"userid"`
	Transactions []MultyTX `json:"transactions"`
}

func newTxRecord(userID, txID, txHash, txOutScript, txAddress, txStatus string, txOutID int, txBlockHeight int64, txOutAmount float64) TxRecord {
	return TxRecord{
		UserID: userID,
		Transactions: []MultyTX{
			MultyTX{
				TxID:          txID,
				TxHash:        txHash,
				TxOutID:       txOutID,
				TxOutAmount:   txOutAmount,
				TxOutScript:   txOutScript,
				TxAddress:     txAddress,
				TxBlockHeight: txBlockHeight,
				TxStatus:      txStatus,
			},
		},
	}
}

func newMultyTX(txID, txHash, txOutScript, txAddress, txStatus string, txOutID int, txBlockHeight int64, txOutAmount float64) MultyTX {
	return MultyTX{
		TxID:          txID,
		TxHash:        txHash,
		TxOutID:       txOutID,
		TxOutAmount:   txOutAmount,
		TxOutScript:   txOutScript,
		TxAddress:     txAddress,
		TxBlockHeight: txBlockHeight,
		TxStatus:      txStatus,
	}
}

const (
	TxStatusAppearedInMempoolIncoming = "incoming in mempool"
	TxStatusAppearedInBlockIncoming   = "incoming in block"

	TxStatusAppearedInMempoolOutcoming = "spend in mempool"
	TxStatusAppearedInBlockOutcoming   = "spend in block"

	TxStatusInBlockConfirmed = "in block confirmed"
)

const (
	SixBlockConfirmation     = 6
	SixPlusBlockConfirmation = 7
)

func blockConfirmations(hash *chainhash.Hash) {
	blockVerbose, err := rpcClient.GetBlockVerbose(hash)
	blockHeight := blockVerbose.Height

	sel := bson.M{"transactions.txblockheight": bson.M{"$lte": blockHeight - SixBlockConfirmation, "$gte": blockHeight - SixPlusBlockConfirmation}}
	update := bson.M{
		"$set": bson.M{
			"transactions.$.txstatus": TxStatusInBlockConfirmed,
		},
	}
	err = txsData.Update(sel, update)
	if err != nil {
		log.Errorf("blockConfirmations:txsData.Update: %s", err.Error())
	}

	query := bson.M{"transactions.txblockheight": blockHeight + SixBlockConfirmation}

	var records []TxRecord
	txsData.Find(query).All(&records)
	for _, usertxs := range records {

		txMsq := BtcTransactionWithUserID{
			UserID: usertxs.UserID,
			NotificationMsg: &BtcTransaction{
				TransactionType: TxStatusInBlockConfirmed,
			},
		}
		sendNotifyToClients(&txMsq)
	}

}
