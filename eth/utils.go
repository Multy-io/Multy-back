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
	ethpb "github.com/Multy-io/Multy-back/ns-eth-protobuf"
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
			if addr.Address == tx.From && user.Wallets[i].NetworkID == networkID {
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
			if addr.Address == tx.To && user.Wallets[i].NetworkID == networkID {
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
				"gasprice":    tx.GasPrice,
				"gaslimit":    tx.GasLimit,
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
				"gasprice":    tx.GasPrice,
				"gaslimit":    tx.GasLimit,
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

func processMultisig(tx *store.TransactionETH, networtkID int, nsqProducer *nsq.Producer, ethcli *ETHConn) (string, error) {

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
		return "", errors.New("processMultisig: wrong networkID")
	}

	tx.Multisig.Contract = tx.To
	multyTX := &store.TransactionETH{}
	if tx.Status == store.TxStatusAppearedInBlockIncoming || tx.Status == store.TxStatusAppearedInMempoolIncoming || tx.Status == store.TxStatusInBlockConfirmedIncoming {
		log.Debugf("saveTransaction new incoming tx to %v", tx.To)

		sel := bson.M{"hash": tx.Hash}
		err := multisigStore.Find(sel).One(nil)
		if err == mgo.ErrNotFound {
			multyTX = ParseMultisigInput(tx, networtkID, multisigStore, txStore, nsqProducer, ethcli)
			if multyTX.Multisig.MethodInvoked == submitTransaction && multyTX.Status == store.TxStatusAppearedInBlockIncoming {
				multyTX.Status = store.TxStatusInBlockConfirmedOutcoming
			}
			err := multisigStore.Insert(multyTX)
			return "", err
		}

		multyTX = ParseMultisigInput(tx, networtkID, multisigStore, txStore, nsqProducer, ethcli)
		if multyTX.Multisig.MethodInvoked == submitTransaction && multyTX.Status == store.TxStatusAppearedInBlockIncoming {
			multyTX.Status = store.TxStatusInBlockConfirmedOutcoming
		}

		if err != nil && err != mgo.ErrNotFound {
			return "", err
		}

		update := bson.M{
			"$set": bson.M{
				"from":                      tx.From,
				"to":                        tx.To,
				"txstatus":                  tx.Status,
				"blockheight":               tx.BlockHeight,
				"blocktime":                 tx.BlockTime,
				"amount":                    tx.Amount,
				"multisig.requestid":        tx.Multisig.RequestID,
				"multisig.return":           tx.Multisig.Return,
				"multisig.invocationstatus": tx.Multisig.InvocationStatus,
				"multisig.confirmed":        tx.Multisig.Confirmed,
			},
		}

		err = multisigStore.Update(sel, update)

		return tx.Multisig.MethodInvoked, err
	}
	return "", nil
}

func ParseMultisigInput(tx *store.TransactionETH, networtkID int, multisigStore, txStore *mgo.Collection, nsqProducer *nsq.Producer, ethcli *ETHConn) *store.TransactionETH { // method

	owners, _ := FetchContractOwners(currencies.Ether, networtkID, tx.Multisig.Contract)
	tx.Multisig.Owners = owners
	tx.Multisig.MethodInvoked = fetchMethod(tx.Multisig.Input)

	users := findContractOwners(tx.To)
	contract, err := fetchMultisig(users, tx.To)
	if err != nil {
		log.Errorf("ParseMultisigInput:fetchMultisig: %v", err.Error())
	}

	switch tx.Multisig.MethodInvoked {
	case submitTransaction: // "c6427474": "submitTransaction(address,uint256,bytes)"
		// Fetch contract owners, send notfy to owners about transation. status: waiting for confirmations
		// find in db if one one confirmation needed DONE internal transaction
		log.Debugf("submitTransaction:  Input :%v Return :%v ", tx.Multisig.Input, tx.Multisig.Return)
		if tx.BlockTime != 0 {
			i, _ := new(big.Int).SetString(tx.Multisig.Return, 16)
			if i == nil || tx.Multisig.Return == "" {
				log.Errorf("ParseMultisigInput:confirmTransaction:empty return from contract %v", tx.Hash)
				tx.Multisig.InvocationStatus = false
				return tx
			}

			if !tx.Multisig.InvocationStatus {
				tx.Status = store.TxStatusInBlockMethodInvocationFail
				sel := bson.M{"txhash": tx.Hash}
				update := bson.M{
					"$set": bson.M{
						"txstatus": store.TxStatusInBlockMethodInvocationFail,
					},
				}
				txStore.UpdateAll(sel, update)
			}
			tx.Multisig.RequestID = i.Int64()

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
					txToUser := *tx
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
						txToUser.Multisig = nil
						_ = txStore.Insert(txToUser)
					}
					isOurUser = false

					sel = bson.M{"multisig.contractAddress": address}
					usersData.Find(sel).One(&user)
					// internal transaction to multisig
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
							txToUser.Multisig.Contract = multisig.ContractAddress
							break
						}
					}

					if isOurUser {
						_ = multisigStore.Insert(txToUser)
					}

				}

				if err != nil && err != mgo.ErrNotFound {
					// database error
					log.Errorf("ParseMultisigInput:confirmTransaction:multisigStore.Find %v", err.Error())
				}

				// notify on submit transaction

				for _, user := range users {
					msg := store.WsMessage{
						Type:    store.NotifyTxSubmitted,
						To:      user.UserID,
						Date:    time.Now().Unix(),
						Payload: "ok",
					}
					ethcli.WsServer.BroadcastToAll(store.MsgReceive+":"+user.UserID, msg)
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

		address, amount := parseSubmitInput(tx.Multisig.Input)
		tx.From = tx.To
		tx.To = address
		tx.Amount = amount

		return tx

	case confirmTransaction: // "c01a8c84": "confirmTransaction(uint256)"
		log.Warnf("confirmTransaction: %v", tx.Multisig.Input)
		i, _ := new(big.Int).SetString(tx.Multisig.Input[10:], 16)
		sel := bson.M{"multisig.requestid": i.Int64(), "multisig.contract": tx.Multisig.Contract, "multisig.methodinvoked": submitTransaction}
		log.Warnf("confirmTransaction:sel %v", sel)

		originTx := store.TransactionETH{}
		err := multisigStore.Find(sel).One(&originTx)
		if err != nil {
			log.Errorf("ParseMultisigInput:confirmTransaction:multisigStore.Find %v requestid:%v  contract:%v ", err.Error(), i.Int64(), contract.ContractAddress)
			return tx
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
		_, err = multisigStore.UpdateAll(sel, update)
		if err != nil {
			log.Errorf("ParseMultisigInput:confirmTransaction:multisigStore.Update %v requestid:%v  contract:%v ", err.Error(), originTx.Multisig.RequestID, contract.ContractAddress)
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
			_, err = multisigStore.UpdateAll(sel, update)
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
				// internal transaction to multisig
				for _, multisig := range user.Multisigs {
					if multisig.ContractAddress == outputAddress {
						tx.Status = store.TxStatusAppearedInBlockIncoming
						txToUser.From = contract.ContractAddress
						txToUser.To = multisig.ContractAddress
						txToUser.PoolTime = time.Now().Unix()
						txToUser.Multisig.MethodInvoked = executeTransaction
						isOurUser = true
						txToUser.Amount = amount
						txToUser.Hash = tx.Hash
						txToUser.Status = store.TxStatusInBlockConfirmedIncoming
						txToUser.IsInternal = true
						txToUser.Multisig.Contract = multisig.ContractAddress
						txToUser.Multisig.MethodInvoked = store.SubmitTransaction
						break
					}
				}

				if isOurUser {
					_ = multisigStore.Insert(txToUser)
				}
			}
			if err != nil && err != mgo.ErrNotFound {
				// database error
				log.Errorf("ParseMultisigInput:confirmTransaction:multisigStore.Find %v", err.Error())
			}

			// notify on submit transaction
			for _, user := range users {
				msg := store.WsMessage{
					Type:    store.NotifyTxSubmitted,
					To:      user.UserID,
					Date:    time.Now().Unix(),
					Payload: "ok",
				}
				ethcli.WsServer.BroadcastToAll(store.MsgReceive+":"+user.UserID, msg)
			}
		}

		if tx.BlockTime != 0 && !tx.Multisig.InvocationStatus {
			tx.Status = store.TxStatusInBlockMethodInvocationFail
			sel := bson.M{"txhash": tx.Hash}
			update := bson.M{
				"$set": bson.M{
					"txstatus": store.TxStatusInBlockMethodInvocationFail,
				},
			}
			txStore.UpdateAll(sel, update)
		}

		return tx

	case revokeConfirmation: // "20ea8d86": "revokeConfirmation(uint256)"
		// TODO: send notfy to owners about -1 confirmation. store confirmations in db

		tx.Multisig.Owners = []store.OwnerHistory{}
		log.Debugf("revokeConfirmation: %v", tx.Multisig.Input)

		requestid, err := parseRevokeInput(tx.Multisig.Input)
		if err != nil {
			log.Errorf("ParseMultisigInput:revokeConfirmation: parseRevokeInput %v requestid:%v  contract:%v ", err.Error(), requestid, contract.ContractAddress)
		} else {
			sel := bson.M{"multisig.requestid": requestid, "multisig.contract": contract.ContractAddress}

			originTx := store.TransactionETH{}
			err = multisigStore.Find(sel).One(&originTx)
			if err != nil {
				log.Errorf("ParseMultisigInput:revokeConfirmation:multisigStore.Find %v requestid:%v  contract:%v ", err.Error(), requestid, contract.ContractAddress)
			}
			ownerHistorys := []store.OwnerHistory{}
			for _, ownerHistory := range tx.Multisig.Owners {
				if ownerHistory.Address == originTx.From {
					ownerHistorys = append(ownerHistorys, store.OwnerHistory{
						Address:            tx.From,
						ConfirmationStatus: store.MultisigOwnerStatusRevoked,
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

			_, err = multisigStore.UpdateAll(sel, update)
			if err != nil {
				log.Errorf("ParseMultisigInput:revokeConfirmation:multisigStore.Update %v requestid:%v  contract:%v ", err.Error(), requestid, contract.ContractAddress)
			}
		}

		if tx.BlockTime != 0 && !tx.Multisig.InvocationStatus {
			tx.Status = store.TxStatusInBlockMethodInvocationFail
			sel := bson.M{"txhash": tx.Hash}
			update := bson.M{
				"$set": bson.M{
					"txstatus": store.TxStatusInBlockMethodInvocationFail,
				},
			}
			txStore.UpdateAll(sel, update)
		}

		return tx
	case "0x": // incoming transaction
		if tx.BlockTime != 0 && !tx.Multisig.InvocationStatus {
			tx.Status = store.TxStatusInBlockMethodInvocationFail
			sel := bson.M{"txhash": tx.Hash}
			update := bson.M{
				"$set": bson.M{
					"txstatus": store.TxStatusInBlockMethodInvocationFail,
				},
			}
			txStore.UpdateAll(sel, update)
		}
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

func FetchUserAddresses(currencyID, networkID int, user store.User, addreses []string) ([]store.AddressExtended, error) {
	addresses := []store.AddressExtended{}
	fetched := map[string]store.AddressExtended{}

	for _, address := range addreses {
		fetched[address] = store.AddressExtended{
			Address:    address,
			Associated: false,
		}
	}

	for _, wallet := range user.Wallets {
		for _, addres := range wallet.Adresses {
			if wallet.CurrencyID == currencyID && wallet.NetworkID == networkID {
				for addr, fetchAddr := range fetched {
					if addr == addres.Address {
						fetchAddr.Associated = true
						fetchAddr.WalletIndex = wallet.WalletIndex
						fetchAddr.AddressIndex = addres.AddressIndex
						fetchAddr.UserID = user.UserID
						fetched[addres.Address] = fetchAddr
					}
				}
			}
		}
	}

	for _, addr := range fetched {
		addresses = append(addresses, addr)
	}

	return addresses, nil
}

func FetchContractOwners(currencyID, networkID int, contractaddress string) ([]store.OwnerHistory, error) {
	oh := []store.OwnerHistory{}

	sel := bson.M{"multisig.contractAddress": contractaddress}
	user := store.User{}
	_ = usersData.Find(sel).One(&user)

	for _, multisig := range user.Multisigs {
		for _, owner := range multisig.Owners {
			if multisig.CurrencyID == currencyID && multisig.NetworkID == networkID && multisig.ContractAddress == contractaddress {
				oh = append(oh, store.OwnerHistory{
					Address: owner.Address,
				})
			}
		}
	}
	return oh, nil
}

func fetchMethod(input string) string {
	method := input
	if len(input) < 10 {
		method = "0x"
	} else {
		method = input[:10]
	}

	return method
}

func fetchMultisig(users []store.User, contract string) (*store.Multisig, error) {
	if len(users) > 0 {
		for _, m := range users[0].Multisigs {
			if m.ContractAddress == contract {
				return &m, nil
			}
		}
	}

	return &store.Multisig{}, errors.New("fetchMultisig: contract have no multy users :" + contract)
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
	// 266 is minimal length of valid input for this kind of transactions
	if len(input) >= 266 {
		// crop method name from input data
		in := input[10:]
		re := regexp.MustCompile(`.{64}`) // Every 64 chars
		parts := re.FindAllString(in, -1) // Split the string into 64 chars blocks.

		// 4 is minimal count of parts for correct method invocation
		if len(parts) >= 4 {
			address = strings.ToLower("0x" + parts[0][24:])
			a, _ := new(big.Int).SetString(parts[1], 16)
			amount = a.String()
		}
	}

	return address, amount
}

func parseRevokeInput(input string) (int64, error) {
	in := input[10:]
	re := regexp.MustCompile(`.{64}`) // Every 64 chars
	parts := re.FindAllString(in, -1) // Split the string into 64 chars blocks.

	if len(parts) > 0 {
		i, ok := new(big.Int).SetString(input, 16)
		if !ok {
			return 0, fmt.Errorf("bad input %v", input)
		} else {
			return i.Int64(), nil
		}

	}

	return 0, fmt.Errorf("low len input %v", input)
}

func signatuteToStatus(signature string) int {
	switch signature {
	case submitTransaction: // "c6427474": "submitTransaction(address,uint256,bytes)"
		return store.NotifyPaymentReq
	case confirmTransaction: // "c01a8c84": "confirmTransaction(uint256)"
		return store.NotifyConfirmTx

	case revokeConfirmation: // "20ea8d86": "revokeConfirmation(uint256)"
		return store.NotifyRevokeTx

	case "0x": // incoming transaction
		return store.NotifyIncomingTx

	default:
		return store.NotifyPaymentReq
	}
}

func msToUserData(addresses []string, usersData *mgo.Collection) map[string]store.User {
	users := map[string]store.User{} // ms attached address to user
	for _, address := range addresses {
		user := store.User{}
		err := usersData.Find(bson.M{"wallets.addresses.address": strings.ToLower(address)}).One(&user)
		if err != nil {
			break
		}
		// attachedAddress = strings.ToLower(address)
		users[strings.ToLower(address)] = user
	}
	return users
}

// Fetch invite code from undeployed multisigs
func fetchInviteUndeployed(users map[string]store.User) string {
	invitecode := ""
	ownersCount := 0
	for _, msUser := range users {
		for _, ms := range msUser.Multisigs {
			for _, owner := range ms.Owners {
				for addres := range users {
					if addres == owner.Address {
						ownersCount++
						if ownersCount == ms.OwnersCount {
							invitecode = ms.InviteCode
							break
						}
					}
				}
			}
			ownersCount = 0
		}
	}
	return invitecode

}
