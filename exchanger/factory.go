package exchanger

import "github.com/Multy-io/Multy-back/exchanger/changelly"


type FactoryExchanger struct {
	exchangers []CommonExchangerInterface
}

func (fe *FactoryExchanger) GetSupportedExchangers() ([]Exchanger, error) {

}

func (fe *FactoryExchanger) GetExchanger(exchangerName string) (CommonExchangerInterface, error) {
	return &changelly.ExchangerChangelly{}, nil
}

type CommonExchangerInterface interface {
	GetName() string
	Init(config interface{}) error
	GetSupportedCurrencies() []CurrencyExchanger
	GetTransactionMinimumAmount(from CurrencyExchanger, to CurrencyExchanger) (float64, error)
	GetExchangeAmount(from CurrencyExchanger, to CurrencyExchanger, amount float64) (float64, error)
	CreateTransaction(from CurrencyExchanger, to CurrencyExchanger, amount float64, address string) (
		ExchangeTransaction, error)
}