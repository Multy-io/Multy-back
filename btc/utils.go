/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import (
	"time"

	"math"

	pb "github.com/Appscrunch/Multy-back/node-streamer/btc"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

const ( // currency id  nsq
	TxStatusAppearedInMempoolIncoming = 1
	TxStatusAppearedInBlockIncoming   = 2

	TxStatusAppearedInMempoolOutcoming = 3
	TxStatusAppearedInBlockOutcoming   = 4

	TxStatusInBlockConfirmedIncoming  = 5
	TxStatusInBlockConfirmedOutcoming = 6

	SatoshiInBitcoint = 100000000
)

var SatoshiToBitcoin = float64(100000000)

func newAddresAmount(address string, amount int64) store.AddresAmount {
	return store.AddresAmount{
		Address: address,
		Amount:  amount,
	}
}

func (c *Client) rawTxByTxid(txid string) (*btcjson.TxRawResult, error) {
	hash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return nil, err
	}
	previousTxVerbose, err := c.RpcClient.GetRawTransactionVerbose(hash)
	if err != nil {
		return nil, err
	}
	return previousTxVerbose, nil
}

// setTransactionInfo set fee, inputs and outputs
func (c *Client) setTransactionInfo(multyTx *store.MultyTX, txVerbose *btcjson.TxRawResult, blockHeight int64, isReSync bool) error {
	inputs := []store.AddresAmount{}
	outputs := []store.AddresAmount{}
	var inputSum float64
	var outputSum float64

	for _, out := range txVerbose.Vout {
		for _, address := range out.ScriptPubKey.Addresses {
			amount := int64(out.Value * SatoshiToBitcoin)
			outputs = append(outputs, newAddresAmount(address, amount))
		}
		outputSum += out.Value
	}
	for _, input := range txVerbose.Vin {
		hash, err := chainhash.NewHashFromStr(input.Txid)
		if err != nil {
			log.Errorf("setTransactionInfo:chainhash.NewHashFromStr: %s", err.Error())
		}
		previousTxVerbose, err := c.RpcClient.GetRawTransactionVerbose(hash)
		if err != nil {
			log.Errorf("setTransactionInfo:RpcClient.GetRawTransactionVerbose: %s", err.Error())
		}

		for _, address := range previousTxVerbose.Vout[input.Vout].ScriptPubKey.Addresses {
			amount := int64(previousTxVerbose.Vout[input.Vout].Value * SatoshiToBitcoin)
			inputs = append(inputs, newAddresAmount(address, amount))
		}
		inputSum += previousTxVerbose.Vout[input.Vout].Value
	}
	fee := int64((inputSum - outputSum) * SatoshiToBitcoin)

	if blockHeight == -1 || isReSync {
		multyTx.MempoolTime = txVerbose.Time
	}

	if blockHeight != -1 {
		multyTx.BlockTime = txVerbose.Blocktime
	}
	multyTx.TxInputs = inputs
	multyTx.TxOutputs = outputs
	multyTx.TxFee = fee

	return nil
}

/*

Main process BTC transaction method

can be called from:
	- Mempool
	- New block
	- Re-sync

*/

// HACK is a wrapper for processTransaction. in future it will be in separate file
// func ProcessTransaction(blockChainBlockHeight int64, txVerbose *btcjson.TxRawResult, isReSync bool) {
// 	processTransaction(blockChainBlockHeight, txVerbose, isReSync)
// }

// func GetRawTransactionVerbose(txHash *chainhash.Hash) (*btcjson.TxRawResult, error) {
// 	return RpcClient.GetRawTransactionVerbose(txHash)
// }

// func GetBlockHeight() (int64, error) {
// 	return RpcClient.GetBlockCount()
// }

// func ProcessTransaction(blockChainBlockHeight int64, txVerbose *btcjson.TxRawResult, isReSync bool, usersData *map[string]string) {
// 	processTransaction(blockChainBlockHeight, txVerbose, isReSync, usersData)
// }

