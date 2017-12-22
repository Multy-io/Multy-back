package btc

import (
	"encoding/json"

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
				}

				var rec TxRecord
				err = txsData.Find(bson.M{"userid": user.UserID}).One(nil)
				switch err {
				case mgo.ErrNotFound:
					// add new
					log.Debugf("New user with tx created. with status %s \n", TxStatusAppearedInBlockOutcoming)
					log.Debugf("On block height %d \n", blockHeight)
					newRec := newTxRecord(user.UserID, blockTxVerbose.Txid, blockTxVerbose.Hash, output.ScriptPubKey.Hex, address, TxStatusAppearedInBlockIncoming, int(output.N), blockHeight, output.Value)
					err = txsData.Insert(newRec)
					if err != nil {
						log.Errorf("parseNewBlock:outputsData.Insert mgo.ErrNotFound: %s", err.Error())
					}
				case nil:
					// append
					sel := bson.M{"userid": user.UserID, "transactions.txid": blockTxVerbose.Txid, "transactions.txaddress": address}
					err = txsData.Find(sel).One(rec)
					switch err {
					case mgo.ErrNotFound:
						//no record like this, create new
						log.Debugf("Tx created with status %s \n", TxStatusAppearedInBlockIncoming)
						log.Debugf("On block height %d \n", blockHeight)
						newTx := newMultyTX(blockTxVerbose.Txid, blockTxVerbose.Hash, output.ScriptPubKey.Hex, address, TxStatusAppearedInBlockIncoming, int(output.N), blockHeight, output.Value)
						sel = bson.M{"userid": rec.UserID}
						update := bson.M{"$push": bson.M{"transactions": newTx}}
						err = txsData.Update(sel, update)
						if err != nil {
							log.Errorf("parseNewBlock:outputsData.Insert case nil: %s", err.Error())
						}
					case nil:
						//record fetched, change status
						log.Debugf("Status changed to %s \n", TxStatusAppearedInBlockIncoming)
						log.Debugf("On block height %d \n", blockHeight)
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
					default:
						//handle error
					}

				default:
					//handle err
					log.Errorf("parseNewBlock:outputsData.Insert default: %s", err.Error())
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
				err := usersData.Find(query).One(&user)
				if err != nil {
					continue
				}

				err = txsData.Find(bson.M{"userid": user.UserID}).One(nil)
				switch err {
				case mgo.ErrNotFound:
					// add new
					log.Debugf("New user with tx created. with status %s \n", TxStatusAppearedInBlockOutcoming)
					log.Debugf("On block height %d \n", blockHeight)
					newRec := newTxRecord(user.UserID, blockTxVerbose.Txid, blockTxVerbose.Hash, previousTxVerbose.Vout[input.Vout].ScriptPubKey.Hex, address, TxStatusAppearedInBlockOutcoming, int(previousTxVerbose.Vout[input.Vout].N), blockHeight, previousTxVerbose.Vout[input.Vout].Value)
					err = txsData.Insert(newRec)
					if err != nil {
						log.Errorf("parseNewBlock:outputsData.Insert mgo.ErrNotFound: %s", err.Error())
					}
				case nil:
					// Tx fetched update status to spend of or push if not exists in db.
					sel := bson.M{"userid": user.UserID, "transactions.txid": blockTxVerbose.Txid, "transactions.txaddress": address}
					err = txsData.Find(sel).One(nil)
					switch err {
					case mgo.ErrNotFound:
						log.Debugf("Tx created with status %s \n", TxStatusAppearedInBlockOutcoming)
						log.Debugf("On block height %d \n", blockHeight)
						// Tx record is not exist case. Create record and push to user entity.
						newTx := newMultyTX(blockTxVerbose.Txid, blockTxVerbose.Hash, previousTxVerbose.Vout[input.Vout].ScriptPubKey.Hex, address, TxStatusAppearedInBlockOutcoming, int(previousTxVerbose.Vout[input.Vout].N), blockHeight, previousTxVerbose.Vout[input.Vout].Value)
						sel = bson.M{"userid": user.UserID}
						update := bson.M{"$push": bson.M{"transactions": newTx}}
						err = txsData.Update(sel, update)
						if err != nil {
							log.Errorf("parseNewBlock:outputsData.Insert case nil: %s", err.Error())
						}
					case nil:
						log.Debugf("Status changed to %s \n", TxStatusAppearedInBlockOutcoming)
						log.Debugf("On block height %d \n", blockHeight)
						// Tx record exists case. Update record status.
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
					default:
						// Handle db error.
						log.Errorf("parseNewBlock:outputsData.Insert default: %s", err.Error())
					}

				default:
					//handle err
					log.Errorf("parseNewBlock:outputsData.Insert default: %s", err.Error())
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
