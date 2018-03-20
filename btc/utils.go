/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"fmt"
	"time"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/btcsuite/btcd/btcjson"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	exRate    *mgo.Collection
	usersData *mgo.Collection

	mempoolRates     *mgo.Collection
	txsData          *mgo.Collection
	spendableOutputs *mgo.Collection

	mempoolRatesTest     *mgo.Collection
	txsDataTest          *mgo.Collection
	spendableOutputsTest *mgo.Collection
)

func newAddresAmount(address string, amount int64) store.AddresAmount {
	return store.AddresAmount{
		Address: address,
		Amount:  amount,
	}
}

func fetchWalletAndAddressIndexes(wallets []store.Wallet, address string) (int, int) {
	var walletIndex int
	var addressIndex int
	for _, wallet := range wallets {
		for _, addr := range wallet.Adresses {
			if addr.Address == address {
				walletIndex = wallet.WalletIndex
				addressIndex = addr.AddressIndex
				break
			}
		}
	}
	return walletIndex, addressIndex
}

/*

Main process BTC transaction method

can be called from:
	- Mempool
	- New block
	- Re-sync

*/

// HACK is a wrapper for processTransaction. in future it will be in separate file
func ProcessTransaction(blockChainBlockHeight int64, txVerbose *btcjson.TxRawResult, isReSync bool) {
	processTransaction(blockChainBlockHeight, txVerbose, isReSync)
}
func processTransaction(blockChainBlockHeight int64, txVerbose *btcjson.TxRawResult, isReSync bool) {
	var multyTx *store.MultyTX = parseRawTransaction(blockChainBlockHeight, txVerbose)
	CreateSpendableOutputs(txVerbose, blockChainBlockHeight)
	DeleteSpendableOutputs(txVerbose, blockChainBlockHeight)
	if multyTx != nil {
		multyTx.BlockHeight = blockChainBlockHeight

		setExchangeRates(multyTx, isReSync, txVerbose.Time)

		log.Debugf("processTransaction:setTransactionInfo %v", multyTx)

		transactions := splitTransaction(*multyTx, blockChainBlockHeight)
		log.Debugf("processTransaction:splitTransaction %v", transactions)

		for _, transaction := range transactions {

			finalizeTransaction(&transaction, txVerbose)
			setUserID(&transaction)
			SaveMultyTransaction(transaction)
			updateWalletAndAddressDate(transaction)
			// if !isReSync {
			// 	go func() {
			// 		select {
			// 		case <-time.After(time.Second):
			// 			go sendNotifyToClients(transaction)
			// 		}
			// 	}()
			// }
		}
	}
}

func setUserID(tx *store.MultyTX) {
	user := store.User{}
	for _, address := range tx.TxAddress {
		query := bson.M{"wallets.addresses.address": address}
		err := usersData.Find(query).One(&user)
		if err != nil {
			log.Errorf("setUserID: usersData.Find: %s", err.Error())
		}
		tx.UserId = user.UserID
	}
}

/*
This method should parse raw transaction from BTC node

_________________________
Inputs:
* blockChainBlockHeight int64 - could be:
-1 in case of mempool call
-1 in case of block transaction
max chain height in case of resync

*txVerbose - raw BTC transaction
_________________________
Output:
* multyTX - multy transaction Structure

*/
func parseRawTransaction(blockChainBlockHeight int64, txVerbose *btcjson.TxRawResult) *store.MultyTX {
	multyTx := store.MultyTX{}

	err := parseInputs(txVerbose, blockChainBlockHeight, &multyTx)
	if err != nil {
		log.Errorf("parseRawTransaction:parseInputs: %s", err.Error())
	}

	err = parseOutputs(txVerbose, blockChainBlockHeight, &multyTx)
	if err != nil {
		log.Errorf("parseRoawTransaction:parseOutputs: %s", err.Error())
	}

	if multyTx.TxID != "" {
		return &multyTx
	} else {
		return nil
	}
}

