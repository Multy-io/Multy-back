package changelly

import (
	"fmt"
	"testing"
)

func TestExchangerChangelly_GetRequestHash(t *testing.T) {
	testName := "testName"
	testApiKey := "testApiKey"
	testApiSecret := "testApiSecret"
	expectedHash := "d1742e0349bf3800a155e140323e5bf40436bef0135eb6871263dd707dc4351a6480485527a1ab" +
		"fc8add9f837f7341ecb11c41a369c4ab3d55685a201d59111e"

	requestPacket := rpcPacket{
		Id: 1,
		Jsonrpc: "2.0",
		Method: "testMethod",
		Params: map[string]string{},
	}

	api := ExchangerChangelly{
		name: testName,
		config: InitConfig{
			apiUrl: "",
			apiKey: testApiKey,
			apiSecret: testApiSecret,
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
	api := ExchangerChangelly{
		name: "testName",
		config: InitConfig{
			apiUrl: "http://api.changelly.com",
			apiKey: "e277668dacd24629836b4c5f289aa52d",
			apiSecret: "57ef86f42d6790fc9b02f281a43e500a5639f067e4b25dc043240d891fc4e400",
		},
	}

	response, _ := api.GetSupportedCurrencies()
	fmt.Println(response)
}
