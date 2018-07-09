/*
Copyright 2019 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package eth

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/Appscrunch/Multy-back/currencies"
	pb "github.com/Appscrunch/Multy-back/node-streamer/eth"
	"github.com/Appscrunch/Multy-back/store"
	nsq "github.com/bitly/go-nsq"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func setGRPCHandlers(cli pb.NodeCommuunicationsClient, nsqProducer *nsq.Producer, networtkID int, wa chan pb.WatchAddress, mempool sync.Map) {

	mempoolCh := make(chan interface{})
	log.Errorf("\n\n\nnetwortkIDnetwortkIDnetwortkIDnetwortkIDnetwortkID %v\n\n\n ", networtkID)
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

			mempoolCh <- store.MempoolRecord{
				Category: int(mpRec.Category),
				HashTX:   mpRec.HashTX,
			}
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

		for {
			mpRec, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("setGRPCHandlers: client.EventAddMempoolRecord:stream.Recv: %s", err.Error())
			}
			mempoolCh <- store.MempoolRecord{
				Category: int(mpRec.Category),
				HashTX:   mpRec.HashTX,
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

		for {
			mpRec, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("initGrpcClient: cli.EventDeleteMempool:stream.Recv: %s", err.Error())
			}

			mempoolCh <- mpRec.Hash

			if err != nil {
				log.Errorf("setGRPCHandlers:mpRates.Remove: %s", err.Error())
			} else {
				// log.Debugf("Tx removed: %s", mpRec.Hash)
			}
		}

	}()

	go func() {
		stream, err := cli.AddMultisig(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.AddMultisig: %s", err.Error())
		}

		for {
			multisigTx, err := stream.Recv()
			if err == io.EOF {
				break
			}
			// Add or notify about error
			if err != nil {
				log.Errorf("initGrpcClient: cli.AddMultisig:stream.Recv: %s", err.Error())
			}
			log.Debugf("initGrpcClient: cli.AddMultisig:stream.Recv:")
			users := map[string]store.User{}
			multisig := generatedMultisigTxToStore(multisigTx)
			multisig.CurrencyID = currencies.Ether

			//TODO: fix problem with net id

			if networtkID == 0 {
				multisig.NetworkID = currencies.ETHMain
			}

			if networtkID == 1 {
				multisig.NetworkID = currencies.ETHTest
			}

			// feth ussers included as owners in multisig
			for _, address := range multisigTx.Addresses {
				log.Debugf("range multisigTx.Addresses")
				user := store.User{}
				err := usersData.Find(bson.M{"wallets.addresses.address": strings.ToLower(address)}).One(&user)
				if err != nil {
					log.Errorf("cli.AddMultisig:stream.Recv:usersData.Find: not multy user in contrat %v  %v", err.Error(), address)
					break
				}
				users[user.UserID] = user
			}

			for uid, _ := range users {
				log.Warnf("\nuser :%v \n", uid)
			}

			for userid, user := range users {
				addrs, err := FethUserAddresses(currencies.Ether, multisig.NetworkID, user, multisigTx.Addresses)
				if err != nil {
					log.Errorf("createMultisig:FethUserAddresses: %v", err.Error())
				}

				for _, addr := range addrs {
					log.Warnf("addr :%v AddressIndex: %v Associated: %v UserID: %v \n", addr.Address, addr.AddressIndex, addr.Associated, addr.UserID)
				}

				multisig.Owners = addrs

				sel := bson.M{"userID": userid}
				update := bson.M{"$push": bson.M{"multisig": multisig}}

				err = usersData.Update(sel, update)
				if err != nil {
					log.Errorf("cli.AddMultisig:stream.Recv:userStore.Update: %s", err.Error())
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
				log.Errorf("initGrpcClient: cli.NewTx:stream.Recv: %s", err.Error())
			}

			tx := generatedTxDataToStore(gTx)
			setExchangeRates(&tx, gTx.Resync, tx.BlockTime)

			if !gTx.GetMultisig() {
				err = saveTransaction(tx, networtkID, gTx.Resync)
				updateWalletAndAddressDate(tx, networtkID)
				if err != nil {
					log.Errorf("initGrpcClient: saveMultyTransaction: %s", err)
				}

				if !gTx.GetResync() {
					sendNotifyToClients(tx, nsqProducer, networtkID)
				}
			}

			err = processMultisig(&tx, networtkID)
			if err != nil {
				log.Errorf("initGrpcClient: processMultisig: %s", err.Error())
			}

		}
	}()

	go func() {
		stream, err := cli.EventNewBlock(context.Background(), &pb.Empty{})
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

			query := bson.M{"currencyid": currencies.Bitcoin, "networkid": networtkID}
			update := bson.M{
				"$set": bson.M{
					"blockheight": h.GetHeight(),
				},
			}

			err = restoreState.Update(query, update)
			if err == mgo.ErrNotFound {
				restoreState.Insert(store.LastState{
					BlockHeight: h.GetHeight(),
					CurrencyID:  currencies.Bitcoin,
					NetworkID:   networtkID,
				})
			}

			if err != nil {
				log.Errorf("initGrpcClient: restoreState.Update: %s", err.Error())
			}
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

	go func() {

		for {
			switch v := (<-mempoolCh).(type) {
			// default:
			// 	log.Errorf("Not found type: %v", v)
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
