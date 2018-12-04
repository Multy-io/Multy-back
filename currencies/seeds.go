/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package currencies

const (
	MultySeed = iota
	MetamaskSeed
)

var seeds = []int{
	MultySeed,
	MetamaskSeed,
}

func IsSeedPhraseTypeValid(seedPhraseType int) bool {
	switch seedPhraseType {
	case MultySeed,
		MetamaskSeed:
		return true
	}
	return false
}
