package store

import "fmt"

func FetchCoinType(coinTypes []CoinType, currencyID, networkID int) (CoinType, error) {
	for _, ct := range coinTypes {
		if ct.СurrencyID == currencyID && ct.NetworkID == networkID {
			return ct, nil
		}
	}
	return CoinType{}, fmt.Errorf("fetchCoinType: no such coin in config")
}
