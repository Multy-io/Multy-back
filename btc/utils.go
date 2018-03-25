/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/Appscrunch/Multy-back/currencies"
	btcpb "github.com/Appscrunch/Multy-back/node-streamer/btc"
	"github.com/Appscrunch/Multy-back/store"
	nsq "github.com/bitly/go-nsq"
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

func updateWalletAndAddressDate(tx store.MultyTX) error {

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
			return errors.New("updateWalletAndAddressDate:usersData.Update: " + err.Error())
		}

		// update wallets last action time
		// Set status to OK if some money transfered to this address
		sel = bson.M{"userID": walletOutput.UserId, "wallets.walletIndex": walletOutput.WalletIndex, "wallets.addresses.address": walletOutput.Address.Address}
		update = bson.M{
			"$set": bson.M{
				"wallets.$.status":         store.WalletStatusOK,
				"wallets.$.lastActionTime": time.Now().Unix(),
			},
		}
		err = usersData.Update(sel, update)
		if err != nil {
			return errors.New("updateWalletAndAddressDate:usersData.Update: " + err.Error())
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
			return errors.New("updateWalletAndAddressDate:usersData.Update: " + err.Error())
		}

		// update wallets last action time
		sel = bson.M{"userID": walletInput.UserId, "wallets.walletIndex": walletInput.WalletIndex, "wallets.addresses.address": walletInput.Address.Address}
		update = bson.M{
			"$set": bson.M{
				"wallets.$.lastActionTime": time.Now().Unix(),
			},
		}
		err = usersData.Update(sel, update)
		if err != nil {
			return errors.New("updateWalletAndAddressDate:usersData.Update: " + err.Error())
		}
	}

	return nil
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

func sendNotifyToClients(tx store.MultyTX, nsqProducer *nsq.Producer) {

	for _, walletOutput := range tx.WalletsOutput {
		txMsq := BtcTransactionWithUserID{
			UserID: walletOutput.UserId,
			NotificationMsg: &BtcTransaction{
				TransactionType: tx.TxStatus,
				Amount:          tx.TxOutAmount,
				TxID:            tx.TxID,
				Address:         walletOutput.Address.Address,
			},
		}
		sendNotify(&txMsq, nsqProducer)
	}

	for _, walletInput := range tx.WalletsInput {
		txMsq := BtcTransactionWithUserID{
			UserID: walletInput.UserId,
			NotificationMsg: &BtcTransaction{
				TransactionType: tx.TxStatus,
				Amount:          tx.TxOutAmount,
				TxID:            tx.TxID,
				Address:         walletInput.Address.Address,
			},
		}
		sendNotify(&txMsq, nsqProducer)
	}
}

