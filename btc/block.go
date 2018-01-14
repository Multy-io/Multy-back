/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

const (
	TxStatusAppearedInMempoolIncoming = "incoming in mempool"
	TxStatusAppearedInBlockIncoming   = "incoming in block"

	TxStatusAppearedInMempoolOutcoming = "spend in mempool"
	TxStatusAppearedInBlockOutcoming   = "spend in block"

	TxStatusInBlockConfirmed = "in block confirmed"

	TxStatusRejectedFromBlock = "rejected block"
)

const (
	SixBlockConfirmation     = 6
	SixPlusBlockConfirmation = 7
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
						TransactionType: TxStatusAppearedInBlockIncoming,
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
							TransactionType: TxStatusAppearedInBlockOutcoming,
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

	for _, txHash := range allBlockTransactions {

		blockTxVerbose, err := rpcClient.GetRawTransactionVerbose(&txHash)
		if err != nil {
			log.Errorf("parseNewBlock:rpcClient.GetRawTransactionVerbose: %s", err.Error())
			continue
		}

		// apear as output
		err = parseOutput(blockTxVerbose, blockHeight, TxStatusInBlockConfirmed)
		if err != nil {
			log.Errorf("parseNewBlock:parseOutput: %s", err.Error())
		}

		// apear as input
		err = parseInput(blockTxVerbose, blockHeight, TxStatusAppearedInBlockOutcoming)
		if err != nil {
			log.Errorf("parseNewBlock:parseInput: %s", err.Error())
		}
	}
}

