/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package eth

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	pb "github.com/Multy-io/Multy-ETH-node-service/node-streamer"
	"github.com/Multy-io/Multy-back/store"
	"github.com/onrik/ethrpc"
	"gopkg.in/mgo.v2/bson"
)

const (
	MultiSigFactory    = "0xf8f73808"
	submitTransaction  = "0xc6427474"
	confirmTransaction = "0xc01a8c84"
	revokeConfirmation = "0x20ea8d86"
	executeTransaction = "0xee22610b"
)

type FactoryInfo struct {
	Confirmations  int64
	FactoryAddress string
	TxOfCreation   string
	Contract       string
	Addresses      []string
	IsFailed       bool
}

type Multisig struct {
	FactoryAddress string
	UsersContracts sync.Map // concrete multysig contract as a key. FactoryAddress as value
}

func newETHtx(hash, from, to string, amount float64, gas, gasprice, nonce int) store.TransactionETH {
	return store.TransactionETH{}
}

func (client *Client) SendRawTransaction(rawTX string) (string, error) {
	hash, err := client.Rpc.EthSendRawTransaction(rawTX)
	if err != nil {
		log.Errorf("SendRawTransaction:rpc.EthSendRawTransaction: %s", err.Error())
		return hash, err
	}
	return hash, err
}

func (client *Client) GetAddressBalance(address string) (big.Int, error) {
	balance, err := client.Rpc.EthGetBalance(address, "latest")
	if err != nil {
		log.Errorf("GetAddressBalance:rpc.EthGetBalance: %s", err.Error())
		return balance, err
	}
	return balance, err
}

func (client *Client) GetTxByHash(hash string) (bool, error) {
	tx, err := client.Rpc.EthGetTransactionByHash(hash)
	if tx == nil {
		return false, err
	} else {
		return true, err
	}
}

func (client *Client) GetGasPrice() (big.Int, error) {
	gas, err := client.Rpc.EthGasPrice()
	if err != nil {
		log.Errorf("GetGasPrice:rpc.EthGetBalance: %s", err.Error())
		return gas, err
	}
	return gas, err
}

func (client *Client) GetAddressPendingBalance(address string) (big.Int, error) {
	balance, err := client.Rpc.EthGetBalance(address, "pending")
	if err != nil {
		log.Errorf("GetAddressPendingBalance:rpc.EthGetBalance: %s", err.Error())
		return balance, err
	}
	log.Errorf("GetAddressPendingBalance %v", balance.String())
	return balance, err
}

func (client *Client) GetAllTxPool() (map[string]interface{}, error) {
	return client.Rpc.TxPoolContent()
}

func (client *Client) GetBlockHeight() (int, error) {
	return client.Rpc.EthBlockNumber()
}

func (client *Client) GetCode(address string) (string, error) {
	return client.Rpc.EthGetCode(address, "latest")
}

func (client *Client) GetAddressNonce(address string) (int, error) {
	return client.Rpc.EthGetTransactionCount(address, "latest")
}

func (client *Client) ResyncAddress(txid string) error {
	tx, err := client.Rpc.EthGetTransactionByHash(txid)
	if err != nil {
		return err
	}
	client.parseETHTransaction(*tx, int64(*tx.BlockNumber), true)
	return nil
}

func (client *Client) ResyncMultisig(txid string) error {
	tx, err := client.Rpc.EthGetTransactionByHash(txid)
	if err != nil {
		return err
	}
	client.parseETHMultisig(*tx, int64(*tx.BlockNumber), true)
	return nil
}

func (client *Client) parseETHTransaction(rawTX ethrpc.Transaction, blockHeight int64, isResync bool) {
	var fromUser store.AddressExtended
	var toUser store.AddressExtended

	if udFrom, ok := client.UsersData.Load(rawTX.From); ok {
		fromUser = udFrom.(store.AddressExtended)
	}

	if udTo, ok := client.UsersData.Load(rawTX.To); ok {
		toUser = udTo.(store.AddressExtended)
	}

	if fromUser.UserID == toUser.UserID && fromUser.UserID == "" {
		// not our users tx
		return
	}

	tx := rawToGenerated(rawTX)
	tx.Resync = isResync

	block, err := client.Rpc.EthGetBlockByHash(rawTX.BlockHash, false)
	if err != nil {
		if blockHeight == -1 {
			tx.TxpoolTime = time.Now().Unix()
		} else {
			tx.BlockTime = time.Now().Unix()
		}
		tx.BlockHeight = blockHeight
	} else {
		tx.BlockTime = int64(block.Timestamp)
		tx.BlockHeight = int64(block.Number)
	}

	if blockHeight == -1 {
		tx.TxpoolTime = time.Now().Unix()
	}

	// log.Infof("tx - %v", tx)

	/*
		Fetching tx status and send
	*/
	// from v1 to v1
	if fromUser.UserID == toUser.UserID && fromUser.UserID != "" {
		tx.UserID = fromUser.UserID
		tx.WalletIndex = int32(fromUser.WalletIndex)
		tx.AddressIndex = int32(fromUser.AddressIndex)

		tx.Status = store.TxStatusAppearedInBlockOutcoming
		if blockHeight == -1 {
			tx.Status = store.TxStatusAppearedInMempoolOutcoming
		}

		// send to multy-back
		client.TransactionsCh <- tx
	}

	// from v1 to v2 outgoing
	if fromUser.UserID != "" {
		tx.UserID = fromUser.UserID
		tx.WalletIndex = int32(fromUser.WalletIndex)
		tx.AddressIndex = int32(fromUser.AddressIndex)
		tx.Status = store.TxStatusAppearedInBlockOutcoming
		if blockHeight == -1 {
			tx.Status = store.TxStatusAppearedInMempoolOutcoming
		}

		// send to multy-back
		client.TransactionsCh <- tx
	}

	// from v1 to v2 incoming
	if toUser.UserID != "" {
		tx.UserID = toUser.UserID
		tx.WalletIndex = int32(toUser.WalletIndex)
		tx.AddressIndex = int32(toUser.AddressIndex)
		tx.Status = store.TxStatusAppearedInBlockIncoming
		if blockHeight == -1 {
			tx.Status = store.TxStatusAppearedInMempoolIncoming
		}

		// send to multy-back
		client.TransactionsCh <- tx
	}

}

