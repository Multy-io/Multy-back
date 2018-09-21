/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	btcpb "github.com/Multy-io/Multy-BTC-node-service/node-streamer"
	"github.com/Multy-io/Multy-back/currencies"
	"github.com/Multy-io/Multy-back/store"
	nsq "github.com/bitly/go-nsq"
	_ "github.com/jekabolt/slflog"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	exRate    *mgo.Collection
	usersData *mgo.Collection

	txsData          *mgo.Collection
	spendableOutputs *mgo.Collection
	spentOutputs     *mgo.Collection

	txsDataTest          *mgo.Collection
	spendableOutputsTest *mgo.Collection
	spentOutputsTest     *mgo.Collection

	restoreState *mgo.Collection
)

func updateWalletAndAddressDate(tx store.MultyTX, networkID int) error {
	for _, wallet := range tx.WalletsInput {
		sel := bson.M{"userID": wallet.UserId, "wallets.addresses.address": wallet.Address.Address}
		user := store.User{}
		err := usersData.Find(sel).One(&user)
		update := bson.M{}

		var ok bool

		for i := range user.Wallets {
			for j, addr := range user.Wallets[i].Adresses {
				if addr.Address == wallet.Address.Address {
					ok = true
					update = bson.M{
						"$set": bson.M{
							"wallets." + strconv.Itoa(i) + ".lastActionTime":                                   time.Now().Unix(),
							"wallets." + strconv.Itoa(i) + ".addresses." + strconv.Itoa(j) + ".lastActionTime": time.Now().Unix(),
							"wallets." + strconv.Itoa(i) + ".status":                                           store.WalletStatusOK,
						},
					}
					break
				}
			}
		}
		if ok {
			err = usersData.Update(sel, update)
			if err != nil {
				return errors.New("updateWalletAndAddressDate:usersData.Update: " + err.Error())
			}
		}

	}
	for _, wallet := range tx.WalletsOutput {
		sel := bson.M{"userID": wallet.UserId, "wallets.addresses.address": wallet.Address.Address}
		user := store.User{}
		err := usersData.Find(sel).One(&user)
		update := bson.M{}

		var ok bool

		for i := range user.Wallets {
			for j, addr := range user.Wallets[i].Adresses {
				if addr.Address == wallet.Address.Address {
					ok = true
					update = bson.M{
						"$set": bson.M{
							"wallets." + strconv.Itoa(i) + ".lastActionTime":                                   time.Now().Unix(),
							"wallets." + strconv.Itoa(i) + ".addresses." + strconv.Itoa(j) + ".lastActionTime": time.Now().Unix(),
							"wallets." + strconv.Itoa(i) + ".status":                                           store.WalletStatusOK,
						},
					}
					break
				}
			}
		}
		if ok {
			err = usersData.Update(sel, update)
			if err != nil {
				return errors.New("updateWalletAndAddressDate:usersData.Update: " + err.Error())
			}
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

func sendNotifyToClients(tx store.MultyTX, nsqProducer *nsq.Producer, netid int) {
	for _, walletOutput := range tx.WalletsOutput {
		txMsq := store.TransactionWithUserID{
			UserID: walletOutput.UserId,
			NotificationMsg: &store.WsTxNotify{
				CurrencyID:      currencies.Bitcoin,
				NetworkID:       netid,
				Address:         walletOutput.Address.Address,
				Amount:          strconv.Itoa(int(tx.TxOutAmount)),
				TxID:            tx.TxID,
				TransactionType: tx.TxStatus,
				From:            walletOutput.Address.Address,
				To:              tx.TxAddress[0],
				WalletIndex:     walletOutput.WalletIndex,
			},
		}
		if walletOutput.Address.Address != tx.TxAddress[0] {
			sendNotify(&txMsq, nsqProducer)
		}
	}

	for _, walletInput := range tx.WalletsInput {
		txMsq := store.TransactionWithUserID{
			UserID: walletInput.UserId,
			NotificationMsg: &store.WsTxNotify{
				CurrencyID:      currencies.Bitcoin,
				NetworkID:       netid,
				Address:         walletInput.Address.Address,
				Amount:          strconv.Itoa(int(tx.TxOutAmount)),
				TxID:            tx.TxID,
				TransactionType: tx.TxStatus,
				WalletIndex:     walletInput.WalletIndex,
				From:            walletInput.Address.Address,
				To:              tx.TxAddress[0],
			},
		}
		if walletInput.Address.Address != tx.TxAddress[0] {
			sendNotify(&txMsq, nsqProducer)
		}
	}

	for _, txInputs := range tx.TxInputs {
		txMsq := store.TransactionWithUserID{
			UserID: tx.UserId,
			NotificationMsg: &store.WsTxNotify{
				CurrencyID:      currencies.Bitcoin,
				NetworkID:       netid,
				Address:         tx.TxAddress[0],
				Amount:          strconv.Itoa(int(tx.TxOutAmount)),
				TxID:            tx.TxID,
				TransactionType: tx.TxStatus,
				WalletIndex:     100,
				To:              tx.TxAddress[0],
				From:            txInputs.Address,
			},
		}
		if tx.TxAddress[0] != txInputs.Address {
			sendNotify(&txMsq, nsqProducer)
		}
	}
}

func sendNotify(txMsq *store.TransactionWithUserID, nsqProducer *nsq.Producer) {
	newTxJSON, err := json.Marshal(txMsq)
	if err != nil {
		log.Errorf("sendNotifyToClients: [%+v] %s\n", txMsq, err.Error())
		return
	}
	err = nsqProducer.Publish(store.TopicTransaction, newTxJSON)
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
		TxID:         gSpOut.TxID,
		TxOutID:      int(gSpOut.TxOutID),
		TxOutAmount:  gSpOut.TxOutAmount,
		TxOutScript:  gSpOut.TxOutScript,
		Address:      gSpOut.Address,
		UserID:       gSpOut.UserID,
		TxStatus:     int(gSpOut.TxStatus),
		WalletIndex:  int(gSpOut.WalletIndex),
		AddressIndex: int(gSpOut.AddressIndex),
	}
}

func saveMultyTransaction(tx store.MultyTX, networtkID int, resync bool) error {

	txStore := &mgo.Collection{}
	switch networtkID {
	case currencies.Main:
		txStore = txsData
	case currencies.Test:
		txStore = txsDataTest
	default:
		return errors.New("saveMultyTransaction: wrong networkID")
	}

	// fetchedTxs := []store.MultyTX{}
	// query := bson.M{"txid": tx.TxID}
	// txStore.Find(query).All(&fetchedTxs)

	// Doubling txs fix on -1
	if tx.BlockHeight == -1 {
		multyTX := store.MultyTX{}
		if len(tx.WalletsInput) != 0 {
			sel := bson.M{"userid": tx.WalletsInput[0].UserId, "txid": tx.TxID, "walletsoutput.walletindex": tx.WalletsInput[0].WalletIndex}
			err := txStore.Find(sel).One(multyTX)
			if err == nil && multyTX.BlockHeight > -1 {
				return nil
			}
		}
		if len(tx.WalletsOutput) > 0 {
			sel := bson.M{"userid": tx.WalletsOutput[0].UserId, "txid": tx.TxID, "walletsoutput.walletindex": tx.WalletsOutput[0].WalletIndex}
			err := txStore.Find(sel).One(multyTX)
			if err == nil && multyTX.BlockHeight > -1 {
				return nil
			}
		}
	}

	// Doubling txs fix on a asynchronous err
	if len(tx.WalletsInput) != 0 {
		sel := bson.M{"userid": tx.UserId, "txid": tx.TxID, "walletsoutput.walletindex": tx.WalletsInput[0].WalletIndex, "mempooltime": tx.MempoolTime}
		err := txStore.Find(sel).One(nil)
		if err == nil {
			return nil
		}
	}
	if len(tx.WalletsOutput) > 0 {
		sel := bson.M{"userid": tx.UserId, "txid": tx.TxID, "walletsoutput.walletindex": tx.WalletsOutput[0].WalletIndex, "mempooltime": tx.MempoolTime}
		err := txStore.Find(sel).One(nil)
		if err == nil {
			return nil
		}
	}

	// This is splited transaction! That means that transaction's WalletsInputs and WalletsOutput have the same WalletIndex!
	//Here we have outgoing transaction for exact wallet!
	multyTX := store.MultyTX{}
	if tx.WalletsInput != nil && len(tx.WalletsInput) > 0 {

		sel := bson.M{"userid": tx.WalletsInput[0].UserId, "txid": tx.TxID, "txaddress": tx.TxAddress[0]}

		err := txStore.Find(sel).One(&multyTX)
		if err == mgo.ErrNotFound {
			// initial insertion
			err := txStore.Insert(tx)
			return err
		}
		if err != nil && err != mgo.ErrNotFound {
			// database error
			return err
		}

		update := bson.M{
			"$set": bson.M{
				"txstatus":      tx.TxStatus,
				"blockheight":   tx.BlockHeight,
				"confirmations": tx.Confirmations,
				"blocktime":     tx.BlockTime,
				"walletsoutput": tx.WalletsOutput,
				"walletsinput":  tx.WalletsInput,
			},
		}
		err = txStore.Update(sel, update)
		if err != nil {
			log.Errorf("saveMultyTransaction:txsData.Update %s", err.Error())
		}
		return err

		// sel := bson.M{"userid": tx.WalletsInput[0].UserId, "transactions.txid": tx.TxID, "transactions.walletsinput.walletindex": tx.WalletsInput[0].WalletIndex}
		// sel := bson.M{"userid": tx.WalletsInput[0].UserId, "txid": tx.TxID, "walletsinput.walletindex": tx.WalletsInput[0].WalletIndex} // last
		sel = bson.M{"userid": tx.WalletsInput[0].UserId, "txid": tx.TxID, "walletsoutput.walletindex": tx.WalletsInput[0].WalletIndex}
		if tx.BlockHeight != -1 {
			sel = bson.M{"userid": tx.WalletsInput[0].UserId, "txid": tx.TxID, "walletsinput.walletindex": tx.WalletsInput[0].WalletIndex}
		}

		err = txStore.Find(sel).One(&multyTX)
		if err == mgo.ErrNotFound {
			// initial insertion
			err := txStore.Insert(tx)
			return err
		}
		if err != nil && err != mgo.ErrNotFound {
			// database error
			return err
		}

		update = bson.M{
			"$set": bson.M{
				"txstatus":      tx.TxStatus,
				"blockheight":   tx.BlockHeight,
				"confirmations": tx.Confirmations,
				"blocktime":     tx.BlockTime,
				"walletsoutput": tx.WalletsOutput,
				"walletsinput":  tx.WalletsInput,
			},
		}
		err = txStore.Update(sel, update)
		return err
	} else if tx.WalletsOutput != nil && len(tx.WalletsOutput) > 0 {
		// sel := bson.M{"userid": tx.WalletsOutput[0].UserId, "transactions.txid": tx.TxID, "transactions.walletsoutput.walletindex": tx.WalletsOutput[0].WalletIndex}
		sel := bson.M{"userid": tx.WalletsOutput[0].UserId, "txid": tx.TxID, "txaddress": tx.TxAddress[0]}
		err := txStore.Find(sel).One(&multyTX)
		if err == mgo.ErrNotFound {
			// initial insertion
			err := txStore.Insert(tx)
			return err
		}
		if err != nil && err != mgo.ErrNotFound {
			// database error
			return err
		}

		update := bson.M{
			"$set": bson.M{
				"txstatus":      tx.TxStatus,
				"blockheight":   tx.BlockHeight,
				"confirmations": tx.Confirmations,
				"blocktime":     tx.BlockTime,
				"walletsoutput": tx.WalletsOutput,
				"walletsinput":  tx.WalletsInput,
			},
		}
		err = txStore.Update(sel, update)
		if err != nil {
			log.Errorf("saveMultyTransaction:txsData.Update %s", err.Error())
		}
		return err
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
