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

	"github.com/Multy-io/Multy-back/currencies"
	ethpb "github.com/Multy-io/Multy-back/node-streamer/eth"
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

func updateWalletAndAddressDate(tx store.TransactionETH, networkID int) error {

	sel := bson.M{"userID": tx.UserID, "wallets.addresses.address": tx.From}
	user := store.User{}
	err := usersData.Find(sel).One(&user)
	update := bson.M{}

	var ok bool

	for i := range user.Wallets {
		for j, addr := range user.Wallets[i].Adresses {
			if addr.Address == tx.From {
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

	sel = bson.M{"userID": tx.UserID, "wallets.addresses.address": tx.To}
	user = store.User{}
	err = usersData.Find(sel).One(&user)
	update = bson.M{}

	ok = false

	for i := range user.Wallets {
		for j, addr := range user.Wallets[i].Adresses {
			if addr.Address == tx.To {
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

	return nil
}

func GetReSyncExchangeRate(time int64) ([]store.ExchangeRatesRecord, error) {
	selCCCAGG := bson.M{
		"stockexchange": "CCCAGG",
		"timestamp":     bson.M{"$lt": time},
	}
	stocksCCCAGG := store.ExchangeRatesRecord{}
	err := exRate.Find(selCCCAGG).Sort("-timestamp").One(&stocksCCCAGG)
	return []store.ExchangeRatesRecord{stocksCCCAGG}, err
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
		rates, err := GetReSyncExchangeRate(tx.BlockTime)
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

func sendNotifyToClients(tx store.TransactionETH, nsqProducer *nsq.Producer, netid int) {
	//TODO: make correct notify

	if tx.Status == store.TxStatusAppearedInBlockIncoming || tx.Status == store.TxStatusAppearedInMempoolIncoming || tx.Status == store.TxStatusInBlockConfirmedIncoming {
		txMsq := store.TransactionWithUserID{
			UserID: tx.UserID,
			NotificationMsg: &store.WsTxNotify{
				CurrencyID:      currencies.Ether,
				NetworkID:       netid,
				Address:         tx.To,
				Amount:          tx.Amount,
				TxID:            tx.Hash,
				TransactionType: tx.Status,
				WalletIndex:     tx.WalletIndex,
			},
		}
		sendNotify(&txMsq, nsqProducer)
	}

	if tx.Status == store.TxStatusAppearedInBlockOutcoming || tx.Status == store.TxStatusAppearedInMempoolOutcoming || tx.Status == store.TxStatusInBlockConfirmedOutcoming {
		txMsq := store.TransactionWithUserID{
			UserID: tx.UserID,
			NotificationMsg: &store.WsTxNotify{
				CurrencyID:      currencies.Ether,
				NetworkID:       netid,
				Address:         tx.From,
				Amount:          tx.Amount,
				TxID:            tx.Hash,
				TransactionType: tx.Status,
				WalletIndex:     tx.WalletIndex,
			},
		}
		sendNotify(&txMsq, nsqProducer)
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
