/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package eth

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/Appscrunch/Multy-back/currencies"
	ethpb "github.com/Appscrunch/Multy-back/node-streamer/eth"
	"github.com/Appscrunch/Multy-back/store"
	_ "github.com/KristinaEtc/slflog"
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
	spentOutputs     *mgo.Collection

	mempoolRatesTest     *mgo.Collection
	txsDataTest          *mgo.Collection
	spendableOutputsTest *mgo.Collection
	spentOutputsTest     *mgo.Collection
)

//TODO: make an update
func updateWalletAndAddressDate(tx store.MultyTX, networkID int) error {
	//TODO: make an update
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

		//TODO: fix "wallets.$.status":store.WalletStatusOK,
		// update wallets last action time
		// Set status to OK if some money transfered to this address
		user := store.User{}
		sel = bson.M{"userID": walletOutput.UserId, "wallets.walletIndex": walletOutput.WalletIndex, "wallets.addresses.address": walletOutput.Address.Address, "wallets.networkID": networkID, "wallets.currencyID": currencies.Bitcoin}
		err = usersData.Find(sel).One(&user)
		if err != nil {
			return errors.New("updateWalletAndAddressDate:usersData.Update: " + err.Error())
		}

		//TODO: fix hardcode if wallet.NetworkID == networkID && walletOutput.WalletIndex == walletindex && currencies.Bitcoin == currencyID {
		var flag bool
		var position int
		for i, wallet := range user.Wallets {
			if wallet.NetworkID == networkID && wallet.WalletIndex == walletOutput.WalletIndex && wallet.CurrencyID == currencies.Bitcoin {
				position = i
				flag = true
				break
			}
		}

		if flag {
			update = bson.M{
				"$set": bson.M{
					"wallets." + strconv.Itoa(position) + ".status":         store.WalletStatusOK,
					"wallets." + strconv.Itoa(position) + ".lastActionTime": time.Now().Unix(),
				},
			}
			err = usersData.Update(sel, update)
			if err != nil {
				return errors.New("updateWalletAndAddressDate:usersData.Update: " + err.Error())
			}

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
	//TODO:Make eth ex rate
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

func setExchangeRates(tx *store.TransactionETH, isReSync bool, TxTime int64) {
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

func sendNotifyToClients(tx store.TransactionETH, nsqProducer *nsq.Producer) {
	//TODO: make correct notify

	if tx.Status == store.TxStatusAppearedInBlockIncoming || tx.Status == store.TxStatusAppearedInMempoolIncoming || tx.Status == store.TxStatusInBlockConfirmedIncoming {
		txMsq := BtcTransactionWithUserID{
			UserID: tx.UserID,
			NotificationMsg: &BtcTransaction{
				TransactionType: tx.Status,
				Amount:          0, //TODO: tx.Amount
				TxID:            tx.Hash,
				Address:         tx.To,
			},
		}
		sendNotify(&txMsq, nsqProducer)
	}

	if tx.Status == store.TxStatusAppearedInBlockOutcoming || tx.Status == store.TxStatusAppearedInMempoolOutcoming || tx.Status == store.TxStatusInBlockConfirmedOutcoming {
		txMsq := BtcTransactionWithUserID{
			UserID: tx.UserID,
			NotificationMsg: &BtcTransaction{
				TransactionType: tx.Status,
				Amount:          0, //TODO: tx.Amount
				TxID:            tx.Hash,
				Address:         tx.From,
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

func generatedTxDataToStore(tx *ethpb.ETHTransaction) store.TransactionETH {
	return store.TransactionETH{
		UserID:       tx.UserID,
		WalletIndex:  int(tx.WalletIndex),
		AddressIndex: int(tx.AddressIndex),
		Hash:         tx.Hash,
		From:         tx.From,
		To:           tx.To,
		Amount:       tx.Amount,
		GasPrice:     tx.GasPrice,
		GasLimit:     tx.GasLimit,
		Nonce:        int(tx.Nonce),
		Status:       int(tx.Status),
		BlockTime:    tx.BlockTime,
		PoolTime:     tx.TxpoolTime,
		BlockHeight:  tx.BlockHeight,
	}
}

func saveTransaction(tx store.TransactionETH, networtkID int, resync bool) error {

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

	// This is splited transaction! That means that transaction's WalletsInputs and WalletsOutput have the same WalletIndex!
	//Here we have outgoing transaction for exact wallet!
	multyTX := store.TransactionETH{}
	if tx.Status == store.TxStatusAppearedInBlockIncoming || tx.Status == store.TxStatusAppearedInMempoolIncoming || tx.Status == store.TxStatusInBlockConfirmedIncoming {
		log.Debugf("saveTransaction new incoming tx to %v", tx.To)
		sel := bson.M{"userid": tx.UserID, "hash": tx.Hash, "walletindex": tx.WalletIndex}
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
				"txstatus":    tx.Status,
				"blockheight": tx.BlockHeight,
				"blocktime":   tx.BlockTime,
			},
		}
		err = txStore.Update(sel, update)
		return err
	} else if tx.Status == store.TxStatusAppearedInBlockOutcoming || tx.Status == store.TxStatusAppearedInMempoolOutcoming || tx.Status == store.TxStatusInBlockConfirmedOutcoming {
		log.Debugf("saveTransaction new outcoming tx  %v", tx.From)
		sel := bson.M{"userid": tx.UserID, "hash": tx.Hash, "walletindex": tx.WalletIndex}
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
				"txstatus":    tx.Status,
				"blockheight": tx.BlockHeight,
				"blocktime":   tx.BlockTime,
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
