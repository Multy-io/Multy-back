/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package store

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Multy-io/Multy-back/currencies"
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
	TableTxsDataBTCMain          string
	TableSpendableOutputsBTCMain string
	TableSpentOutputsBTCMain     string

	// BTC test
	TableMempoolRatesBTCTest     string
	TableTxsDataBTCTest          string
	TableSpendableOutputsBTCTest string
	TableSpentOutputsBTCTest     string

	// ETH main
	TableMultisigTxsMain string
	TableTxsDataETHMain  string

	// ETH main
	TableMultisigTxsTest string
	TableTxsDataETHTest  string

	//RestoreState
	DBRestoreState string
	TableState     string

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
	// FindUserTxs(query bson.M, userTxs *TxRecord) error
	// InsertTxStore(userTxs TxRecord) error
	FindUserErr(query bson.M) error
	FindUserAddresses(query bson.M, sel bson.M, ws *WalletsSelect) error
	InsertExchangeRate(ExchangeRates, string) error
	GetExchangeRatesDay() ([]RatesAPIBitstamp, error)

	//TODo update this method by eth
	GetAllWalletTransactions(userid string, currencyID, networkID int, walletTxs *[]MultyTX) error
	GetAllWalletEthTransactions(userid string, currencyID, networkID int, walletTxs *[]TransactionETH) error
	GetAllAddressTransactions(address string, currencyID, networkID int, walletTxs *[]TransactionETH) error
	GetAllMultisigEthTransactions(contractAddress string, currencyID, networkID int, walletTxs *[]TransactionETH) error

	// GetAllSpendableOutputs(query bson.M) (error, []SpendableOutputs)
	GetAddressSpendableOutputs(address string, currencyID, networkID int) ([]SpendableOutputs, error)
	DeleteWallet(userid, address string, walletindex, currencyID, networkID, assetType int) error

	FindAllUserETHTransactions(sel bson.M) ([]TransactionETH, error)
	FindUserDataChain(CurrencyID, NetworkID int) (map[string]AddressExtended, error)
	FindUsersContractsChain(CurrencyID, NetworkID int) (map[string]string, error)

	FethUserAddresses(currencyID, networkID int, userid string, addreses []string) (AddressExtended, error)

	FindMultisig(userid, invitecode string) (*Multisig, error)
	JoinMultisig(userid string, multisig *Multisig) error
	LeaveMultisig(userid, invitecode string) error
	KickMultisig(address, invitecode string) error
	DeleteMultisig(invitecode string) error
	CheckInviteCode(invitecode string) bool
	InviteCodeInfo(invitecode string) InviteCodeInfo
	IsRelatedAddress(userid, address string) bool
	CheckMultisigCurrency(invitecode string, currencyid, networkid int) bool
	ViewTransaction(txid, address string, currencyid, networkid int) error
	DeclineTransaction(txid, address string, currencyid, networkid int) error

	FindMultisigUsers(invitecode string) []User
	UpdateMultisigOwners(userid, invitecode string, owners []AddressExtended, deployStatus int) error

	DeleteHistory(CurrencyID, NetworkID int, Address string) error

	FethLastSyncBlockState(networkid, currencyid int) (int64, error)
	// MsToUserData(addresses []string) map[string]User
	// sToUserData(addresses []string) map[string]store.User

	CheckTx(tx string) bool
}

