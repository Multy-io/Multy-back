/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package ethereum

type MultyETHTransaction struct {
	Hash     string  `json:"hash"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	Amount   float64 `json:"amount"`
	Gas      int     `json:"gas"`
	GasPrice int     `json:"gasprice"`
	Nonce    int     `json:"nonce"`
}
