/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package store

import (
	"errors"
	"strconv"
	"time"

	"github.com/Appscrunch/Multy-back/currencies"
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
	TableStockExchangeRate = "TableStockExchangeRate"
)

// Conf is a struct for database configuration
type Conf struct {
	Address             string
	DBUsers             string
	DBFeeRates          string
	DBTx                string
	DBStockExchangeRate string

	// BTC main
	TableMempoolRatesBTCMain     string
	TableTxsDataBTCMain          string
	TableSpendableOutputsBTCMain string
	TableSpentOutputsBTCMain     string

	// BTC test
	TableMempoolRatesBTCTest     string
	TableTxsDataBTCTest          string
	TableSpendableOutputsBTCTest string
	TableSpentOutputsBTCTest     string

	// ETH main
	TableMempoolRatesETHMain string
	TableTxsDataETHMain      string

	// ETH main
	TableMempoolRatesETHTest string
	TableTxsDataETHTest      string

	//Authentification
	Username string
	Password string
}

type UserStore interface {
	GetUserByDevice(device bson.M, user *User)
	Update(sel, update bson.M) error
	Insert(user User) error
	Close() error
	FindUser(query bson.M, user *User) error
	UpdateUser(sel bson.M, user *User) error
	GetAllRates(currencyID, networkID int, sortBy string, rates *[]MempoolRecord) error
	// FindUserTxs(query bson.M, userTxs *TxRecord) error
	// InsertTxStore(userTxs TxRecord) error
	FindUserErr(query bson.M) error
	FindUserAddresses(query bson.M, sel bson.M, ws *WalletsSelect) error
	InsertExchangeRate(ExchangeRates, string) error
	GetExchangeRatesDay() ([]RatesAPIBitstamp, error)

	//TODo update this method by eth
	GetAllWalletTransactions(userid string, currencyID, networkID int, walletTxs *[]MultyTX) error
	GetAllWalletEthTransactions(userid string, currencyID, networkID int, walletTxs *[]TransactionETH) error
	// GetAllSpendableOutputs(query bson.M) (error, []SpendableOutputs)
	GetAddressSpendableOutputs(address string, currencyID, networkID int) ([]SpendableOutputs, error)
	DeleteWallet(userid string, walletindex, currencyID, networkID int) error
	// DropTest()
	DeleteMempool()

	FindAllUserETHTransactions(sel bson.M) ([]TransactionETH, error)
	FindUserDataChain(CurrencyID, NetworkID int) (map[string]AddressExtended, error)
}

type MongoUserStore struct {
	config    *Conf
	session   *mgo.Session
	usersData *mgo.Collection

	// btc main
	BTCMainRatesData        *mgo.Collection
	BTCMainTxsData          *mgo.Collection
	BTCMainSpendableOutputs *mgo.Collection

	// btc test
	BTCTestRatesData        *mgo.Collection
	BTCTestTxsData          *mgo.Collection
	BTCTestSpendableOutputs *mgo.Collection

	//eth main
	ETHMainRatesData *mgo.Collection
	ETHMainTxsData   *mgo.Collection

	//eth test
	ETHTestRatesData *mgo.Collection
	ETHTestTxsData   *mgo.Collection

	stockExchangeRate *mgo.Collection
	ethTxHistory      *mgo.Collection
	ETHTest           *mgo.Collection
}

func InitUserStore(conf Conf) (UserStore, error) {
	uStore := &MongoUserStore{
		config: &conf,
	}

	addr := []string{conf.Address}

	mongoDBDial := &mgo.DialInfo{
		Addrs:    addr,
		Username: conf.Username,
		Password: conf.Password,
	}

	session, err := mgo.DialWithInfo(mongoDBDial)
	if err != nil {
		return nil, err
	}

	uStore.session = session
	uStore.usersData = uStore.session.DB(conf.DBUsers).C(TableUsers)
	uStore.stockExchangeRate = uStore.session.DB(conf.DBStockExchangeRate).C(TableStockExchangeRate)

	// BTC main
	uStore.BTCMainRatesData = uStore.session.DB(conf.DBFeeRates).C(conf.TableMempoolRatesBTCMain)
	uStore.BTCMainTxsData = uStore.session.DB(conf.DBTx).C(conf.TableTxsDataBTCMain)
	uStore.BTCMainSpendableOutputs = uStore.session.DB(conf.DBTx).C(conf.TableSpendableOutputsBTCMain)

	// BTC test
	uStore.BTCTestRatesData = uStore.session.DB(conf.DBFeeRates).C(conf.TableMempoolRatesBTCTest)
	uStore.BTCTestTxsData = uStore.session.DB(conf.DBTx).C(conf.TableTxsDataBTCTest)
	uStore.BTCTestSpendableOutputs = uStore.session.DB(conf.DBTx).C(conf.TableSpendableOutputsBTCTest)

	// ETH main
	uStore.ETHMainRatesData = uStore.session.DB(conf.DBFeeRates).C(conf.TableMempoolRatesETHMain)
	uStore.ETHMainTxsData = uStore.session.DB(conf.DBTx).C(conf.TableTxsDataETHMain)

	// ETH test
	uStore.ETHTestRatesData = uStore.session.DB(conf.DBFeeRates).C(conf.TableMempoolRatesETHTest)
	uStore.ETHTestTxsData = uStore.session.DB(conf.DBTx).C(conf.TableTxsDataETHTest)

	return uStore, nil
}

