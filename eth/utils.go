/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package eth

import (
	"errors"
	"math/big"
	"regexp"
	"sync"
	"time"

	pb "github.com/Multy-io/Multy-back/node-streamer/eth"
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
	Addresses     []string
	Confirmations int64
	Contract      string
}

type Multisig struct {
	FactoryAddress string
	UsersContracts map[string][]Owner // concrete multysig contract as a string
	m              sync.Mutex
}

type Owner struct {
	UserID  string `json:"userid"`
	Address string `json:"address"`
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

func (client *Client) parseETHTransaction(rawTX ethrpc.Transaction, blockHeight int64, isResync bool) {
	var fromUser store.AddressExtended
	var toUser store.AddressExtended

	client.UserDataM.Lock()
	ud := *client.UsersData
	client.UserDataM.Unlock()

	if udFrom, ok := ud[rawTX.From]; ok {
		fromUser = udFrom
	}

	if udTo, ok := ud[rawTX.To]; ok {
		toUser = udTo
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

func rawToGenerated(rawTX ethrpc.Transaction) pb.ETHTransaction {
	return pb.ETHTransaction{
		Hash:     rawTX.Hash,
		From:     rawTX.From,
		To:       rawTX.To,
		Amount:   rawTX.Value.String(),
		GasPrice: int64(rawTX.GasPrice.Int64()),
		GasLimit: int64(rawTX.Gas),
		Nonce:    int32(rawTX.Nonce),
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

func parseFactoryInput(in string) (FactoryInfo, error) {
	// fetch method id by hash
	fi := FactoryInfo{}
	if in[:10] == MultiSigFactory {
		in := in[10:]

		c := in[64:128]
		confirmations, _ := new(big.Int).SetString(c, 10)
		fi.Confirmations = confirmations.Int64()

		in = in[192:]

		contractAddresses := []string{}
		re := regexp.MustCompile(`.{64}`) // Every 64 chars
		parts := re.FindAllString(in, -1) // Split the string into 64 chars blocks.

		for _, address := range parts {
			contractAddresses = append(contractAddresses, "0x"+address[24:])
		}
		fi.Addresses = contractAddresses

		return fi, nil
	}

	return fi, errors.New("Wrong method name")
}
