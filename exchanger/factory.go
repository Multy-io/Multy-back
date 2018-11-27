package exchanger

import (
	"github.com/Multy-io/Multy-back/exchanger/changelly"
	"github.com/Multy-io/Multy-back/exchanger/common"
)


type FactoryExchanger struct {
	exchangers []CommonExchangerInterface
}

func (fe FactoryExchanger) GetSupportedExchangers() []common.Exchanger {
	var exchangers []common.Exchanger
	// TODO: it might be a good idea to configure this list via .config file when exchangers will be > 1
	exchangers = append(exchangers, common.Exchanger{
		Name: changelly.ExchangeChangellyCanonicalName,
	})

	return exchangers
}

func (fe *FactoryExchanger) GetExchanger(exchangerName string) (CommonExchangerInterface, error) {
	var exchanger CommonExchangerInterface

	switch exchangerName {
	case changelly.ExchangeChangellyCanonicalName:
		exchanger = &changelly.ExchangerChangelly{}
		var dummy changelly.InitConfig
		exchanger.Init(dummy)
		break
	}

	return exchanger, nil
}

type CommonExchangerInterface interface {
	Init(config interface{}) error
	GetName() string
	GetSupportedCurrencies() []common.CurrencyExchanger
	GetTransactionMinimumAmount(from common.CurrencyExchanger, to common.CurrencyExchanger) (float64, error)
	GetExchangeAmount(from common.CurrencyExchanger, to common.CurrencyExchanger, amount float64) (float64, error)
	CreateTransaction(from common.CurrencyExchanger, to common.CurrencyExchanger, amount float64, address string) (
		common.ExchangeTransaction, error)
}