/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package store

import (
	"time"

	"github.com/graarh/golang-socketio"
)

const (
	TxStatusAppearedInMempoolIncoming = 1
	TxStatusAppearedInBlockIncoming   = 2

	TxStatusAppearedInMempoolOutcoming = 3
	TxStatusAppearedInBlockOutcoming   = 4

	TxStatusInBlockConfirmedIncoming  = 5
	TxStatusInBlockConfirmedOutcoming = 6

	// ws notification topic
	TopicTransaction = "TransactionUpdate"
	TopicNewIncoming = "NewIncoming"
)

// User represents a single app user
type User struct {
	UserID    string     `bson:"userID"`  // User uqnique identifier
	Devices   []Device   `bson:"devices"` // All user devices
	Wallets   []Wallet   `bson:"wallets"` // All user addresses in all chains
	Multisigs []Multisig `bson:"multisig"`
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
	DeviceID       string `bson:"deviceID"`       // Device uqnique identifier
	PushToken      string `bson:"pushToken"`      // Firebase
	JWT            string `bson:"JWT"`            // Device JSON Web Token
	LastActionTime int64  `bson:"lastActionTime"` // Last action time from current device
	LastActionIP   string `bson:"lastActionIP"`   // IP from last session
	AppVersion     string `bson:"appVersion"`     // Mobile app verson
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
	// Sub-net of currency 0 - main 1 - test
	NetworkID int `bson:"networkID"`

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

type Multisig struct {
	CurrencyID      int               `bson:"currencyid"`
	NetworkID       int               `bson:"networkid"`
	Confirmations   int               `bson:"confirmations"`
	WalletName      string            `bson:"walletname"`
	FactoryAddress  string            `bson:"factoryaddress"`
	ContractAddress string            `bson:"contractaddress"`
	TxOfCreation    string            `bson:"txofcreation"`
	LastActionTime  int64             `bson:"lastactiontime"`
	DateOfCreation  int64             `bson:"dateofcreation"`
	Owners          []AddressExtended `bson:"owners"`
	DeployStatus    bool              `bson:"deploystatus"`
	Status          string            `bson:"status"`
	InviteCode      string            `bson:"invitecode"`
	OwnersCount     int               `bson:"ownerscount"`
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

type WalletForTx struct {
	UserId      string           `json:"userid"`
	WalletIndex int              `json:"walletindex"`
	Address     AddressForWallet `json:"address"`
}

type AddressForWallet struct {
	AddressIndex    int    `json:"addressindex"`
	AddressOutIndex int    `json:"addresoutindex"`
	Address         string `json:"address"`
	Amount          int64  `json:"amount"`
}

// the way how user transations store in db
type MultyTX struct {
	UserId            string                `json:"userid"`
	TxID              string                `json:"txid"`
	TxHash            string                `json:"txhash"`
	TxOutScript       string                `json:"txoutscript"`
	TxAddress         []string              `json:"addresses"` //this is major addresses of the transaction (if send - inputs addresses of our user, if get - outputs addresses of our user)
	TxStatus          int                   `json:"txstatus"`
	TxOutAmount       int64                 `json:"txoutamount"`
	BlockTime         int64                 `json:"blocktime"`
	BlockHeight       int64                 `json:"blockheight"`
	Confirmations     int                   `json:"confirmations"`
	TxFee             int64                 `json:"txfee"`
	MempoolTime       int64                 `json:"mempooltime"`
	StockExchangeRate []ExchangeRatesRecord `json:"stockexchangerate"`
	TxInputs          []AddresAmount        `json:"txinputs"`
	TxOutputs         []AddresAmount        `json:"txoutputs"`
	WalletsInput      []WalletForTx         `json:"walletsinput"`  //here we storing all wallets and addresses that took part in Inputs of the transaction
	WalletsOutput     []WalletForTx         `json:"walletsoutput"` //here we storing all wallets and addresses that took part in Outputs of the transaction
}

type BTCResync struct {
	Txs    []MultyTX
	SpOuts []SpendableOutputs
}
type ResyncTx struct {
	Hash        string
	BlockHeight int
}

type WsTxNotify struct {
	CurrencyID      int    `json:"currencyid"`
	NetworkID       int    `json:"networkid"`
	Address         string `json:"address"`
	Amount          string `json:"amount"`
	TxID            string `json:"txid"`
	TransactionType int    `json:"transactionType"`
	WalletIndex     int    `json:"walletindex"`
	From            string `json:"from"`
	To              string `json:"to"`
	Multisig        string `json:"multisig"`
}

type TransactionWithUserID struct {
	NotificationMsg *WsTxNotify
	UserID          string
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

// ExchangeRates stores exchange rates
type ExchangeRates struct {
	EURtoBTC float64 `json:"eur_btc"`
	USDtoBTC float64 `json:"usd_btc"`
	ETHtoBTC float64 `json:"eth_btc"`

	ETHtoUSD float64 `json:"eth_usd"`
	ETHtoEUR float64 `json:"eth_eur"`

	BTCtoUSD float64 `json:"btc_usd"`
}

type RatesAPIBitstamp struct {
	Date  string `json:"date"`
	Price string `json:"price"`
}
type SpendableOutputs struct {
	TxID              string                `json:"txid"`
	TxOutID           int                   `json:"txoutid"`
	TxOutAmount       int64                 `json:"txoutamount"`
	TxOutScript       string                `json:"txoutscript"`
	Address           string                `json:"address"`
	UserID            string                `json:"userid"`
	WalletIndex       int                   `json:"walletindex"`
	AddressIndex      int                   `json:"addressindex"`
	TxStatus          int                   `json:"txstatus"`
	StockExchangeRate []ExchangeRatesRecord `json:"stockexchangerate"`
}

type WalletETH struct {
	// Currency of wallet.
	CurrencyID int `bson:"currencyID"`
	// Currency of wallet.
	NetworkID int `bson:"networkID"`

	//wallet identifier
	WalletIndex int `bson:"walletIndex"`

	//wallet identifier
	WalletName string `bson:"walletName"`

	LastActionTime int64 `bson:"lastActionTime"`

	DateOfCreation int64 `bson:"dateOfCreation"`

	// All addresses assigned to this wallet.
	Adresses []Address `bson:"addresses"`

	// Wallet status
	Status string `bson:"status"`

	// Balance of the eth wallet in wei
	Balance int64 `bson:"balance"`

	// Nonce of the wallet - index of the last transaction
	Nonce int64 `bson:"nonce"`
}

type TransactionETH struct {
	UserID            string                `json:"userid,omitempty"`
	WalletIndex       int                   `json:"walletindex,omitempty"`
	AddressIndex      int                   `json:"addressindex,omitempty"`
	Hash              string                `json:"txhash"`
	From              string                `json:"from"`
	To                string                `json:"to"`
	Amount            string                `json:"txoutamount"`
	Input             string                `json:"input"`
	GasPrice          int64                 `json:"gasprice"`
	GasLimit          int64                 `json:"gaslimit"`
	Nonce             int                   `json:"nonce"`
	Status            int                   `json:"txstatus" bson:"txstatus"`
	BlockTime         int64                 `json:"blocktime"`
	PoolTime          int64                 `json:"mempooltime"`
	BlockHeight       int64                 `json:"blockheight"`
	Confirmations     int                   `json:"confirmations"`
	Contract          string                `json:"contract,omitempty"`
	Index             int64                 `json:"index,omitempty"`
	MethodInvoked     string                `json:"methodinvoked,omitempty"`
	InvocationStatus  bool                  `json:"invocationstatus,omitempty"`
	Return            string                `json:"return,omitempty"`
	Owners            []OwnerHistory        `json:"owners,omitempty"`
	Confirmed         bool                  `json:"confirmed,omitempty"`
	IsInternal        bool                  `json:"isinternal,omitempty"`
	StockExchangeRate []ExchangeRatesRecord `json:"stockexchangerate"`
}

type OwnerHistory struct {
	Address          string `json:"address"`
	ConfirmationTX   string `json:"confirmationtx"`
	Confirmed        bool   `json:"confirmed"`
	Seen             bool   `json:"seen"`
	ConfirmationTime int64  `json:"confirmationTime"`
	SeenTime         int64  `json:"seenTime"`
}

type CoinType struct {
	СurrencyID int `bson:"currencyID"`
	NetworkID  int `bson:"networkID"`
	GRPCUrl    string
}

type MempoolRecord struct {
	Category int    `json:"category"`
	HashTX   string `json:"hashTX"`
}

type DeleteSpendableOutput struct {
	UserID  string
	TxID    string
	Address string
}

type DonationInfo struct {
	FeatureCode     int
	DonationAddress string
}

type AddressExtended struct {
	UserID       string
	Address      string // etereum asociated to contract address
	Associated   bool   // is associated
	Creator      bool
	WalletIndex  int
	AddressIndex int
}

type ServerConfig struct {
	BranchName string `json:"branch"`
	CommitHash string `json:"commit"`
	Build      string `json:"build_time"`
	Tag        string `json:"tag"`
}

// Donation Statuses
// 0 - Pending
// 1 - Active
// 2 - Closed
// 3 - Canceled
type Donation struct {
	FeatureID int    `json:"id"`
	Address   string `json:"address"`
	Amount    int64  `json:"amount"`
	Status    int    `json:"status"`
}

type ServiceInfo struct {
	Branch    string
	Commit    string
	Buildtime string
	Lasttag   string
}

type Receiver struct {
	ID         string `json:"userid"`
	UserCode   string `json:"usercode"`
	CurrencyID int    `json:"currencyid"`
	NetworkID  int    `json:"networkid"`
	Address    string `json:"address"`
	Amount     string `json:"amount"`
	Socket     *gosocketio.Channel
}

type Sender struct {
	ID       string `json:"userid"`
	UserCode string `json:"usercode"`
	Visible  map[string]bool
	Socket   *gosocketio.Channel
}

type ReceiverInData struct {
	ID         string `json:"userid"`
	CurrencyID int    `json:"currencyid"`
	Amount     int64  `json:"amount"`
	UserCode   string `json:"usercode"`
}

type SenderInData struct {
	Code    string   `json:"usercode"`
	UserID  string   `json:"userid"`
	Visible []string `json:"userid"`
}

type PaymentData struct {
	FromID     string `json:"fromid"`
	ToID       string `json:"toid"`
	CurrencyID int    `json:"currencyid"`
	Amount     int64  `json:"amount"`
}

type NearVisible struct {
	IDs []string `json:"ids"`
}

type RawHDTx struct {
	CurrencyID int    `json:"currencyid"`
	NetworkID  int    `json:"networkID"`
	UserCode   string `json:"usercode"`
	JWT        string `json:"JWT"`
	Payload    `json:"payload"`
}

type Payload struct {
	Address      string `json:"address"`
	AddressIndex int    `json:"addressindex"`
	WalletIndex  int    `json:"walletindex"`
	Transaction  string `json:"transaction"`
	IsHD         bool   `json:"ishd"`
}

type LastState struct {
	BlockHeight int64 `bson:"blockheight"`
	CurrencyID  int   `bson:"currencyid"`
	NetworkID   int   `bson:"networkid"`
}

type WsMessage struct {
	Type    string      `bson:"type"`
	From    string      `bson:"from"`
	To      string      `bson:"to"`
	Date    int64       `bson:"date"`
	Status  int         `bson:"status"`
	Payload interface{} `bson:"payload"`
}
type WsResponse struct {
	Message string      `bson:"message"`
	Payload interface{} `bson:"payload"`
}

type MultisigMsg struct {
	UserID      string `bson:"userid"`
	Address     string `bson:"address"`
	InviteCode  string `bson:"invitecode"`
	WalletIndex int    `bson:"walletindex"`
	CurrencyID  int    `bson:"currencyid"`
	NetworkID   int    `bson:"networkid"`
}