type MongoUserStore struct {
	config    *Conf
	session   *mgo.Session
	usersData *mgo.Collection

	// btc main
	BTCMainTxsData          *mgo.Collection
	BTCMainSpendableOutputs *mgo.Collection
	BTCMainSpentOutputs     *mgo.Collection

	// btc test
	BTCTestTxsData          *mgo.Collection
	BTCTestSpendableOutputs *mgo.Collection
	BTCTestSpentOutputs     *mgo.Collection

	//eth main
	// ETHMainRatesData *mgo.Collection
	ETHMainTxsData *mgo.Collection

	//eth test
	// ETHTestRatesData *mgo.Collection
	ETHTestTxsData *mgo.Collection

	//eth multisig test
	ETHTestMultisigTxsData *mgo.Collection

	//eth multisig main
	ETHMainMultisigTxsData *mgo.Collection

	stockExchangeRate *mgo.Collection
	ethTxHistory      *mgo.Collection
	ETHTest           *mgo.Collection

	RestoreState *mgo.Collection
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
	uStore.BTCMainTxsData = uStore.session.DB(conf.DBTx).C(conf.TableTxsDataBTCMain)
	uStore.BTCMainSpendableOutputs = uStore.session.DB(conf.DBTx).C(conf.TableSpendableOutputsBTCMain)
	uStore.BTCMainSpentOutputs = uStore.session.DB(conf.DBTx).C(conf.TableSpentOutputsBTCMain)

	// BTC test
	uStore.BTCTestTxsData = uStore.session.DB(conf.DBTx).C(conf.TableTxsDataBTCTest)
	uStore.BTCTestSpendableOutputs = uStore.session.DB(conf.DBTx).C(conf.TableSpendableOutputsBTCTest)
	uStore.BTCTestSpentOutputs = uStore.session.DB(conf.DBTx).C(conf.TableSpentOutputsBTCTest)

	// ETH main
	uStore.ETHMainTxsData = uStore.session.DB(conf.DBTx).C(conf.TableTxsDataETHMain)

	// ETH test
	uStore.ETHTestTxsData = uStore.session.DB(conf.DBTx).C(conf.TableTxsDataETHTest)

	//eth multisig test
	uStore.ETHTestMultisigTxsData = uStore.session.DB(conf.DBTx).C(conf.TableMultisigTxsTest)

	//eth multisig main
	uStore.ETHMainMultisigTxsData = uStore.session.DB(conf.DBTx).C(conf.TableMultisigTxsMain)

	uStore.RestoreState = uStore.session.DB(conf.DBRestoreState).C(conf.TableState)

	return uStore, nil
}

