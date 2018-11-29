/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/

package exchanger

import (
	"testing"
)

func TestFactoryExchanger_GetSupportedExchangers(t *testing.T) {
	factory := FactoryExchanger{}
	exchangers := factory.GetSupportedExchangers()

	// for correct feature support, at least 1 exchange should be configured
	if len(exchangers) < 0 {
		t.Errorf("At least 1 exchanger should be configured, got 0")
	}

	targetExchanger := exchangers[0]
	if targetExchanger.Name == "" {
		t.Errorf("Zero Exchanger name should be configured and provided")
	}
}

func TestFactoryExchanger_GetExchanger(t *testing.T) {
	factory := FactoryExchanger{}
	exchangers := factory.GetSupportedExchangers()

	targetExchanger := exchangers[0]
	targetExchangerApi, err := factory.GetExchanger(targetExchanger.Name)
	if err != nil {
		t.Errorf("Something went wrong on get exchanger initialization, [%+v] \n", err.Error())
	}

	if targetExchangerApi.GetName() == "" {
		t.Errorf("Failed to get desired exchanger from factory, got [%+v] \n", targetExchangerApi)
	}

	checkExchangerApi, err := factory.GetExchanger(targetExchanger.Name)
	if &targetExchangerApi !=  &checkExchangerApi {
		t.Errorf("Factory should not re-create objects")
	}
}