func (c *Client) ProcessTransaction(blockChainBlockHeight int64, txVerbose *btcjson.TxRawResult, isReSync bool) {
	multyTx, related := c.ParseRawTransaction(blockChainBlockHeight, txVerbose)
	if related {
		log.Debugf("ProcessTransaction...")
		c.CreateSpendableOutputs(txVerbose, blockChainBlockHeight)
		c.DeleteSpendableOutputs(txVerbose, blockChainBlockHeight)
	}

	if multyTx != nil {
		multyTx.BlockHeight = blockChainBlockHeight
		log.Debugf("ProcessTransaction... on blockHeight %d", blockChainBlockHeight)

		c.setTransactionInfo(multyTx, txVerbose, blockChainBlockHeight, isReSync)

		transactions := c.splitTransaction(*multyTx, blockChainBlockHeight)

		for _, transaction := range transactions {
			finalizeTransaction(&transaction, txVerbose)
			saveMultyTransaction(transaction, isReSync, c.TransactionsCh)
		}
	}
}

/*
This method should parse raw transaction from BTC node

_________________________
Inputs:
* blockChainBlockHeight int64 - could be:
-1 in case of mempool call
-1 in case of block transaction
max chain height in case of resync

*txVerbose - raw BTC transaction
_________________________
Output:
* multyTX - multy transaction Structure

*/
func (c *Client) ParseRawTransaction(blockChainBlockHeight int64, txVerbose *btcjson.TxRawResult) (*store.MultyTX, bool) {
	multyTx := store.MultyTX{}

	err := c.parseInputs(txVerbose, blockChainBlockHeight, &multyTx)
	if err != nil {
		log.Errorf("ParseRawTransaction:parseInputs: %s", err.Error())
	}

	err = c.parseOutputs(txVerbose, blockChainBlockHeight, &multyTx)
	if err != nil {
		log.Errorf("ParseRawTransaction:parseOutputs: %s", err.Error())
	}

	if multyTx.TxID != "" {
		return &multyTx, true
	} else {
		return nil, false
	}
}

/*
This method need if we have one transaction with more the one u wser'sallet
That means that from one btc transaction we should build more the one Multy Transaction
*/
func (c *Client) splitTransaction(multyTx store.MultyTX, blockHeight int64) []store.MultyTX {

	transactions := []store.MultyTX{}

	currentBlockHeight, err := c.RpcClient.GetBlockCount()
	if err != nil {
		log.Errorf("splitTransaction:getBlockCount: %s", err.Error())
	}

	blockDiff := currentBlockHeight - blockHeight

	//This is implementatios for single wallet transaction for multi addresses not for multi wallets!
	if multyTx.WalletsInput != nil && len(multyTx.WalletsInput) > 0 {
		outgoingTx := newEntity(multyTx)
		//By that code we are erasing WalletOutputs for new splited transaction
		outgoingTx.WalletsOutput = []store.WalletForTx{}

		for _, walletOutput := range multyTx.WalletsOutput {
			var isTheSameWallet = false
			for _, walletInput := range outgoingTx.WalletsInput {
				if walletInput.UserId == walletOutput.UserId && walletInput.WalletIndex == walletOutput.WalletIndex {
					isTheSameWallet = true
				}
			}
			if isTheSameWallet {
				outgoingTx.WalletsOutput = append(outgoingTx.WalletsOutput, walletOutput)
			}
		}

		setTransactionStatus(&outgoingTx, blockDiff, currentBlockHeight, true)
		transactions = append(transactions, outgoingTx)
	}

	if multyTx.WalletsOutput != nil && len(multyTx.WalletsOutput) > 0 {
		for _, walletOutput := range multyTx.WalletsOutput {
			var alreadyAdded = false
			for i := 0; i < len(transactions); i++ {
				//Check if our output wallet is in the inputs
				if transactions[i].WalletsInput != nil && len(transactions[i].WalletsInput) > 0 {
					for _, splitedInput := range transactions[i].WalletsInput {
						if splitedInput.UserId == walletOutput.UserId && splitedInput.WalletIndex == walletOutput.WalletIndex {
							alreadyAdded = true
						}
					}
				}

				if transactions[i].WalletsOutput != nil && len(transactions[i].WalletsOutput) > 0 {
					for j := 0; j < len(transactions[i].WalletsOutput); j++ {
						if walletOutput.UserId == transactions[i].WalletsOutput[j].UserId && walletOutput.WalletIndex == transactions[i].WalletsOutput[j].WalletIndex { //&& walletOutput.Address.Address != transactions[i].WalletsOutput[j].Address.Address Don't think this ckeck we need
							//We have the same wallet index in output but different addres
							alreadyAdded = true
							if &transactions[i] == nil {
								transactions[i].WalletsOutput = append(transactions[i].WalletsOutput, walletOutput)
								log.Errorf("splitTransaction error allocate memory\n")
							}
						}
					}
				}
			}

			if alreadyAdded {
				continue
			} else {
				//Add output transaction here
				incomingTx := newEntity(multyTx)
				incomingTx.WalletsInput = nil
				incomingTx.WalletsOutput = []store.WalletForTx{}
				incomingTx.WalletsOutput = append(incomingTx.WalletsOutput, walletOutput)
				setTransactionStatus(&incomingTx, blockDiff, currentBlockHeight, false)
				transactions = append(transactions, incomingTx)
			}
		}

	}

	return transactions
}

