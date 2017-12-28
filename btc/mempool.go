package btc

import (
	"fmt"
	"time"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func parseMempoolTransaction(inTx *btcjson.TxRawResult) {
	var user store.User
	// parse every new transaction out from mempool and notify user with websocket
	for _, out := range inTx.Vout {
		for _, address := range out.ScriptPubKey.Addresses {

			query := bson.M{"wallets.addresses.address": address}
			err := usersData.Find(query).One(&user)
			if err != nil {
				continue
			}
			log.Debugf("[IS OUR USER] parseMempoolTransaction: usersData.Find = %s", address)

			txMsq := BtcTransactionWithUserID{
				UserID: user.UserID,
				NotificationMsg: &BtcTransaction{
					TransactionType: txInMempool,
					Amount:          out.Value,
					TxID:            inTx.Txid,
					Address:         address,
				},
			}
			sendNotifyToClients(&txMsq)
		}
	}

	// parse every new transaction in from mempool and notify user with websocket
	for _, input := range inTx.Vin {
		txHash, err := chainhash.NewHashFromStr(input.Txid)
		if err != nil {
			log.Errorf("parseMempoolTransaction: chainhash.NewHashFromStr = %s", err)
		}
		previousTx, err := rpcClient.GetRawTransactionVerbose(txHash)
		if err != nil {
			log.Errorf("parseMempoolTransaction:rpcClient.GetRawTransactionVerbose: %s", err.Error())
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
						TransactionType: txOutMempool,
						Amount:          out.Value,
						TxID:            inTx.Txid,
						Address:         address,
					},
				}
				sendNotifyToClients(&txMsq)
			}
		}
	}
}

func mempoolTransaction(inTx *btcjson.TxRawResult) {
	log.Debugf("[MEMPOOL TX]")
	var user store.User
	mempoolTimeUnixNano := time.Now().Unix()
	// apear as output
	for _, output := range inTx.Vout {
		for _, address := range output.ScriptPubKey.Addresses {
			query := bson.M{"wallets.addresses.address": address}
			err := usersData.Find(query).One(&user)
			if err != nil {
				// Is not our user.
				continue
			}
			fmt.Println("[ITS OUR USER] ", user.UserID)

			sel := bson.M{"userid": user.UserID, "transactions.txid": inTx.Txid, "transactions.txaddress": address}
			err = txsData.Find(sel).One(nil)
			if err == mgo.ErrNotFound {
				//appending transaction to user entity
				newTx := newMultyTX(inTx.Txid, inTx.Hash, output.ScriptPubKey.Hex, address, TxStatusAppearedInMempoolIncoming, output.Value, int(output.N), mempoolTimeUnixNano, -1, []StockExchangeRate{})
				sel = bson.M{"userid": user.UserID}
				update := bson.M{"$push": bson.M{"transactions": newTx}}
				err = txsData.Update(sel, update)
				if err != nil {
					log.Errorf("txsData.Update add new tx to user: %s", err.Error())
				}
				continue
			} else if err != nil && err != mgo.ErrNotFound {
				log.Errorf("mempoolTransaction: txsData.Find: %s", err.Error())
				continue
			}

			sel = bson.M{"userid": user.UserID, "transactions.txid": inTx.Txid, "transactions.txaddress": address}
			update := bson.M{
				"$set": bson.M{
					"transactions.$.txstatus":      TxStatusAppearedInMempoolIncoming,
					"transactions.$.txblockheight": -1,
				},
			}
			err = txsData.Update(sel, update)
			if err != nil {
				log.Errorf("mempoolTransaction: parseNewBlock:outputsData.Insert case nil: %s", err.Error())
			}
		}
	}

	// apear as input
	for _, input := range inTx.Vin {
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
				// Is not our user
			} else {
				log.Debugf("[ITS OUR USER] %s", user.UserID)
			}

			// Is our user already have this transactions.
			sel := bson.M{"userid": user.UserID, "transactions.txid": previousTxVerbose.Txid, "transactions.txaddress": address}
			err = txsData.Find(sel).One(nil)
			if err == mgo.ErrNotFound {
				// User have no transaction like this. Add to DB.
				newTx := newMultyTX(previousTxVerbose.Txid, previousTxVerbose.Hash, previousTxVerbose.Vout[input.Vout].ScriptPubKey.Hex, address, TxStatusAppearedInMempoolOutcoming, previousTxVerbose.Vout[input.Vout].Value, int(previousTxVerbose.Vout[input.Vout].N), mempoolTimeUnixNano, -1, []StockExchangeRate{})
				sel = bson.M{"userid": user.UserID}
				update := bson.M{"$push": bson.M{"transactions": newTx}}
				err = txsData.Update(sel, update)
				if err != nil {
					log.Errorf("txsData.Update add new tx to user: %s", err.Error())
				}
				continue
			} else if err != nil && err != mgo.ErrNotFound {
				log.Errorf("[ERR]txsData.Find: %s", err.Error())
				continue
			}

			// User have this transaction but with another status.
			// Update statsus and block height.
			sel = bson.M{"userid": user.UserID, "transactions.txid": previousTxVerbose.Txid, "transactions.txaddress": address}
			update := bson.M{
				"$set": bson.M{
					"transactions.$.txstatus":      TxStatusAppearedInMempoolOutcoming,
					"transactions.$.txblockheight": -1,
				},
			}
			err = txsData.Update(sel, update)
			if err != nil {
				log.Errorf("parseNewBlock:outputsData.Insert case nil: %s", err.Error())
			}
		}
	}

}
