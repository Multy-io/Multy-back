package client

import (
	"sync"
	"testing"

	"github.com/Multy-io/Multy-back/store"
)

func TestFetchMempool(t *testing.T) {
	// less than quantile map len case
	var q int = 3
	emptyMap := sync.Map{}
	VeryFastRate := fetchMempool(emptyMap, q)
	if VeryFastRate != store.ETHStandardVeryFastFeeRate {
		t.Errorf("TestFetchMempool: bad fetch for empty map VeryFastRate != ETHStandardVeryFastFeeRate %v", VeryFastRate)
	}

	filledMap := sync.Map{}
	var gasPrice int64 = 1
	filledMap.Store("h1", gasPrice)
	filledMap.Store("h2", gasPrice)
	filledMap.Store("h3", gasPrice)
	filledMap.Store("h4", gasPrice)
	VeryFastRate = fetchMempool(filledMap, q)
	if VeryFastRate != gasPrice {
		t.Errorf("TestFetchMempool: bad fetch for filled map VeryFastRate != gasPrice %v", VeryFastRate)
	}
}
