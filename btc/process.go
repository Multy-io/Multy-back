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

var Cert = `-----BEGIN CERTIFICATE-----
MIICPDCCAZ2gAwIBAgIQf8XOycg2EQ8wHpXsZJSy7jAKBggqhkjOPQQDBDAjMREw
DwYDVQQKEwhnZW5jZXJ0czEOMAwGA1UEAxMFYW50b24wHhcNMTcxMTI2MTY1ODQ0
WhcNMjcxMTI1MTY1ODQ0WjAjMREwDwYDVQQKEwhnZW5jZXJ0czEOMAwGA1UEAxMF
YW50b24wgZswEAYHKoZIzj0CAQYFK4EEACMDgYYABAGuHzCFKsJwlFwmtx5QMT/r
YJ/ap9E2QlUsCnMUCn1ho0wLJkpIgNQWs1zcaKTMGZNpwwLemCHke9sX06h/MdAG
CwGf1CY5kafyl7dTTlmD10sBA7UD1RXDjYnmYQhB1Z1MUNXKWXe4jCv7DnWmFEnc
+s5N1NXJx1PNzx/EcsCkRJcMraNwMG4wDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB
/wQFMAMBAf8wSwYDVR0RBEQwQoIFYW50b26CCWxvY2FsaG9zdIcEfwAAAYcQAAAA
AAAAAAAAAAAAAAAAAYcEwKgAeYcQ/oAAAAAAAAByhcL//jB99jAKBggqhkjOPQQD
BAOBjAAwgYgCQgCfs9tYHA1nvU5HSdNeHSZCR1WziHYuZHmGE7eqAWQjypnVbFi4
pccvzDFvESf8DG4FVymK4E2T/RFnD9qUDiMzPQJCATkCMzSKcyYlsL7t1ZgQLwAK
UpQl3TYp8uTf+UWzBz0uoEbB4CFeE2G5ZzrVK4XWZK615sfVFSorxHOOZaLwZEEL
-----END CERTIFICATE-----`

var connCfg = &rpcclient.ConnConfig{
	Host:         "192.168.0.121:18334",
	User:         "multy",
	Pass:         "multy",
	Endpoint:     "ws",
	Certificates: []byte(Cert),
}

func RunProcess() error {
	fmt.Println("[DEBUG] RunProcess()")

	db, err := mgo.Dial("192.168.0.121:27017")
	fmt.Println(err)

	usersData = db.DB("userDB").C("userCollection")

	ntfnHandlers := rpcclient.NotificationHandlers{
		OnBlockConnected: func(hash *chainhash.Hash, height int32, t time.Time) {
			log.Printf("[DEBUG] OnBlockConnected: %v (%d) %v", hash, height, t)
			go getAndParseNewBlock(hash)
		},
		OnTxAcceptedVerbose: func(txDetails *btcjson.TxRawResult) {
			log.Printf("[DEBUG] OnTxAcceptedVerbose: new transaction id = %v", txDetails.Txid)
			// notify on new in
			// notify on new out
			parseMempoolTransaction(txDetails)
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

	rpcClient.WaitForShutdown()
	return nil
}
