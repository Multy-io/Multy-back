package btc

import (
	"log"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

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

func parseRawTransaction(inTx *btcjson.TxRawResult) error {

	memPoolTx := MultyMempoolTx{size: inTx.Size, hash: inTx.Hash, txid: inTx.Txid}

	inputs := inTx.Vin

	var inputSum float64
	var outputSum float64

	for j := 0; j < len(inputs); j++ {
		input := inputs[j]

		inputNum := input.Vout

		txCHash, errCHash := chainhash.NewHashFromStr(input.Txid)

		if errCHash != nil {
			log.Fatal(errCHash)
		}

		previousTx, err := rpcClient.GetRawTransactionVerbose(txCHash)
		if err != nil {
			log.Println("ERR GetRawTransactionVerbose [previous]: ", err.Error())
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

	rec := newRecord(int(memPoolTx.feeRate), memPoolTx.hash)

	err := mempoolRates.Insert(rec)
	if err != nil {
		log.Println("ERR mempoolRates.Insert: ", err.Error())
		return err
	}

	memPool = append(memPool, memPoolTx)

	log.Printf("New Multy MemPool Size is: %d", len(memPool))
	return nil
}