/*
This method need if we have one transaction with more the one u wser'sallet
That means that from one btc transaction we should build more the one Multy Transaction
*/
func splitTransaction(multyTx store.MultyTX, blockHeight int64) []store.MultyTX {
	// transactions := make([]store.MultyTX, 1)
	transactions := []store.MultyTX{}

	// currentBlockHeight, err := rpcClient.GetBlockCount()
	currentBlockHeight := int64(0)

	blockDiff := currentBlockHeight - blockHeight

	//This is implementatios for single wallet transaction for multi addresses not for multi wallets!
	if multyTx.WalletsInput != nil && len(multyTx.WalletsInput) > 0 {
		outgoingTx := newEntity(multyTx)
		//By that code we are erasing WalletOutputs for new splited transaction
		outgoingTx.WalletsOutput = []store.WalletForTx{}

		for _, walletOutput := range multyTx.WalletsOutput {
			var isTheSameWallet = false
			for _, walletInput := range outgoingTx.WalletsInput {
				if walletInput.UserId == walletOutput.UserId && walletInput.WalletIndex == walletOutput.WalletIndex {
					isTheSameWallet = true
				}
			}
			if isTheSameWallet {
				outgoingTx.WalletsOutput = append(outgoingTx.WalletsOutput, walletOutput)
			}
		}

		setTransactionStatus(&outgoingTx, blockDiff, currentBlockHeight, true)
		transactions = append(transactions, outgoingTx)
	}

	if multyTx.WalletsOutput != nil && len(multyTx.WalletsOutput) > 0 {
		for _, walletOutput := range multyTx.WalletsOutput {
			var alreadyAdded = false
			for i := 0; i < len(transactions); i++ {
				//Check if our output wallet is in the inputs
				//var walletOutputExistInInputs = false
				if transactions[i].WalletsInput != nil && len(transactions[i].WalletsInput) > 0 {
					for _, splitedInput := range transactions[i].WalletsInput {
						if splitedInput.UserId == walletOutput.UserId && splitedInput.WalletIndex == walletOutput.WalletIndex {
							alreadyAdded = true
						}
					}
				}

				if transactions[i].WalletsOutput != nil && len(transactions[i].WalletsOutput) > 0 {
					for j := 0; j < len(transactions[i].WalletsOutput); j++ {
						if walletOutput.UserId == transactions[i].WalletsOutput[j].UserId && walletOutput.WalletIndex == transactions[i].WalletsOutput[j].WalletIndex { //&& walletOutput.Address.Address != transactions[i].WalletsOutput[j].Address.Address Don't think this ckeck we need
							//We have the same wallet index in output but different addres
							alreadyAdded = true
							if &transactions[i] == nil {
								transactions[i].WalletsOutput = append(transactions[i].WalletsOutput, walletOutput)
								log.Errorf("splitTransaction error allocate memory")
							}
							log.Errorf("splitTransaction ! no ! error allocate memory")
						}
					}
				}

			}

			if alreadyAdded {
				continue
			} else {
				//Add output transaction here
				incomingTx := newEntity(multyTx)
				incomingTx.WalletsInput = nil
				incomingTx.WalletsOutput = []store.WalletForTx{}
				incomingTx.WalletsOutput = append(incomingTx.WalletsOutput, walletOutput)
				setTransactionStatus(&incomingTx, blockDiff, currentBlockHeight, false)
				transactions = append(transactions, incomingTx)
			}
		}

	}

	return transactions
}

func newEntity(multyTx store.MultyTX) store.MultyTX {
	newTx := store.MultyTX{
		TxID:              multyTx.TxID,
		TxHash:            multyTx.TxHash,
		TxOutScript:       multyTx.TxOutScript,
		TxAddress:         multyTx.TxAddress,
		TxStatus:          multyTx.TxStatus,
		TxOutAmount:       multyTx.TxOutAmount,
		BlockTime:         multyTx.BlockTime,
		BlockHeight:       multyTx.BlockHeight,
		TxFee:             multyTx.TxFee,
		MempoolTime:       multyTx.MempoolTime,
		StockExchangeRate: multyTx.StockExchangeRate,
		TxInputs:          multyTx.TxInputs,
		TxOutputs:         multyTx.TxOutputs,
		WalletsInput:      multyTx.WalletsInput,
		WalletsOutput:     multyTx.WalletsOutput,
	}
	return newTx
}

