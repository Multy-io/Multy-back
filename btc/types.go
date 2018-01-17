/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package btc

func newRecord(category int, hashTX string) Record {
	return Record{
		Category: category,
		HashTX:   hashTX,
	}
}

type Record struct {
	Category int    `json:"category"`
	HashTX   string `json:"hashTX"`
}
