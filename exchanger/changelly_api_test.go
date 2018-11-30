/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/

package exchanger

import (
	"math/rand"
	"testing"
	"time"
)

func TestExchangerChangelly_GetName(t *testing.T) {
	testName := "testName"

	api := ExchangerChangelly{
		name: testName,
	}

	if api.GetName() != testName {
		t.Errorf("Invalid exchanger api name, expected [%s], got [%s] ", testName, api.GetName())
	}
}

func TestExchangerChangelly_GetRequestHash(t *testing.T) {
	testName := "testName"
	testApiKey := "testApiKey"
	testApiSecret := "testApiSecret"
	expectedHash := "d1742e0349bf3800a155e140323e5bf40436bef0135eb6871263dd707dc4351a6480485527a1ab" +
		"fc8add9f837f7341ecb11c41a369c4ab3d55685a201d59111e"

	requestPacket := rpcPacket{
		Id:      "1",
		Jsonrpc: "2.0",
		Method:  "testMethod",
		Params:  map[string]string{},
	}

	api := ExchangerChangelly{
		name: testName,
		config: InitConfig{
			ApiUrl:    "",
			ApiKey:    testApiKey,
			ApiSecret: testApiSecret,
		},
	}

	encodedHash, err := api.GetRequestHash(requestPacket)
	if err != nil {
		t.Errorf("An error occured on request hashing [%+v] \n", err.Error())
	}

	if encodedHash != expectedHash {
		t.Errorf("Invalid hash sum calculated, expected [%s], calculated [%s]", expectedHash, encodedHash)
	}
}

func TestExchangerChangelly_GetSupportedCurrencies(t *testing.T) {
	t.Skip("integration test, SKIP")
	api := getSemiConfiguredApi()

	supportedCurrencies, err := api.GetSupportedCurrencies()
	if err != nil {
		t.Errorf("Got error in api response, [%s]", err.Error())
	}

	if len(supportedCurrencies) < 1 {
		t.Errorf("At least one  currency should be returned from api")
	}
}

func TestExchangerChangelly_GetExchangeAmount(t *testing.T) {
	t.Skip("integration test, SKIP")
	api := getSemiConfiguredApi()

	supportedCurrencies, err := api.GetSupportedCurrencies()
	if err != nil {
		t.Errorf("Got error in api response, [%s]", err.Error())
	}

	currencyFrom, currencyTo := getRandomCurrencyPairFromSlice(supportedCurrencies)

	exchangeAmountOrigin := 100.0
	exchangeAmountConverted, err := api.GetExchangeAmount(currencyFrom, currencyTo, exchangeAmountOrigin)
	if err != nil {
		t.Errorf("Got error in api response, [%s]", err.Error())
	}

	if exchangeAmountConverted < 0 {
		t.Errorf("Converted amount could not be less than 0, got [%v]", exchangeAmountConverted)
	}
}

func TestExchangerChangelly_CreateTransaction(t *testing.T) {
	t.Skip("integration test, SKIP")
	api := getSemiConfiguredApi()

	currencyFrom := CurrencyExchanger{Name: "btc"}
	currencyTo := CurrencyExchanger{Name: "eth"}
	exchangeAmount := 1000.0
	dummyAddress := "0xe6001AEb462B880A202597CAA3ad064093dD4880"

	transaction, err := api.CreateTransaction(currencyFrom, currencyTo, exchangeAmount, dummyAddress)
	if err != nil {
		t.Errorf("Got error on transaction creation, [%s]", err.Error())
	}

	if len(transaction.PayInAddress) < 10 {
		t.Errorf("PayIn Address is too short and pretty strange, [%s]", transaction.PayInAddress)
	}

	if transaction.PayOutAddress != dummyAddress {
		t.Errorf("PayOut Address miss match, expected [%s], got [%s]", dummyAddress, transaction.PayOutAddress)
	}
}

func getRandomCurrencyPairFromSlice(supportedCurrencies []CurrencyExchanger) (CurrencyExchanger, CurrencyExchanger) {
	rand.Seed(time.Now().Unix())
	indexCurrencyFrom := rand.Intn(len(supportedCurrencies))
	currencyFrom := supportedCurrencies[indexCurrencyFrom]
	supportedCurrencies = append(supportedCurrencies[:indexCurrencyFrom], supportedCurrencies[indexCurrencyFrom+1:]...)
	indexCurrencyTo := rand.Intn(len(supportedCurrencies))
	currencyTo := supportedCurrencies[indexCurrencyTo]

	return currencyFrom, currencyTo
}

func getSemiConfiguredApi() ExchangerChangelly {
	return ExchangerChangelly{
		name: "testName",
		config: InitConfig{
			ApiUrl:    "https://api.changelly.com",
			ApiKey:    "testKey",
			ApiSecret: "testSecret",
		},
	}
}
