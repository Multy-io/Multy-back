package btc

import (
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
			log.Debugf("[IS OUR USER] parseMempoolTransaction: usersData.Find = %s", address)

			txMsq := BtcTransactionWithUserID{
				UserID: user.UserID,
				NotificationMsg: &BtcTransaction{
					TransactionType: TxStatusAppearedInMempoolIncoming,
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
						TransactionType: TxStatusAppearedInMempoolOutcoming,
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

	// apear as output
	err := parseOutput(inTx, -1, TxStatusAppearedInMempoolIncoming)
	if err != nil {
		log.Errorf("mempoolTransaction:parseOutput: %s", err.Error())
	}

	// apear as input
	err = parseInput(inTx, -1, TxStatusAppearedInMempoolOutcoming)
	if err != nil {
		log.Errorf("mempoolTransaction:parseInput: %s", err.Error())
	}

}
