package btc

import (
	"fmt"

	"gopkg.in/mgo.v2"

	"github.com/Appscrunch/Multy-back/currencies"
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

func SetWsHandlers(cli *gosocketio.Client, networkID int) {

	// spendable outputs
	cli.On("newSpout", func(c *gosocketio.Channel, spOut store.SpendableOutputs) {
		log.Infof("Add spendable output%s", spOut.Address)

		user := store.User{}
		sel := bson.M{"wallets.addresses.address": spOut.Address}
		err := usersData.Find(sel).One(&user)
		if err != nil && err != mgo.ErrNotFound {
			log.Errorf("SetWsHandlers: cli.On newIncomingTx: %s", err)
			return
		}
		for _, wallet := range user.Wallets {
			for _, address := range wallet.Adresses {
				if address.Address == spOut.Address {
					spOut.AddressIndex = address.AddressIndex
					spOut.WalletIndex = wallet.WalletIndex
				}
			}
		}

		switch networkID {
		case currencies.Main:
			err = spendableOutputs.Insert(spOut)
			if err != nil {
				log.Errorf("SetWsHandlers: spendableOutputs.Insert: %s", err)
			}
		case currencies.Test:
			err = spendableOutputsTest.Insert(spOut)
			if err != nil {
				log.Errorf("SetWsHandlers: spendableOutputs.Insert: %s", err)
			}
		}

	})

	cli.On("deleteSpout", func(c *gosocketio.Channel, delSpOut store.DeleteSpendableOutput) {
		log.Infof("Delete spendable output%s", delSpOut.Address)
		spOutsCollection := &mgo.Collection{}

		switch networkID {
		case currencies.Main:
			spOutsCollection = spendableOutputs
		case currencies.Test:
			spOutsCollection = spendableOutputsTest
		}

		sel := bson.M{"txid": delSpOut.TxID, "userid": delSpOut.UserID, "address": delSpOut.Address}
		err := spOutsCollection.Find(sel).One(nil)
		if err == mgo.ErrNotFound {
			log.Errorf("SetWsHandlers: cli.On deleteSpout: spOutsCollection.Find: %s", err)
		} else {
			err := spOutsCollection.Remove(sel)
			if err != nil {
				log.Errorf("SetWsHandlers: cli.On deleteSpout: spOutsCollection.Remove: %s", err)
			}
		}

	})

	// Tx history
	cli.On("newOutcomingTx", func(c *gosocketio.Channel, outTx store.MultyTX) {
		log.Infof("New outcoming transaction %s", outTx.TxID)

		user := store.User{}
		setExchangeRates(&outTx, true, outTx.MempoolTime)
		for _, in := range outTx.WalletsInput {
			sel := bson.M{"wallets.addresses.address": in.Address.Address}
			err := usersData.Find(sel).One(&user)
			if err == mgo.ErrNotFound {
				continue
			} else if err != nil && err != mgo.ErrNotFound {
				log.Errorf("SetWsHandlers: cli.On newIncomingTx: %s", err)
			}

			for _, wallet := range user.Wallets {
				for i := 0; i < len(wallet.Adresses); i++ {
					if wallet.Adresses[i].Address == in.Address.Address {
						in.WalletIndex = wallet.WalletIndex
						in.Address.AddressIndex = wallet.Adresses[i].AddressIndex
					}
				}
			}
		}

		for _, out := range outTx.WalletsOutput {
			sel := bson.M{"wallets.addresses.address": out.Address.Address}
			err := usersData.Find(sel).One(&user)
			if err == mgo.ErrNotFound {
				continue
			} else if err != nil && err != mgo.ErrNotFound {
				log.Errorf("SetWsHandlers: cli.On newIncomingTx: %s", err)
			}

			for _, wallet := range user.Wallets {
				for i := 0; i < len(wallet.Adresses); i++ {
					if wallet.Adresses[i].Address == out.Address.Address {
						out.WalletIndex = wallet.WalletIndex
						out.Address.AddressIndex = wallet.Adresses[i].AddressIndex
					}
				}
			}
		}

		switch networkID {
		case currencies.Main:
			err := txsData.Insert(outTx)
			if err != nil {
				log.Errorf("SetWsHandlers: txsData.Insert: %s", err)
			}
		case currencies.Test:
			err := txsDataTest.Insert(outTx)
			if err != nil {
				log.Errorf("SetWsHandlers: txsData.Insert: %s", err)
			}
		}

	})

	cli.On("newIncomingTx", func(c *gosocketio.Channel, inTx store.MultyTX) {
		log.Infof("New incoming transaction %s", inTx.TxID)
		// TODO: handle tx history in
		user := store.User{}
		setExchangeRates(&inTx, true, inTx.MempoolTime)
		for _, in := range inTx.WalletsInput {
			sel := bson.M{"wallets.addresses.address": in.Address.Address}
			err := usersData.Find(sel).One(&user)
			if err == mgo.ErrNotFound {
				continue
			} else if err != nil && err != mgo.ErrNotFound {
				log.Errorf("SetWsHandlers: cli.On newIncomingTx: %s", err)
			}

			for _, wallet := range user.Wallets {
				for i := 0; i < len(wallet.Adresses); i++ {
					if wallet.Adresses[i].Address == in.Address.Address {
						in.WalletIndex = wallet.WalletIndex
						in.Address.AddressIndex = wallet.Adresses[i].AddressIndex
					}
				}
			}
		}

		for _, out := range inTx.WalletsOutput {
			sel := bson.M{"wallets.addresses.address": out.Address.Address}
			err := usersData.Find(sel).One(&user)
			if err == mgo.ErrNotFound {
				continue
			} else if err != nil && err != mgo.ErrNotFound {
				log.Errorf("SetWsHandlers: cli.On newIncomingTx: %s", err)
			}

			for _, wallet := range user.Wallets {
				for i := 0; i < len(wallet.Adresses); i++ {
					if wallet.Adresses[i].Address == out.Address.Address {
						out.WalletIndex = wallet.WalletIndex
						out.Address.AddressIndex = wallet.Adresses[i].AddressIndex
					}
				}
			}
		}

		switch networkID {
		case currencies.Main:
			err := txsData.Insert(inTx)
			if err != nil {
				log.Errorf("SetWsHandlers: txsData.Insert: %s", err)
			}
		case currencies.Test:
			err := txsDataTest.Insert(inTx)
			if err != nil {
				log.Errorf("SetWsHandlers: txsData.Insert: %s", err)
			}
		}

		// fmt.Println(
		// 	inTx.BlockHeight, "BlockHeight\n",
		// 	inTx.BlockTime, "BlockTime\n",
		// 	inTx.Confirmations, "Confirmations\n",
		// 	inTx.MempoolTime, "MempoolTime\n",
		// 	inTx.StockExchangeRate, "StockExchangeRate\n",
		// 	inTx.TxAddress, "TxAddress\n",
		// 	inTx.TxFee, "TxFee\n",
		// 	inTx.TxHash, "TxHash\n",
		// 	inTx.TxID, "TxID\n",
		// 	inTx.TxInputs, "TxInputs\n",
		// 	inTx.TxOutputs, "TxOutputs\n",
		// 	inTx.TxOutAmount, "TxOutAmount\n",
		// 	inTx.TxOutScript, "TxOutScript\n",
		// 	inTx.TxStatus, "TxStatus\n",
		// 	inTx.UserId, "UserId\n",
		// 	inTx.WalletsInput, "WalletsInput\n",
		// 	inTx.WalletsOutput, "WalletsOutput\n",
		// )
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
