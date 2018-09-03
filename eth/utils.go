/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package eth

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Multy-io/Multy-back/currencies"
	ethpb "github.com/Multy-io/Multy-back/node-streamer/eth"
	"github.com/Multy-io/Multy-back/store"
	nsq "github.com/bitly/go-nsq"
	_ "github.com/jekabolt/slflog"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	MultiSigFactory    = "0xf8f73808"
	submitTransaction  = "0xc6427474"
	confirmTransaction = "0xc01a8c84"
	revokeConfirmation = "0x20ea8d86"
	executeTransaction = "0xee22610b"
	// GasLimit           = map[string]string{
	// 	submitTransaction:  "7039920",
	// 	revokeConfirmation: "7039920",
	// 	confirmTransaction: "7039920",
	// 	executeTransaction: "7039920",
	// }
)

var (
	exRate    *mgo.Collection
	usersData *mgo.Collection

	txsData      *mgo.Collection
	multisigData *mgo.Collection

	txsDataTest      *mgo.Collection
	multisigDataTest *mgo.Collection

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

func sendNotifyToClients(tx store.TransactionETH, nsqProducer *nsq.Producer, netid int, userid ...string) {
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
				From:            tx.From,
				To:              tx.To,
				Multisig:        tx.Multisig.Contract,
			},
		}
		if len(userid) > 0 {
			txMsq.UserID = userid[0]
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
				From:            tx.From,
				To:              tx.To,
				Multisig:        tx.Multisig.Contract,
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
		UserID:       tx.GetUserID(),
		WalletIndex:  int(tx.GetWalletIndex()),
		AddressIndex: int(tx.GetAddressIndex()),
		Hash:         tx.GetHash(),
		From:         tx.GetFrom(),
		To:           tx.GetTo(),
		Amount:       tx.GetAmount(),
		GasPrice:     tx.GetGasPrice(),
		GasLimit:     tx.GetGasLimit(),
		Nonce:        int(tx.GetNonce()),
		Status:       int(tx.GetStatus()),
		BlockTime:    tx.GetBlockTime(),
		PoolTime:     tx.GetTxpoolTime(),
		BlockHeight:  tx.GetBlockHeight(),
		Multisig: &store.MultisigTx{
			Contract:         tx.GetContract(),
			MethodInvoked:    tx.GetMethodInvoked(),
			InvocationStatus: tx.GetInvocationStatus(),
			Return:           tx.GetReturn(),
			Input:            tx.GetInput(),
		},
		// Contract:         tx.GetContract(),
		// MethodInvoked:    tx.GetMethodInvoked(),
		// InvocationStatus: tx.GetInvocationStatus(),
		// Return:           tx.GetReturn(),
		// Input:            tx.GetInput(),
	}
}

