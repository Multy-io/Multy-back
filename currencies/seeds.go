/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package currencies

const (
	MultySeed    = 0
	MetamaskSeed = 1
)

var SeedPhraseTypes = map[int]bool{
	MultySeed:    true, // multy seed
	MetamaskSeed: true, // metamask seed
}