func SaveMultyTransaction(tx store.MultyTX) {
	// This is splited transaction! That means that transaction's WalletsInputs and WalletsOutput have the same WalletIndex!
	//Here we have outgoing transaction for exact wallet!
	multyTX := store.MultyTX{}
	if tx.WalletsInput != nil && len(tx.WalletsInput) > 0 {
		// sel := bson.M{"userid": tx.WalletsInput[0].UserId, "transactions.txid": tx.TxID, "transactions.walletsinput.walletindex": tx.WalletsInput[0].WalletIndex}
		sel := bson.M{"userid": tx.WalletsInput[0].UserId, "txid": tx.TxID, "walletsinput.walletindex": tx.WalletsInput[0].WalletIndex}
		err := txsData.Find(sel).One(&multyTX)
		if err == mgo.ErrNotFound {
			// initial insertion
			err := txsData.Insert(tx)
			if err != nil {
				log.Errorf("parseInput.Update add new tx to user: %s", err.Error())
			}
			return
		}
		if err != nil && err != mgo.ErrNotFound {
			// database error
			log.Errorf("saveMultyTransaction:txsData.Find %s", err.Error())
			return
		}

		update := bson.M{
			"$set": bson.M{
				"txstatus":      tx.TxStatus,
				"blockheight":   tx.BlockHeight,
				"confirmations": tx.Confirmations,
				"blocktime":     tx.BlockTime,
			},
		}
		err = txsData.Update(sel, update)
		if err != nil {
			log.Errorf("saveMultyTransaction:txsData.Update %s", err.Error())
		}
		return
	} else if tx.WalletsOutput != nil && len(tx.WalletsOutput) > 0 {
		// sel := bson.M{"userid": tx.WalletsOutput[0].UserId, "transactions.txid": tx.TxID, "transactions.walletsoutput.walletindex": tx.WalletsOutput[0].WalletIndex}
		sel := bson.M{"userid": tx.WalletsOutput[0].UserId, "txid": tx.TxID, "walletsoutput.walletindex": tx.WalletsOutput[0].WalletIndex}
		err := txsData.Find(sel).One(&multyTX)
		if err == mgo.ErrNotFound {
			// initial insertion
			err := txsData.Insert(tx)
			if err != nil {
				log.Errorf("parseInput.Update add new tx to user: %s", err.Error())
			}
			return
		}
		if err != nil && err != mgo.ErrNotFound {
			// database error
			log.Errorf("saveMultyTransaction:txsData.Find %s", err.Error())
			return
		}

		update := bson.M{
			"$set": bson.M{
				"txstatus":      tx.TxStatus,
				"blockheight":   tx.BlockHeight,
				"confirmations": tx.Confirmations,
				"blocktime":     tx.BlockTime,
			},
		}
		err = txsData.Update(sel, update)
		if err != nil {
			log.Errorf("saveMultyTransaction:txsData.Update %s", err.Error())
		}
		return
	}
}

// func sendNotify(txMsq *BtcTransactionWithUserID) {
// 	newTxJSON, err := json.Marshal(txMsq)
// 	if err != nil {
// 		log.Errorf("sendNotifyToClients: [%+v] %s\n", txMsq, err.Error())
// 		return
// 	}

// 	err = nsqProducer.Publish(TopicTransaction, newTxJSON)
// 	if err != nil {
// 		log.Errorf("nsq publish new transaction: [%+v] %s\n", txMsq, err.Error())
// 		return
// 	}
// 	return
// }

// func sendNotifyToClients(tx store.MultyTX) {

// 	for _, walletOutput := range tx.WalletsOutput {
// 		txMsq := BtcTransactionWithUserID{
// 			UserID: walletOutput.UserId,
// 			NotificationMsg: &BtcTransaction{
// 				TransactionType: tx.TxStatus,
// 				Amount:          tx.TxOutAmount,
// 				TxID:            tx.TxID,
// 				Address:         walletOutput.Address.Address,
// 			},
// 		}
// 		sendNotify(&txMsq)
// 	}

// 	for _, walletInput := range tx.WalletsInput {
// 		txMsq := BtcTransactionWithUserID{
// 			UserID: walletInput.UserId,
// 			NotificationMsg: &BtcTransaction{
// 				TransactionType: tx.TxStatus,
// 				Amount:          tx.TxOutAmount,
// 				TxID:            tx.TxID,
// 				Address:         walletInput.Address.Address,
// 			},
// 		}
// 		sendNotify(&txMsq)
// 	}
// }