func blockConfirmations(hash *chainhash.Hash) {
	blockVerbose, err := rpcClient.GetBlockVerbose(hash)
	blockHeight := blockVerbose.Height

	sel := bson.M{"transactions.blockheight": bson.M{"$lte": blockHeight - SixBlockConfirmation, "$gte": blockHeight - SixPlusBlockConfirmation}}
	update := bson.M{
		"$set": bson.M{
			"transactions.$.txstatus": TxStatusInBlockConfirmed,
		},
	}
	err = txsData.Update(sel, update)
	if err != nil {
		log.Errorf("blockConfirmations:txsData.Update: %s", err.Error())
	}

	query := bson.M{"transactions.blockheight": blockHeight + SixBlockConfirmation}

	var records []store.TxRecord
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

func parseOutput(txVerbose *btcjson.TxRawResult, blockHeight int64, txStatus string) error {
	user := store.User{}
	blockTimeUnixNano := time.Now().Unix()

	for _, output := range txVerbose.Vout {
		for _, address := range output.ScriptPubKey.Addresses {
			query := bson.M{"wallets.addresses.address": address}
			err := usersData.Find(query).One(&user)
			if err != nil {
				continue
				// is not our user
			}
			fmt.Println("[ITS OUR USER] ", user.UserID)

			walletIndex := fetchWalletIndex(user.Wallets, address)

			sel := bson.M{"userID": user.UserID, "wallets.walletIndex": walletIndex}
			update := bson.M{
				"$set": bson.M{
					"wallets.$.status":         store.WalletStatusOK,
					"wallets.$.lastActionTime": time.Now().Unix(),
				},
			}

			err = usersData.Update(sel, update)
			if err != nil {
				log.Errorf("parseOutput:restClient.userStore.Update: %s", err.Error())
			}

			sel = bson.M{"userID": user.UserID, "wallets.addresses.address": address}
			update = bson.M{
				"$set": bson.M{
					"wallets." + strconv.Itoa(walletIndex) + ".addresses.$.data": time.Now().Unix(),
				},
			}
			err = usersData.Update(sel, update)
			if err != nil {
				log.Errorf("parseOutput:restClient.userStore.Update: %s", err.Error())
			}

			inputs, outputs, fee, err := txInfo(txVerbose)
			if err != nil {
				log.Errorf("parseInput:txInfo:output: %s", err.Error())
				continue
			}

			exRates, err := GetLatestExchangeRate()
			if err != nil {
				log.Errorf("parseOutput:GetLatestExchangeRate: %s", err.Error())
			}

			sel = bson.M{"userid": user.UserID, "transactions.txid": txVerbose.Txid, "transactions.txaddress": address}
			err = txsData.Find(sel).One(nil)
			if err == mgo.ErrNotFound {
				txOutAmount := int64(100000000 * output.Value)
				newTx := newMultyTX(txVerbose.Txid, txVerbose.Hash, output.ScriptPubKey.Hex, address, txStatus, int(output.N), walletIndex, txOutAmount, blockTimeUnixNano, blockHeight, fee, exRates, inputs, outputs)
				sel = bson.M{"userid": user.UserID}
				update := bson.M{"$push": bson.M{"transactions": newTx}}
				err = txsData.Update(sel, update)
				if err != nil {
					log.Errorf("parseInput.Update add new tx to user: %s", err.Error())
				}
				continue
			} else if err != nil && err != mgo.ErrNotFound {
				log.Errorf("parseInput:txsData.Find: %s", err.Error())
				continue
			}

			sel = bson.M{"userid": user.UserID, "transactions.txid": txVerbose.Txid, "transactions.txaddress": address}
			update = bson.M{
				"$set": bson.M{
					"transactions.$.txstatus":          txStatus,
					"transactions.$.blockheight":       blockHeight,
					"transactions.$.txfee":             fee,
					"transactions.$.stockexchangerate": exRates,
					"transactions.$.txinputs":          inputs,
					"transactions.$.txoutputs":         outputs,
					"transactions.$.blocktime":         blockTimeUnixNano,
				},
			}

			err = txsData.Update(sel, update)
			if err != nil {
				log.Errorf("parseInput:outputsData.Insert case nil: %s", err.Error())
			}
		}
	}
	return nil
}

func parseInput(txVerbose *btcjson.TxRawResult, blockHeight int64, txStatus string) error {
	user := store.User{}
	blockTimeUnixNano := time.Now().Unix()

	for _, input := range txVerbose.Vin {

		previousTxVerbose, err := rawTxByTxid(input.Txid)
		if err != nil {
			log.Errorf("parseInput:rawTxByTxid: %s", err.Error())
			continue
		}

		for _, address := range previousTxVerbose.Vout[input.Vout].ScriptPubKey.Addresses {
			query := bson.M{"wallets.addresses.address": address}
			// Is it's our user transaction
			err := usersData.Find(query).One(&user)
			if err != nil {
				continue
				// Is not our user
			}

			log.Debugf("[ITS OUR USER] %s", user.UserID)

			inputs, outputs, fee, err := txInfo(txVerbose)
			if err != nil {
				log.Errorf("parseInput:txInfo:input: %s", err.Error())
				continue
			}
			exRates, err := GetLatestExchangeRate()
			if err != nil {
				log.Errorf("parseOutput:GetLatestExchangeRate: %s", err.Error())
			}

			walletIndex := fetchWalletIndex(user.Wallets, address)

			sel := bson.M{"userID": user.UserID, "wallets.walletIndex": walletIndex}
			update := bson.M{
				"$set": bson.M{
					"wallets.$.lastActionTime": time.Now().Unix(),
				},
			}
			err = usersData.Update(sel, update)
			if err != nil {
				log.Errorf("parseOutput:restClient.userStore.Update: %s", err.Error())
			}

			sel = bson.M{"userID": user.UserID, "wallets.addresses.address": address}
			update = bson.M{
				"$set": bson.M{
					"wallets." + strconv.Itoa(walletIndex) + ".addresses.$.data": time.Now().Unix(),
				},
			}
			err = usersData.Update(sel, update)
			if err != nil {
				log.Errorf("parseOutput:restClient.userStore.Update: %s", err.Error())
			}

			// Is our user already have this transactions.
			sel = bson.M{"userid": user.UserID, "transactions.txid": txVerbose.Txid, "transactions.txaddress": address}
			err = txsData.Find(sel).One(nil)
			if err == mgo.ErrNotFound {
				// User have no transaction like this. Add to DB.
				txOutAmount := int64(100000000 * previousTxVerbose.Vout[input.Vout].Value)
				newTx := newMultyTX(txVerbose.Txid, txVerbose.Hash, previousTxVerbose.Vout[input.Vout].ScriptPubKey.Hex, address, txStatus, int(previousTxVerbose.Vout[input.Vout].N), walletIndex, txOutAmount, blockTimeUnixNano, blockHeight, fee, exRates, inputs, outputs)
				sel = bson.M{"userid": user.UserID}
				update := bson.M{"$push": bson.M{"transactions": newTx}}
				err = txsData.Update(sel, update)
				if err != nil {
					log.Errorf("parseInput:txsData.Update add new tx to user: %s", err.Error())
				}
				continue
			} else if err != nil && err != mgo.ErrNotFound {
				log.Errorf("parseInput:txsData.Find: %s", err.Error())
				continue
			}

			// User have this transaction but with another status.
			// Update statsus, block height, exchange rate,block time, inputs and outputs.
			sel = bson.M{"userid": user.UserID, "transactions.txid": txVerbose.Txid, "transactions.txaddress": address}
			update = bson.M{
				"$set": bson.M{
					"transactions.$.txstatus":    txStatus,
					"transactions.$.blockheight": blockHeight,
					"transactions.$.txinputs":    inputs,
					"transactions.$.txoutputs":   outputs,
					"transactions.$.blocktime":   blockTimeUnixNano,
				},
			}
			err = txsData.Update(sel, update)
			if err != nil {
				log.Errorf("parseInput:txsData.Update: %s", err.Error())
			}
		}
	}
	return nil
}

func GetLatestExchangeRate() ([]store.ExchangeRatesRecord, error) {
	selGdax := bson.M{
		"stockexchange": "Gdax",
	}
	selPoloniex := bson.M{
		"stockexchange": "Poloniex",
	}
	stocksGdax := store.ExchangeRatesRecord{}
	err := exRate.Find(selGdax).Sort("-timestamp").One(&stocksGdax)
	if err != nil {
		return nil, err
	}

	stocksPoloniex := store.ExchangeRatesRecord{}
	err = exRate.Find(selPoloniex).Sort("-timestamp").One(&stocksPoloniex)
	if err != nil {
		return nil, err
	}
	return []store.ExchangeRatesRecord{stocksPoloniex, stocksGdax}, nil

}
