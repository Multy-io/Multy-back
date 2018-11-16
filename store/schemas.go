/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package store

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	ethpb "github.com/Multy-io/Multy-ETH-node-service/node-streamer"
	"github.com/graarh/golang-socketio"
)

const (
	TxStatusAppearedInMempoolIncoming = 1
	TxStatusAppearedInBlockIncoming   = 2

	TxStatusAppearedInMempoolOutcoming = 3
	TxStatusAppearedInBlockOutcoming   = 4

	TxStatusInBlockConfirmedIncoming  = 5
	TxStatusInBlockConfirmedOutcoming = 6

	TxStatusInBlockMethodInvocationFail = 7
	TxStatusTxRejectedIncoming          = 8
	TxStatusTxRejectedOutgoing          = 9

	MultisigStatusWaitingForJoin = 1
	MultisigStatusAllJoined      = 2
	MultisigStatusDeployPending  = 3
	MultisigStatusRejected       = 4
	MultisigStatusDeployed       = 5

	MultisigOwnerStatusWaiting   = 0
	MultisigOwnerStatusSeen      = 1
	MultisigOwnerStatusConfirmed = 2
	MultisigOwnerStatusDeclined  = 3
	MultisigOwnerStatusRevoked   = 4

	// ws notification topic
	TopicTransaction = "TransactionUpdate"
	TopicNewIncoming = "NewIncoming"

	MsgSend    = "message:send"
	MsgRecieve = "message:recieve"

	JoinMultisig       = 1
	LeaveMultisig      = 2
	DeleteMultisig     = 3
	KickMultisig       = 4
	CheckMultisig      = 5
	ViewTransaction    = 6
	DeclineTransaction = 7

	NotifyDeploy      = 8
	NotifyPaymentReq  = 9
	NotifyIncomingTx  = 10
	NotifyConfirmTx   = 11
	NotifyRevokeTx    = 12
	NotifyResyncEnd   = 13
	NotifyTxSubmitted = 14

	AssetTypeMultyAddress    = 0
	AssetTypeMultisig        = 1
	AssetTypeImportedAddress = 2

	MultiSigFactory    = "0xf8f73808"
	SubmitTransaction  = "0xc6427474"
	ConfirmTransaction = "0xc01a8c84"
	RevokeConfirmation = "0x20ea8d86"
	ExecuteTransaction = "0xee22610b"

	MaximumAvalibeEmptyWallets = 20
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

	IsImported   bool `bson:"isImported"`
	BrokenStatus int  `bson:"brokenStatus"`
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

type WalletParams struct {
	CurrencyID   int            `json:"currencyID"`
	NetworkID    int            `json:"networkID"`
	Address      string         `json:"address"`
	AddressIndex int            `json:"addressIndex"`
	WalletIndex  int            `json:"walletIndex"`
	WalletName   string         `json:"walletName"`
	IsImported   bool           `json:"isImported"`
	Multisig     MultisigWallet `json:"multisig"`
}

type MultisigWallet struct {
	IsMultisig         bool   `json:"isMultisig"`
	SignaturesRequired int    `json:"signaturesRequired"`
	OwnersCount        int    `json:"ownersCount"`
	InviteCode         string `json:"inviteCode"`
	IsImported         bool   `json:"isImported"`
	ContractAddress    string `json:"contractAddress"`
}

type EtherscanResp struct {
	Status  string               `json:"status"`
	Message string               `json:"message"`
	Result  []ethpb.ERC20History `json:"result"`
}
type ERC20TokenTransferTx struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	BlockHash         string `json:"blockHash"`
	From              string `json:"from"`
	ContractAddress   string `json:"contractAddress"`
	To                string `json:"to"`
	Value             string `json:"value"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	TokenDecimal      string `json:"tokenDecimal"`
	TransactionIndex  string `json:"transactionIndex"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	Input             string `json:"input"`
	Confirmations     string `json:"confirmations"`
}

type TokenBalance struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

type TransactionETH struct {
	UserID            string                `json:"userid,omitempty"`
	WalletIndex       int                   `json:"walletindex,omitempty"`
	AddressIndex      int                   `json:"addressindex,omitempty"`
	Hash              string                `json:"txhash"`
	From              string                `json:"from"`
	To                string                `json:"to"`
	Amount            string                `json:"txoutamount"`
	GasPrice          int64                 `json:"gasprice"`
	GasLimit          int64                 `json:"gaslimit"`
	Nonce             int                   `json:"nonce"`
	Status            int                   `json:"txstatus" bson:"txstatus"`
	BlockTime         int64                 `json:"blocktime"`
	PoolTime          int64                 `json:"mempooltime"`
	BlockHeight       int64                 `json:"blockheight"`
	Confirmations     int                   `json:"confirmations"`
	IsInternal        bool                  `json:"isinternal"`
	Multisig          *MultisigTx           `json:"multisig,omitempty"`
	ERC20Token        *ERC20Tx              `json:"erc20Token,omitempty"`
	StockExchangeRate []ExchangeRatesRecord `json:"stockexchangerate"`
}

