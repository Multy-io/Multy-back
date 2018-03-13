package btc

import (
	"fmt"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/graarh/golang-socketio"
	"gopkg.in/mgo.v2/bson"
)

const (
	EventAddNewAddress        = "newAddress"
	EventResyncAddress        = "resync"
	EventSendRawTx            = "sendRaw"
	EventGetAllMempool        = "getAllMempool"
	EventMempool              = "mempool"
	EventDeleteMempoolOnBlock = "deleteMempool"
	Room                      = "node"
)

func SetWsHandlers(cli *gosocketio.Client) {


	cli.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		fmt.Printf("\n\n\n\n Ws Connected to Service Node\n\n\n\n\n")
	})

	cli.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		fmt.Printf("\n\n\n\n Ws Disconnected from Service Node\n\n\n\n\n")
	})
	
	cli.On("newSpout", func(c *gosocketio.Channel, spOut store.SpendableOutputs) {
		// TODO: handle new sp outs
		fmt.Println(spOut, " \nspouts\n")
	})

	cli.On("deleteSpout", func(c *gosocketio.Channel, delSpOut store.DeleteSpendableOutput) {
		// TODO: handle delete sp outs
		fmt.Println(delSpOut, " \ndeleteSpout\n")
	})

	// Tx history
	cli.On("newOutcomingTx", func(c *gosocketio.Channel, outTx store.MultyTX) {
		// TODO: handle tx history out
		fmt.Println(outTx, " \nnewOutcomingTx\n")
	})

	cli.On("newIncomingTx", func(c *gosocketio.Channel, inTx store.MultyTX) {
		// TODO: handle tx history in
		fmt.Println(inTx, " \nnewIncomingTx\n")
	})

	// Add tx and feerate to mempool
	cli.On(EventMempool, func(c *gosocketio.Channel, recs []store.MempoolRecord) {
		fmt.Println(recs)
		InsertMempoolRecords(recs...)
	})

	// Mempool delete on block
	cli.On(EventDeleteMempoolOnBlock, func(c *gosocketio.Channel, hash string) {
		query := bson.M{"hashtx": hash}
		err := mempoolRates.Remove(query)
		if err != nil {
			log.Errorf("parseNewBlock:mempoolRates.Remove: %s", err.Error())
		} else {
			log.Debugf("Tx removed: %s", hash)
		}
	})
}