func parseInputs(txVerbose *btcjson.TxRawResult, blockHeight int64, multyTx *store.MultyTX) error {
	//NEW LOGIC
	// user := store.User{}
	//Ranging by inputs
	/*
		for _, input := range txVerbose.Vin {

		//getting previous verbose transaction from BTC Node for checking addresses
		// previousTxVerbose, err := rawTxByTxid(input.Txid)
		// if err != nil {
		// 	log.Errorf("parseInput:rawTxByTxid: %s", err.Error())
		// 	continue
		// }


					for _, txInAddress := range previousTxVerbose.Vout[input.Vout].ScriptPubKey.Addresses {
						query := bson.M{"wallets.addresses.address": txInAddress}

						err := usersData.Find(query).One(&user)
						if err != nil {
							continue
							// is not our user
						}
						fmt.Println("[ITS OUR USER] ", user.UserID)

						walletIndex, addressIndex := fetchWalletAndAddressIndexes(user.Wallets, txInAddress)

						txInAmount := int64(SatoshiInBitcoint * previousTxVerbose.Vout[input.Vout].Value)

						currentWallet := store.WalletForTx{UserId: user.UserID, WalletIndex: walletIndex}

						if multyTx.WalletsInput == nil {
							multyTx.WalletsInput = []store.WalletForTx{}
						}

						currentWallet.Address = store.AddressForWallet{Address: txInAddress, AddressIndex: addressIndex, Amount: txInAmount}
						multyTx.WalletsInput = append(multyTx.WalletsInput, currentWallet)

						multyTx.TxID = txVerbose.Txid
						multyTx.TxHash = txVerbose.Hash

					}

			}
	*/
	return nil
}

func parseOutputs(txVerbose *btcjson.TxRawResult, blockHeight int64, multyTx *store.MultyTX) error {

	user := store.User{}

	for _, output := range txVerbose.Vout {
		for _, txOutAddress := range output.ScriptPubKey.Addresses {
			query := bson.M{"wallets.addresses.address": txOutAddress}

			err := usersData.Find(query).One(&user)
			if err != nil {
				continue
				// is not our user
			}
			fmt.Println("[ITS OUR USER] ", user.UserID)

			walletIndex, addressIndex := fetchWalletAndAddressIndexes(user.Wallets, txOutAddress)

			currentWallet := store.WalletForTx{UserId: user.UserID, WalletIndex: walletIndex}

			if multyTx.TxOutputs == nil {
				multyTx.TxOutputs = []store.AddresAmount{}
			}

			if multyTx.WalletsOutput == nil {
				multyTx.WalletsOutput = []store.WalletForTx{}
			}

			currentWallet.Address = store.AddressForWallet{Address: txOutAddress, AddressIndex: addressIndex, Amount: int64(SatoshiInBitcoint * output.Value)}
			multyTx.WalletsOutput = append(multyTx.WalletsOutput, currentWallet)

			multyTx.TxOutputs = append(multyTx.TxOutputs, store.AddresAmount{Address: txOutAddress, Amount: int64(SatoshiInBitcoint * output.Value)})

			multyTx.TxID = txVerbose.Txid
			multyTx.TxHash = txVerbose.Hash
		}
	}
	return nil
}

func updateWalletAndAddressDate(tx store.MultyTX) {

	for _, walletOutput := range tx.WalletsOutput {

		// update addresses last action time
		sel := bson.M{"userID": walletOutput.UserId, "wallets.addresses.address": walletOutput.Address.Address}
		update := bson.M{
			"$set": bson.M{
				"wallets.$.addresses.$[].lastActionTime": time.Now().Unix(),
			},
		}
		err := usersData.Update(sel, update)
		if err != nil {
			log.Errorf("updateWalletAndAddressDate:usersData.Update: %s", err.Error())
		}

		// update wallets last action time
		// Set status to OK if some money transfered to this address
		sel = bson.M{"userID": walletOutput.UserId, "wallets.walletIndex": walletOutput.WalletIndex}
		update = bson.M{
			"$set": bson.M{
				"wallets.$.status":         store.WalletStatusOK,
				"wallets.$.lastActionTime": time.Now().Unix(),
			},
		}
		err = usersData.Update(sel, update)
		if err != nil {
			log.Errorf("updateWalletAndAddressDate:usersData.Update: %s", err.Error())
		}

	}

	for _, walletInput := range tx.WalletsInput {
		// update addresses last action time
		sel := bson.M{"userID": walletInput.UserId, "wallets.addresses.address": walletInput.Address.Address}
		update := bson.M{
			"$set": bson.M{
				"wallets.$.addresses.$[].lastActionTime": time.Now().Unix(),
			},
		}
		err := usersData.Update(sel, update)
		if err != nil {
			log.Errorf("updateWalletAndAddressDate:usersData.Update: %s", err.Error())
		}

		// update wallets last action time
		sel = bson.M{"userID": walletInput.UserId, "wallets.walletIndex": walletInput.WalletIndex}
		update = bson.M{
			"$set": bson.M{
				"wallets.$.lastActionTime": time.Now().Unix(),
			},
		}
		err = usersData.Update(sel, update)
		if err != nil {
			log.Errorf("updateWalletAndAddressDate:usersData.Update: %s", err.Error())
		}
	}
}

