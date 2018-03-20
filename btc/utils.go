/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"time"

	"github.com/Appscrunch/Multy-back/store"
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

// TODO: update date
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
