/*
Copyright 2019 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"context"
	"io"
	"sync"
	"time"

	pb "github.com/Multy-io/Multy-BTC-node-service/node-streamer"
	"github.com/Multy-io/Multy-back/currencies"
	"github.com/Multy-io/Multy-back/store"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (btcCli *BTCConn) setGRPCHandlers(networtkID, accuracyRange int) {

	var client pb.NodeCommunicationsClient
	var wa chan pb.WatchAddress
	var mempool sync.Map

	nsqProducer := btcCli.NsqProducer
	resync := btcCli.Resync

	switch networtkID {
	case currencies.Main:
		client = btcCli.CliMain
		wa = btcCli.WatchAddressMain
		mempool = btcCli.BtcMempool
	case currencies.Test:
		client = btcCli.CliTest
		wa = btcCli.WatchAddressTest
		mempool = btcCli.BtcMempoolTest

	}

	mempoolCh := make(chan interface{})
	// initial fill mempool respectively network id
	go func() {
		stream, err := client.EventGetAllMempool(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventGetAllMempool: %s", err.Error())
		}

		for {
			mpRec, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("setGRPCHandlers: client.EventGetAllMempool: %s", err.Error())
			}

			mempoolCh <- store.MempoolRecord{
				Category: int(mpRec.GetCategory()),
				HashTX:   mpRec.GetHashTX(),
			}

		}
	}()

	// add transaction on every new tx on node
	go func() {
		stream, err := client.EventAddMempoolRecord(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventAddMempoolRecord: %s", err.Error())
			// return nil, err
		}

		for {
			mpRec, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("setGRPCHandlers: client.EventAddMempoolRecord:stream.Recv: %s", err.Error())
			}

			mempoolCh <- store.MempoolRecord{
				Category: int(mpRec.GetCategory()),
				HashTX:   mpRec.GetHashTX(),
			}

			if err != nil {
				log.Errorf("initGrpcClient: mpRates.Insert: %s", err.Error())
			}
		}
	}()

	go func() {
		stream, err := client.EventNewBlock(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventNewBlock: %s", err.Error())
			// return nil, err
		}

		for {
			h, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("setGRPCHandlers: client.EventNewBlock:stream.Recv: %s", err.Error())
			}

			sel := bson.M{"currencyid": currencies.Bitcoin, "networkid": networtkID}
			update := bson.M{
				"$set": bson.M{
					"blockheight": h.GetHeight(),
				},
			}

			_, err = restoreState.Upsert(sel, update)
			if err != nil {
				log.Errorf("restoreState.Upsert: %v", err.Error())
			}

			// check for rejected transactions
			var txStore *mgo.Collection
			var nsCli pb.NodeCommunicationsClient
			switch networtkID {
			case currencies.Main:
				txStore = txsData
				nsCli = btcCli.CliMain
			case currencies.Test:
				txStore = txsDataTest
				nsCli = btcCli.CliTest
			}

			query := bson.M{
				"$or": []bson.M{
					bson.M{"$and": []bson.M{
						bson.M{"blockheight": 0},
						bson.M{"txstatus": bson.M{"$nin": []int{store.TxStatusTxRejectedOutgoing, store.TxStatusTxRejectedIncoming}}},
					}},

					bson.M{"$and": []bson.M{
						bson.M{"blockheight": bson.M{"$lt": h.GetHeight()}},
						bson.M{"blockheight": bson.M{"$gt": h.GetHeight() - int64(accuracyRange)}},
					}},
				},
			}

			txs := []store.TransactionETH{}
			txStore.Find(query).All(&txs)

			hashes := &pb.TxsToCheck{}

			for _, tx := range txs {
				hashes.Hash = append(hashes.Hash, tx.Hash)
			}

			txToReject, err := nsCli.CheckRejectTxs(context.Background(), hashes)
			if err != nil {
				log.Errorf("setGRPCHandlers: CheckRejectTxs: %s", err.Error())
			}

			// Set status to rejected in db
			if len(txToReject.GetRejectedTxs()) > 0 {

				for _, hash := range txToReject.GetRejectedTxs() {
					// reject incoming
					query := bson.M{"$and": []bson.M{
						bson.M{"hash": hash},
						bson.M{"txstatus": bson.M{"$in": []int{store.TxStatusAppearedInMempoolIncoming,
							store.TxStatusAppearedInBlockIncoming,
							store.TxStatusInBlockConfirmedIncoming}}},
					}}

					update := bson.M{
						"$set": bson.M{
							"txstatus": store.TxStatusTxRejectedIncoming,
						},
					}
					_, err := txStore.UpdateAll(query, update)
					if err != nil {
						log.Errorf("setGRPCHandlers: cli.EventNewBlock:txStore.UpdateAll:Incoming: %s", err.Error())
					}

					query = bson.M{"$and": []bson.M{
						bson.M{"hash": hash},
						bson.M{"txstatus": bson.M{"$in": []int{store.TxStatusAppearedInMempoolOutcoming,
							store.TxStatusAppearedInBlockOutcoming,
							store.TxStatusInBlockConfirmedOutcoming}}},
					}}
					update = bson.M{
						"$set": bson.M{
							"txstatus": store.TxStatusTxRejectedOutgoing,
						},
					}
					_, err = txStore.UpdateAll(query, update)
					if err != nil {
						log.Errorf("setGRPCHandlers: cli.EventNewBlock:txStore.UpdateAll:Outcoming: %s", err.Error())
					}
				}

				if err != nil {
					log.Errorf("initGrpcClient: restoreState.Update: %s", err.Error())
				}
			}

		}
	}()

	//deleting mempool record on block
	go func() {
		stream, err := client.EventDeleteMempool(context.Background(), &pb.Empty{})
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
				log.Errorf("initGrpcClient: cli.EventDeleteMempool:stream.Recv: %s", err.Error())
			}

			mempoolCh <- mpRec.GetHash()

			if err != nil {
				log.Errorf("setGRPCHandlers:mpRates.Remove: %s", err.Error())
			} else {
				// log.Debugf("Tx removed: %s", mpRec.Hash)
			}
		}

	}()

	// new spendable output
	go func() {
		stream, err := client.EventAddSpendableOut(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventGetAllMempool: %s", err.Error())
		}

		spOutputs := &mgo.Collection{}
		spend := &mgo.Collection{}
		switch networtkID {
		case currencies.Main:
			spOutputs = spendableOutputs
			spend = spentOutputs
		case currencies.Test:
			spOutputs = spendableOutputsTest
			spend = spentOutputsTest
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

			query := bson.M{"userid": gSpOut.GetUserID(), "txid": gSpOut.GetTxID(), "address": gSpOut.GetAddress()}
			err = spend.Find(query).One(nil)

			if err == mgo.ErrNotFound {
				user := store.User{}
				sel := bson.M{"wallets.addresses.address": gSpOut.GetAddress()}
				err = usersData.Find(sel).One(&user)
				if err != nil && err != mgo.ErrNotFound {
					log.Errorf("SetWsHandlers: cli.On newIncomingTx: %s", err)
					return
				}

				spOut := generatedSpOutsToStore(gSpOut)

				log.Infof("Add spendable output : %v", gSpOut.String())

				exRates, err := GetLatestExchangeRate()
				if err != nil {
					log.Errorf("initGrpcClient: GetLatestExchangeRate: %s", err.Error())
				}
				spOut.StockExchangeRate = exRates

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

		}

	}()

	// delete spendable output
	go func() {
		stream, err := client.EventDeleteSpendableOut(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventGetAllMempool: %s", err.Error())
		}
		spOutputs := &mgo.Collection{}
		spend := &mgo.Collection{}
		switch networtkID {
		case currencies.Main:
			spOutputs = spendableOutputs
			spend = spentOutputs
		case currencies.Test:
			spOutputs = spendableOutputsTest
			spend = spentOutputsTest
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

			i := 0
			for {
				//insert to spend collection
				var once sync.Once

				once.Do(func() {
					err := spend.Insert(del)
					if err != nil {
						log.Errorf("DeleteSpendableOutputs:spend.Insert: %s", err)
					}
				})

				query := bson.M{"userid": del.UserID, "txid": del.TxID, "address": del.Address}
				log.Infof("-------- query delete %v\n", query)
				err = spOutputs.Remove(query)
				if err != nil {
					log.Errorf("DeleteSpendableOutputs:spendableOutputs.Remove: %s", err.Error())
				} else {
					log.Infof("delete success √: %v", query)
					break
				}
				i++
				if i == 4 {
					break
				}
				time.Sleep(time.Second * 3)
			}
			log.Debugf("DeleteSpendableOutputs:spendableOutputs.Remove: %s", err)
		}
	}()

	// add to transaction history record and send ws notification on tx
	go func() {
		stream, err := client.NewTx(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventGetAllMempool: %s", err.Error())
		}

		for {
			gTx, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("initGrpcClient: cli.NewTx:stream.Recv: %s", err.Error())
			}
			tx := generatedTxDataToStore(gTx)

			setExchangeRates(&tx, gTx.GetResync(), tx.MempoolTime)
			setUserID(&tx)
			// setTxInfo(&tx)
			user := store.User{}
			// set wallet index and address index in input
			for i := 0; i < len(tx.WalletsInput); i++ {
				sel := bson.M{"wallets.addresses.address": tx.WalletsInput[i].Address.Address}
				err := usersData.Find(sel).One(&user)
				if err == mgo.ErrNotFound {
					continue
				} else if err != nil && err != mgo.ErrNotFound {
					log.Errorf("initGrpcClient: cli.On newIncomingTx: %s", err)
				}
				for _, wallet := range user.Wallets {
					for _, addr := range wallet.Adresses {
						if addr.Address == tx.WalletsInput[i].Address.Address {
							tx.WalletsInput[i].WalletIndex = wallet.WalletIndex
							tx.WalletsInput[i].Address.AddressIndex = addr.AddressIndex
						}
					}
				}
			}

			for i := 0; i < len(tx.WalletsOutput); i++ {
				sel := bson.M{"wallets.addresses.address": tx.WalletsOutput[i].Address.Address}
				err := usersData.Find(sel).One(&user)
				if err == mgo.ErrNotFound {
					continue
				} else if err != nil && err != mgo.ErrNotFound {
					log.Errorf("initGrpcClient: cli.On newIncomingTx: %s", err)
				}

				for _, wallet := range user.Wallets {
					for _, addr := range wallet.Adresses {
						if addr.Address == tx.WalletsOutput[i].Address.Address {
							tx.WalletsOutput[i].WalletIndex = wallet.WalletIndex
							tx.WalletsOutput[i].Address.AddressIndex = addr.AddressIndex
						}
					}
				}
			}

			log.Infof("New tx history in- %v out-%v\n", tx.WalletsInput, tx.WalletsOutput)

			err = saveMultyTransaction(tx, networtkID, gTx.Resync)
			if err != nil {
				log.Errorf("initGrpcClient: saveMultyTransaction: %s", err)
			}
			updateWalletAndAddressDate(tx, networtkID)
			if !gTx.Resync {
				sendNotifyToClients(tx, nsqProducer, networtkID)
			}
		}
	}()

	// Resync tx history and spendable outputs
	go func() {
		spOutputs := &mgo.Collection{}
		spend := &mgo.Collection{}
		switch networtkID {
		case currencies.Main:
			spOutputs = spendableOutputs
			spend = spentOutputs
		case currencies.Test:
			spOutputs = spendableOutputsTest
			spend = spentOutputsTest
		default:
			log.Errorf("setGRPCHandlers: wrong networkID:")
		}

		stream, err := client.ResyncAddress(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.EventGetAllMempool: %s", err.Error())
		}

		for {
			rTxs, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("initGrpcClient: cli.NewTx:stream.Recv: %s", err.Error())
			}

			if rTxs.GetDeleteFromQueue() != "" {
				resync.Delete(rTxs.GetDeleteFromQueue())
			}

			// tx history
			if rTxs.GetTxs() != nil {
				for _, gTx := range rTxs.Txs {
					tx := generatedTxDataToStore(gTx)
					setExchangeRates(&tx, gTx.Resync, tx.MempoolTime)
					setUserID(&tx)
					user := store.User{}
					// set wallet index and address index in input
					for i := 0; i < len(tx.WalletsInput); i++ {
						sel := bson.M{"wallets.addresses.address": tx.WalletsInput[i].Address.Address}
						err := usersData.Find(sel).One(&user)
						if err == mgo.ErrNotFound {
							continue
						} else if err != nil && err != mgo.ErrNotFound {
							log.Errorf("initGrpcClient: cli.On newIncomingTx: %s", err)
						}

						for _, wallet := range user.Wallets {
							for _, addr := range wallet.Adresses {
								if addr.Address == tx.WalletsInput[i].Address.Address {
									tx.WalletsInput[i].WalletIndex = wallet.WalletIndex
									tx.WalletsInput[i].Address.AddressIndex = addr.AddressIndex
								}
							}
						}
					}
					// set wallet index and address index in output
					for i := 0; i < len(tx.WalletsOutput); i++ {
						sel := bson.M{"wallets.addresses.address": tx.WalletsOutput[i].Address.Address}
						err := usersData.Find(sel).One(&user)
						if err == mgo.ErrNotFound {
							continue
						} else if err != nil && err != mgo.ErrNotFound {
							log.Errorf("initGrpcClient: cli.On newIncomingTx: %s", err)
						}

						for _, wallet := range user.Wallets {
							for _, addr := range wallet.Adresses {
								if addr.Address == tx.WalletsOutput[i].Address.Address {
									tx.WalletsOutput[i].WalletIndex = wallet.WalletIndex
									tx.WalletsOutput[i].Address.AddressIndex = addr.AddressIndex
								}
							}
						}
					}
					err = saveMultyTransaction(tx, networtkID, gTx.Resync)
					if err != nil {
						log.Errorf("initGrpcClient: saveMultyTransaction: %s", err)
					}
					updateWalletAndAddressDate(tx, networtkID)
				}
			}

			// sp outs
			if rTxs.GetSpOuts() != nil {
				for _, gSpOut := range rTxs.SpOuts {
					query := bson.M{"userid": gSpOut.UserID, "txid": gSpOut.TxID, "address": gSpOut.Address}
					err = spend.Find(query).One(nil)
					if err == mgo.ErrNotFound {
						user := store.User{}
						sel := bson.M{"wallets.addresses.address": gSpOut.Address}
						err = usersData.Find(sel).One(&user)
						if err != nil && err != mgo.ErrNotFound {
							log.Errorf("SetWsHandlers: cli.On newIncomingTx: %s", err)
							return
						}
						spOut := generatedSpOutsToStore(gSpOut)
						log.Infof("Add spendable output : %v", gSpOut.String())
						exRates, err := GetLatestExchangeRate()
						if err != nil {
							log.Errorf("initGrpcClient: GetLatestExchangeRate: %s", err.Error())
						}
						spOut.StockExchangeRate = exRates

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
				}
			}

			// del sp outs
			if rTxs.GetSpOutDelete() != nil {
				for _, del := range rTxs.SpOutDelete {
					i := 0
					for {
						//insert to spend collection
						err = spend.Insert(del)
						if err != nil {
							log.Errorf("DeleteSpendableOutputs:spend.Insert: %s", err)
						}
						query := bson.M{"userid": del.UserID, "txid": del.TxID, "address": del.Address}
						log.Infof("-------- query delete %v\n", query)
						err = spOutputs.Remove(query)
						if err != nil {
							log.Errorf("DeleteSpendableOutputs:spendableOutputs.Remove: %s", err.Error())
						} else {
							log.Infof("delete success √: %v", query)
							break
						}
						i++
						if i == 4 {
							break
						}
						time.Sleep(time.Second * 3)
					}
				}
			}

			if rTxs.GetTxs() != nil && len(rTxs.Txs) > 0 {
				for _, spout := range rTxs.GetSpOuts() {
					resync.Delete(spout.Address)
				}
				if len(rTxs.SpOuts) > 0 {

					msg := store.WsMessage{
						Type:    store.NotifyResyncEnd,
						To:      rTxs.SpOuts[0].UserID,
						Date:    time.Now().Unix(),
						Payload: "",
					}
					btcCli.WsServer.BroadcastToAll(store.MsgRecieve+":"+rTxs.SpOuts[0].UserID, msg)
				}
			}

		}

	}()

	// watch for channel and push to node
	go func() {
		for {
			select {
			case addr := <-wa:
				a := addr
				rp, err := client.EventAddNewAddress(context.Background(), &a)
				if err != nil {
					log.Errorf("NewAddressNode: cli.EventAddNewAddress %s\n", err.Error())
				}
				log.Debugf("EventAddNewAddress Reply %s", rp)

				rp, err = client.EventResyncAddress(context.Background(), &pb.AddressToResync{
					Address:      addr.GetAddress(),
					UserID:       addr.GetUserID(),
					WalletIndex:  addr.GetWalletIndex(),
					AddressIndex: addr.GetWalletIndex(),
				})
				if err != nil {
					log.Errorf("EventResyncAddress: cli.EventResyncAddress %s\n", err.Error())
				}
				log.Debugf("EventResyncAddress Reply %s", rp)

			}
		}
	}()

	go func() {
		for {
			switch v := (<-mempoolCh).(type) {
			case string:
				// delete tx from pool
				mempool.Delete(v)
			case store.MempoolRecord:
				// add tx to pool
				mempool.Store(v.HashTX, v.Category)
			}
		}
	}()

}