func newEntity(multyTx store.MultyTX) store.MultyTX {
	newTx := store.MultyTX{
		TxID:              multyTx.TxID,
		TxHash:            multyTx.TxHash,
		TxOutScript:       multyTx.TxOutScript,
		TxAddress:         multyTx.TxAddress,
		TxStatus:          multyTx.TxStatus,
		TxOutAmount:       multyTx.TxOutAmount,
		BlockTime:         multyTx.BlockTime,
		BlockHeight:       multyTx.BlockHeight,
		TxFee:             multyTx.TxFee,
		MempoolTime:       multyTx.MempoolTime,
		StockExchangeRate: multyTx.StockExchangeRate,
		TxInputs:          multyTx.TxInputs,
		TxOutputs:         multyTx.TxOutputs,
		WalletsInput:      multyTx.WalletsInput,
		WalletsOutput:     multyTx.WalletsOutput,
	}
	return newTx
}

func saveMultyTransaction(tx store.MultyTX, resync bool, TransactionsCh chan pb.BTCTransaction) {
	// This is splited transaction! That means that transaction's WalletsInputs and WalletsOutput have the same WalletIndex!
	//Here we have outgoing transaction for exact wallet!
	if tx.WalletsInput != nil && len(tx.WalletsInput) > 0 {
		if len(tx.WalletsInput) > 0 && len(tx.WalletsOutput) > 0 {
			var amount int64
			if len(tx.TxAddress) != 0 {
				for _, in := range tx.TxInputs {
					if in.Address == tx.TxAddress[0] {
						amount += in.Amount
					}
				}
			}
			if amount == 0 {
				for _, out := range tx.TxOutputs {
					if out.Address == tx.TxAddress[0] {
						amount += out.Amount
					}
				}
			}
			tx.TxOutAmount = amount

			// if tx.TxOutScript == "" {
			// 	if tx.TxStatus == store.TxStatusAppearedInMempoolIncoming {
			// 		tx.TxStatus = store.TxStatusAppearedInMempoolOutcoming
			// 	}

			// 	if tx.TxStatus == store.TxStatusAppearedInBlockIncoming {
			// 		tx.TxStatus = store.TxStatusAppearedInBlockOutcoming
			// 	}

			// 	if tx.TxStatus == store.TxStatusInBlockConfirmedIncoming {
			// 		tx.TxStatus = store.TxStatusInBlockConfirmedOutcoming
			// 	}
			// }
		}

		//HACK: fetching userid like this
		for _, input := range tx.WalletsInput {
			if input.UserId != "" {
				tx.UserId = input.UserId
				break
			}
		}

		outcomingTx := storeTxToGenerated(tx)
		// send to outcomingTx to chan
		TransactionsCh <- outcomingTx

		return
	} else if tx.WalletsOutput != nil && len(tx.WalletsOutput) > 0 {
		if len(tx.WalletsInput) > 0 && len(tx.WalletsOutput) > 0 {
			var amount int64
			if len(tx.TxAddress) != 0 {
				for _, in := range tx.TxInputs {
					if in.Address == tx.TxAddress[0] {
						amount += in.Amount
					}
				}
			}
			if amount == 0 {
				for _, out := range tx.TxOutputs {
					if out.Address == tx.TxAddress[0] {
						amount += out.Amount
					}
				}
			}

			tx.TxOutAmount = amount

			// if tx.TxOutScript == "" {
			// 	if tx.TxStatus == store.TxStatusAppearedInMempoolIncoming {
			// 		tx.TxStatus = store.TxStatusAppearedInMempoolOutcoming
			// 	}

			// 	if tx.TxStatus == store.TxStatusAppearedInBlockIncoming {
			// 		tx.TxStatus = store.TxStatusAppearedInBlockOutcoming
			// 	}

			// 	if tx.TxStatus == store.TxStatusInBlockConfirmedIncoming {
			// 		tx.TxStatus = store.TxStatusInBlockConfirmedOutcoming
			// 	}
			// }
		}

		//HACK: fetching userid like this
		for _, output := range tx.WalletsOutput {
			if output.UserId != "" {
				tx.UserId = output.UserId
				break
			}
		}
		// fmt.Println("[DEBUG] newIncomingTx\n")

		incomingTx := storeTxToGenerated(tx)
		incomingTx.Resync = resync
		// send to incomingTx to chan
		TransactionsCh <- incomingTx

		return
	}
}