func saveTransaction(tx store.TransactionETH, networtkID int, resync bool) error {

	txStore := &mgo.Collection{}
	switch networtkID {
	case currencies.ETHMain:
		txStore = txsData
	case currencies.ETHTest:
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

func processMultisig(tx *store.TransactionETH, networtkID int, nsqProducer *nsq.Producer) error {

	multisigStore := &mgo.Collection{}
	txStore := &mgo.Collection{}
	switch networtkID {
	case currencies.ETHMain:
		multisigStore = multisigData
		txStore = txsData
	case currencies.ETHTest:
		multisigStore = multisigDataTest
		txStore = txsDataTest
	default:
		return errors.New("processMultisig: wrong networkID")
	}

	//TODO: fix netids
	if networtkID == 0 {
		networtkID = currencies.ETHMain
	}
	if networtkID == 1 {
		networtkID = currencies.ETHTest
	}

	tx.Multisig.Contract = tx.To
	multyTX := &store.TransactionETH{}
	if tx.Status == store.TxStatusAppearedInBlockIncoming || tx.Status == store.TxStatusAppearedInMempoolIncoming || tx.Status == store.TxStatusInBlockConfirmedIncoming {
		log.Debugf("saveTransaction new incoming tx to %v", tx.To)
		sel := bson.M{"hash": tx.Hash}
		err := multisigStore.Find(sel).One(nil)
		if err == mgo.ErrNotFound {
			multyTX = ParseMultisigInput(tx, networtkID, multisigStore, txStore, nsqProducer)
			err := multisigStore.Insert(multyTX)
			return err
		}

		multyTX = ParseMultisigInput(tx, networtkID, multisigStore, txStore, nsqProducer)
		if multyTX.Multisig.MethodInvoked == "0xc6427474" && multyTX.Status == store.TxStatusAppearedInBlockIncoming {
			multyTX.Status = store.TxStatusInBlockConfirmedOutcoming
		}

		if err != nil && err != mgo.ErrNotFound {
			// database error
			return err
		}

		update := bson.M{
			"$set": bson.M{
				"txstatus":                  tx.Status,
				"blockheight":               tx.BlockHeight,
				"blocktime":                 tx.BlockTime,
				"amount":                    tx.Amount,
				"multisig.index":            multyTX.Multisig.Index,
				"multisig.return":           tx.Multisig.Return,
				"multisig.invocationstatus": tx.Multisig.InvocationStatus,
				"multisig.confirmed":        tx.Multisig.Confirmed,
			},
		}

		err = multisigStore.Update(sel, update)

		return err
	}
	return nil
}

func ParseMultisigInput(tx *store.TransactionETH, networtkID int, multisigStore, txStore *mgo.Collection, nsqProducer *nsq.Producer) *store.TransactionETH { // method

	owners, _ := FethContractOwners(currencies.Ether, networtkID, tx.Multisig.Contract)
	tx.Multisig.Owners = owners
	tx.Multisig.MethodInvoked = fethMethod(tx.Multisig.Input)

	users := findContractOwners(tx.To)
	contract, err := fethMultisig(users, tx.To)
	if err != nil {
		log.Errorf("ParseMultisigInput:fethMultisig: %v", err.Error())
	}

	fmt.Println("-------------------------- ", owners)

	switch tx.Multisig.MethodInvoked {
	case submitTransaction: // "c6427474": "submitTransaction(address,uint256,bytes)"
		// Feth contract owners, send notfy to owners about transation. status: waiting for confirmations
		// find in db if one one confirmation needed DONE internal transaction
		log.Debugf("submitTransaction:  Input :%v Return :%v ", tx.Multisig.Input, tx.Multisig.Return)
		if tx.BlockTime != 0 {
			i, _ := new(big.Int).SetString(tx.Multisig.Return, 16)
			tx.Multisig.Index = i.Int64()

			address, amount := parseSubmitInput(tx.Multisig.Input)

			tx.Amount = amount

			// fakeit intrrnal transaction history
			if contract.Confirmations == 1 {
				tx.Multisig.Confirmed = true
				sel := bson.M{"txhash": tx.Hash}
				err := txStore.Find(sel).One(nil)
				if err == mgo.ErrNotFound {
					// initial insertion
					user := store.User{}
					txToUser := tx
					txToUser.From = contract.ContractAddress
					txToUser.To = address

					sel := bson.M{"wallets.addresses.address": address}
					_ = usersData.Find(sel).One(&user)

					isOurUser := false
					// internal tansaction to wallet
					for _, wallet := range user.Wallets {
						for _, adr := range wallet.Adresses {
							if adr.Address == address {
								txToUser.From = contract.ContractAddress
								txToUser.To = adr.Address
								txToUser.PoolTime = time.Now().Unix()
								txToUser.Multisig.MethodInvoked = executeTransaction
								isOurUser = true
								txToUser.UserID = user.UserID
								txToUser.WalletIndex = wallet.WalletIndex
								txToUser.AddressIndex = adr.AddressIndex
								txToUser.Amount = amount
								txToUser.Hash = tx.Hash
								txToUser.Status = store.TxStatusInBlockConfirmedOutcoming
								txToUser.IsInternal = true
								break
							}
						}
					}

					if isOurUser {
						_ = txStore.Insert(tx)
					}
					isOurUser = false

					sel = bson.M{"multisig.contractAddress": address}
					usersData.Find(sel).One(&user)
					// internal tansaction to multisig
					for _, multisig := range user.Multisigs {
						if multisig.ContractAddress == address {
							txToUser.From = contract.ContractAddress
							txToUser.To = multisig.ContractAddress
							txToUser.PoolTime = time.Now().Unix()
							txToUser.Multisig.MethodInvoked = executeTransaction
							isOurUser = true
							txToUser.Amount = amount
							txToUser.Hash = tx.Hash
							txToUser.Status = store.TxStatusInBlockConfirmedIncoming
							txToUser.IsInternal = true
							break
						}
					}

					if isOurUser {
						_ = multisigStore.Insert(tx)
					}

				}

				if err != nil && err != mgo.ErrNotFound {
					// database error
					log.Errorf("ParseMultisigInput:confirmTransaction:multisigStore.Find %v", err.Error())
				}

			}
		}

		// update confirmations history
		ownerHistorys := []store.OwnerHistory{}
		for _, ownerHistory := range tx.Multisig.Owners {
			if ownerHistory.Address == tx.From {
				ownerHistorys = append(ownerHistorys, store.OwnerHistory{
					Address:            tx.From,
					ConfirmationTX:     tx.Hash,
					ConfirmationStatus: store.MultisigOwnerStatusConfirmed,
					ConfirmationTime:   time.Now().Unix(),
					SeenTime:           time.Now().Unix(),
				})
			} else {
				ownerHistorys = append(ownerHistorys, ownerHistory)
			}
		}

		tx.Multisig.Owners = ownerHistorys

		//TODO: notifications
		// //notify users
		// for _, user := range users {
		// 	sendNotifyToClients(*tx, nsqProducer, networtkID, user.UserID)
		// }

		return tx

	case confirmTransaction: // "c01a8c84": "confirmTransaction(uint256)"
		//TODO: send notfy to owners about +1 confirmation. store confiramtions id db

		//TODO: outcoming tranastions for multisig
		log.Debugf("confirmTransaction: %v", tx.Multisig.Input)
		i, _ := new(big.Int).SetString(tx.Multisig.Input[10:], 16)
		sel := bson.M{"multisig.index": i.Int64(), "multisig.contract": tx.Multisig.Contract}

		originTx := store.TransactionETH{}
		err := multisigStore.Find(sel).One(&originTx)
		if err != nil {
			log.Errorf("ParseMultisigInput:confirmTransaction:multisigStore.Find %v index:%v  contract:%v ", err.Error(), i.Int64(), contract.ContractAddress)
		}

		//todo update only on block and exec true
		ownerHistorys := []store.OwnerHistory{}

		for _, ownerHistory := range originTx.Multisig.Owners {
			if ownerHistory.Address == tx.From {
				ownerHistorys = append(ownerHistorys, store.OwnerHistory{
					Address:            tx.From,
					ConfirmationTX:     tx.Hash,
					ConfirmationStatus: store.MultisigOwnerStatusConfirmed,
					ConfirmationTime:   time.Now().Unix(),
					SeenTime:           time.Now().Unix(),
				})
			} else {
				ownerHistorys = append(ownerHistorys, ownerHistory)
			}
		}

		update := bson.M{
			"$set": bson.M{
				"multisig.owners": ownerHistorys,
			},
		}

		// update confirmations history
		err = multisigStore.Update(sel, update)
		if err != nil {
			log.Errorf("ParseMultisigInput:confirmTransaction:multisigStore.Update %v index:%v  contract:%v ", err.Error(), originTx.Multisig.Index, contract.ContractAddress)
		}

		tx.Multisig.Owners = []store.OwnerHistory{}

		confirmations := 0
		for _, oh := range ownerHistorys {
			if oh.ConfirmationStatus == store.MultisigOwnerStatusConfirmed {
				confirmations++
			}
		}

		// Internal transaction contract to user
		if contract.Confirmations <= confirmations && tx.BlockTime != 0 {
			tx.Multisig.Confirmed = true
			//update owners history
			sel := bson.M{"hash": originTx.Hash}
			update := bson.M{
				"$set": bson.M{
					"multisig.owners":    ownerHistorys,
					"multisig.confirmed": true,
				},
			}
			err = multisigStore.Update(sel, update)
			if err != nil {
				log.Errorf("ParseMultisigInput:confirmTransaction:multisigStore.Update:contract.Confirmations %v contract:%v ", err.Error(), contract.ContractAddress)
			}

			sel = bson.M{"hash": originTx.Hash, "isinternal": true}
			err := txStore.Find(sel).One(nil)
			txToUser := tx
			if err == mgo.ErrNotFound {
				// initial insertion

				log.Debugf("Internal transaction:", MultiSigFactory)

				isOurUser := false

				user := store.User{}
				outputAddress, amount := parseSubmitInput(originTx.Multisig.Input)

				// internal transaction contract to addres
				sel := bson.M{"wallets.addresses.address": outputAddress}
				_ = usersData.Find(sel).One(&user)
				for _, wallet := range user.Wallets {
					for _, adr := range wallet.Adresses {
						if adr.Address == outputAddress {
							txToUser.From = contract.ContractAddress
							txToUser.To = adr.Address
							txToUser.PoolTime = time.Now().Unix()
							txToUser.Multisig.MethodInvoked = executeTransaction
							isOurUser = true
							txToUser.UserID = user.UserID
							txToUser.WalletIndex = wallet.WalletIndex
							txToUser.AddressIndex = adr.AddressIndex
							txToUser.Amount = amount
							txToUser.Hash = tx.Hash
							txToUser.Status = store.TxStatusInBlockConfirmedIncoming
							txToUser.IsInternal = true
							break
						}
					}
				}

				if isOurUser {
					log.Warnf("not our user")
					_ = txStore.Insert(txToUser)
				}

				// contract to contract history
				isOurUser = false
				sel = bson.M{"multisig.contractAddress": outputAddress}
				usersData.Find(sel).One(&user)
				// internal tansaction to multisig
				for _, multisig := range user.Multisigs {
					if multisig.ContractAddress == outputAddress {
						txToUser.From = contract.ContractAddress
						txToUser.To = multisig.ContractAddress
						txToUser.PoolTime = time.Now().Unix()
						txToUser.Multisig.MethodInvoked = executeTransaction
						isOurUser = true
						txToUser.Amount = amount
						txToUser.Hash = tx.Hash
						txToUser.Status = store.TxStatusInBlockConfirmedIncoming
						txToUser.IsInternal = true
						break
					}
				}

				if isOurUser {
					_ = multisigStore.Insert(tx)
				}
			}
			if err != nil && err != mgo.ErrNotFound {
				// database error
				log.Errorf("ParseMultisigInput:confirmTransaction:multisigStore.Find %v", err.Error())
			}
		}

		return tx

	case revokeConfirmation: // "20ea8d86": "revokeConfirmation(uint256)"
		// TODO: send notfy to owners about -1 confirmation. store confirmations in db

		tx.Multisig.Owners = []store.OwnerHistory{}

		log.Debugf("revokeConfirmation: %v", tx.Multisig.Input)
		i, _ := new(big.Int).SetString(tx.Multisig.Input, 16)

		sel := bson.M{"multisig.index": i.Int64(), "multisig.contract": contract.ContractAddress}

		originTx := store.TransactionETH{}
		err := multisigStore.Find(sel).One(&originTx)
		if err != nil {
			log.Errorf("ParseMultisigInput:revokeConfirmation:multisigStore.Find %v index:%v  contract:%v ", err.Error(), i.Int64(), contract.ContractAddress)
		}
		ownerHistorys := []store.OwnerHistory{}
		for _, ownerHistory := range tx.Multisig.Owners {
			if ownerHistory.Address == originTx.From {
				ownerHistorys = append(ownerHistorys, store.OwnerHistory{
					Address:            tx.From,
					ConfirmationStatus: store.MultisigOwnerStatusDeclined,
					ConfirmationTime:   time.Now().Unix(),
					SeenTime:           time.Now().Unix(),
				})
			}
			ownerHistorys = append(ownerHistorys, ownerHistory)
		}

		update := bson.M{
			"$set": bson.M{
				"multisig.owners": ownerHistorys,
			},
		}

		err = multisigStore.Update(sel, update)
		if err != nil {
			log.Errorf("ParseMultisigInput:revokeConfirmation:multisigStore.Update %v index:%v  contract:%v ", err.Error(), i.Int64(), contract.ContractAddress)
		}

		return tx
	case "0x": // incoming transaction
		// TODO: notify owners about new transation
		log.Debugf("incoming transaction: %v", tx.Multisig.Input)
		return tx

	default:
		log.Errorf("wrong method: ", tx.Multisig.Input)
		return tx
		// wrong method
	}

}

func generatedMultisigTxToStore(mul *ethpb.Multisig, currenyid, networkid int) store.Multisig {
	return store.Multisig{
		CurrencyID:      currenyid,
		NetworkID:       networkid,
		Confirmations:   int(mul.GetConfirmations()),
		ContractAddress: mul.GetContract(),
		TxOfCreation:    mul.GetTxOfCreation(),
		FactoryAddress:  mul.GetFactoryAddress(),
		LastActionTime:  time.Now().Unix(),
		DateOfCreation:  time.Now().Unix(),
		DeployStatus:    int(mul.GetDeployStatus()),
		Status:          store.WalletStatusOK,
	}
}

func FethUserAddresses(currencyID, networkID int, user store.User, addreses []string) ([]store.AddressExtended, error) {
	addresses := []store.AddressExtended{}
	fethed := map[string]store.AddressExtended{}

	for _, address := range addreses {
		fethed[address] = store.AddressExtended{
			Address:    address,
			Associated: false,
		}
	}

	for _, wallet := range user.Wallets {
		for _, addres := range wallet.Adresses {
			if wallet.CurrencyID == currencyID && wallet.NetworkID == networkID {
				for addr, fethAddr := range fethed {
					if addr == addres.Address {
						fethAddr.Associated = true
						fethAddr.WalletIndex = wallet.WalletIndex
						fethAddr.AddressIndex = addres.AddressIndex
						fethAddr.UserID = user.UserID
						fethed[addres.Address] = fethAddr
					}
				}
			}
		}
	}

	for _, addr := range fethed {
		addresses = append(addresses, addr)
	}

	return addresses, nil
}

func FethContractOwners(currencyID, networkID int, contractaddress string) ([]store.OwnerHistory, error) {
	oh := []store.OwnerHistory{}

	sel := bson.M{"multisig.contractAddress": contractaddress}
	user := store.User{}
	_ = usersData.Find(sel).One(&user)

	for _, multisig := range user.Multisigs {
		for _, owner := range multisig.Owners {
			if multisig.CurrencyID == currencyID && multisig.NetworkID == networkID {
				oh = append(oh, store.OwnerHistory{
					Address: owner.Address,
				})
			}
		}
	}
	return oh, nil
}

func fethMethod(input string) string {
	method := input
	if len(input) < 10 {
		method = "0x"
	} else {
		method = input[:10]
	}

	return method
}

func fethMultisig(users []store.User, contract string) (*store.Multisig, error) {
	if len(users) > 0 {
		for _, m := range users[0].Multisigs {
			if m.ContractAddress == contract {
				return &m, nil
			}
		}
	}

	return &store.Multisig{}, errors.New("fethMultisig: contract have no multy users :" + contract)
}

func findContractOwners(contractAddress string) []store.User {
	users := []store.User{}
	err := usersData.Find(bson.M{"multisig.contractAddress": strings.ToLower(contractAddress)}).All(&users)
	if err != nil {
		log.Errorf("cli.AddMultisig:stream.Recv:usersData.Find: not multy user in contrat %v  %v", err.Error(), contractAddress)
	}
	return users
}

func parseSubmitInput(input string) (string, string) {
	address := ""
	amount := ""
	if len(input) >= 266 {
		in := input[10:]
		re := regexp.MustCompile(`.{64}`) // Every 64 chars
		parts := re.FindAllString(in, -1) // Split the string into 64 chars blocks.

		if len(parts) == 4 {
			address = strings.ToLower("0x" + parts[0][24:])
			a, _ := new(big.Int).SetString(parts[1], 16)
			amount = a.String()
		}
	}

	return address, amount
}
