package store

import (
	"testing"

	"github.com/Multy-io/Multy-back/currencies"
)

func TestSum(t *testing.T) {
	mainURL := "main"
	testURL := "test"
	cts := []CoinType{
		CoinType{
			СurrencyID: currencies.Ether,
			NetworkID:  currencies.ETHMain,
			GRPCUrl:    mainURL,
		},
		CoinType{
			СurrencyID: currencies.Ether,
			NetworkID:  currencies.ETHTest,
			GRPCUrl:    testURL,
		},
	}
	ct, err := FetchCoinType(cts, currencies.Ether, currencies.ETHMain)
	if err != nil {
		t.Errorf("FetchCoinType error while fetching err: %v ", err.Error())
	}
	if ct.GRPCUrl != mainURL {
		t.Errorf("Fetch was incorrect, got url: %v, want: %v.", ct.GRPCUrl, mainURL)
	}
}
