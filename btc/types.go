/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

func CreateBtcTransactionWithUserID(addr, userId, txType, txId string, amount float64) BtcTransactionWithUserID {
	return BtcTransactionWithUserID{
		UserID: userId,
		NotificationMsg: &BtcTransaction{
			TransactionType: txType,
			Amount:          amount,
			TxID:            txId,
			Address:         addr,
		},
	}
}

func newRecord(category int, hashTX string) Record {
	return Record{
		Category: category,
		HashTX:   hashTX,
	}
}

type Record struct {
	Category int    `json:"category"`
	HashTX   string `json:"hashTX"`
}

func newTxInfo(txType, txHash, address string, amount float64) TxInfo {
	return TxInfo{
		Type:    txType,
		TxHash:  txHash,
		Address: address,
		Amount:  amount,
		// timestamp
	}
}

type TxInfo struct {
	Type    string  `json:"type"`
	TxHash  string  `json:"txhash"`
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}
