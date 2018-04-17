/*
Copyright 2019 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package eth

import (
	"context"
	"io"

	"gopkg.in/mgo.v2"

	"github.com/Appscrunch/Multy-back/currencies"
	pb "github.com/Appscrunch/Multy-back/node-streamer/eth"
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
				log.Errorf("initGrpcClient: cli.NewTx:stream.Recv: %s", err.Error())
			}
			tx := generatedTxDataToStore(gTx)
			setExchangeRates(&tx, gTx.Resync, tx.BlockTime)

			err = saveTransaction(tx, networtkID, gTx.Resync)
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
