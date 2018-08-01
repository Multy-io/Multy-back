package client

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jekabolt/slf"
)

// PoloniexAPI wraps websocket connection for PoloniexAPI
type BitfinexAPI struct {
	conn    *websocket.Conn
	rates   *StockRate
	indexes map[int]string

	log slf.StructuredLogger
}

// PoloniexSocketEvent is a Poloniex json parser structure
type BitfinexSocketEvent struct {
	Data struct {
		Price     string `json:"last"`
		ProductID string `json:"symbol"`
	} `json:"params"`
}

type BitfinexSub struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Pair    string `json:"pair"`
}

type BitfinexIndex struct {
	Pair string `json:"pair"`
	ID   int    `json:"id"`
}
type BitfinexPriceIndex struct {
	Price float64 `json:"price"`
	ID    int     `json:"id"`
}

func round(val float64) int {
	return int(val)
}

func (eChart *exchangeChart) newBitfinexAPI(log slf.StructuredLogger) (*BitfinexAPI, error) {
	bitfinexAPI := &BitfinexAPI{rates: eChart.rates.exchangeBitfinex, log: log.WithField("api", "bitfinex")}

	c, err := newWebSocketConn(bitfinexAPIAddr, log)
	if err != nil {
		eChart.log.Errorf("new Bitfinex connection: %s", err.Error())
		c, err = reconnectWebSocketConn(gdaxAPIAddr, log)
		if err != nil {
			eChart.log.Errorf("Bitfinex connection: %s", err.Error())
			return nil, err
		}
	}

	bitfinexAPI.conn = c
	return bitfinexAPI, nil
}

func (bitfinex *BitfinexAPI) subscribe() {
	subscribtions := []string{
		`{"event":"subscribe","channel":"ticker","pair":"EOSUSD"}`,
		`{"event":"subscribe","channel":"ticker","pair":"BTCUSD"}`,
		`{"event":"subscribe","channel":"ticker","pair":"ETHUSD"}`,
	}

	bitfinex.conn.WriteMessage(websocket.TextMessage, []byte(subscribtions[0]))
	bitfinex.conn.WriteMessage(websocket.TextMessage, []byte(subscribtions[1]))
	bitfinex.conn.WriteMessage(websocket.TextMessage, []byte(subscribtions[2]))

	ids := map[int]string{} // id to ticker
	for i := 0; i < len(subscribtions); {
		_, message, err := bitfinex.conn.ReadMessage()
		var rateRaw interface{}
		err = json.Unmarshal(message, &rateRaw)
		if err != nil {
			bitfinex.log.Errorf("unmarshal error %s", err.Error())
			continue
		}

		switch rateRaw.(type) {
		case map[string]interface{}:
		case []interface{}:
			rawResponse := rateRaw.([]interface{})
			if len(rawResponse) > 2 {
				subTo := BitfinexSub{}
				json.Unmarshal([]byte(subscribtions[i]), &subTo)
				id := round(rawResponse[0].(float64))
				ids[id] = subTo.Pair

				switch subTo.Pair {
				case "EOSUSD":
					bitfinex.rates.exchange.EOStoUSD = rawResponse[7].(float64)
				case "BTCUSD":
					bitfinex.rates.exchange.BTCtoUSD = rawResponse[7].(float64)
					bitfinex.rates.exchange.USDtoBTC = 1 / rawResponse[7].(float64)
				case "ETHUSD":
					bitfinex.rates.exchange.ETHtoUSD = rawResponse[7].(float64)
				default:
					bitfinex.log.Warnf("unknown rate: %+v", rateRaw)
				}

				i++
			}
		}
	}
	bitfinex.indexes = ids

}

func (bitfinex *BitfinexAPI) listen() {
	time.Sleep(time.Second * 7)
	for {
		_, message, err := bitfinex.conn.ReadMessage()
		if err != nil {
			bitfinex.log.Errorf("read message: %s\n", err.Error())
			cn, err := reconnectWebSocketConn(bitfinexAPIAddr, bitfinex.log)
			if err != nil {
				bitfinex.log.Errorf("gdax reconnection: %s\n", err.Error())
			}
			bitfinex.conn = cn
			continue
		}
		// bitfinex.log.Warnf("--dsdsdsds---- %v ", string(message))

		var rateRaw interface{}
		err = json.Unmarshal(message, &rateRaw)
		if err != nil {
			bitfinex.log.Errorf("unmarshal error %s", err.Error())
			continue
		}

		switch rateRaw.(type) {
		case map[string]interface{}:

		case []interface{}:
			rawResponse := rateRaw.([]interface{})
			if len(rawResponse) > 2 {
				// bitfinex.log.Warnf("res %v , id %v\n", rawResponse[7], rawResponse[0])
				id := round(rawResponse[0].(float64))
				bitfinex.updateRate(&BitfinexPriceIndex{
					Price: rawResponse[7].(float64),
					ID:    id,
				})
			}

		}
	}
}

func (bitfinex *BitfinexAPI) updateRate(rateRaw *BitfinexPriceIndex) {

	bitfinex.rates.m.Lock()
	ticker := bitfinex.indexes[rateRaw.ID]
	switch ticker {
	case "EOSUSD":
		bitfinex.rates.exchange.EOStoUSD = rateRaw.Price
	case "BTCUSD":
		bitfinex.rates.exchange.BTCtoUSD = rateRaw.Price
		bitfinex.rates.exchange.USDtoBTC = 1 / rateRaw.Price
	case "ETHUSD":
		bitfinex.rates.exchange.ETHtoUSD = rateRaw.Price
	default:
		bitfinex.log.Warnf("unknown rate: %+v", rateRaw)
	}
	bitfinex.rates.m.Unlock()
}
