/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package ethereum

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/onrik/ethrpc"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func newETHtx(hash, from, to string, amount float64, gas, gasprice, nonce int) store.MultyETHTransaction {
	return store.MultyETHTransaction{
		Hash:     hash,
		From:     from,
		To:       to,
		Amount:   amount,
		Gas:      gas,
		GasPrice: gasprice,
		Nonce:    nonce,
	}
}

func parseRawTransaction(rawTX ethrpc.Transaction, pending bool) store.MultyETHTransaction {
	tx := newETHtx(rawTX.Hash, rawTX.From, rawTX.To, (float64(rawTX.Value.Int64()) / 1000000000000000000), rawTX.Gas, int(rawTX.GasPrice.Int64()), rawTX.Nonce)
	str := "from block"
	if pending {
		str = "from txpool"
	}
	fmt.Println(tx, str, "\n")
	return tx
}

func rawToMultyETH(rawTX ethrpc.Transaction) store.MultyETHTransaction {
	return newETHtx(rawTX.Hash, rawTX.From, rawTX.To, (float64(rawTX.Value.Int64()) / 1000000000000000000), rawTX.Gas, int(rawTX.GasPrice.Int64()), rawTX.Nonce)
}

func (client *Client) SendRawTransaction(rawTX string) (string, error) {
	hash, err := client.rpc.EthSendRawTransaction(rawTX)
	if err != nil {
		client.log.Errorf("SendRawTransaction:rpc.EthSendRawTransaction: %s", err.Error())
		return hash, err
	}
	return hash, err
}

func (client *Client) GetAddressBalance(address string) (big.Int, error) {
	balance, err := client.rpc.EthGetBalance(address, "pending")
	if err != nil {
		client.log.Errorf("GetAddressBalance:rpc.EthGetBalance: %s", err.Error())
		return balance, err
	}
	return balance, err
}

/*
func resyncETHaddress(address string) {
	// Make a get request
	rs, err := http.Get("https://google.com")
	// Process response
	if err != nil {
		panic(err) // More idiomatic way would be to print the error and die unless it's a serious error
	}
	defer rs.Body.Close()

	bodyBytes, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		panic(err)
	}

	bodyString := string(bodyBytes)
}
*/

func (client *Client) parseETHTransaction(rawTX ethrpc.Transaction, blockHeight int64) {
	fromUser := store.User{}
	toUser := store.User{}

	query := bson.M{"wallets.addresses.address": rawTX.From}
	err := client.db.FindUser(query, &fromUser)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("parseETHTransaction: usersData.Find:", err)
	}

	query = bson.M{"wallets.addresses.address": rawTX.To}
	err = client.db.FindUser(query, &toUser)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("parseETHTransaction: usersData.Find:", err)
	}

	if fromUser.UserID == toUser.UserID && fromUser.UserID == "" {
		return
	}

	tx := rawToMultyETH(rawTX)
	mempool := false
	if blockHeight == -1 {
		tx.TxPoolTime = time.Now().Unix()
		mempool = true
	} else {
		tx.BlockTime = time.Now().Unix()
	}

	// from v1 to v1
	if fromUser.UserID == toUser.UserID && fromUser.UserID != "" {

		tx.UserID = fromUser.UserID

		tx.Status = TxStatusAppearedInBlockOutcoming
		if blockHeight == -1 {
			tx.Status = TxStatusAppearedInMempoolOutcoming
		}

		sel := bson.M{"hash": tx.Hash, "userid": tx.UserID}
		err := client.db.FindETHTransaction(sel)
		if err != nil && err != mgo.ErrNotFound {
			log.Printf("parseETHTransaction: client.db.FindETHTransaction", err)
		}
		if err == mgo.ErrNotFound {
			// insert
			err := client.db.AddEthereumTransaction(tx)
			if err != nil {
				log.Printf("parseETHTransaction: AddEthereumTransaction", err)
			}
		}

		//update
		update := isMempoolUpdate(mempool, tx.Status)

		err = client.db.UpdateEthereumTransaction(sel, update)
		if err != nil {
			log.Printf("parseETHTransaction: AddEthereumTransaction", err)
		}
		return
	}

	// from v1 to v2 outgoing
	if fromUser.UserID != "" {
		tx.UserID = fromUser.UserID
		tx.Status = TxStatusAppearedInBlockOutcoming
		if blockHeight == -1 {
			tx.Status = TxStatusAppearedInMempoolOutcoming
		}

		sel := bson.M{"hash": tx.Hash, "userid": tx.UserID}
		err := client.db.FindETHTransaction(sel)
		if err != nil && err != mgo.ErrNotFound {
			log.Printf("parseETHTransaction: client.db.FindETHTransaction", err)
		}
		if err == mgo.ErrNotFound {
			// insert
			err := client.db.AddEthereumTransaction(tx)
			if err != nil {
				log.Printf("parseETHTransaction: AddEthereumTransaction", err)
			}
		} else {
			//update
			update := isMempoolUpdate(mempool, tx.Status)
			err := client.db.UpdateEthereumTransaction(sel, update)
			if err != nil {
				log.Printf("parseETHTransaction: AddEthereumTransaction", err)
			}
		}
	}

	// from v1 to v2 incoming
	if toUser.UserID != "" {
		tx.UserID = toUser.UserID
		tx.Status = TxStatusAppearedInBlockIncoming
		if blockHeight == -1 {
			tx.Status = TxStatusAppearedInMempoolIncoming
		}

		sel := bson.M{"hash": tx.Hash, "userid": tx.UserID}
		err := client.db.FindETHTransaction(sel)
		if err != nil && err != mgo.ErrNotFound {
			log.Printf("parseETHTransaction: client.db.FindETHTransaction", err)
		}
		if err == mgo.ErrNotFound {
			// insert
			err := client.db.AddEthereumTransaction(tx)
			if err != nil {
				log.Printf("parseETHTransaction: AddEthereumTransaction", err)
			}

		} else {
			// update
			update := isMempoolUpdate(mempool, tx.Status)
			err := client.db.UpdateEthereumTransaction(sel, update)
			if err != nil {
				log.Printf("parseETHTransaction: AddEthereumTransaction", err)
			}
		}

	}

}

func isMempoolUpdate(mempool bool, status int) bson.M {
	if mempool {
		return bson.M{
			"$set": bson.M{
				"status": status,
			},
		}
	}
	return bson.M{
		"$set": bson.M{
			"status":    status,
			"blocktime": time.Now().Unix(),
		},
	}
}
