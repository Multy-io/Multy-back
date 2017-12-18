package btc

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

const (
	btcToSatoshi = 100000000
)

func getAllMempool() {
	mempool, err := rpcClient.GetRawMempoolVerbose()
	if err != nil {
		log.Printf("[ERR] getAllMempool: rpcClient.GetRawMempoolVerbose: %s \n", err.Error())
	}
	for hash, txInfo := range mempool {
		rec := newRecord(int(txInfo.Fee/float64(txInfo.Size))*btcToSatoshi, hash)
		err = mempoolRates.Insert(rec)
		if err != nil {
			log.Println("[ERR] getAllMempool: mempoolRates.Insert: ", err.Error())
			continue
		}
	}
	count, err := mempoolRates.Count()
	if err != nil {
		log.Println("[ERR] getAllMempool: mempoolRates.Count: ", err.Error())
		return
	}
	fmt.Println("Total mempool size is ", count)
}

func newTxToDB(hash string) {
	txHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		log.Println("[ERR] newTxToDB: chainhash.NewHashFromStr : ", err.Error())
		return
	}
	tx, err := rpcClient.GetRawTransaction(txHash)
	if err != nil {
		log.Println("[ERR] newTxToDB: rpcClient.GetTransaction : ", err.Error())
		return
	}

	var inputSum float64
	var outputSum float64

	for _, output := range tx.MsgTx().TxOut {
		outputSum += float64(output.Value)
	}
	for _, input := range tx.MsgTx().TxIn {
		previousTx, err := rpcClient.GetRawTransactionVerbose(&input.PreviousOutPoint.Hash)
		if err != nil {
			log.Println("[ERR] newTxToDB: rpcClient.GetRawTransactionVerbose : ", err.Error())
			return
		}
		inputSum += previousTx.Vout[input.PreviousOutPoint.Index].Value
	}

	fee := inputSum - outputSum

	rec := newRecord(int(fee/float64(tx.MsgTx().SerializeSize()))*btcToSatoshi, tx.Hash().String())
	err = mempoolRates.Insert(rec)
	if err != nil {
		log.Println("[ERR] newTxToDB: mempoolRates.Insert: ", err.Error())
	}
}

/*
func getAllMempool() {
	rawMempoolTxs, err := rpcClient.GetRawMempool()
	if err != nil {
		log.Printf("[ERR] getAllMempool: rpcClient.GetRawMempool: %s ", err.Error())
	}
	log.Printf("[DEBUG] len mempool = %d \n", len(rawMempoolTxs))
	for _, txHash := range rawMempoolTxs {
		getRawTx(txHash)
	}
}

func getRawTx(hash *chainhash.Hash) {
	rawTx, err := rpcClient.GetRawTransactionVerbose(hash)
	if err != nil {
		log.Println("[ERR] getRawTx:GetRawTransactionVerbose: ", err.Error())
		return
	}
	parseRawTransaction(rawTx)
}

func parseRawTransaction(inTx *btcjson.TxRawResult) {

	memPoolTx := MultyMempoolTx{size: inTx.Size, hash: inTx.Hash, txid: inTx.Txid}

	inputs := inTx.Vin

	var inputSum float64
	var outputSum float64

	for j := 0; j < len(inputs); j++ {
		input := inputs[j]

		inputNum := input.Vout

		txCHash, err := chainhash.NewHashFromStr(input.Txid)
		if err != nil {
			log.Println("[ERR] chainhash.NewHashFromStr : ", err.Error())
		}

		previousTx, err := rpcClient.GetRawTransactionVerbose(txCHash)
		if err != nil {
			log.Println("[ERR] GetRawTransactionVerbose [previous]: ", err.Error())
			continue
		}

		previousOut := previousTx.Vout

		inputSum += previousOut[inputNum].Value

		addressesInputs := previousOut[inputNum].ScriptPubKey.Addresses

		inputAdr := MultyAddress{addressesInputs, previousOut[inputNum].Value}

		memPoolTx.inputs = append(memPoolTx.inputs, inputAdr)
	}

	outputs := inTx.Vout

	var txOutputs []MultyAddress

	for _, output := range outputs {
		addressesOuts := output.ScriptPubKey.Addresses
		outputSum += output.Value

		txOutputs = append(txOutputs, MultyAddress{addressesOuts, output.Value})
	}
	memPoolTx.outputs = txOutputs

	memPoolTx.amount = inputSum
	memPoolTx.fee = inputSum - outputSum

	memPoolTx.feeRate = int32(memPoolTx.fee / float64(memPoolTx.size) * 100000000)

	rec := newRecord(int(memPoolTx.feeRate), memPoolTx.txid)

	err := mempoolRates.Insert(rec)
	if err != nil {
		log.Println("[ERR] mempoolRates.Insert: ", err.Error())
		return
	}

	memPool = append(memPool, memPoolTx)

	log.Printf("New Multy MemPool Size is: %d [txid] - %s ", len(memPool), memPoolTx.txid)

}
*/
