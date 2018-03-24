/*
Copyright 2019 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"context"
	"fmt"
	"io"

	"gopkg.in/mgo.v2"

	"github.com/Appscrunch/Multy-back/currencies"
	pb "github.com/Appscrunch/Multy-back/node-streamer/btc"
	"github.com/Appscrunch/Multy-back/store"
	nsq "github.com/bitly/go-nsq"
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

			// set wallet index and addres index
			for _, wallet := range user.Wallets {
				for _, address := range wallet.Adresses {
					if address.Address == spOut.Address {
						spOut.AddressIndex = address.AddressIndex
						spOut.WalletIndex = wallet.WalletIndex
					}
				}
			}
			//TODO: add exRates

			query := bson.M{"userid": spOut.UserID, "txid": spOut.TxID, "address": spOut.Address}
			err = spOutputs.Find(query).One(nil)
			if err == mgo.ErrNotFound {
				//insertion
				err := spOutputs.Insert(spOut)
				if err != nil {
					log.Errorf("Create spOutputs:txsData.Insert: %s", err.Error())
				}
				continue
			}
			if err != nil && err != mgo.ErrNotFound {
				log.Errorf("Create spOutputs:spOutputs.Find %s", err.Error())
				continue
			}

			update := bson.M{
				"$set": bson.M{
					"txstatus": spOut.TxStatus,
				},
			}
			err = spOutputs.Update(query, update)
			if err != nil {
				log.Errorf("CreateSpendableOutputs:spendableOutputs.Update: %s", err.Error())
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

			query := bson.M{"userid": del.UserID, "txid": del.TxID, "address": del.Address}
			err = spOutputs.Remove(query)
			if err != nil {
				log.Errorf("DeleteSpendableOutputs:spendableOutputs.Remove: %s", err.Error())
			}
			log.Debugf("DeleteSpendableOutputs:spendableOutputs.Remove: %s", err)

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

			tx := generatedTxDataToStore(gTx)
			fmt.Println("-------out", gTx.WalletsOutput)
			fmt.Println("-------in", gTx.WalletsInput)

			user := store.User{}
			setExchangeRates(&tx, true, tx.MempoolTime)
			setUserID(&tx)

			//TODO: wrap to func
			// set wallet index and address index in input
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

			// set wallet index and address index in output
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

			log.Infof("[DEBUG] Our tx %v \n", tx)
			// err = txData.Insert(tx)
			err = saveMultyTransaction(tx, networtkID)
			if err != nil {
				log.Errorf("initGrpcClient: saveMultyTransaction: %s", err)
			}
			updateWalletAndAddressDate(tx)
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
				log.Debugf("EventAddNewAddress Reply %s", rp)

				rp, err = cli.EventResyncAddress(context.Background(), &pb.AddressToResync{
					Address: addr.Address,
				})
				if err != nil {
					log.Errorf("EventResyncAddress: cli.EventResyncAddress %s\n", err.Error())
				}
				log.Debugf("EventResyncAddress Reply %s", rp)

			}
		}
	}()

}
