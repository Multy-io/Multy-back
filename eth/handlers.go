/*
Copyright 2019 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package eth

import (
	"context"
	"io"
	"sync"
	"time"

	pb "github.com/Multy-io/Multy-back/ns-eth-protobuf"
	"github.com/Multy-io/Multy-back/currencies"
	"github.com/Multy-io/Multy-back/store"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (ethcli *ETHConn) setGRPCHandlers(networtkID int, accuracyRange int) {
	mempoolCh := make(chan interface{})
	// initial fill mempool respectively network id

	var client pb.NodeCommunicationsClient
	var wa chan pb.WatchAddress
	var mempool sync.Map

	nsqProducer := ethcli.NsqProducer

	switch networtkID {
	case currencies.ETHMain:
		client = ethcli.CliMain
		wa = ethcli.WatchAddressMain
		mempool = ethcli.Mempool
	case currencies.ETHTest:
		client = ethcli.CliTest
		wa = ethcli.WatchAddressTest
		mempool = ethcli.MempoolTest

	}

	go func() {
		stream, err := client.EventGetAllMempool(context.Background(), &pb.Empty{})
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
				Category: int(mpRec.Category),
				HashTX:   mpRec.HashTX,
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

			mempoolCh <- mpRec.Hash

			if err != nil {
				log.Errorf("setGRPCHandlers:mpRates.Remove: %s", err.Error())
			} else {
				// log.Debugf("Tx removed: %s", mpRec.Hash)
			}
		}

	}()

	go func() {
		stream, err := client.AddMultisig(context.Background(), &pb.Empty{})
		if err != nil {
			log.Errorf("setGRPCHandlers: cli.AddMultisig: %s", err.Error())
		}

		// Loop:
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

			log.Warnf("Multisig recved on address contract %v ", multisigTx.Contract)

			multisig := generatedMultisigTxToStore(multisigTx, currencies.Ether, networtkID)

			log.Warnf("\n\n\n\n\n multisigTx.DeployStatus = %v", multisigTx.DeployStatus)

			users := msToUserData(multisigTx.Addresses, usersData)

			invitecode := fetchInviteUndeployed(users)
			log.Warnf("\ninvitecode %v", invitecode)

			emptyCode := false
			if invitecode == "" {
				log.Errorf("cli.AddMultisig:stream.Recv:not found contract transaction %v", multisigTx.Addresses)
				emptyCode = true
			}

			msUser := store.User{}
			err = usersData.Find(bson.M{"multisig.inviteCode": invitecode}).One(&msUser)
			doubleInvited := false
			for _, checkMs := range msUser.Multisigs {
				if checkMs.InviteCode == invitecode {
					if checkMs.ContractAddress != "" {
						doubleInvited = true
					}
				}
			}

			if !doubleInvited && !emptyCode {
				for _, user := range users {
					addrs, err := FetchUserAddresses(currencies.Ether, multisig.NetworkID, user, multisigTx.Addresses)
					if err != nil {
						log.Errorf("createMultisig:FetchUserAddresses: %v", err.Error())
					}

					for _, addr := range addrs {
						log.Warnf("addr :%v AddressIndex: %v Associated: %v UserID: %v \n", addr.Address, addr.AddressIndex, addr.Associated, addr.UserID)
					}

					multisig.Owners = addrs

					sel := bson.M{"userID": user.UserID, "multisig.inviteCode": invitecode}
					update := bson.M{"$set": bson.M{
						"multisig.$.confirmations":   multisig.Confirmations,
						"multisig.$.contractAddress": multisig.ContractAddress,
						"multisig.$.txOfCreation":    multisig.TxOfCreation,
						"multisig.$.factoryAddress":  multisig.FactoryAddress,
						"multisig.$.lastActionTime":  multisig.LastActionTime,
						"multisig.$.deployStatus":    multisig.DeployStatus,
					}}

					err = usersData.Update(sel, update)
					if err != nil {
						log.Errorf("cli.AddMultisig:stream.Recv:userStore.Update: %s", err.Error())
					}
					multisig.InviteCode = invitecode
					msg := store.WsMessage{
						Type:    store.NotifyDeploy,
						To:      user.UserID,
						Date:    time.Now().Unix(),
						Payload: multisig,
					}
					ethcli.WsServer.BroadcastToAll(store.MsgRecieve+":"+user.UserID, msg)
				}
			}
		}
		log.Panicf("AddMultisig")
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
				log.Errorf("initGrpcClient: cli.NewTx:stream.Recv: %s", err.Error())
			}

			log.Warnf("new tx for uid %v ", gTx.GetUserID())

			tx := generatedTxDataToStore(gTx)
			setExchangeRates(&tx, gTx.Resync, tx.BlockTime)

			err = saveTransaction(tx, networtkID, gTx.Resync)
			updateWalletAndAddressDate(tx, networtkID)
			if err != nil {
				log.Errorf("initGrpcClient: saveMultyTransaction: %s", err)
			}

			if !gTx.GetResync() {
				sendNotifyToClients(tx, nsqProducer, networtkID)
			}
			// process multisig txs
			if gTx.Multisig {
				methodInvoked, err := processMultisig(&tx, networtkID, nsqProducer, ethcli)
				log.Warnf("methodInvoked %v tx.Multisig.Return %v ", methodInvoked, tx.Multisig.Return)
				// ws notify about all kinds of ms transactions
				sel := bson.M{"multisig.contractAddress": tx.Multisig.Contract}
				users := []store.User{}
				err = usersData.Find(sel).All(&users)
				if err != nil {
					log.Errorf("initGrpcClient:gTx.Multisig:usersData.Find: %s", err.Error())
				}
				for _, user := range users {
					msg := store.WsMessage{
						Type:    signatuteToStatus(methodInvoked),
						To:      user.UserID,
						Date:    time.Now().Unix(),
						Payload: gTx,
					}
					ethcli.WsServer.BroadcastToAll(store.MsgRecieve+":"+user.UserID, msg)
				}

				if err != nil {
					log.Errorf("initGrpcClient: processMultisig: %s", err.Error())
				}
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

			height := h.GetHeight()
			sel := bson.M{"currencyid": currencies.Ether, "networkid": networtkID}
			update := bson.M{
				"$set": bson.M{
					"blockheight": height,
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
			case currencies.ETHMain:
				txStore = txsData
				nsCli = ethcli.CliMain
			case currencies.ETHTest:
				txStore = txsDataTest
				nsCli = ethcli.CliTest
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

					// reject outcoming
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
