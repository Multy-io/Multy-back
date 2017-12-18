package btc

import (
	"log"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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
			log.Printf("[DEBUG] [IS OUR USER] parseMempoolTransaction: usersData.Find = %s", address)

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
			log.Printf("[ERR] parseMempoolTransaction: chainhash.NewHashFromStr = %s", err)
		}
		previousTx, err := rpcClient.GetRawTransactionVerbose(txHash)
		if err != nil {
			log.Printf("[ERR] parseMempoolTransaction:rpcClient.GetRawTransactionVerbose: %s ", err.Error())
		}

		for _, out := range previousTx.Vout {
			for _, address := range out.ScriptPubKey.Addresses {
				query := bson.M{"wallets.addresses.address": address}
				err := usersData.Find(query).One(&user)
				if err != nil {
					continue
				}
				log.Printf("[DEBUG] [IS OUR USER]-AS-OUT parseMempoolTransaction: usersData.Find = %s", address)

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
