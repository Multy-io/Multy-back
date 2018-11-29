package exchanger

import (
	"github.com/Multy-io/Multy-back/exchanger/changelly"
	"github.com/Multy-io/Multy-back/exchanger/common"
)


type FactoryExchanger struct {
	Exchangers 	[]CommonExchangerInterface
	Config 		[]common.BasicExchangeConfiguration
}

func (fe *FactoryExchanger) SetExchangersConfig(config []common.BasicExchangeConfiguration) {
	fe.Config = config
}

func (fe *FactoryExchanger) GetSupportedExchangers() []common.Exchanger {
	var exchangers []common.Exchanger
	// TODO: it might be a good idea to configure this list via .config file when exchangers will be > 1
	exchangers = append(exchangers, common.Exchanger{
		Name: changelly.ExchangeChangellyCanonicalName,
	})

	return exchangers
}

func (fe *FactoryExchanger) GetExchanger(exchangerName string) (CommonExchangerInterface, error) {
	for _, _exchanger := range fe.Exchangers {
		if _exchanger.GetName() == exchangerName {
			return _exchanger, nil
		}
	}

	var exchanger CommonExchangerInterface

	switch exchangerName {
	case changelly.ExchangeChangellyCanonicalName:
		exchanger = &changelly.ExchangerChangelly{}
		break
	}

	exchanger.Init(fe.getConfigByExchangerName(exchangerName))
	fe.Exchangers = append(fe.Exchangers, exchanger)

	return exchanger, nil
}

func (fe *FactoryExchanger) getConfigByExchangerName(exchangerName string) interface{} {
	var targetConfig interface{}

	for _, config := range fe.Config {
		if config.Name == exchangerName {
			targetConfig = config.Config
			break
		}
	}

	return targetConfig
}

type CommonExchangerInterface interface {
	Init(config interface{}) error
	GetName() string
	GetSupportedCurrencies() ([]common.CurrencyExchanger, error)
	GetTransactionMinimumAmount(from common.CurrencyExchanger, to common.CurrencyExchanger) (float64, error)
	GetExchangeAmount(from common.CurrencyExchanger, to common.CurrencyExchanger, amount float64) (float64, error)
	CreateTransaction(from common.CurrencyExchanger, to common.CurrencyExchanger, amount float64, address string) (
		common.ExchangeTransaction, error)
}