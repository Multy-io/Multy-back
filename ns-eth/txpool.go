/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package nseth

import (
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Multy-io/Multy-back/store"
	"github.com/Multy-io/Multy-back/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/jekabolt/slf"
)

const gWei = 1000000000

// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockHash        common.Hash     `json:"blockHash"`
	BlockNumber      *hexutil.Big    `json:"blockNumber"`
	From             common.Address  `json:"from"`
	Gas              hexutil.Uint64  `json:"gas"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Hash             common.Hash     `json:"hash"`
	Input            hexutil.Bytes   `json:"input"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	To               *common.Address `json:"to"`
	TransactionIndex hexutil.Uint    `json:"transactionIndex"`
	Value            *hexutil.Big    `json:"value"`
	V                *hexutil.Big    `json:"v"`
	R                *hexutil.Big    `json:"r"`
	S                *hexutil.Big    `json:"s"`
}

func (c *Client) AddTransactionToTxpool(txHash string) {
	rawTx, err := c.Rpc.EthGetTransactionByHash(txHash)
	if err != nil {
		log.Errorf("c.Rpc.EthGetTransactionByHash:Get TX Err: %s", err.Error())
	}
	c.parseETHTransaction(*rawTx, -1, false)

	c.parseETHMultisig(*rawTx, -1, false)

	// add txpool record
	if rawTx.GasPrice.IsUint64() {
		c.Mempool.Store(rawTx.Hash, rawTx.GasPrice.Uint64())
	}

	if strings.ToLower(rawTx.To) == strings.ToLower(c.Multisig.FactoryAddress) {
		go func() {
			fi, err := parseFactoryInput(rawTx.Input)
			if err != nil {
				log.Errorf("txpoolTransaction:parseFactoryInput: %s", err.Error())
			}
			fi.TxOfCreation = txHash
			fi.FactoryAddress = c.Multisig.FactoryAddress
			fi.DeployStatus = int64(store.MultisigStatusDeployPending)
			c.NewMultisigStream <- *fi
		}()
	}

}

func (c *Client) DeleteTxpoolTransaction(txHash string) {
	c.Mempool.Delete(txHash)
}

func (c *Client) ReloadTxPool() error {

	mp, err := c.GetAllTxPool()
	if err != nil {
		return err
	}

	// We get responыe as sample down https://github.com/ethereum/go-ethereum/wiki/Management-APIs#example-14 and convert response to map[string]map[string]map[string]*RPCTransaction
	/***
	"pending": {
		"0x00000000C0293c8cA34Dac9BCC0F953532D34e4d": { //address
			"685615": { // nonce
				... // transaction body, see RPCTransaction for details
			}
		}
	},
	"queued": {
		// same as for "pending"
	}
	***/
	var mempoolTx map[string]map[string]map[string]*RPCTransaction
	err = json.Unmarshal(mp, &mempoolTx)
	if err != nil {
		log.Errorf("can'not unmarshal err: %v", err)
	}
	var mempool *sync.Map = &sync.Map{}
	var length uint64 = 0
	// For each address, find a transaction with smaller nonce, filtering out txs with nonce ridiculously large
	for _, addrTxs := range mempoolTx["pending"] {
		var nonce_min uint64 = math.MaxUint64
		var nonce_min_str string = ""
		for nonce := range addrTxs {
			nonce_uint, err := strconv.ParseUint(nonce, 10, 64)
			if err != nil {
				log.Errorf("Impossible to convert nonce to int. Expected value: %v, Error: %v", nonce, err)
			}
			if nonce_min > nonce_uint {
				nonce_min = nonce_uint
				nonce_min_str = nonce
			}
		}
		var gasPrice uint64 = addrTxs[nonce_min_str].GasPrice.ToInt().Uint64()
		var hash string = addrTxs[nonce_min_str].Hash.String()
		mempool.Store(hash, gasPrice)
		length++
	}
	log.Debugf("length adderss with tx : %v", length)
	length = 0
	mempool.Range(func(_, _ interface{}) bool {
		length++
		return true
	})

	log.Debugf("load mempool. len : %v", length)
	c.Mempool = mempool
	return nil
}

func (c *Client) EstimateTransactionGasPrice() types.TransactionFeeRateEstimation {
	return estimateTransactionGasPriceFromTxpool(c.Mempool, uint64(gWei))
}

func estimateTransactionGasPriceFromTxpool(mempool *sync.Map, minReturnValue uint64) types.TransactionFeeRateEstimation {
	var fees []uint64
	mempool.Range(func(_, v interface{}) bool {
		amount, err := v.(uint64)
		if err {
			fees = append(fees, amount)
		} else {
			log.Errorf("estimateTransactionGasPriceFromTxpool can not convert gasPrice to uint: %v", err)
		}
		return true
	})
	sort.Slice(fees, func(i, j int) bool { return fees[i] > fees[j] })

	slf.WithContext("estimateTransactionGasPrice").WithCaller(slf.CallerShort).Debugf("ETH feerate:mempool size: = %d", len(fees))

	// if mempool tx size > 1300, use sorted first 1300 mempool Transaction for estimate gas price
	if len(fees) > 1300 {
		fees = fees[:1300]
	}

	// Estimate ferate if mempool size  more 500
	if len(fees) > 500 {
		var firstPack int = len(fees) / 10
		var step int = (len(fees) - firstPack) / 4
		return types.TransactionFeeRateEstimation{
			VeryFast: max(average(fees[:firstPack]), minReturnValue),
			Fast:     max(average(fees[firstPack:(firstPack+1*step)]), minReturnValue),
			Medium:   max(average(fees[(firstPack+1*step):(firstPack+2*step)]), minReturnValue),
			Slow:     max(average(fees[(firstPack+2*step):(firstPack+3*step)]), minReturnValue),
			VerySlow: max(average(fees[(firstPack+3*step):]), minReturnValue),
		}
	}
	return types.TransactionFeeRateEstimation{
		VerySlow: 9 * gWei,
		Slow:     10 * gWei,
		Medium:   14 * gWei,
		Fast:     20 * gWei,
		VeryFast: 25 * gWei,
	}
}

func average(fees []uint64) uint64 {
	var total uint64 = 0
	for _, value := range fees {
		total += value
	}
	return uint64(total / uint64(len(fees)))
}

func max(x, y uint64) uint64 {
	if x < y {
		return y
	}
	return x
}