func storeTxToGenerated(tx store.MultyTX) pb.BTCTransaction {
	outs := []*pb.BTCTransaction_AddresAmount{}
	for _, output := range tx.TxOutputs {
		outs = append(outs, &pb.BTCTransaction_AddresAmount{
			Address: output.Address,
			Amount:  output.Amount,
		})
	}

	ins := []*pb.BTCTransaction_AddresAmount{}
	for _, inputs := range tx.TxInputs {

		ins = append(ins, &pb.BTCTransaction_AddresAmount{
			Address: inputs.Address,
			Amount:  inputs.Amount,
		})
	}

	walletsOutput := []*pb.BTCTransaction_WalletForTx{}
	for _, wOutput := range tx.WalletsOutput {
		walletsOutput = append(walletsOutput, &pb.BTCTransaction_WalletForTx{
			Userid:     wOutput.UserId,
			Address:    wOutput.Address.Address,
			Amount:     wOutput.Address.Amount,
			TxOutIndex: int32(wOutput.Address.AddressOutIndex),
		})
	}

	walletsInput := []*pb.BTCTransaction_WalletForTx{}
	for _, wInput := range tx.WalletsInput {
		walletsInput = append(walletsInput, &pb.BTCTransaction_WalletForTx{
			Userid:     wInput.UserId,
			Address:    wInput.Address.Address,
			Amount:     wInput.Address.Amount,
			TxOutIndex: int32(wInput.Address.AddressOutIndex),
		})
	}

	return pb.BTCTransaction{
		UserID:        tx.UserId,
		TxID:          tx.TxID,
		TxHash:        tx.TxHash,
		TxOutScript:   tx.TxOutScript,
		TxAddress:     tx.TxAddress,
		TxStatus:      int32(tx.TxStatus),
		TxOutAmount:   tx.TxOutAmount,
		BlockTime:     tx.BlockTime,
		BlockHeight:   tx.BlockHeight,
		Confirmations: int32(tx.Confirmations),
		TxFee:         tx.TxFee,
		MempoolTime:   tx.MempoolTime,
		TxInputs:      ins,
		TxOutputs:     outs,
		WalletsInput:  walletsInput,
		WalletsOutput: walletsOutput,
	}
}

func (c *Client) parseInputs(txVerbose *btcjson.TxRawResult, blockHeight int64, multyTx *store.MultyTX) error {
	//Ranging by inputs
	for _, input := range txVerbose.Vin {

		//getting previous verbose transaction from BTC Node for checking addresses
		previousTxVerbose, err := c.rawTxByTxid(input.Txid)
		if err != nil {
			log.Errorf("parseInputs:rawTxByTxid: %s", err.Error())
			continue
		}

		for _, txInAddress := range previousTxVerbose.Vout[input.Vout].ScriptPubKey.Addresses {
			// check the ownership of the transaction to our users
			c.UserDataM.Lock()
			ud := *c.UsersData
			addressEx, ok := ud[txInAddress]
			c.UserDataM.Unlock()
			if !ok {
				continue
			}

			txInAmount := int64(SatoshiToBitcoin * previousTxVerbose.Vout[input.Vout].Value)

			currentWallet := store.WalletForTx{
				UserId:      addressEx.UserID,
				WalletIndex: addressEx.WalletIndex,
				Address: store.AddressForWallet{
					AddressIndex:    addressEx.AddressIndex,
					Address:         txInAddress,
					Amount:          txInAmount,
					AddressOutIndex: int(input.Vout),
				},
			}

			multyTx.WalletsInput = append(multyTx.WalletsInput, currentWallet)

			multyTx.TxID = txVerbose.Txid
			multyTx.TxHash = txVerbose.Hash

		}

	}

	return nil
}

