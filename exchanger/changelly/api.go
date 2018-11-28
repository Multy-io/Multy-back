package changelly

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Multy-io/Multy-back/exchanger/common"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	ExchangeChangellyCanonicalName  = "changelly"
	RpcGetCurrencies = "getCurrencies"
	RpcGetTransactionMinimumAmount = "getMinAmount"
	RpcGetExchangeAmount = "getExchangeAmount"
	RpcCreateTransaction = "createTransaction"
)

type InitConfig struct {
	apiUrl string
	apiKey string
	apiSecret string
}

type rpcPacket struct {
	Id string					`json:"id"`
	Jsonrpc string				`json:"jsonrpc"`
	Method string				`json:"method""`
	Params map[string]string	`json:"params"`
}

type rpcPacketResponse struct {
	Id string					`json:"id"`
	Jsonrpc string				`json:"jsonrpc"`
	Result interface{}			`json:"result"`
	Error struct{
		Code int32 		`json:"code"`
		Message string	`json:"message"`
	}							`json:"error,omitempty"`
}

type ExchangerChangelly struct {
	name string
	config InitConfig
}

func (ec *ExchangerChangelly) Init(config interface{}) error {
	ec.name = ExchangeChangellyCanonicalName
	ec.config = config.(InitConfig)

	return nil
}

func (ec *ExchangerChangelly) GetName() string {
	return ec.name
}

func (ec *ExchangerChangelly) GetSupportedCurrencies() ([]common.CurrencyExchanger, error) {
	var supportedCurrencies []common.CurrencyExchanger

	responseData, err := ec.sendRequest(RpcGetCurrencies, map[string]string{})
	if err == nil {
		var responsePacket rpcPacketResponse
		err = json.Unmarshal(responseData, &responsePacket)
		if err != nil {
			return supportedCurrencies, err
		}

		for _, currencyName := range responsePacket.Result.([]interface{}) {
			supportedCurrencies = append(supportedCurrencies, common.CurrencyExchanger{
				Name: currencyName.(string),
			})
		}
	}

	return supportedCurrencies, err
}

func (ec *ExchangerChangelly) GetTransactionMinimumAmount(from common.CurrencyExchanger,
	to common.CurrencyExchanger) (float64, error) {

		return 0.0, nil
}

func (ec *ExchangerChangelly) GetExchangeAmount(from common.CurrencyExchanger,
	to common.CurrencyExchanger, amount float64) (float64, error) {
		var amountConverted float64

		responseData, err := ec.sendRequest(RpcGetExchangeAmount, map[string]string{
			"from": from.Name,
			"to": to.Name,
			"amount": fmt.Sprintf("%f", amount),
		})

		if err == nil {
			var responsePacket rpcPacketResponse
			err = json.Unmarshal(responseData, &responsePacket)
			if err != nil {
				return amountConverted, err
			}

			responseResult, _ := responsePacket.Result.(string)
			amountConverted, err = strconv.ParseFloat(responseResult, 64)
			if err != nil {
				return amountConverted, err
			}
		}

		return amountConverted, nil
}

func (ec *ExchangerChangelly) CreateTransaction(from common.CurrencyExchanger, to common.CurrencyExchanger,
	amount float64, address string) (common.ExchangeTransaction, error) {
		var transaction common.ExchangeTransaction

		responseData, err := ec.sendRequest(RpcCreateTransaction, map[string]string{
			"from": from.Name,
			"to": to.Name,
			"amount": fmt.Sprintf("%f", amount),
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
		Id: "1",
		Jsonrpc: "2.0",
		Method: methodName,
		Params: params,
	}

	requestHash, err := ec.GetRequestHash(requestData)
	if err != nil {
		return []byte{}, err
	}

	requestDataJson, err := json.Marshal(requestData)
	if err != nil {
		return []byte{}, err
	}

	request, err := http.NewRequest("POST", ec.config.apiUrl, bytes.NewBuffer(requestDataJson))
	if err != nil {
		return []byte{}, err
	}

	request.Header.Set("api-key", ec.config.apiKey)
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
		hmac512 := hmac.New(sha512.New, []byte(ec.config.apiSecret))
		hmac512.Write(inputBytes)

		encodedHash = hex.EncodeToString(hmac512.Sum(nil))
	}

	return encodedHash, err
}

