package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	topicExchangeAll          = "exchangeAll"
	topicExchangeUpdate       = "exchangeUpdate"
	topicBTCTransactionUpdate = "btcTransaction"

	topicEthTransactionUpdate = "ethTransaction"
)

type EventGetExchangeReq struct {
	From string
	To   string
}

type EventGetExchangeResp struct {
	Currency string
	Value    float64
	Error    string
}

func processGetExchangeEvent(req EventGetExchangeReq) EventGetExchangeResp {
	log.Printf("[DEBUG] processGetExchangeEvent\n")

	reqURI := fmt.Sprintf("https://min-api.cryptocompare.com/data/price?fsym=%s&tsyms=%s", req.From, req.To)
	_, err := url.ParseRequestURI(reqURI)
	if err != nil {
		log.Printf("[ERR] processGetExchangeEvent: wrong reqURI: [%s], %s\n", reqURI, err.Error())
		return EventGetExchangeResp{
			Error: "Bad request",
		}
	}

	log.Printf("[DEBUG] processGetExchangeEvent: reqURI=%s\n", reqURI)
	resp, err := http.Get(reqURI)
	if err != nil {
		log.Printf("[ERR] processGetExchangeEvent: get exchange: [%s], err=%s\n", reqURI, err.Error())
		return EventGetExchangeResp{
			Error: "Bad request",
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERR] processGetExchangeEvent: get exchange: response status code=%d\n", resp.StatusCode)
		return EventGetExchangeResp{
			Error: "Internal server error",
		}
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERR] processGetExchangeEvent: get exchange: get response body: %s\n", err.Error())
		return EventGetExchangeResp{
			Error: "Internal server error",
		}
	}

	log.Printf("[DEBUG] processGetExchangeEvent: resp=[%s]\n", string(bodyBytes))

	var respMsg map[string]float64
	if err := json.Unmarshal(bodyBytes, &respMsg); err != nil {
		log.Printf("[ERR] processGetExchangeEvent: parse responce=%s\n", err.Error())
		return EventGetExchangeResp{
			Error: "Internal server error",
		}
	}

	log.Printf("[DEBUG] processGetExchangeEvent: respMsg=[%+v]\n", respMsg)

	value, ok := respMsg[req.To]
	if !ok {
		return EventGetExchangeResp{
			Error: "Internal server error",
		}
	}

	log.Printf("[DEBUG] processGetExchangeEvent: value=%.6f\n", value)

	return EventGetExchangeResp{
		Currency: req.From,
		Value:    value,
		Error:    "",
	}
}