func (c *Client) parseOutputs(txVerbose *btcjson.TxRawResult, blockHeight int64, multyTx *store.MultyTX) error {

	//Ranging by outputs
	for _, output := range txVerbose.Vout {
		for _, txOutAddress := range output.ScriptPubKey.Addresses {

			c.UserDataM.Lock()
			ud := *c.UsersData
			addressEx, ok := ud[txOutAddress]
			c.UserDataM.Unlock()
			if !ok {
				continue
			}

			currentWallet := store.WalletForTx{
				UserId:      addressEx.UserID,
				WalletIndex: addressEx.WalletIndex,
				Address: store.AddressForWallet{
					AddressIndex:    addressEx.AddressIndex,
					Address:         txOutAddress,
					Amount:          int64(SatoshiToBitcoin * output.Value),
					AddressOutIndex: int(output.N),
				},
			}

			multyTx.WalletsOutput = append(multyTx.WalletsOutput, currentWallet)

			multyTx.TxID = txVerbose.Txid
			multyTx.TxHash = txVerbose.Hash
		}
	}
	return nil
}

func setTransactionStatus(tx *store.MultyTX, blockDiff int64, currentBlockHeight int64, fromInput bool) {
	transactionTime := time.Now().Unix()
	if blockDiff > currentBlockHeight {
		//This call was made from memPool
		tx.Confirmations = 0
		if fromInput {
			tx.TxStatus = TxStatusAppearedInMempoolOutcoming
			tx.MempoolTime = transactionTime
			tx.BlockTime = -1
		} else {
			tx.TxStatus = TxStatusAppearedInMempoolIncoming
			tx.MempoolTime = transactionTime
			tx.BlockTime = -1
		}
	} else if blockDiff >= 0 && blockDiff < 6 {
		//This call was made from block or resync
		//Transaction have no enough confirmations
		tx.Confirmations = int(blockDiff + 1)
		if fromInput {
			tx.TxStatus = TxStatusAppearedInBlockOutcoming
			tx.BlockTime = transactionTime
		} else {
			tx.TxStatus = TxStatusAppearedInBlockIncoming
			tx.BlockTime = transactionTime
		}
	} else if blockDiff >= 6 && blockDiff < currentBlockHeight {
		//This call was made from resync
		//Transaction have enough confirmations
		tx.Confirmations = int(blockDiff + 1)
		if fromInput {
			tx.TxStatus = TxStatusInBlockConfirmedOutcoming
			//TODO add block time for re-sync
		} else {
			tx.TxStatus = TxStatusInBlockConfirmedIncoming
		}
	}
}

func finalizeTransaction(tx *store.MultyTX, txVerbose *btcjson.TxRawResult) {

	if tx.TxAddress == nil {
		tx.TxAddress = []string{}
	}

	if tx.TxStatus == TxStatusInBlockConfirmedOutcoming || tx.TxStatus == TxStatusAppearedInBlockOutcoming || tx.TxStatus == TxStatusAppearedInMempoolOutcoming {
		for _, walletInput := range tx.WalletsInput {
			tx.TxOutAmount += walletInput.Address.Amount
			tx.TxAddress = append(tx.TxAddress, walletInput.Address.Address)
		}

		for i := 0; i < len(tx.WalletsOutput); i++ {
			//Here we descreasing amount of the current transaction
			tx.TxOutAmount -= tx.WalletsOutput[i].Address.Amount
			for _, output := range txVerbose.Vout {
				for _, outAddr := range output.ScriptPubKey.Addresses {
					if tx.WalletsOutput[i].Address.Address == outAddr {
						tx.WalletsOutput[i].Address.AddressOutIndex = int(output.N)
						tx.TxOutScript = txVerbose.Vout[output.N].ScriptPubKey.Hex
					}
				}
			}
		}
	} else {
		for i := 0; i < len(tx.WalletsOutput); i++ {
			tx.TxOutAmount += tx.WalletsOutput[i].Address.Amount
			tx.TxAddress = append(tx.TxAddress, tx.WalletsOutput[i].Address.Address)

			for _, output := range txVerbose.Vout {
				for _, outAddr := range output.ScriptPubKey.Addresses {
					if tx.WalletsOutput[i].Address.Address == outAddr {
						tx.WalletsOutput[i].Address.AddressOutIndex = int(output.N)
						tx.TxOutScript = txVerbose.Vout[output.N].ScriptPubKey.Hex
					}
				}
			}
		}
		//TxOutIndexes we need only for incoming transactions
	}
}