func setTransactionStatus(tx *store.MultyTX, blockDiff int64, currentBlockHeight int64, fromInput bool) {
	transactionTime := time.Now().Unix()
	if blockDiff > currentBlockHeight {
		//This call was made from memPool
		tx.Confirmations = 0
		if fromInput {
			tx.TxStatus = TxStatusAppearedInMempoolOutcoming
			tx.MempoolTime = transactionTime
			tx.BlockTime = -1
		} else {
			tx.TxStatus = TxStatusAppearedInMempoolIncoming
			tx.MempoolTime = transactionTime
			tx.BlockTime = -1
		}

	} else if blockDiff >= 0 && blockDiff < 6 {
		//This call was made from block or resync
		//Transaction have no enough confirmations
		tx.Confirmations = int(blockDiff + 1)
		if fromInput {
			tx.TxStatus = TxStatusAppearedInBlockOutcoming
			tx.BlockTime = transactionTime
		} else {
			tx.TxStatus = TxStatusAppearedInBlockIncoming
			tx.BlockTime = transactionTime
		}
	} else if blockDiff >= 6 && blockDiff < currentBlockHeight {
		//This call was made from resync
		//Transaction have enough confirmations
		tx.Confirmations = int(blockDiff + 1)
		if fromInput {
			tx.TxStatus = TxStatusInBlockConfirmedOutcoming
			//TODO add block time for re-sync
		} else {
			tx.TxStatus = TxStatusInBlockConfirmedIncoming
		}
	}
}

func finalizeTransaction(tx *store.MultyTX, txVerbose *btcjson.TxRawResult) {

	if tx.TxAddress == nil {
		tx.TxAddress = []string{}
	}

	if tx.TxStatus == TxStatusInBlockConfirmedOutcoming || tx.TxStatus == TxStatusAppearedInBlockOutcoming || tx.TxStatus == TxStatusAppearedInMempoolOutcoming {
		for _, walletInput := range tx.WalletsInput {
			tx.TxOutAmount += walletInput.Address.Amount
			tx.TxAddress = append(tx.TxAddress, walletInput.Address.Address)
		}

		for i := 0; i < len(tx.WalletsOutput); i++ {
			//Here we descreasing amount of the current transaction
			tx.TxOutAmount -= tx.WalletsOutput[i].Address.Amount

			for _, output := range txVerbose.Vout {
				for _, outAddr := range output.ScriptPubKey.Addresses {
					if tx.WalletsOutput[i].Address.Address == outAddr {
						tx.WalletsOutput[i].Address.AddressOutIndex = int(output.N)
						tx.TxOutScript = txVerbose.Vout[output.N].ScriptPubKey.Hex
					}
				}
			}

		}
	} else {
		for i := 0; i < len(tx.WalletsOutput); i++ {
			tx.TxOutAmount += tx.WalletsOutput[i].Address.Amount
			tx.TxAddress = append(tx.TxAddress, tx.WalletsOutput[i].Address.Address)

			for _, output := range txVerbose.Vout {
				for _, outAddr := range output.ScriptPubKey.Addresses {
					if tx.WalletsOutput[i].Address.Address == outAddr {
						tx.WalletsOutput[i].Address.AddressOutIndex = int(output.N)
						tx.TxOutScript = txVerbose.Vout[output.N].ScriptPubKey.Hex
					}
				}
			}
		}
		//TxOutIndexes we need only for incoming transactions
	}
}

