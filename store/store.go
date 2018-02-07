/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package store

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	errType        = errors.New("wrong database type")
	errEmplyConfig = errors.New("empty configuration for datastore")
)

// Default table names
const (
	TableUsers             = "UserCollection"
	TableFeeRates          = "Rates" // and send those two fields there
	TableBTC               = "BTC"
	TableStockExchangeRate = "TableStockExchangeRate"
)

// Conf is a struct for database configuration
type Conf struct {
	Address string

	// TODO: move to one database
	DBUsers             string
	DBFeeRates          string
	DBTx                string
	DBStockExchangeRate string
}

type UserStore interface {
	GetUserByDevice(device bson.M, user *User)
	Update(sel, update bson.M) error
	Insert(user User) error
	Close() error
	FindUser(query bson.M, user *User) error
	UpdateUser(sel bson.M, user *User) error
	GetAllRates(sortBy string, rates *[]RatesRecord) error //add to rates store
	FindUserTxs(query bson.M, userTxs *TxRecord) error
	InsertTxStore(userTxs TxRecord) error
	FindUserErr(query bson.M) error
	FindUserAddresses(query bson.M, sel bson.M, ws *WalletsSelect) error
	InsertExchangeRate(ExchangeRates, string) error
	GetExchangeRatesDay() ([]RatesAPIBitstamp, error)
	GetAllWalletTransactions(query bson.M, walletTxs *[]MultyTX) error
	GetAllSpendableOutputs(query bson.M) (error, []SpendableOutputs)
	GetAddressSpendableOutputs(query bson.M) ([]SpendableOutputs, error)
	DeleteWallet(userid string, walletindex int) error
	GetEthereumTransationHistory(query bson.M) ([]MultyETHTransaction, error)
	AddEthereumTransaction(tx MultyETHTransaction) error
	UpdateEthereumTransaction(sel, update bson.M) error
	FindETHTransaction(sel bson.M) error
	DropTest()
}

type MongoUserStore struct {
	config            *Conf
	session           *mgo.Session
	usersData         *mgo.Collection
	ratesData         *mgo.Collection
	txsData           *mgo.Collection
	spendableOutputs  *mgo.Collection
	stockExchangeRate *mgo.Collection
	ethTxHistory      *mgo.Collection
}

func InitUserStore(conf Conf) (UserStore, error) {
	uStore := &MongoUserStore{
		config: &conf,
	}
	session, err := mgo.Dial(conf.Address)
	if err != nil {
		return nil, err
	}
	uStore.session = session
	uStore.usersData = uStore.session.DB(conf.DBUsers).C(TableUsers)
	uStore.ratesData = uStore.session.DB(conf.DBFeeRates).C(TableFeeRates)
	uStore.txsData = uStore.session.DB(conf.DBTx).C(TableBTC)
	uStore.stockExchangeRate = uStore.session.DB(conf.DBStockExchangeRate).C(TableStockExchangeRate)
	// TODO: make varribles in a config
	uStore.spendableOutputs = uStore.session.DB(conf.DBTx).C("SpendableOutputs")
	uStore.ethTxHistory = uStore.session.DB(conf.DBTx).C("ETH")

	return uStore, nil
}

func (mStore *MongoUserStore) DropTest() {
	mStore.usersData.DropCollection()
	mStore.txsData.DropCollection()
	mStore.spendableOutputs.DropCollection()
}

func (mStore *MongoUserStore) FindETHTransaction(sel bson.M) error {
	err := mStore.ethTxHistory.Find(sel).One(nil)
	return err
}

func (mStore *MongoUserStore) UpdateEthereumTransaction(sel, update bson.M) error {
	err := mStore.ethTxHistory.Update(sel, update)
	return err
}

func (mStore *MongoUserStore) AddEthereumTransaction(tx MultyETHTransaction) error {
	err := mStore.ethTxHistory.Insert(tx)
	return err
}

func (mStore *MongoUserStore) GetEthereumTransationHistory(query bson.M) ([]MultyETHTransaction, error) {
	allTxs := []MultyETHTransaction{}
	err := mStore.ethTxHistory.Find(query).All(&allTxs)
	return allTxs, err
}

