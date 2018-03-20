/*
Copyright 2019 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"

	"gopkg.in/mgo.v2"

	"github.com/Appscrunch/Multy-back/currencies"
	pb "github.com/Appscrunch/Multy-back/node-streamer"
	"github.com/Appscrunch/Multy-back/store"
	nsq "github.com/bitly/go-nsq"
	"github.com/graarh/golang-socketio"
	"gopkg.in/mgo.v2/bson"
)

func setGRPCHandlers(cli pb.NodeCommuunicationsClient, nsqProducer *nsq.Producer, networtkID int, wa chan pb.WatchAddress) {

	// initial fill mempool respectively network id
	go func() {
		stream, err := cli.EventGetAllMempool(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventGetAllMempool: %s", err.Error())
			// return nil, err
		}

		for {
			mpRec, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("setGRPCHandlers: client.EventGetAllMempool: %s", err.Error())
			}

			mpRates := &mgo.Collection{}
			switch networtkID {
			case currencies.Main:
				mpRates = mempoolRates
			case currencies.Test:
				mpRates = mempoolRatesTest
			default:
				log.Errorf("setGRPCHandlers: wrong networkID:")
			}

			err = mpRates.Insert(store.MempoolRecord{
				Category: int(mpRec.Category),
				HashTX:   mpRec.HashTX,
			})
			if err != nil {
				log.Errorf("initGrpcClient: mpRates.Insert: %s", err.Error())
			}
		}
	}()

	// add transaction on every new tx on node
	go func() {
		stream, err := cli.EventAddMempoolRecord(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventAddMempoolRecord: %s", err.Error())
			// return nil, err
		}

		mpRates := &mgo.Collection{}
		switch networtkID {
		case currencies.Main:
			mpRates = mempoolRates
		case currencies.Test:
			mpRates = mempoolRatesTest
		default:
			log.Errorf("setGRPCHandlers: wrong networkID:")
		}

		for {
			mpRec, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("setGRPCHandlers: client.EventAddMempoolRecord:stream.Recv: %s", err.Error())
			}
			err = mpRates.Insert(store.MempoolRecord{
				Category: int(mpRec.Category),
				HashTX:   mpRec.HashTX,
			})
			if err != nil {
				log.Errorf("initGrpcClient: mpRates.Insert: %s", err.Error())
			}
		}
	}()

	//deleting mempool record on block
	go func() {

		stream, err := cli.EventDeleteMempool(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventGetAllMempool: %s", err.Error())
			// return nil, err
		}

		mpRates := &mgo.Collection{}
		switch networtkID {
		case currencies.Main:
			mpRates = mempoolRates
		case currencies.Test:
			mpRates = mempoolRatesTest
		default:
			log.Errorf("setGRPCHandlers: wrong networkID:")
		}

		for {
			mpRec, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("initGrpcClient: cli.EventDeleteMempool:stream.Recv: %s", err.Error())
			}

			query := bson.M{"hashtx": mpRec.Hash}
			err = mpRates.Remove(query)

			if err != nil {
				log.Errorf("setGRPCHandlers:mpRates.Remove: %s", err.Error())
			} else {
				log.Debugf("Tx removed: %s", mpRec.Hash)
			}
		}

	}()

	// new spendable output
	go func() {
		stream, err := cli.EventAddSpendableOut(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventGetAllMempool: %s", err.Error())
		}

		spOutputs := &mgo.Collection{}
		switch networtkID {
		case currencies.Main:
			spOutputs = spendableOutputs
		case currencies.Test:
			spOutputs = spendableOutputsTest
		default:
			log.Errorf("setGRPCHandlers: wrong networkID:")
		}

		for {
			gSpOut, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("initGrpcClient: cli.EventAddSpendableOut:stream.Recv: %s", err.Error())
			}

			user := store.User{}
			sel := bson.M{"wallets.addresses.address": gSpOut.Address}
			err = usersData.Find(sel).One(&user)
			if err != nil && err != mgo.ErrNotFound {
				log.Errorf("SetWsHandlers: cli.On newIncomingTx: %s", err)
				return
			}
			spOut := generatedSpOutsToStore(gSpOut)

			for _, wallet := range user.Wallets {
				for _, address := range wallet.Adresses {
					if address.Address == spOut.Address {
						spOut.AddressIndex = address.AddressIndex
						spOut.WalletIndex = wallet.WalletIndex
					}
				}
			}

			//TODO: add exRates
			err = spOutputs.Insert(spOut)
			if err != nil {
				log.Errorf("SetWsHandlers: spendableOutputs.Insert: %s", err)
			}

		}

	}()

	// delete spendable output
	go func() {
		stream, err := cli.EventDeleteSpendableOut(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventGetAllMempool: %s", err.Error())
		}

		spOutputs := &mgo.Collection{}
		switch networtkID {
		case currencies.Main:
			spOutputs = spendableOutputs
		case currencies.Test:
			spOutputs = spendableOutputsTest
		default:
			log.Errorf("setGRPCHandlers: wrong networkID:")
		}

		for {
			del, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Errorf("initGrpcClient: cli.EventDeleteMempool:stream.Recv: %s", err.Error())
			}

			sel := bson.M{"txid": del.TxID, "userid": del.UserID, "address": del.Address}
			err = spOutputs.Find(sel).One(nil)
			if err == mgo.ErrNotFound {
				log.Errorf("SetWsHandlers: cli.On deleteSpout: spOutsCollection.Find: %s", err)
			} else {
				err = spOutputs.Remove(sel)
				if err != nil {
					log.Errorf("SetWsHandlers: cli.On deleteSpout: spOutsCollection.Remove: %s", err)
				}
			}
		}
	}()

	// add to transaction history record and send ws notification on tx
	go func() {
		stream, err := cli.NewTx(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventGetAllMempool: %s", err.Error())
		}

		for {
			gTx, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Errorf("initGrpcClient: cli.EventDeleteMempool:stream.Recv: %s", err.Error())
			}

			fmt.Printf("[DEBUG] Our tx %v addr %v \n", gTx.UserID, gTx.TxAddress)
			tx := generatedTxDataToStore(gTx)

			user := store.User{}
			setExchangeRates(&tx, true, tx.MempoolTime)
			for _, in := range tx.WalletsInput {
				sel := bson.M{"wallets.addresses.address": in.Address.Address}
				err := usersData.Find(sel).One(&user)
				if err == mgo.ErrNotFound {
					continue
				} else if err != nil && err != mgo.ErrNotFound {
					log.Errorf("initGrpcClient: cli.On newIncomingTx: %s", err)
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

			for _, out := range tx.WalletsOutput {
				sel := bson.M{"wallets.addresses.address": out.Address.Address}
				err := usersData.Find(sel).One(&user)
				if err == mgo.ErrNotFound {
					continue
				} else if err != nil && err != mgo.ErrNotFound {
					log.Errorf("initGrpcClient: cli.On newIncomingTx: %s", err)
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

			// err = txData.Insert(tx)
			err = saveMultyTransaction(tx, networtkID)
			if err != nil {
				log.Errorf("initGrpcClient: saveMultyTransaction: %s", err)
			}
			sendNotifyToClients(tx, nsqProducer)

		}
	}()

	// watch for channel and push to node
	go func() {
		for {
			select {
			case addr := <-wa:
				a := addr
				rp, err := cli.EventAddNewAddress(context.Background(), &a)
				if err != nil {
					log.Errorf("NewAddressNode: cli.EventAddNewAddress %s\n", err.Error())
				}
				log.Debugf("EventAddNewAddress Reply %s", rp.Message)

				rp, err = cli.EventResyncAddress(context.Background(), &pb.AddressToResync{
					Address: addr.Address,
				})
				if err != nil {
					log.Errorf("EventResyncAddress: cli.EventResyncAddress %s\n", err.Error())
				}
				log.Debugf("EventResyncAddress Reply %s", rp.Message)

			}
		}
	}()

}

func sendNotifyToClients(tx store.MultyTX, nsqProducer *nsq.Producer) {

	for _, walletOutput := range tx.WalletsOutput {
		txMsq := BtcTransactionWithUserID{
			UserID: walletOutput.UserId,
			NotificationMsg: &BtcTransaction{
				TransactionType: tx.TxStatus,
				Amount:          tx.TxOutAmount,
				TxID:            tx.TxID,
				Address:         walletOutput.Address.Address,
			},
		}
		sendNotify(&txMsq, nsqProducer)
	}

	for _, walletInput := range tx.WalletsInput {
		txMsq := BtcTransactionWithUserID{
			UserID: walletInput.UserId,
			NotificationMsg: &BtcTransaction{
				TransactionType: tx.TxStatus,
				Amount:          tx.TxOutAmount,
				TxID:            tx.TxID,
				Address:         walletInput.Address.Address,
			},
		}
		sendNotify(&txMsq, nsqProducer)
	}
}

func sendNotify(txMsq *BtcTransactionWithUserID, nsqProducer *nsq.Producer) {
	newTxJSON, err := json.Marshal(txMsq)
	if err != nil {
		log.Errorf("sendNotifyToClients: [%+v] %s\n", txMsq, err.Error())
		return
	}

	err = nsqProducer.Publish(TopicTransaction, newTxJSON)
	if err != nil {
		log.Errorf("nsq publish new transaction: [%+v] %s\n", txMsq, err.Error())
		return
	}

	return
}

func generatedTxDataToStore(gSpOut *pb.BTCTransaction) store.MultyTX {
	outs := []store.AddresAmount{}
	for _, output := range gSpOut.TxOutputs {
		outs = append(outs, store.AddresAmount{
			Address: output.Address,
			Amount:  output.Amount,
		})
	}

	ins := []store.AddresAmount{}
	for _, inputs := range gSpOut.TxInputs {
		ins = append(ins, store.AddresAmount{
			Address: inputs.Address,
			Amount:  inputs.Amount,
		})
	}

	return store.MultyTX{
		UserId:        gSpOut.UserID,
		TxID:          gSpOut.TxID,
		TxHash:        gSpOut.TxHash,
		TxOutScript:   gSpOut.TxOutScript,
		TxAddress:     gSpOut.TxAddress,
		TxStatus:      int(gSpOut.TxStatus),
		TxOutAmount:   gSpOut.TxOutAmount,
		BlockTime:     gSpOut.BlockTime,
		BlockHeight:   gSpOut.BlockHeight,
		Confirmations: int(gSpOut.Confirmations),
		TxFee:         gSpOut.TxFee,
		MempoolTime:   gSpOut.MempoolTime,
		TxInputs:      ins,
		TxOutputs:     outs,
	}
}

func generatedSpOutsToStore(gSpOut *pb.AddSpOut) store.SpendableOutputs {
	return store.SpendableOutputs{
		TxID:        gSpOut.TxID,
		TxOutID:     int(gSpOut.TxOutID),
		TxOutAmount: gSpOut.TxOutAmount,
		TxOutScript: gSpOut.TxOutScript,
		Address:     gSpOut.Address,
		UserID:      gSpOut.UserID,
		TxStatus:    int(gSpOut.TxStatus),
	}
}

func saveMultyTransaction(tx store.MultyTX, networtkID int) error {

	txsdata := &mgo.Collection{}
	switch networtkID {
	case currencies.Main:
		txsdata = txsData
	case currencies.Test:
		txsdata = txsDataTest
	default:
		log.Errorf("setGRPCHandlers: wrong networkID:")
	}

	// This is splited transaction! That means that transaction's WalletsInputs and WalletsOutput have the same WalletIndex!
	//Here we have outgoing transaction for exact wallet!
	multyTX := store.MultyTX{}
	if tx.WalletsInput != nil && len(tx.WalletsInput) > 0 {
		// sel := bson.M{"userid": tx.WalletsInput[0].UserId, "transactions.txid": tx.TxID, "transactions.walletsinput.walletindex": tx.WalletsInput[0].WalletIndex}
		sel := bson.M{"userid": tx.WalletsInput[0].UserId, "txid": tx.TxID, "walletsinput.walletindex": tx.WalletsInput[0].WalletIndex}
		err := txsdata.Find(sel).One(&multyTX)
		if err == mgo.ErrNotFound {
			// initial insertion
			err := txsdata.Insert(tx)
			if err != nil {
				return err
			}

		}
		if err != nil && err != mgo.ErrNotFound {
			// database error
			return err
		}

		update := bson.M{
			"$set": bson.M{
				"txstatus":      tx.TxStatus,
				"blockheight":   tx.BlockHeight,
				"confirmations": tx.Confirmations,
				"blocktime":     tx.BlockTime,
			},
		}
		err = txsdata.Update(sel, update)
		if err != nil {
			return err
		}
		return nil
	} else if tx.WalletsOutput != nil && len(tx.WalletsOutput) > 0 {
		// sel := bson.M{"userid": tx.WalletsOutput[0].UserId, "transactions.txid": tx.TxID, "transactions.walletsoutput.walletindex": tx.WalletsOutput[0].WalletIndex}
		sel := bson.M{"userid": tx.WalletsOutput[0].UserId, "txid": tx.TxID, "walletsoutput.walletindex": tx.WalletsOutput[0].WalletIndex}
		err := txsdata.Find(sel).One(&multyTX)
		if err == mgo.ErrNotFound {
			// initial insertion
			err := txsdata.Insert(tx)
			if err != nil {
				return err
			}
			return nil
		}
		if err != nil && err != mgo.ErrNotFound {
			// database error
			return err
		}

		update := bson.M{
			"$set": bson.M{
				"txstatus":      tx.TxStatus,
				"blockheight":   tx.BlockHeight,
				"confirmations": tx.Confirmations,
				"blocktime":     tx.BlockTime,
			},
		}
		err = txsdata.Update(sel, update)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func SetWsHandlers(cli *gosocketio.Client, networkID int) {

	cli.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		log.Errorf("\n\n\n\n Ws Connected to Service Node\n\n\n\n\n")
	})

	cli.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		debug.PrintStack()
		log.Errorf("\n\n\n\n Ws Disconnected from Service Node\n\n\n\n\n")
	})

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
	// cli.On(EventMempool, func(c *gosocketio.Channel, recs []store.MempoolRecord) {
	// 	fmt.Println(recs)
	// 	InsertMempoolRecords(recs...)
	// })

	// // Mempool delete on block
	// cli.On(EventDeleteMempoolOnBlock, func(c *gosocketio.Channel, hash string) {
	// 	query := bson.M{"hashtx": hash}
	// 	err := mempoolRates.Remove(query)
	// 	if err != nil {
	// 		log.Errorf("parseNewBlock:mempoolRates.Remove: %s", err.Error())
	// 	} else {
	// 		log.Debugf("Tx removed: %s", hash)
	// 	}
	// })
}