func (mStore *MongoUserStore) FindUserDataChain(CurrencyID, NetworkID int) (map[string]AddressExtended, error) {
	users := []User{}
	usersData := map[string]AddressExtended{} // addres -> userid
	err := mStore.usersData.Find(nil).All(&users)
	if err != nil {
		return usersData, err
	}
	for _, user := range users {
		for _, wallet := range user.Wallets {
			if wallet.CurrencyID == CurrencyID && wallet.NetworkID == NetworkID {
				for _, address := range wallet.Adresses {
					usersData[address.Address] = AddressExtended{
						UserID:       user.UserID,
						WalletIndex:  wallet.WalletIndex,
						AddressIndex: address.AddressIndex,
					}
				}
			}
		}
	}
	return usersData, nil
}

func (mStore *MongoUserStore) DeleteMempool() {
	mStore.BTCMainRatesData.DropCollection()
	mStore.BTCTestRatesData.DropCollection()
}

func (mStore *MongoUserStore) FindAllUserETHTransactions(sel bson.M) ([]TransactionETH, error) {
	allTxs := []TransactionETH{}
	err := mStore.ethTxHistory.Find(sel).All(&allTxs)
	return allTxs, err
}
func (mStore *MongoUserStore) FindETHTransaction(sel bson.M) error {
	err := mStore.ethTxHistory.Find(sel).One(nil)
	return err
}

func (mStore *MongoUserStore) DeleteWallet(userid string, walletindex, currencyID, networkID int) error {
	user := User{}
	sel := bson.M{"userID": userid, "wallets.networkID": networkID, "wallets.currencyID": currencyID, "wallets.walletIndex": walletindex}
	err := mStore.usersData.Find(bson.M{"userID": userid}).One(&user)
	var position int
	if err == nil {
		for i, wallet := range user.Wallets {
			if wallet.NetworkID == networkID && wallet.WalletIndex == walletindex && wallet.CurrencyID == currencyID {
				position = i
				break
			}
		}
		update := bson.M{
			"$set": bson.M{
				"wallets." + strconv.Itoa(position) + ".status": WalletStatusDeleted,
			},
		}
		return mStore.usersData.Update(sel, update)
	}

	return err

}

// func (mStore *MongoUserStore) GetAllSpendableOutputs(query bson.M) (error, []SpendableOutputs) {
// 	spOuts := []SpendableOutputs{}
// 	err := mStore.spendableOutputs.Find(query).All(&spOuts)
// 	return err, spOuts
// }
func (mStore *MongoUserStore) GetAddressSpendableOutputs(address string, currencyID, networkID int) ([]SpendableOutputs, error) {
	spOuts := []SpendableOutputs{}
	var err error

	query := bson.M{"address": address}

	switch currencyID {
	case currencies.Bitcoin:
		if networkID == currencies.Main {
			err = mStore.BTCMainSpendableOutputs.Find(query).All(&spOuts)
		}
		if networkID == currencies.Test {
			err = mStore.BTCTestSpendableOutputs.Find(query).All(&spOuts)
		}
	case currencies.Litecoin:
		if networkID == currencies.Main {

		}
		if networkID == currencies.Test {

		}
	}

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

func (mStore *MongoUserStore) GetAllRates(currencyID, networkID int, sortBy string, rates *[]MempoolRecord) error {
	switch currencyID {
	case currencies.Bitcoin:
		if networkID == currencies.Main {
			return mStore.BTCMainRatesData.Find(nil).Sort(sortBy).All(rates)
		}
		if networkID == currencies.Test {
			return mStore.BTCTestRatesData.Find(nil).Sort(sortBy).All(rates)
		}
	case currencies.Ether:
		if networkID == currencies.ETHMain {
			return mStore.ETHMainRatesData.Find(nil).Sort(sortBy).All(rates)
		}
		if networkID == currencies.ETHTest {
			return mStore.ETHTestRatesData.Find(nil).Sort(sortBy).All(rates)
		}
	}
	return nil
}

func (mStore *MongoUserStore) InsertExchangeRate(eRate ExchangeRates, exchangeStock string) error {
	eRateRecord := &ExchangeRatesRecord{
		Exchanges:     eRate,
		Timestamp:     time.Now().Unix(),
		StockExchange: exchangeStock,
	}

	return mStore.stockExchangeRate.Insert(eRateRecord)
}

// GetExchangeRatesDay returns exchange rates for last day with time interval equal to hour
func (mStore *MongoUserStore) GetExchangeRatesDay() ([]RatesAPIBitstamp, error) {
	// not implemented
	return nil, nil
}

func (mStore *MongoUserStore) GetAllWalletTransactions(userid string, currencyID, networkID int, walletTxs *[]MultyTX) error {
	switch currencyID {
	case currencies.Bitcoin:
		query := bson.M{"userid": userid}
		if networkID == currencies.Main {
			return mStore.BTCMainTxsData.Find(query).All(walletTxs)
		}
		if networkID == currencies.Test {
			return mStore.BTCTestTxsData.Find(query).All(walletTxs)
		}
	}
	return nil
}

func (mStore *MongoUserStore) GetAllWalletEthTransactions(userid string, currencyID, networkID int, walletTxs *[]TransactionETH) error {
	switch currencyID {
	case currencies.Ether:
		query := bson.M{"userid": userid}
		if networkID == currencies.ETHMain {
			return mStore.ETHMainTxsData.Find(query).All(walletTxs)
		}
		if networkID == currencies.ETHTest {
			err := mStore.ETHTestTxsData.Find(query).All(walletTxs)
			return err
		}

	}
	return nil
}

func (mStore *MongoUserStore) Close() error {
	mStore.session.Close()
	return nil
}
