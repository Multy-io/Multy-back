/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package ethereum

import (
	"fmt"
	"math/big"

	"github.com/onrik/ethrpc"
)

func newETHtx(hash, from, to string, amount float64, gas, gasprice, nonce int) MultyETHTransaction {
	return MultyETHTransaction{
		Hash:     hash,
		From:     from,
		To:       to,
		Amount:   amount,
		Gas:      gas,
		GasPrice: gasprice,
		Nonce:    nonce,
	}
}

func parseRawTransaction(rawTX ethrpc.Transaction, pending bool) MultyETHTransaction {
	tx := newETHtx(rawTX.Hash, rawTX.From, rawTX.To, (float64(rawTX.Value.Int64()) / 1000000000000000000), rawTX.Gas, int(rawTX.GasPrice.Int64()), rawTX.Nonce)
	str := "from block"
	if pending {
		str = "from txpool"
	}
	fmt.Println(tx, str, "\n")
	return tx
}

func (client *Client) SendRawTransaction(rawTX string) (string, error) {
	hash, err := client.rpc.EthSendRawTransaction(rawTX)
	if err != nil {
		client.log.Errorf("SendRawTransaction:rpc.EthSendRawTransaction: %s", err.Error())
		return hash, err
	}
	return hash, err
}

func (client *Client) GetAddressBalance(address string) (big.Int, error) {
	balance, err := client.rpc.EthGetBalance(address, "pending")
	if err != nil {
		client.log.Errorf("GetAddressBalance:rpc.EthGetBalance: %s", err.Error())
		return balance, err
	}
	return balance, err
}