func (mStore *MongoUserStore) CheckTx(tx string) bool {
	query := bson.M{"txid": tx}
	// sp := SpendableOutputs{}
	err := mStore.usersData.Find(query).One(nil)
	if err != nil {
		return true
	}
	return false
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

func (mStore *MongoUserStore) FindUsersContractsChain(CurrencyID, NetworkID int) (map[string]string, error) {
	users := []User{}
	UsersContracts := map[string]string{} // addres -> factory address
	err := mStore.usersData.Find(nil).All(&users)
	if err != nil {
		return UsersContracts, err
	}
	for _, user := range users {
		for _, multisig := range user.Multisigs {
			if multisig.CurrencyID == CurrencyID && multisig.NetworkID == NetworkID {
				UsersContracts[multisig.ContractAddress] = multisig.FactoryAddress
			}
		}
	}
	return UsersContracts, nil
}

func (mStore *MongoUserStore) FethUserAddresses(currencyID, networkID int, userid string, addreses []string) (AddressExtended, error) {
	user := User{}
	err := mStore.usersData.Find(bson.M{"userID": userid}).One(&user)
	if err != nil {
		return AddressExtended{}, err
	}
	addresses := []AddressExtended{}

	for _, wallet := range user.Wallets {
		for _, addres := range wallet.Adresses {
			if wallet.CurrencyID == currencyID && wallet.NetworkID == networkID {
				for _, fethAddr := range addreses {

					ae := AddressExtended{
						Address:    fethAddr,
						Associated: false,
					}
					if fethAddr == addres.Address {
						ae.Associated = true
						ae.WalletIndex = wallet.WalletIndex
						ae.AddressIndex = addres.AddressIndex
						ae.UserID = userid
					}
					addresses = append(addresses, ae)
				}

			}
		}
	}
	return AddressExtended{}, nil
}

func (mStore *MongoUserStore) DeleteHistory(CurrencyID, NetworkID int, Address string) error {

	switch CurrencyID {
	case currencies.Bitcoin:
		if NetworkID == currencies.Main {
			mStore.BTCMainTxsData.Remove(bson.M{"txaddress": Address})
			mStore.BTCMainSpendableOutputs.RemoveAll(bson.M{
				"address": Address,
			})
			mStore.BTCMainSpentOutputs.RemoveAll(bson.M{
				"address": Address,
			})
			return nil
		}
		if NetworkID == currencies.Test {
			mStore.BTCTestTxsData.RemoveAll(bson.M{"txaddress": Address})
			mStore.BTCTestSpendableOutputs.RemoveAll(bson.M{
				"address": Address,
			})
			mStore.BTCTestSpentOutputs.RemoveAll(bson.M{
				"address": Address,
			})
			return nil
		}
	case currencies.Ether:
		if NetworkID == currencies.ETHMain {

		}
		if NetworkID == currencies.ETHTest {

		}
	}
	return nil
}

func (mStore *MongoUserStore) FethLastSyncBlockState(networkid, currencyid int) (int64, error) {
	ls := LastState{}
	sel := bson.M{"networkid": networkid, "currencyid": currencyid}
	err := mStore.RestoreState.Find(sel).Sort("blockheight").One(&ls)
	return ls.BlockHeight, err
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

func (mStore *MongoUserStore) DeleteWallet(userid, address string, walletindex, currencyID, networkID, assetType int) error {
	var err error
	switch assetType {
	case AssetTypeMultyAddress:
		user := User{}
		sel := bson.M{"userID": userid, "wallets.networkID": networkID, "wallets.currencyID": currencyID, "wallets.walletIndex": walletindex}
		err = mStore.usersData.Find(bson.M{"userID": userid}).One(&user)
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
	case AssetTypeImportedAddress:
		query := bson.M{"userID": userid, "wallets.addresses.address": address, "true": true}
		update := bson.M{
			"$set": bson.M{
				"wallets.$.status": WalletStatusDeleted,
			},
		}
		err := mStore.usersData.Update(query, update)
		if err != nil {
			return errors.New("DeleteWallet:restClient.userStore.Update:AssetTypeImportedAddress " + err.Error())
		}
		return err
	case AssetTypeMultisig:

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
			err := mStore.ETHMainTxsData.Find(query).All(walletTxs)
			return err
		}
		if networkID == currencies.ETHTest {
			err := mStore.ETHTestTxsData.Find(query).All(walletTxs)
			return err
		}

	}
	return nil
}

func (mStore *MongoUserStore) GetAllAddressTransactions(address string, currencyID, networkID int, walletTxs *[]TransactionETH) error {
	switch currencyID {
	case currencies.Ether:
		query := bson.M{
			"$or": []bson.M{
				bson.M{"to": address},
				bson.M{"from": address},
			},
			"userid": "imported",
		}

		if networkID == currencies.ETHMain {
			err := mStore.ETHMainTxsData.Find(query).All(walletTxs)
			return err
		}
		if networkID == currencies.ETHTest {
			err := mStore.ETHTestTxsData.Find(query).All(walletTxs)
			return err
		}
	}
	return nil
}

func (mStore *MongoUserStore) GetAllMultisigEthTransactions(contractAddress string, currencyID, networkID int, multisigTxs *[]TransactionETH) error {
	switch currencyID {
	case currencies.Ether:
		query := bson.M{
			"$or": []bson.M{
				bson.M{"to": contractAddress},
				bson.M{"from": contractAddress},
			},
		}
		if networkID == currencies.ETHMain {
			return mStore.ETHMainMultisigTxsData.Find(query).All(multisigTxs)
		}
		if networkID == currencies.ETHTest {
			err := mStore.ETHTestMultisigTxsData.Find(query).All(multisigTxs)
			return err
		}

	}
	return nil
}

func (mStore *MongoUserStore) FindMultisig(userid, invitecode string) (*Multisig, error) {

	users := []User{}
	multisig := Multisig{}

	// // // only accept one address from one user in multisig
	// sel := bson.M{"userID": userid, "multisig.inviteCode": invitecode}
	// err := mStore.usersData.Find(sel).One(nil)
	// if err == mgo.ErrNotFound {
	// 	return &multisig, errors.New("User: " + userid + " don't have this multsig")
	// }

	sel := bson.M{"multisig.inviteCode": invitecode}
	err := mStore.usersData.Find(sel).All(&users)
	if err != nil {
		fmt.Println("No such multisigs with this invite code")
		return &multisig, errors.New("No such multisigs with this invite code")
	}

	if len(users) > 0 {
		for _, mu := range users[0].Multisigs {
			if mu.InviteCode == invitecode {
				return &mu, nil
			}
		}
	}
	if len(users) == 0 {
		return &multisig, errors.New("No such multisigs with this invite code")
	}

	return &multisig, nil
}

func (mStore *MongoUserStore) JoinMultisig(userid string, multisig *Multisig) error {
	sel := bson.M{"userID": userid}
	update := bson.M{"$push": bson.M{"multisig": multisig}}
	return mStore.usersData.Update(sel, update)
}
func (mStore *MongoUserStore) LeaveMultisig(userid, invitecode string) error {
	sel := bson.M{"userID": userid}
	user := User{}

	multisigs := []Multisig{}
	err := mStore.usersData.Find(sel).One(&user)
	if err != nil {
		return err
	}
	for _, multisig := range user.Multisigs {
		if multisig.InviteCode != invitecode {
			multisigs = append(multisigs, multisig)
		}
	}
	update := bson.M{"$set": bson.M{"multisig": multisigs}}
	return mStore.usersData.Update(sel, update)
}
func (mStore *MongoUserStore) KickMultisig(address, invitecode string) error {
	sel := bson.M{"wallets.addresses.address": address}
	user := User{}

	multisigs := []Multisig{}
	err := mStore.usersData.Find(sel).One(&user)
	if err != nil {
		return err
	}
	for _, multisig := range user.Multisigs {
		if multisig.InviteCode != invitecode {
			multisigs = append(multisigs, multisig)
		}
	}
	update := bson.M{"$set": bson.M{"multisig": multisigs}}
	return mStore.usersData.Update(sel, update)
}

func (mStore *MongoUserStore) DeleteMultisig(invitecode string) error {
	sel := bson.M{"multisig.inviteCode": invitecode}
	users := []User{}
	mStore.usersData.Find(sel).All(&users)
	var err error
	for _, user := range users {
		okMultisigs := []Multisig{}
		for _, multisig := range user.Multisigs {
			if multisig.InviteCode != invitecode {
				okMultisigs = append(okMultisigs, multisig)
			}
		}
		sel = bson.M{"userID": user.UserID}
		update := bson.M{"$set": bson.M{"multisig": okMultisigs}}
		err = mStore.usersData.Update(sel, update)

	}
	return err
}

func (mStore *MongoUserStore) CheckInviteCode(invitecode string) bool {
	sel := bson.M{"multisig.inviteCode": invitecode}
	err := mStore.usersData.Find(sel).One(nil)
	if err == mgo.ErrNotFound {
		return true
	}
	return false
}

func (mStore *MongoUserStore) InviteCodeInfo(invitecode string) InviteCodeInfo {
	sel := bson.M{"multisig.inviteCode": invitecode}
	user := User{}
	inCodeInfo := InviteCodeInfo{}
	_ = mStore.usersData.Find(sel).One(&user)
	for _, multisig := range user.Multisigs {
		if multisig.InviteCode == invitecode {
			inCodeInfo = InviteCodeInfo{
				CurrencyID: multisig.CurrencyID,
				NetworkID:  multisig.NetworkID,
				Exists:     true,
			}
		}
	}
	return inCodeInfo
}

func (mStore *MongoUserStore) CheckMultisigCurrency(invitecode string, currencyid, networkid int) bool {
	sel := bson.M{"multisig.inviteCode": invitecode}
	user := User{}
	inCodeInfo := InviteCodeInfo{}
	_ = mStore.usersData.Find(sel).One(&user)
	for _, multisig := range user.Multisigs {
		if multisig.InviteCode == invitecode {
			inCodeInfo = InviteCodeInfo{
				CurrencyID: multisig.CurrencyID,
				NetworkID:  multisig.NetworkID,
				Exists:     true,
			}
		}
	}

	if inCodeInfo.Exists && inCodeInfo.NetworkID == networkid && inCodeInfo.CurrencyID == currencyid {
		return true
	}
	return false
}

func (mStore *MongoUserStore) ViewTransaction(txid, address string, currencyid, networkid int) error {
	switch currencyid {
	case currencies.Ether:
		update := bson.M{"$set": bson.M{
			"multisig.owners.$.confirmationStatus": MultisigOwnerStatusSeen,
			"multisig.owners.$.seenTime":           time.Now().Unix(),
		}}
		sel := bson.M{"hash": txid, "multisig.owners.address": address}
		ms := TransactionETH{}
		if networkid == currencies.ETHMain {
			err := mStore.ETHMainMultisigTxsData.Find(sel).One(&ms)
			if ms.Multisig != nil {
				for _, owner := range ms.Multisig.Owners {
					if owner.Address == address && owner.ConfirmationTime != 0 {
						return errors.New("transaction already seen")
					}
				}
			}

			err = mStore.ETHMainMultisigTxsData.Update(sel, update)
			return err
		}
		if networkid == currencies.ETHTest {
			err := mStore.ETHTestMultisigTxsData.Find(sel).One(&ms)
			if ms.Multisig.Owners != nil {
				for _, owner := range ms.Multisig.Owners {
					if owner.Address == address && owner.ConfirmationTime != 0 {
						return errors.New("transaction already seen")
					}
				}
			}
			err = mStore.ETHTestMultisigTxsData.Update(sel, update)
			return err
		}
	}
	return nil
}

func (mStore *MongoUserStore) DeclineTransaction(txid, address string, currencyid, networkid int) error {
	switch currencyid {
	case currencies.Ether:
		sel := bson.M{"hash": txid, "multisig.owners.address": address}
		fmt.Println(sel)
		update := bson.M{"$set": bson.M{
			"multisig.owners.$.confirmationStatus": MultisigOwnerStatusDeclined,
		}}
		if networkid == currencies.ETHMain {
			err := mStore.ETHMainMultisigTxsData.Update(sel, update)
			return err
		}
		if networkid == currencies.ETHTest {
			err := mStore.ETHTestMultisigTxsData.Update(sel, update)
			return err
		}
	}
	return nil
}

func (mStore *MongoUserStore) IsRelatedAddress(userid, address string) bool {
	sel := bson.M{"userID": userid, "wallets.addresses.address": address}
	err := mStore.usersData.Find(sel).One(nil)
	if err == mgo.ErrNotFound {
		return false
	}
	return true
}

func (mStore *MongoUserStore) FindMultisigUsers(invitecode string) []User {
	sel := bson.M{"multisig.inviteCode": invitecode}
	users := []User{}
	mStore.usersData.Find(sel).All(&users)
	return users
}
func (mStore *MongoUserStore) UpdateMultisigOwners(userid, invitecode string, owners []AddressExtended, deployStatus int) error {
	sel := bson.M{"userID": userid, "multisig.inviteCode": invitecode}
	update := bson.M{"$set": bson.M{
		"multisig.$.owners":       owners,
		"multisig.$.deployStatus": deployStatus,
	}}
	return mStore.usersData.Update(sel, update)
}

func (mStore *MongoUserStore) Close() error {
	mStore.session.Close()
	return nil
}
