package btc

import (
	"fmt"
	"log"
	"time"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"gopkg.in/mgo.v2/bson"
)

func getAndParseNewBlock(hash *chainhash.Hash) {
	log.Printf("[DEBUG] getNewBlock()")
	blockMSG, err := rpcClient.GetBlock(hash)
	if err != nil {
		log.Println("[ERR] getAndParseNewBlock: ", err.Error())
	}

	// tx speed remover on block
	BlockTxHashes, err := blockMSG.TxHashes() // txHashes of all block tx's
	if err != nil {
		fmt.Printf("[ERR] getAndParseNewBlock(): TxHashes: %s\n", err.Error())
	}

	log.Println("[DEBUG] hash iteration logic transactions")

	var (
		user      *store.User
		txHashStr string
	)

	// txHash is concrete block transaction hash
	for _, txHash := range BlockTxHashes {
		txHashStr = txHash.String()

		blockTx, err := parseBlockTransaction(&txHash)
		if err != nil {
			log.Println("[ERR] parseBlockTransaction:  ", err.Error())
		}

		for addr, _ := range blockTx.Outputs {

			log.Printf("[DEBUG] addr=[%s]\n", addr)
			if err := usersData.Find(bson.M{"wallets.addresses.address": addr}).One(user); err != nil {
				log.Printf("[ERR] getAndParseNewBlock: usersData.Find: %s\n", err.Error())
				continue
			}
			log.Printf("[DEBUG] user=%+v", user)
			log.Printf("[DEBUG] getAndParseNewBlock: Find hashStr=%s/user.UserID=%s\n", txHashStr, user.UserID)
			// check this out

			// !notify users that their transactions was applied in a block

			for _, wallet := range user.Wallets {
				for _, addr := range wallet.Adresses {
					if output, ok := blockTx.Outputs[addr.Address]; !ok {
						continue
					} else {

						// got output with our address; notify user about it
						log.Println("[DEBUG] getAndParseNewBlock: address=", addr)
						addUserTransactionsToDB(user.UserID, output)
						chToClient <- CreateBtcTransactionWithUserID(addr.Address, user.UserID, txOut+" block", txHashStr, output.Amount)
					}
				}
			}
		}
		log.Println("[DEBUG] hash iteration logic transactions done")

		for _, tx := range blockMSG.Transactions {
			for index, memTx := range memPool {
				if memTx.hash == tx.TxHash().String() {
					//TODO remove transaction from mempool
					//TODO update fee rates table
					//TODO check if tx is of our client
					//TODO is so -> notify client
					memPool = append(memPool[:index], memPool[index+1:]...)
				}
			}
		}
	}
	log.Println("[DEBUG] getNewBlock() done")
}

func parseBlockTransaction(txHash *chainhash.Hash) (*store.BTCTransaction, error) {
	log.Printf("[DEBUG] parseBlockTransaction()")
	currentRaw, err := rpcClient.GetRawTransactionVerbose(txHash)
	if err != nil {
		fmt.Printf("[ERR] parseBlockTransaction: %s\n", err.Error())
		return nil, err
	}

	blockTx := serealizeBTCTransaction(currentRaw)
	log.Printf("[DEBUG] done parseBlockTransaction: %+v\n", blockTx.Outputs)

	return blockTx, nil
}

func serealizeBTCTransaction(currentRaw *btcjson.TxRawResult) *store.BTCTransaction {
	blockTx := store.BTCTransaction{
		Hash: currentRaw.Hash,
		Txid: currentRaw.Txid,
		Time: time.Now(),
	}

	outputsRaw := currentRaw.Vout

	allOutputs := make(map[string]*store.BtcOutput, 0)

	for _, output := range outputsRaw {
		addressesOuts := output.ScriptPubKey.Addresses
		if len(addressesOuts) == 0 {
			log.Println("[WARN] serealizeBTCTransaction: len(addressesOuts)==0")
			continue
		}
		allOutputs[addressesOuts[0]] = &store.BtcOutput{
			Address:     addressesOuts[0],
			Amount:      output.Value,
			TxIndex:     output.N,
			TxOutScript: output.ScriptPubKey.Hex,
		}
		log.Printf("[DEBUG] allOutputs[addressesOuts[0]]=%+v\n", allOutputs[addressesOuts[0]])
	}
	blockTx.Outputs = allOutputs
	log.Printf("[DEBUG] blocktx=%+v\n", blockTx)

	return &blockTx
}

func addUserTransactionsToDB(userID string, output *store.BtcOutput) {
	log.Print("[DEBUG] addUserTransactionsToDB")

	sel := bson.M{"userID": userID}
	update := bson.M{"$push": bson.M{"transactions": *output}}
	err := usersData.Update(sel, update)
	if err != nil {
		fmt.Printf("[ERR] push transaction to db: %s\n", err.Error())
	}
	log.Print("[DEBUG] addUserTransactionsToDB done")
}