func sendNotify(txMsq *BtcTransactionWithUserID, nsqProducer *nsq.Producer) {
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

func generatedTxDataToStore(gSpOut *btcpb.BTCTransaction) store.MultyTX {
	outs := []store.AddresAmount{}
	for _, output := range gSpOut.TxOutputs {
		outs = append(outs, store.AddresAmount{
			Address: output.Address,
			Amount:  output.Amount,
		})
	}

	ins := []store.AddresAmount{}
	for _, inputs := range gSpOut.TxInputs {
		ins = append(ins, store.AddresAmount{
			Address: inputs.Address,
			Amount:  inputs.Amount,
		})
	}

	// gSpOut.WalletsInput
	// gSpOut.WalletsOutput

	wInputs := []store.WalletForTx{}
	for _, walletOutputs := range gSpOut.WalletsOutput {
		wInputs = append(wInputs, store.WalletForTx{
			UserId: walletOutputs.Userid,
			Address: store.AddressForWallet{
				Address:         walletOutputs.Address,
				Amount:          walletOutputs.Amount,
				AddressOutIndex: int(walletOutputs.TxOutIndex),
			},
		})
	}

	wOutputs := []store.WalletForTx{}
	for _, walletInputs := range gSpOut.WalletsInput {
		wOutputs = append(wOutputs, store.WalletForTx{
			UserId: walletInputs.Userid,
			Address: store.AddressForWallet{
				Address:         walletInputs.Address,
				Amount:          walletInputs.Amount,
				AddressOutIndex: int(walletInputs.TxOutIndex),
			},
		})
	}

	return store.MultyTX{
		UserId:        gSpOut.UserID,
		TxID:          gSpOut.TxID,
		TxHash:        gSpOut.TxHash,
		TxOutScript:   gSpOut.TxOutScript,
		TxAddress:     gSpOut.TxAddress,
		TxStatus:      int(gSpOut.TxStatus),
		TxOutAmount:   gSpOut.TxOutAmount,
		BlockTime:     gSpOut.BlockTime,
		BlockHeight:   gSpOut.BlockHeight,
		Confirmations: int(gSpOut.Confirmations),
		TxFee:         gSpOut.TxFee,
		MempoolTime:   gSpOut.MempoolTime,
		TxInputs:      ins,
		TxOutputs:     outs,
		WalletsInput:  wInputs,
		WalletsOutput: wOutputs,
	}
}

func generatedSpOutsToStore(gSpOut *btcpb.AddSpOut) store.SpendableOutputs {
	return store.SpendableOutputs{
		TxID:        gSpOut.TxID,
		TxOutID:     int(gSpOut.TxOutID),
		TxOutAmount: gSpOut.TxOutAmount,
		TxOutScript: gSpOut.TxOutScript,
		Address:     gSpOut.Address,
		UserID:      gSpOut.UserID,
		TxStatus:    int(gSpOut.TxStatus),
	}
}

func saveMultyTransaction(tx store.MultyTX, networtkID int) error {

	txsdata := &mgo.Collection{}
	switch networtkID {
	case currencies.Main:
		txsdata = txsData
	case currencies.Test:
		txsdata = txsDataTest
	default:
		return errors.New("saveMultyTransaction: wrong networkID")
	}
	// This is splited transaction! That means that transaction's WalletsInputs and WalletsOutput have the same WalletIndex!
	//Here we have outgoing transaction for exact wallet!
	multyTX := store.MultyTX{}
	if tx.WalletsInput != nil && len(tx.WalletsInput) > 0 {
		// sel := bson.M{"userid": tx.WalletsInput[0].UserId, "transactions.txid": tx.TxID, "transactions.walletsinput.walletindex": tx.WalletsInput[0].WalletIndex}
		sel := bson.M{"userid": tx.WalletsInput[0].UserId, "txid": tx.TxID, "walletsinput.walletindex": tx.WalletsInput[0].WalletIndex}
		err := txsdata.Find(sel).One(&multyTX)
		if err == mgo.ErrNotFound {
			// initial insertion
			err := txsdata.Insert(tx)
			if err != nil {
				log.Errorf("parseInput.Update add new tx to user: %s", err.Error())
			}
			return nil
		}
		if err != nil && err != mgo.ErrNotFound {
			// database error

			return errors.New("saveMultyTransaction:txsdata.Find " + err.Error())
		}

		update := bson.M{
			"$set": bson.M{
				"txstatus":      tx.TxStatus,
				"blockheight":   tx.BlockHeight,
				"confirmations": tx.Confirmations,
				"blocktime":     tx.BlockTime,
			},
		}
		err = txsdata.Update(sel, update)
		if err != nil {
			log.Errorf("saveMultyTransaction:txsdata.Update %s", err.Error())
		}
		return nil
	} else if tx.WalletsOutput != nil && len(tx.WalletsOutput) > 0 {
		// sel := bson.M{"userid": tx.WalletsOutput[0].UserId, "transactions.txid": tx.TxID, "transactions.walletsoutput.walletindex": tx.WalletsOutput[0].WalletIndex}
		sel := bson.M{"userid": tx.WalletsOutput[0].UserId, "txid": tx.TxID, "walletsoutput.walletindex": tx.WalletsOutput[0].WalletIndex}
		err := txsdata.Find(sel).One(&multyTX)
		if err == mgo.ErrNotFound {
			// initial insertion
			err := txsdata.Insert(tx)
			if err != nil {
				log.Errorf("parseInput.Update add new tx to user: %s", err.Error())
			}
			return nil
		}
		if err != nil && err != mgo.ErrNotFound {
			// database error
			return errors.New("saveMultyTransaction:txsdata.Find: " + err.Error())
		}

		update := bson.M{
			"$set": bson.M{
				"txstatus":      tx.TxStatus,
				"blockheight":   tx.BlockHeight,
				"confirmations": tx.Confirmations,
				"blocktime":     tx.BlockTime,
			},
		}
		err = txsdata.Update(sel, update)
		if err != nil {
			log.Errorf("saveMultyTransaction:txsData.Update %s", err.Error())
		}
		return nil
	}
	return nil
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

func setTxInfo(tx *store.MultyTX) {
	user := store.User{}
	// set wallet index and address index in input
	for _, in := range tx.WalletsInput {
		sel := bson.M{"wallets.addresses.address": in.Address.Address}
		err := usersData.Find(sel).One(&user)
		if err == mgo.ErrNotFound {
			continue
		} else if err != nil && err != mgo.ErrNotFound {
			log.Errorf("initGrpcClient: cli.On newIncomingTx: %s", err)
		}

		for _, wallet := range user.Wallets {
			for i := 0; i < len(wallet.Adresses); i++ {
				if wallet.Adresses[i].Address == in.Address.Address {
					in.WalletIndex = wallet.WalletIndex
					in.Address.AddressIndex = wallet.Adresses[i].AddressIndex
				}
			}
		}
	}

	// set wallet index and address index in output
	for _, out := range tx.WalletsOutput {
		sel := bson.M{"wallets.addresses.address": out.Address.Address}
		err := usersData.Find(sel).One(&user)
		if err == mgo.ErrNotFound {
			continue
		} else if err != nil && err != mgo.ErrNotFound {
			log.Errorf("initGrpcClient: cli.On newIncomingTx: %s", err)
		}

		for _, wallet := range user.Wallets {
			for i := 0; i < len(wallet.Adresses); i++ {
				if wallet.Adresses[i].Address == out.Address.Address {
					out.WalletIndex = wallet.WalletIndex
					out.Address.AddressIndex = wallet.Adresses[i].AddressIndex
				}
			}
		}
	}
}
