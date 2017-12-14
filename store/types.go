package store

import "time"

// User represents a single app user
type User struct {
	UserID          string                     `bson:"userID"`  // User uqnique identifier
	Devices         []Device                   `bson:"devices"` // All user devices
	Wallets         []Wallet                   `bson:"wallets"` // All user addresses in all chains
	BTCTransactions map[string]*BTCTransaction //////// fix name kritinaaaaaaa
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
	DeviceID       string    `bson:"deviceID"`       // Device uqnique identifier (MAC address of device)
	PushToken      string    `bson:"pushToken"`      // Firebase
	JWT            string    `bson:"JWT"`            // Device JSON Web Token
	LastActionTime time.Time `bson:"lastActionTime"` // Last action time from current device
	LastActionIP   string    `bson:"lastActionIP"`   // IP from last session
	DeviceType     int       `bson:"deviceType"`     // 1 - IOS, 2 - Android
}

// Wallet Specifies a concrete wallet of user.
type Wallet struct {
	// Currency of wallet.
	CurrencyID int `bson:"currencyID"`

	//wallet identifier
	WalletIndex int `bson:"walletIndex"`

	//wallet identifier
	WalletName string `bson:"walletName"`

	LastActionTime time.Time `bson:"lastActionTime"`

	DateOfCreation time.Time `bson:"dateOfCreation"`

	// All addresses assigned to this wallet.
	Adresses []Address `bson:"addresses"`
}

type RatesRecord struct {
	Category int    `json:"category" bson:"category"`
	TxHash   string `json:"txHash" bson:"txHash"`
}

type Address struct {
	AddressIndex int    `json:"addressIndex" bson:"addressIndex"`
	Address      string `json:"address" bson:"address"`
}
