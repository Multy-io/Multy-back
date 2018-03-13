/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

import "github.com/Appscrunch/Multy-back/store"

func newRecord(category int, hashTX string) store.MempoolRecord {
	return store.MempoolRecord{
		Category: category,
		HashTX:   hashTX,
	}
}