func (c *Client) CreateSpendableOutputs(tx *btcjson.TxRawResult, blockHeight int64) {
	log.Debugf("CreateSpendableOutputs")
	for _, output := range tx.Vout {
		if len(output.ScriptPubKey.Addresses) > 0 {
			address := output.ScriptPubKey.Addresses[0]

			c.UserDataM.Lock()
			ud := *c.UsersData
			addressEx, ok := ud[address]
			c.UserDataM.Unlock()

			if !ok {
				continue
			}

			txStatus := store.TxStatusAppearedInBlockIncoming
			if blockHeight == -1 {
				txStatus = store.TxStatusAppearedInMempoolIncoming
			}

			amount := int64(output.Value * SatoshiToBitcoin)
			spendableOutput := store.SpendableOutputs{
				TxID:         tx.Txid,
				TxOutID:      int(output.N),
				TxOutAmount:  amount,
				TxOutScript:  output.ScriptPubKey.Hex,
				Address:      address,
				UserID:       addressEx.UserID,
				TxStatus:     txStatus,
				WalletIndex:  addressEx.WalletIndex,
				AddressIndex: addressEx.AddressIndex,
			}

			spOut := spOutToGenerated(spendableOutput)
			//send to channel of creation of spendable output
			c.AddSpOut <- spOut

		}
	}
}
func spOutToGenerated(spOut store.SpendableOutputs) pb.AddSpOut {
	return pb.AddSpOut{
		TxID:         spOut.TxID,
		TxOutID:      int32(spOut.TxOutID),
		TxOutAmount:  spOut.TxOutAmount,
		TxOutScript:  spOut.TxOutScript,
		Address:      spOut.Address,
		UserID:       spOut.UserID,
		TxStatus:     int32(spOut.TxStatus),
		WalletIndex:  int32(spOut.WalletIndex),
		AddressIndex: int32(spOut.AddressIndex),
	}
}

func (c *Client) DeleteSpendableOutputs(tx *btcjson.TxRawResult, blockHeight int64) {
	log.Debugf("DeleteSpendableOutputs")
	for _, input := range tx.Vin {
		previousTx, err := c.rawTxByTxid(input.Txid)
		if err != nil {
			log.Errorf("DeleteSpendableOutputs:rawTxByTxid: %s", err.Error())
		}

		if previousTx == nil {
			continue
		}

		if len(previousTx.Vout[input.Vout].ScriptPubKey.Addresses) > 0 {
			address := previousTx.Vout[input.Vout].ScriptPubKey.Addresses[0]

			c.UserDataM.Lock()
			ud := *c.UsersData
			addressEx, ok := ud[address]
			c.UserDataM.Unlock()

			if !ok {
				continue
			}
			reqDelete := store.DeleteSpendableOutput{
				UserID:  addressEx.UserID,
				TxID:    previousTx.Txid,
				Address: address,
			}
			// send to client and removefrom db

			del := delSpOutToGenerated(reqDelete)
			c.DelSpOut <- del

			// fmt.Printf("[DEBUG] deleteSpout %v \n", del.String())

		}
	}
}
func delSpOutToGenerated(del store.DeleteSpendableOutput) pb.ReqDeleteSpOut {
	return pb.ReqDeleteSpOut{
		UserID:  del.UserID,
		TxID:    del.TxID,
		Address: del.Address,
	}
}

func newMempoolRecord(category int, hashTX string) store.MempoolRecord {
	return store.MempoolRecord{
		Category: category,
		HashTX:   hashTX,
	}
}

func (c *Client) rawTxToMempoolRec(inTx *btcjson.TxRawResult) store.MempoolRecord {
	var inputSum float64
	var outputSum float64
	for _, input := range inTx.Vin {
		txCHash, err := chainhash.NewHashFromStr(input.Txid)
		if err != nil {
			log.Errorf("newTxToDB: chainhash.NewHashFromStr: %s", err.Error())
		}

		c.RpcClientM.Lock()
		previousTx, err := c.RpcClient.GetRawTransactionVerbose(txCHash)
		c.RpcClientM.Unlock()
		if err != nil {
			log.Errorf("newTxToDB: rpcClient.GetTransaction: %s", err.Error())
		}
		inputSum += previousTx.Vout[input.Vout].Value
	}
	for _, output := range inTx.Vout {
		outputSum += output.Value
	}
	fee := inputSum - outputSum

	floatFee := fee / float64(inTx.Size) * 100000000

	//It's some kind of Round function to prefent 0 FeeRates while casting from float to int
	intFee := int(math.Floor(floatFee + 0.5))

	rec := newMempoolRecord(intFee, inTx.Hash)

	return rec
}
