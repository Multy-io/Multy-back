package btc

import (
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
)

func parseMempoolTransaction(inTx *btcjson.TxRawResult) error {
	for num, out := range inTx.Vout {
		fmt.Println("addres", num, out.ScriptPubKey.Addresses)
	}

	return nil
}

/*
//Here we parsing transaction by getting inputs and outputs addresses
func parseRawMempool(inTx *btcjson.TxRawResult) error {
	memPoolTx := MultyMempoolTx{size: inTx.Size, hash: inTx.Hash, txid: inTx.Txid}

	inputs := inTx.Vin

	var inputSum, outputSum float64 = 0, 0

	for j := 0; j < len(inputs); j++ {
		input := inputs[j]

		inputNum := input.Vout

		txCHash, errCHash := chainhash.NewHashFromStr(input.Txid)

		if errCHash != nil {
			log.Fatal(errCHash)
		}

		oldTx, err := rpcClient.GetRawTransactionVerbose(txCHash)
		if err != nil {
			log.Println("ERR GetRawTransactionVerbose [old]: ", err.Error())
			return err
		}

		oldOutputs := oldTx.Vout

		inputSum += oldOutputs[inputNum].Value

		addressesInputs := oldOutputs[inputNum].ScriptPubKey.Addresses

		inputAdr := MultyAddress{addressesInputs, oldOutputs[inputNum].Value}

		memPoolTx.inputs = append(memPoolTx.inputs, inputAdr)
	}

	outputs := inTx.Vout

	var txOutputs []MultyAddress

	for _, output := range outputs {
		addressesOuts := output.ScriptPubKey.Addresses
		outputSum += output.Value

		txOutputs = append(txOutputs, MultyAddress{addressesOuts, output.Value})
	}
	memPoolTx.outputs = txOutputs

	memPoolTx.amount = inputSum
	memPoolTx.fee = inputSum - outputSum

	memPoolTx.feeRate = int32(memPoolTx.fee / float64(memPoolTx.size) * 100000000)

	// log.Printf("\n **************************** Multy-New Tx Found *******************\n hash: %s, id: %s \n amount: %f , fee: %f , feeRate: %d \n Inputs: %v \n OutPuts: %v \n ****************************Multy-the best wallet*******************", memPoolTx.hash, memPoolTx.txid, memPoolTx.amount, memPoolTx.fee, memPoolTx.feeRate, memPoolTx.inputs, memPoolTx.outputs)
	// memPoolTx.hash, memPoolTx.txid, memPoolTx.amount, memPoolTx.fee, memPoolTx.feeRate, memPoolTx.inputs, memPoolTx.outputs

	var user store.User

	for _, input := range memPoolTx.inputs {
		for _, address := range input.address {
			usersData.Find(bson.M{"wallets.addresses.address": address}).One(&user)
			if user.Wallets != nil {
				chToClient <- CreateBtcTransactionWithUserID(user.UserID, txIn, "not implemented", memPoolTx.hash, input.amount)
				// add UserID related tx's to db
				// rec := newTxInfo(txIn, memPoolTx.hash, address, input.amount)
				// sel := bson.M{"userID": user.UserID}
				// update := bson.M{"$push": bson.M{"transactions": rec}}
				// err := usersData.Update(sel, update)
				// if err != nil {
				// 	fmt.Println(err)
				// }
				// // TODO: parse block
			}
			user = store.User{}
		}
	}

	for _, output := range memPoolTx.outputs {
		for _, address := range output.address {
			usersData.Find(bson.M{"wallets.addresses.address": address}).One(&user)
			if user.Wallets != nil {
				chToClient <- CreateBtcTransactionWithUserID(user.UserID, txOut, "not implemented", memPoolTx.hash, output.amount)
				// add UserID related tx's to db

				// rec := newTxInfo(txOut, memPoolTx.hash, address, output.amount)
				// sel := bson.M{"userID": user.UserID}
				// update := bson.M{"$push": bson.M{"transactions": rec}}
				// err := usersData.Update(sel, update)
				// if err != nil {
				// 	fmt.Println(err)
				// }
				// // TODO: parse block
			}
			user = store.User{}
		}
	}

	rec := newRecord(int(memPoolTx.feeRate), memPoolTx.hash)

	err := mempoolRates.Insert(rec)
	if err != nil {
		log.Println("ERR mempoolRates.Insert: ", err.Error())
		return err
	}

	//TODO save transaction as mem pool tx
	//TODO update fee rates table
	memPool = append(memPool, memPoolTx)

	log.Printf("[DEBUG] parseRawTransaction: new multy mempool; size=%d", len(memPool))
	return nil
}
*/
