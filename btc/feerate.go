package btc

import (
	"github.com/Appscrunch/Multy-back/store"
)

const (
	btcToSatoshi = 100000000
)

func (c *Client) GetAllMempool() ([]store.MempoolRecord, error) {
	allMempool := []store.MempoolRecord{}
	mempool, err := c.RpcClient.GetRawMempoolVerbose()
	if err != nil {
		return allMempool, err
	}
	log.Errorf("MEMPOOL SIZE == %v", len(mempool))
	for hash, txInfo := range mempool {
		allMempool = append(allMempool, newMempoolRecord(int(txInfo.Fee/float64(txInfo.Size)*btcToSatoshi), hash))
	}
	return allMempool, err

}
