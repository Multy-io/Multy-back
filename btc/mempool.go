package btc

import (
	"fmt"

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
	fmt.Println("[MEMPOOL TX] ----------")
	var user store.User
	// apear as output
	for _, output := range inTx.Vout {
		for _, address := range output.ScriptPubKey.Addresses {
			query := bson.M{"wallets.addresses.address": address}
			err := usersData.Find(query).One(&user)
			if err != nil {
				continue
			}

			err = txsData.Find(bson.M{"userid": user.UserID}).One(nil)
			switch err {
			case mgo.ErrNotFound:
				// User first transaction. Create new record in db.
				newRec := newTxRecord(user.UserID, inTx.Txid, inTx.Hash, output.ScriptPubKey.Hex, address, TxStatusAppearedInMempoolIncoming, int(output.N), -1, output.Value)
				err = txsData.Insert(newRec)
				if err != nil {
					log.Errorf("parseNewBlock:outputsData.Insert mgo.ErrNotFound: %s", err.Error())
				}
			case nil:
				sel := bson.M{"userid": user.UserID, "transactions.txid": inTx.Txid, "transactions.txaddress": address}
				err = txsData.Find(sel).One(nil)
				switch err {
				case mgo.ErrNotFound:
					//no record like this, create new
					newTx := newMultyTX(inTx.Txid, inTx.Hash, output.ScriptPubKey.Hex, address, TxStatusAppearedInMempoolIncoming, int(output.N), -1, output.Value)
					sel = bson.M{"userid": user.UserID}
					update := bson.M{"$push": bson.M{"transactions": newTx}}
					err = txsData.Update(sel, update)
					if err != nil {
						log.Errorf("parseNewBlock:outputsData.Insert case nil: %s", err.Error())
					}
				case nil:
					//record fetched, change status
					sel = bson.M{"userid": user.UserID, "transactions.txid": inTx.Txid, "transactions.txaddress": address}
					update := bson.M{
						"$set": bson.M{
							"transactions.$.txstatus":      TxStatusAppearedInMempoolIncoming,
							"transactions.$.txblockheight": -1,
						},
					}
					err = txsData.Update(sel, update)
					if err != nil {
						log.Errorf("parseNewBlock:outputsData.Insert case nil: %s", err.Error())
					}
				default:
					log.Errorf("parseNewBlock:outputsData.Insert default: %s", err.Error())
					//handle error
				}

			default:
				//handle err
				log.Errorf("parseNewBlock:outputsData.Insert default: %s", err.Error())
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
			}

			err = txsData.Find(bson.M{"userid": user.UserID}).One(nil)
			switch err {
			case mgo.ErrNotFound:
				// add new
				newRec := newTxRecord(user.UserID, inTx.Txid, inTx.Hash, previousTxVerbose.Vout[input.Vout].ScriptPubKey.Hex, address, TxStatusAppearedInMempoolOutcoming, int(previousTxVerbose.Vout[input.Vout].N), -1, previousTxVerbose.Vout[input.Vout].Value)
				err = txsData.Insert(newRec)
				if err != nil {
					log.Errorf("parseNewBlock:outputsData.Insert mgo.ErrNotFound: %s", err.Error())
				}
			case nil:
				// append
				sel := bson.M{"userid": user.UserID, "transactions.txid": inTx.Txid, "transactions.txaddress": address}
				err = txsData.Find(sel).One(nil)
				switch err {
				case mgo.ErrNotFound:
					newTx := newMultyTX(inTx.Txid, inTx.Hash, previousTxVerbose.Vout[input.Vout].ScriptPubKey.Hex, address, TxStatusAppearedInMempoolOutcoming, int(previousTxVerbose.Vout[input.Vout].N), -1, previousTxVerbose.Vout[input.Vout].Value)
					sel = bson.M{"userid": user.UserID}
					update := bson.M{"$push": bson.M{"transactions": newTx}}
					err = txsData.Update(sel, update)
					if err != nil {
						log.Errorf("parseNewBlock:outputsData.Insert case mgo.ErrNotFound: %s", err.Error())
					}
				case nil:
					// found. nothing to append
					//	update status
					sel = bson.M{"userid": user.UserID, "transactions.txid": inTx.Txid, "transactions.txaddress": address}
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
				default:
					// error
					log.Errorf("parseNewBlock:outputsData.Insert default: %s", err.Error())
				}
			default:
				//handle err
				log.Errorf("parseNewBlock:outputsData.Insert default: %s", err.Error())
			}

		}

	}
}
