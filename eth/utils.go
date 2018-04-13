/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package eth

import (
	"math/big"
	"time"

	pb "github.com/Appscrunch/Multy-back/node-streamer/eth"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/onrik/ethrpc"
	"gopkg.in/mgo.v2/bson"
)

const ( // currency id  nsq
	TxStatusAppearedInMempoolIncoming = 1
	TxStatusAppearedInBlockIncoming   = 2

	TxStatusAppearedInMempoolOutcoming = 3
	TxStatusAppearedInBlockOutcoming   = 4

	TxStatusInBlockConfirmedIncoming  = 5
	TxStatusInBlockConfirmedOutcoming = 6

	WeiInEthereum = 1000000000000000000
)

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
	balance, err := client.Rpc.EthGetBalance(address, "pending")
	if err != nil {
		log.Errorf("GetAddressBalance:rpc.EthGetBalance: %s", err.Error())
		return balance, err
	}
	return balance, err
}

func (client *Client) GetAllTxPool() (map[string]interface{}, error) {
	return client.Rpc.TxPoolContent()
}

func (client *Client) GetBlockHeight() (int, error) {
	return client.Rpc.EthBlockNumber()
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

	if blockHeight == -1 {
		tx.TxpoolTime = time.Now().Unix()
	} else {
		tx.BlockTime = time.Now().Unix()
	}
	tx.BlockHeight = blockHeight
	// log.Infof("tx - %v", tx)

	/*
		Fetching tx status and send
	*/
	// from v1 to v1
	if fromUser.UserID == toUser.UserID && fromUser.UserID != "" {
		tx.UserID = fromUser.UserID
		tx.WalletIndex = int32(fromUser.WalletIndex)
		tx.AddressIndex = int32(fromUser.AddressIndex)

		tx.Status = TxStatusAppearedInBlockOutcoming
		if blockHeight == -1 {
			tx.Status = TxStatusAppearedInMempoolOutcoming
		}

		// send to multy-back
		client.TransactionsCh <- tx
	}

	// from v1 to v2 outgoing
	if fromUser.UserID != "" {
		tx.UserID = fromUser.UserID
		tx.WalletIndex = int32(fromUser.WalletIndex)
		tx.AddressIndex = int32(fromUser.AddressIndex)
		tx.Status = TxStatusAppearedInBlockOutcoming
		if blockHeight == -1 {
			tx.Status = TxStatusAppearedInMempoolOutcoming
		}

		// send to multy-back
		client.TransactionsCh <- tx
	}

	// from v1 to v2 incoming
	if toUser.UserID != "" {
		tx.UserID = toUser.UserID
		tx.WalletIndex = int32(toUser.WalletIndex)
		tx.AddressIndex = int32(toUser.AddressIndex)
		tx.Status = TxStatusAppearedInBlockIncoming
		if blockHeight == -1 {
			tx.Status = TxStatusAppearedInMempoolIncoming
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
		Amount:   int64((float64(rawTX.Value.Int64()) / 1000000000000000000)), //TODO: fix to normal amount
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
