package btc

import (
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
	mgo "gopkg.in/mgo.v2"

	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"log"
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

var usersData *mgo.Collection

var mempoolRates *mgo.Collection

var Cert = `testcert`

var connCfg = &rpcclient.ConnConfig{
	Host:         "localhost:18334",
	User:         "multy",
	Pass:         "multy",
	Endpoint:     "ws",
	Certificates: []byte(Cert),

	HTTPPostMode: false, // Bitcoin core only supports HTTP POST mode
	DisableTLS:   false, // Bitcoin core does not provide TLS by default

}

func RunProcess() error {
	fmt.Println("[DEBUG] RunProcess()")

	db, err := mgo.Dial("localhost:27017")

	if err != nil {
		log.Printf("[ERR] RunProcess: Cand connect to DB: %s\n", err.Error())
		return err
	}

	usersData = db.DB("userDB").C("userCollection") // all db tables
	mempoolRates = db.DB("BTCMempool").C("Rates")

	// Drop collection on every new start of application
	err = mempoolRates.DropCollection()
	if err != nil {
		log.Printf("[ERR] RunProcess:mempoolRates.DropCollection:%s \n", err.Error())
	}

	ntfnHandlers := rpcclient.NotificationHandlers{
		OnBlockConnected: func(hash *chainhash.Hash, height int32, t time.Time) {
			log.Printf("[DEBUG] OnBlockConnected: %v (%d) %v", hash, height, t)
			go parseNewBlock(hash)

		},
		OnTxAcceptedVerbose: func(txDetails *btcjson.TxRawResult) {
			log.Printf("[DEBUG] OnTxAcceptedVerbose: new transaction id = %v", txDetails.Txid)
			// notify on new in
			// notify on new out
			go parseMempoolTransaction(txDetails)
			//add every new tx from mempool to db
			//feeRate
			go newTxToDB(txDetails.Hash)
		},
	}

	rpcClient, err = rpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		log.Printf("[ERR] RunProcess(): rpcclient.New %s\n", err.Error())
		return err
	}

	// Register for block connect and disconnect notifications.
	if err = rpcClient.NotifyBlocks(); err != nil {
		return err
	}
	log.Println("NotifyBlocks: Registration Complete")

	// Register for new transaction in mempool notifications.
	if err = rpcClient.NotifyNewTransactions(true); err != nil {
		return err
	}
	log.Println("NotifyNewTransactions: Registration Complete")

	// get all mempool and append to db
	go getAllMempool()

	rpcClient.WaitForShutdown()
	return nil
}
