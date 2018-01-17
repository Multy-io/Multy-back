/*
Copyright 2017 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package store

import (
	"time"
)

// User represents a single app user
type User struct {
	UserID  string   `bson:"userID"`  // User uqnique identifier
	Devices []Device `bson:"devices"` // All user devices
	Wallets []Wallet `bson:"wallets"` // All user addresses in all chains
}

type BTCTransaction struct {
	Hash    string                `json:"hash"`
	Txid    string                `json:"txid"`
	Time    time.Time             `json:"time"`
	Outputs map[string]*BtcOutput `json:"outputs"` // addresses to outputs, key = address
}

type BtcOutput struct {
	Address     string  `json:"address"`
	Amount      float64 `json:"amount"`
	TxIndex     uint32  `json:"txIndex"`
	TxOutScript string  `json:"txOutScript"`
}

type TxInfo struct {
	Type    string  `json:"type"`
	TxHash  string  `json:"txhash"`
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}

// Device represents a single users device.
type Device struct {
	DeviceID       string `bson:"deviceID"`       // Device uqnique identifier (MAC address of device)
	PushToken      string `bson:"pushToken"`      // Firebase
	JWT            string `bson:"JWT"`            // Device JSON Web Token
	LastActionTime int64  `bson:"lastActionTime"` // Last action time from current device
	LastActionIP   string `bson:"lastActionIP"`   // IP from last session
	DeviceType     int    `bson:"deviceType"`     // 1 - IOS, 2 - Android
}

const (
	WalletStatusOK      = "ok"
	WalletStatusDeleted = "deleted"
)

// Wallet Specifies a concrete wallet of user.
type Wallet struct {
	// Currency of wallet.
	CurrencyID int `bson:"currencyID"`

	//wallet identifier
	WalletIndex int `bson:"walletIndex"`

	//wallet identifier
	WalletName string `bson:"walletName"`

	LastActionTime int64 `bson:"lastActionTime"`

	DateOfCreation int64 `bson:"dateOfCreation"`

	// All addresses assigned to this wallet.
	Adresses []Address `bson:"addresses"`

	Status string `bson:"status"`
}

type RatesRecord struct {
	Category int    `json:"category" bson:"category"`
	TxHash   string `json:"txHash" bson:"txHash"`
}

type Address struct {
	AddressIndex   int    `json:"addressIndex" bson:"addressIndex"`
	Address        string `json:"address" bson:"address"`
	LastActionTime int64  `json:"lastActionTime" bson:"lastActionTime"`
}
type WalletsSelect struct {
	Wallets []struct {
		Addresses []struct {
			AddressIndex int    `bson:"addressIndex"`
			Address      string `bson:"address"`
		} `bson:"addresses"`
		WalletIndex int `bson:"walletIndex"`
	} `bson:"wallets"`
}

// the way how user transations store in db
type MultyTX struct {
	TxID              string                `json:"txid"`
	TxHash            string                `json:"txhash"`
	TxOutScript       string                `json:"txoutscript"`
	TxAddress         string                `json:"address"`
	TxStatus          int                   `json:"txstatus"`
	TxOutAmount       int64                 `json:"txoutamount"`
	TxOutID           int                   `json:"txoutid"`
	WalletIndex       int                   `json:"walletindex"`
	BlockTime         int64                 `json:"blocktime"`
	BlockHeight       int64                 `json:"blockheight"`
	TxFee             int64                 `json:"txfee"`
	MempoolTime       int64                 `json:"mempooltime"`
	StockExchangeRate []ExchangeRatesRecord `json:"stockexchangerate"`
	TxInputs          []AddresAmount        `json:"txinputs"`
	TxOutputs         []AddresAmount        `json:"txoutputs"`
}
type AddresAmount struct {
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}

type TxRecord struct {
	UserID       string    `json:"userid"`
	Transactions []MultyTX `json:"transactions"`
}

// ExchangeRatesRecord presents record with exchanges from rate stock
// with additional information, such as date and exchange stock
type ExchangeRatesRecord struct {
	Exchanges     ExchangeRates `json:"exchanges"`
	Timestamp     int64         `json:"timestamp"`
	StockExchange string        `json:"stock_exchange"`
}
type SpendableOutputs struct {
	TxID              string                `json:"txid"`
	TxOutID           int                   `json:"txoutid"`
	TxOutAmount       int                   `json:"txoutamount"`
	TxOutScript       string                `json:"txoutscript"`
	AddressIndex      int                   `json:"addressindex"`
	TxStatus          int                   `json:"txstatus"`
	StockExchangeRate []ExchangeRatesRecord `json:"stockexchangerate"`
}

// ExchangeRates stores exchange rates
type ExchangeRates struct {
	EURtoBTC float64 `json:"eur_btc,omitempty"`
	USDtoBTC float64 `json:"usd_btc,omitempty"`
	ETHtoBTC float64 `json:"eth_btc,omitempty"`

	ETHtoUSD float64 `json:"eth_usd,omitempty"`
	ETHtoEUR float64 `json:"eth_eur,omitempty"`

	BTCtoUSD float64 `json:"btc_usd,omitempty"`
}

type RatesAPIBitstamp struct {
	Date  string `json:"date"`
	Price string `json:"price"`
}
