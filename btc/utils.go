/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"github.com/Appscrunch/Multy-back/store"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func newEmptyTx(userID string) store.TxRecord {
	return store.TxRecord{
		UserID:       userID,
		Transactions: []store.MultyTX{},
	}
}
func newAddresAmount(address string, amount int64) store.AddresAmount {
	return store.AddresAmount{
		Address: address,
		Amount:  amount,
	}
}

func newMultyTX(txID, txHash, txOutScript, txAddress string, txStatus, txOutID, walletindex int, txOutAmount, blockTime, blockHeight, fee, mempoolTime int64, stockexchangerate []store.ExchangeRatesRecord, inputs, outputs []store.AddresAmount) store.MultyTX {
	return store.MultyTX{
		TxID:              txID,
		TxHash:            txHash,
		TxOutScript:       txOutScript,
		TxAddress:         txAddress,
		TxStatus:          txStatus,
		TxOutAmount:       txOutAmount,
		TxOutID:           txOutID,
		WalletIndex:       walletindex,
		BlockTime:         blockTime,
		BlockHeight:       blockHeight,
		TxFee:             fee,
		MempoolTime:       mempoolTime,
		StockExchangeRate: stockexchangerate,
		TxInputs:          inputs,
		TxOutputs:         outputs,
	}
}

func rawTxByTxid(txid string) (*btcjson.TxRawResult, error) {
	hash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return nil, err
	}
	previousTxVerbose, err := rpcClient.GetRawTransactionVerbose(hash)
	if err != nil {
		return nil, err
	}
	return previousTxVerbose, nil
}

func fetchWalletIndex(wallets []store.Wallet, address string) int {
	var walletIndex int
	for _, wallet := range wallets {
		for _, addr := range wallet.Adresses {
			if addr.Address == address {
				walletIndex = wallet.WalletIndex
				break
			}
		}
	}
	return walletIndex
}

func txInfo(txVerbose *btcjson.TxRawResult) ([]store.AddresAmount, []store.AddresAmount, int64, error) {

	inputs := []store.AddresAmount{}
	outputs := []store.AddresAmount{}
	var inputSum float64
	var outputSum float64

	for _, out := range txVerbose.Vout {
		for _, address := range out.ScriptPubKey.Addresses {
			amount := int64(out.Value * 100000000)
			outputs = append(outputs, newAddresAmount(address, amount))
		}
		outputSum += out.Value
	}
	for _, input := range txVerbose.Vin {
		hash, err := chainhash.NewHashFromStr(input.Txid)
		if err != nil {
			log.Errorf("txInfo:chainhash.NewHashFromStr: %s", err.Error())
			return nil, nil, 0, err
		}
		previousTxVerbose, err := rpcClient.GetRawTransactionVerbose(hash)
		if err != nil {
			log.Errorf("txInfo:rpcClient.GetRawTransactionVerbose: %s", err.Error())
			return nil, nil, 0, err
		}

		for _, address := range previousTxVerbose.Vout[input.Vout].ScriptPubKey.Addresses {
			amount := int64(previousTxVerbose.Vout[input.Vout].Value * 100000000)
			inputs = append(inputs, newAddresAmount(address, amount))
		}
		inputSum += previousTxVerbose.Vout[input.Vout].Value
	}
	fee := int64((inputSum - outputSum) * 100000000)

	return inputs, outputs, fee, nil
}
