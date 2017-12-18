package btc

import (
	"fmt"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/nsqio/go-nsq"
)

const (
	txInMempool  = "incoming from mempool"
	txOutMempool = "outcoming from mempool"
	txInBlock    = "incoming from block"
	txOutBlock   = "outcoming from block"

	// TopicTransaction is a topic for sending notifies to clients
	TopicTransaction = "btcTransactionUpdate"
)

// Dirty hack - this will be wrapped to a struct
var (
	rpcClient = &rpcclient.Client{}

	nsqProducer *nsq.Producer // a producer for sending data to clients
	rpcConf     *rpcclient.ConnConfig
)

func InitHandlers() (*rpcclient.Client, error) {
	config := nsq.NewConfig()
	p, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		return nil, fmt.Errorf("nsq producer: %s", err.Error())
	}
	nsqProducer = p

	go RunProcess()
	return rpcClient, nil
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
