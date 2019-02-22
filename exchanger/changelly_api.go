/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/

package exchanger

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const (
	ExchangeChangellyCanonicalName = "changelly"
	RpcGetCurrencies               = "getCurrencies"
	RpcGetTransactionMinimumAmount = "getMinAmount"
	RpcGetExchangeAmount           = "getExchangeAmount"
	RpcCreateTransaction           = "createTransaction"
)

type InitConfig struct {
	ApiUrl    string
	ApiKey    string
	ApiSecret string
}

type rpcPacket struct {
	Id      string            `json:"id"`
	Jsonrpc string            `json:"jsonrpc"`
	Method  string            `json:"method""`
	Params  map[string]string `json:"params"`
}

type rpcPacketResponse struct {
	Id      string      `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   struct {
		Code    int32  `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type ExchangerChangelly struct {
	name   string
	config InitConfig
}

func (ec *ExchangerChangelly) Init(config interface{}) error {
	ec.name = ExchangeChangellyCanonicalName
	configMap := config.(map[string]interface{})

	ec.config = InitConfig{
		ApiUrl:    configMap["apiUrl"].(string),
		ApiKey:    configMap["apiKey"].(string),
		ApiSecret: configMap["apiSecret"].(string),
	}

	return nil
}

func (ec *ExchangerChangelly) GetName() string {
	return ec.name
}

func (ec *ExchangerChangelly) GetSupportedCurrencies() ([]CurrencyExchanger, error) {
	var supportedCurrencies []CurrencyExchanger

	responseData, err := ec.sendRequest(RpcGetCurrencies, map[string]string{})
	if err == nil {
		var responsePacket rpcPacketResponse
		err = json.Unmarshal(responseData, &responsePacket)
		if err != nil {
			return supportedCurrencies, err
		}

		for _, currencyName := range responsePacket.Result.([]interface{}) {
			supportedCurrencies = append(supportedCurrencies, CurrencyExchanger{
				Name: currencyName.(string),
			})
		}
	}

	return supportedCurrencies, err
}

func (ec *ExchangerChangelly) GetTransactionMinimumAmount(from CurrencyExchanger,
	to CurrencyExchanger) (string, error) {
	responseData, err := ec.sendRequest(RpcGetTransactionMinimumAmount, map[string]string{
		"from": from.Name,
		"to":   to.Name,
	})

	if err != nil {
		return "", errors.Wrapf(err, "Changelly API request failed: %s", RpcGetTransactionMinimumAmount)
	}

	var responsePacket rpcPacketResponse
	err = json.Unmarshal(responseData, &responsePacket)
	if err != nil {
		return "", errors.Wrapf(err, "Changelly API failed to parse response of %s", RpcGetExchangeAmount)
	}

	return responsePacket.Result.(string), nil
}

func (ec *ExchangerChangelly) GetExchangeAmount(from CurrencyExchanger,
	to CurrencyExchanger, amount string) (string, error) {
	responseData, err := ec.sendRequest(RpcGetExchangeAmount, map[string]string{
		"from":   from.Name,
		"to":     to.Name,
		"amount": amount,
	})

	if err != nil {
		return "", errors.Wrapf(err, "Changelly API request failed: %s", RpcGetExchangeAmount)
	}

	var responsePacket rpcPacketResponse
	err = json.Unmarshal(responseData, &responsePacket)
	if err != nil {
		return "", errors.Wrapf(err, "Changelly API failed to parse response of %s", RpcGetExchangeAmount)
	}

	return responsePacket.Result.(string), nil
}

func (ec *ExchangerChangelly) CreateTransaction(from CurrencyExchanger, to CurrencyExchanger,
	amount string, address string) (ExchangeTransaction, error) {
	var transaction ExchangeTransaction

	responseData, err := ec.sendRequest(RpcCreateTransaction, map[string]string{
		"from":    from.Name,
		"to":      to.Name,
		"amount":  amount,
		"address": address,
	})

	if err == nil {
		var responsePacket rpcPacketResponse
		err = json.Unmarshal(responseData, &responsePacket)
		if err != nil {
			return transaction, err
		}

		if responsePacket.Error.Code != 0 {
			transaction.Error = responsePacket.Error
		} else {
			responsePacketDict := responsePacket.Result.(map[string]interface{})
			transaction.Id = responsePacketDict["id"].(string)
			transaction.PayInAddress = responsePacketDict["payinAddress"].(string)
			transaction.PayOutAddress = responsePacketDict["payoutAddress"].(string)
		}
	}

	return transaction, nil
}

func (ec *ExchangerChangelly) sendRequest(methodName string, params map[string]string) ([]byte, error) {
	var requestData = rpcPacket{
		Id:      "1",
		Jsonrpc: "2.0",
		Method:  methodName,
		Params:  params,
	}

	requestHash, err := ec.GetRequestHash(requestData)
	if err != nil {
		return []byte{}, err
	}

	requestDataJson, err := json.Marshal(requestData)
	if err != nil {
		return []byte{}, err
	}

	request, err := http.NewRequest("POST", ec.config.ApiUrl, bytes.NewBuffer(requestDataJson))
	if err != nil {
		return []byte{}, err
	}

	request.Header.Set("api-key", ec.config.ApiKey)
	request.Header.Set("sign", requestHash)
	request.Header.Set("Content-type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return []byte{}, err
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	return body, nil
}

func (ec *ExchangerChangelly) GetRequestHash(request rpcPacket) (string, error) {
	var encodedHash string
	inputBytes, err := json.Marshal(request)

	if err == nil {
		hmac512 := hmac.New(sha512.New, []byte(ec.config.ApiSecret))
		hmac512.Write(inputBytes)

		encodedHash = hex.EncodeToString(hmac512.Sum(nil))
	}

	return encodedHash, err
}
