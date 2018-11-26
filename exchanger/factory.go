/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/

package exchanger

type FactoryExchanger struct {
	Exchangers []CommonExchangerInterface
	Config     []BasicExchangeConfiguration
}

func (fe *FactoryExchanger) SetExchangersConfig(config []BasicExchangeConfiguration) {
	fe.Config = config
}

func (fe *FactoryExchanger) GetSupportedExchangers() []Exchanger {
	var exchangers []Exchanger
	// TODO: it might be a good idea to configure this list via .config file when exchangers will be > 1
	exchangers = append(exchangers, Exchanger{
		Name: ExchangeChangellyCanonicalName,
	})

	return exchangers
}

func (fe *FactoryExchanger) GetExchanger(exchangerName string) (CommonExchangerInterface, error) {
	for _, exchanger := range fe.Exchangers {
		if exchanger.GetName() == exchangerName {
			return exchanger, nil
		}
	}

	var exchanger CommonExchangerInterface

	switch exchangerName {
	case ExchangeChangellyCanonicalName:
		exchanger = &ExchangerChangelly{}
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
	GetSupportedCurrencies() ([]CurrencyExchanger, error)
	GetTransactionMinimumAmount(from CurrencyExchanger, to CurrencyExchanger) (string, error)
	GetExchangeAmount(from CurrencyExchanger, to CurrencyExchanger, amount string) (string, error)
	CreateTransaction(from CurrencyExchanger, to CurrencyExchanger, amount string, address string) (
		ExchangeTransaction, error)
}
