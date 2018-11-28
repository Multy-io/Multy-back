package common


type Exchanger struct {
	Name string
}

type CurrencyExchanger struct {
	Name string
}

type ExchangeTransaction struct {
	Id string				`json:"id"`
	PayInAddress string		`json:"payinAddress"`
	PayOutAddress string	`json:"payoutAddress"`
	Error interface{}		`json:"error"`
}