func (mStore *MongoUserStore) DeleteWallet(userid string, walletindex int) error {
	sel := bson.M{"userID": userid, "wallets.walletIndex": walletindex}
	update := bson.M{
		"$set": bson.M{
			"wallets.$.status": WalletStatusDeleted,
		},
	}
	return mStore.usersData.Update(sel, update)
}
func (mStore *MongoUserStore) GetAllSpendableOutputs(query bson.M) (error, []SpendableOutputs) {
	spOuts := []SpendableOutputs{}
	err := mStore.spendableOutputs.Find(query).All(&spOuts)
	return err, spOuts
}
func (mStore *MongoUserStore) GetAddressSpendableOutputs(query bson.M) ([]SpendableOutputs, error) {
	spOuts := []SpendableOutputs{}
	err := mStore.spendableOutputs.Find(query).All(&spOuts)
	return spOuts, err
}

func (mStore *MongoUserStore) UpdateUser(sel bson.M, user *User) error {
	return mStore.usersData.Update(sel, user)
}

func (mStore *MongoUserStore) GetUserByDevice(device bson.M, user *User) { // rename GetUserByToken
	mStore.usersData.Find(device).One(user)
	return // why?
}

func (mStore *MongoUserStore) Update(sel, update bson.M) error {
	return mStore.usersData.Update(sel, update)
}

func (mStore *MongoUserStore) FindUser(query bson.M, user *User) error {
	return mStore.usersData.Find(query).One(user)
}
func (mStore *MongoUserStore) FindUserErr(query bson.M) error {
	return mStore.usersData.Find(query).One(nil)
}

func (mStore *MongoUserStore) FindUserAddresses(query bson.M, sel bson.M, ws *WalletsSelect) error {
	return mStore.usersData.Find(query).Select(sel).One(ws)
}

func (mStore *MongoUserStore) Insert(user User) error {
	return mStore.usersData.Insert(user)
}

func (mStore *MongoUserStore) GetAllRates(sortBy string, rates *[]RatesRecord) error {
	return mStore.ratesData.Find(nil).Sort(sortBy).All(rates)
}

func (mStore *MongoUserStore) FindUserTxs(query bson.M, userTxs *TxRecord) error {
	return mStore.txsData.Find(query).One(userTxs)
}

func (mStore *MongoUserStore) InsertTxStore(userTxs TxRecord) error {
	return mStore.txsData.Insert(userTxs)
}

func (mStore *MongoUserStore) InsertExchangeRate(eRate ExchangeRates, exchangeStock string) error {
	eRateRecord := &ExchangeRatesRecord{
		Exchanges:     eRate,
		Timestamp:     time.Now().Unix(),
		StockExchange: exchangeStock,
	}

	return mStore.stockExchangeRate.Insert(eRateRecord)
}

// func (mStore *MongoUserStore) GetLatestExchangeRate() ([]ExchangeRatesRecord, error) {
// 	selGdax := bson.M{
// 		"stockexchange": "Gdax",
// 	}
// 	selPoloniex := bson.M{
// 		"stockexchange": "Poloniex",
// 	}
// 	stocksGdax := ExchangeRatesRecord{}
// 	err := mStore.stockExchangeRate.Find(selGdax).Sort("-timestamp").One(&stocksGdax)
// 	if err != nil {
// 		return nil, err
// 	}

// 	stocksPoloniex := ExchangeRatesRecord{}
// 	err = mStore.stockExchangeRate.Find(selPoloniex).Sort("-timestamp").One(&stocksPoloniex)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return []ExchangeRatesRecord{stocksPoloniex, stocksGdax}, nil

// }

// GetExchangeRatesDay returns exchange rates for last day with time interval equal to hour
func (mStore *MongoUserStore) GetExchangeRatesDay() ([]RatesAPIBitstamp, error) {
	// not implemented
	return nil, nil
}

func (mStore *MongoUserStore) GetAllWalletTransactions(query bson.M, walletTxs *[]MultyTX) error {
	return mStore.txsData.Find(query).All(walletTxs)
}

func (mStore *MongoUserStore) Close() error {
	mStore.session.Close()
	return nil
}
