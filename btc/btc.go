package btc

import (
	"time"

	"github.com/btcsuite/btcd/rpcclient"
)

const (
	txInMempool  = "incoming from mempool"
	txOutMempool = "outcoming from mempool"
	txInBlock    = "incoming from block"
	txOutBlock   = "outcoming from block"
)

// Dirty hack - this will be wrapped to a struct
var (
	rpcClient  = &rpcclient.Client{}
	chToClient chan BtcTransactionWithUserID // a channel for sending data to client
)

func simulateSendNewTransactions() {
	for {
		time.Sleep(time.Second * 2)
		b := BtcTransactionWithUserID{
			NotificationMsg: &BtcTransaction{
				Amount: 5,
			},
			UserID: "555",
		}

		chToClient <- b
	}
}

func InitHandlers() (*rpcclient.Client, chan BtcTransactionWithUserID, error) {
	chToClient = make(chan BtcTransactionWithUserID, 0)
	// go simulateSendNewTransactions()

	go RunProcess()
	return rpcClient, chToClient, nil
}

type BtcTransaction struct {
	TransactionType string  `json:"transactionType"`
	Amount          float64 `json:"amount"`
	TxID            string  `json:"txid"`
	Address         string  `json:"address"`
}

type BtcTransactionWithUserID struct {
	NotificationMsg *BtcTransaction
	UserID          string
}