func (client *Client) parseETHMultisig(rawTX ethrpc.Transaction, blockHeight int64, isResync bool) {
	var fromUser string
	var toUser string

	ud := client.Multisig.UsersContracts

	if _, ok := ud.Load(rawTX.From); ok {
		fromUser = rawTX.From
	}

	if _, ok := ud.Load(rawTX.To); ok {
		toUser = rawTX.To
	}

	if fromUser == toUser && fromUser == "" {
		// not our users tx
		return
	}

	input := rawTX.Input

	if len(rawTX.Input) < 10 {
		input = "0x"
	} else {
		input = input[:10]
	}

	fmt.Println("client.Multisig.UsersContracts ", ud)
	fmt.Println("input", input)

	switch input {
	case submitTransaction: // "c6427474": "submitTransaction(address,uint256,bytes)"
		// TODO: feth contract owners, send notfy to owners about transation. status: waiting for confirmations
		// find in db if one one confirmation needed DONE internal transaction
		log.Debugf("submitTransaction: %v", rawTX.Input)
	case confirmTransaction: // "c01a8c84": "confirmTransaction(uint256)"
		// TODO: send notfy to owners about +1 confirmation. store confiramtions id db
		log.Debugf("confirmTransaction: %v", rawTX.Input)
	case revokeConfirmation: // "20ea8d86": "revokeConfirmation(uint256)"
		// TODO: send notfy to owners about -1 confirmation. store confirmations in db
		log.Debugf("revokeConfirmation: %v", rawTX.Input)
	case executeTransaction: // "ee22610b": "executeTransaction(uint256)"
		// TODO: feth contract owners, send notfy to owners about transation. status: conformed transatcion
		log.Debugf("executeTransaction: %v", rawTX.Input)
	case "0x": // incoming transaction
		// TODO: notify owners about new transation
		log.Debugf("incoming transaction: %v", rawTX.Input)
	default:
		log.Debugf("wrong method:  %v", rawTX.Input)
		// wrong method
	}

	tx := rawToGenerated(rawTX)
	tx.Resync = isResync

	block, err := client.Rpc.EthGetBlockByHash(rawTX.BlockHash, false)
	if err != nil {
		if blockHeight == -1 {
			tx.TxpoolTime = time.Now().Unix()
		} else {
			tx.BlockTime = time.Now().Unix()
		}
		tx.BlockHeight = blockHeight
	} else {
		tx.BlockTime = int64(block.Timestamp)
		tx.BlockHeight = int64(block.Number)
	}

	if blockHeight == -1 {
		tx.TxpoolTime = time.Now().Unix()
	}

	if blockHeight != -1 {
		log.Debugf("GetInvocationStatus")
		invocationStatus, returnValue, err := client.GetInvocationStatus(rawTX.Hash)
		if err != nil {
			log.Errorf("GetInvocationStatus: %v", err.Error())
			return
		}
		tx.InvocationStatus = invocationStatus
		tx.Return = returnValue
	}

	log.Debugf(`GetInvocationStatus invocationStatus:  %v  returnValue  "%v" `, tx.InvocationStatus, tx.Return)
	/*
		Fetching tx status and send
	*/
	tx.Multisig = true

	if fromUser != "" {
		// outgoing tx
		tx.Status = store.TxStatusAppearedInBlockOutcoming
		if blockHeight == -1 {
			tx.Status = store.TxStatusAppearedInMempoolOutcoming
		}
		client.TransactionsCh <- tx
	}

	if toUser != "" {
		// incoming tx
		tx.Status = store.TxStatusAppearedInBlockIncoming
		if blockHeight == -1 {
			tx.Status = store.TxStatusAppearedInMempoolIncoming
		}
		client.TransactionsCh <- tx
	}

	if toUser == fromUser {
		// self tx
		tx.Status = store.TxStatusAppearedInBlockOutcoming
		if blockHeight == -1 {
			tx.Status = store.TxStatusAppearedInMempoolOutcoming
		}
		client.TransactionsCh <- tx
	}

}

func rawToGenerated(rawTX ethrpc.Transaction) pb.ETHTransaction {
	return pb.ETHTransaction{
		Hash:     rawTX.Hash,
		From:     rawTX.From,
		To:       rawTX.To,
		Amount:   rawTX.Value.String(),
		GasPrice: int64(rawTX.GasPrice.Int64()),
		GasLimit: int64(rawTX.Gas),
		Nonce:    int32(rawTX.Nonce),
		Input:    rawTX.Input,
	}
}

func isMempoolUpdate(mempool bool, status int) bson.M {
	if mempool {
		return bson.M{
			"$set": bson.M{
				"status": status,
			},
		}
	}
	return bson.M{
		"$set": bson.M{
			"status":    status,
			"blocktime": time.Now().Unix(),
		},
	}
}