func CreateSpendableOutputs(tx *btcjson.TxRawResult, blockHeight int64) {
	user := store.User{}
	for _, output := range tx.Vout {
		if len(output.ScriptPubKey.Addresses) > 0 {
			address := output.ScriptPubKey.Addresses[0]
			query := bson.M{"wallets.addresses.address": address}
			err := usersData.Find(query).One(&user)
			if err != nil {
				continue
				// is not our user
			}
			walletindex, addressIndex := fetchWalletAndAddressIndexes(user.Wallets, address)

			txStatus := store.TxStatusAppearedInBlockIncoming
			if blockHeight == -1 {
				txStatus = store.TxStatusAppearedInMempoolIncoming
			}

			exRate, err := GetLatestExchangeRate()
			if err != nil {
				log.Errorf("CreateSpendableOutputs:GetLatestExchangeRate: %s", err.Error())
			}

			amount := int64(output.Value * SatoshiInBitcoint)
			spendableOutput := store.SpendableOutputs{
				TxID:              tx.Txid,
				TxOutID:           int(output.N),
				TxOutAmount:       amount,
				TxOutScript:       output.ScriptPubKey.Hex,
				Address:           address,
				UserID:            user.UserID,
				WalletIndex:       walletindex,
				AddressIndex:      addressIndex,
				TxStatus:          txStatus,
				StockExchangeRate: exRate,
			}
			fmt.Println("spendableOutput status %s", spendableOutput.TxStatus)

			query = bson.M{"userid": user.UserID, "txid": tx.Txid, "address": address}
			err = spendableOutputs.Find(query).One(nil)
			if err == mgo.ErrNotFound {
				//insertion
				err := spendableOutputs.Insert(spendableOutput)
				if err != nil {
					log.Errorf("CreateSpendableOutputs:txsData.Insert: %s", err.Error())
				}
				continue
			}
			if err != nil && err != mgo.ErrNotFound {
				log.Errorf("CreateSpendableOutputs:spendableOutputs.Find %s", err.Error())
				continue
			}

			update := bson.M{
				"$set": bson.M{
					"txstatus": txStatus,
				},
			}
			err = spendableOutputs.Update(query, update)
			if err != nil {
				log.Errorf("CreateSpendableOutputs:spendableOutputs.Update: %s", err.Error())
			}
		}
	}
}

func DeleteSpendableOutputs(tx *btcjson.TxRawResult, blockHeight int64) {
	/*
		user := store.User{}
		for _, input := range tx.Vin {
			previousTx, err := rawTxByTxid(input.Txid)
			if err != nil {
				log.Errorf("DeleteSpendableOutputs:rawTxByTxid: %s", err.Error())
			}

			if previousTx == nil {
				continue
			}

			if len(previousTx.Vout[input.Vout].ScriptPubKey.Addresses) > 0 {
				address := previousTx.Vout[input.Vout].ScriptPubKey.Addresses[0]
				query := bson.M{"wallets.addresses.address": address}
				err := usersData.Find(query).One(&user)
				if err != nil {
					continue
					// is not our user
				}
				log.Errorf("\n\n !!!found user:!!    %v  \n\n", user.UserID)
				query = bson.M{"userid": user.UserID, "txid": previousTx.Txid, "address": address}
				log.Debugf("userid ", user.UserID, "txid ", previousTx.Txid, "address ", address)
				err = spendableOutputs.Remove(query)
				if err != nil {
					log.Errorf("DeleteSpendableOutputs:spendableOutputs.Remove: %s", err.Error())
					log.Errorf("\n\n !!!not removed:!!    %v  \n\n")
				}
				log.Debugf("DeleteSpendableOutputs:spendableOutputs.Remove: %s", err)
			}
		}
	*/
}

func GetReSyncExchangeRate(time int64) ([]store.ExchangeRatesRecord, error) {
	selCCCAGG := bson.M{
		"stockexchange": "CCCAGG",
		"timestamp":     bson.M{"$lt": time},
	}
	stocksCCCAGG := store.ExchangeRatesRecord{}
	err := exRate.Find(selCCCAGG).Sort("-timestamp").One(&stocksCCCAGG)
	if err != nil {
		return nil, err
	}
	return []store.ExchangeRatesRecord{stocksCCCAGG}, nil
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

func setExchangeRates(tx *store.MultyTX, isReSync bool, TxTime int64) {
	var err error
	if isReSync {
		rates, err := GetReSyncExchangeRate(TxTime)
		if err != nil {
			log.Errorf("processTransaction:ExchangeRates: %s", err.Error())
		}
		tx.StockExchangeRate = rates
		return
	}
	if !isReSync || err != nil {
		rates, err := GetLatestExchangeRate()
		if err != nil {
			log.Errorf("processTransaction:ExchangeRates: %s", err.Error())
		}
		tx.StockExchangeRate = rates
	}
}

func InsertMempoolRecords(recs ...store.MempoolRecord) {
	for _, rec := range recs {
		err := mempoolRates.Insert(rec)
		if err != nil {
			log.Errorf("getAllMempool: mempoolRates.Insert: %s", err.Error())
			continue
		}
	}
}
