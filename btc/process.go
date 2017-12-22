package btc

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
	mgo "gopkg.in/mgo.v2"

	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type MultyMempoolTx struct {
	hash    string
	inputs  []MultyAddress
	outputs []MultyAddress
	amount  float64
	fee     float64
	size    int32
	feeRate int32
	txid    string
}

type MultyAddress struct {
	address []string
	amount  float64
}

var memPool []MultyMempoolTx

type rpcClientWrapper struct {
	*rpcclient.Client
}

var (
	usersData    *mgo.Collection
	mempoolRates *mgo.Collection
	txsData      *mgo.Collection
)

var Cert = `testcert`

var connCfg = &rpcclient.ConnConfig{
	Host:     "localhost:18334",
	User:     "multy",
	Pass:     "multy",
	Endpoint: "ws",
	//Certificates: []byte(Cert), // add it in InitHandlers function

	HTTPPostMode: false, // Bitcoin core only supports HTTP POST mode
	DisableTLS:   false, // Bitcoin core does not provide TLS by default

}

func RunProcess() error {
	log.Info("Run Process")

	db, err := mgo.Dial("localhost:27017")

	if err != nil {
		log.Errorf("RunProcess: Cand connect to DB: %s", err.Error())
		return err
	}

	usersData = db.DB("userDB").C("userCollection") // all db tables
	mempoolRates = db.DB("BTCMempool").C("Rates")
	txsData = db.DB("Tx").C("BTC")

	// Drop collection on every new start of application
	err = mempoolRates.DropCollection()
	if err != nil {
		log.Errorf("RunProcess:mempoolRates.DropCollection:%s", err.Error())
	}

	ntfnHandlers := rpcclient.NotificationHandlers{
		OnBlockConnected: func(hash *chainhash.Hash, height int32, t time.Time) {
			log.Debugf("OnBlockConnected: %v (%d) %v", hash, height, t)
			go notifyNewBlockTx(hash)
			go blockTransactions(hash)
			go blockConfirmations(hash)
		},
		OnTxAcceptedVerbose: func(txDetails *btcjson.TxRawResult) {
			log.Debugf("OnTxAcceptedVerbose: new transaction id = %v", txDetails.Txid)
			// notify on new in
			// notify on new out
			go parseMempoolTransaction(txDetails)
			//add every new tx from mempool to db
			//feeRate
			go newTxToDB(txDetails)

			go mempoolTransaction(txDetails)

		},
	}

	rpcClient, err = rpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		log.Errorf("RunProcess(): rpcclient.New %s\n", err.Error())
		return err
	}

	// Register for block connect and disconnect notifications.
	if err = rpcClient.NotifyBlocks(); err != nil {
		return err
	}
	log.Info("NotifyBlocks: Registration Complete")

	// Register for new transaction in mempool notifications.
	if err = rpcClient.NotifyNewTransactions(true); err != nil {
		return err
	}
	log.Info("NotifyNewTransactions: Registration Complete")

	// get all mempool and append to db
	go getAllMempool()

	rpcClient.WaitForShutdown()
	return nil
}