type MultisigTx struct {
	Contract         string         `json:"contract,omitempty"`
	MethodInvoked    string         `json:"methodinvoked,omitempty"`
	Input            string         `json:"input"`
	InvocationStatus bool           `json:"invocationstatus"`
	RequestID        int64          `json:"requestid"`
	Return           string         `json:"return,omitempty"`
	Owners           []OwnerHistory `json:"owners,omitempty"`
	Confirmed        bool           `json:"confirmed"`
}

type ERC20Tx struct {
	From            string `json:"from"`
	To              string `json:"to"`
	ContractAddress string `json:"ContractAddress"`
	Value           string `json:"value"`
}

type Multisig struct {
	CurrencyID      int               `bson:"currencyid" json:"currencyid"`
	NetworkID       int               `bson:"networkid" json:"networkid"`
	Confirmations   int               `bson:"confirmations" json:"confirmations"`
	WalletName      string            `bson:"walletName" json:"walletName"`
	FactoryAddress  string            `bson:"factoryAddress" json:"factoryAddress"`
	ContractAddress string            `bson:"contractAddress" json:"contractAddress"`
	TxOfCreation    string            `bson:"txOfCreation" json:"txOfCreation"`
	LastActionTime  int64             `bson:"lastActionTime" json:"lastActionTime"`
	DateOfCreation  int64             `bson:"dateOfCreation" json:"dateOfCreation"`
	Owners          []AddressExtended `bson:"owners" json:"owners"`
	DeployStatus    int               `bson:"deployStatus" json:"deployStatus"`
	Status          string            `bson:"status" json:"status"`
	InviteCode      string            `bson:"inviteCode" json:"inviteCode"`
	OwnersCount     int               `bson:"ownersCount" json:"ownersCount"`
	Imported        bool              `bson:"imported" json:"imported"`
}

type MultisigExtended struct {
	Multisig      Multisig `json:"multisig" bson:"multisig"`
	KickedAddress string   `json:"kickedAddress" bson:"kickedAddress"`
}

type OwnerHistory struct {
	Address            string `bson:"address" json:"address"`
	ConfirmationTX     string `bson:"confirmationtx" json:"confirmationtx"`
	ConfirmationStatus int    `bson:"confirmationStatus" json:"confirmationStatus"`
	ConfirmationTime   int64  `bson:"confirmationTime" json:"confirmationTime"`
	SeenTime           int64  `bson:"seenTime" json:"seenTime"`
}

type CoinType struct {
	СurrencyID    int `bson:"currencyID"`
	NetworkID     int `bson:"networkID"`
	AccuracyRange int
	GRPCUrl       string
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
	UserID       string `bson:"userid" json:"userid"`
	Address      string `bson:"address" json:"address"`       // etereum asociated to contract address
	Associated   bool   `bson:"associated" json:"associated"` // is associated
	Creator      bool   `bson:"creator" json:"creator"`
	WalletIndex  int    `bson:"walletIndex" json:"walletIndex"`
	AddressIndex int    `bson:"addressIndex" json:"addressIndex"`
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

type MobileVersions struct {
	Android struct {
		Hard int `json:"hard"`
		Soft int `json:"soft"`
	} `json:"android"`
	Ios struct {
		Hard int `json:"hard"`
		Soft int `json:"soft"`
	} `json:"ios"`
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
	Type    int         `json:"type"`
	From    string      `json:"from"`
	To      string      `json:"to"`
	Date    int64       `json:"date"`
	Status  int         `json:"status"`
	Payload interface{} `json:"payload"`
}
type WsResponse struct {
	Message string      `bson:"message"`
	Payload interface{} `bson:"payload"`
}

type MultisigMsg struct {
	UserID        string `json:"userid"`
	Address       string `json:"address"`
	InviteCode    string `json:"invitecode"`
	AddressToKick string `json:"addresstokick,omitempty"`
	WalletIndex   int    `json:"walletindex"`
	CurrencyID    int    `json:"currencyid"`
	NetworkID     int    `json:"networkid"`
	TxID          string `json:"txid,omitempty"`
}

type InviteCodeInfo struct {
	CurrencyID int  `json:"currencyid"`
	NetworkID  int  `json:"networkid"`
	Exists     bool `json:"exists"`
}

type BtcComResp struct {
	Data struct {
		TotalCount int `json:"total_count"`
		Page       int `json:"page"`
		List       []struct {
			BlockHeight int    `json:"block_height"`
			Hash        string `json:"hash"`
		} `json:"list"`
	}
	ErrNo  int         `json:"err_no"`
	ErrMsg interface{} `json:"err_msg"`
}

type VerifiedTokenList []struct {
	ContractAddress string
	Ticker          string
	Name            string
}

type BrowserDefault struct {
	URL        string `json:"url"`
	CurrencyID int    `json:"currencyid"`
	NetworkID  int    `json:"networkid"`
}

func (s *MultisigMsg) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := SetField(s, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return errors.New("Provided value type didn't match obj field type")
	}

	structFieldValue.Set(val)
	return nil
}
