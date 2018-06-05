/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package client

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/jekabolt/slf"
	"github.com/gorilla/websocket"
)

const (
	exchangeDdax     = "Gdax"
	exchangePoloniex = "Poloniex"

	backOffLimit = time.Duration(time.Second * 600) // reconnection stop
)

const (
	poloniexAPIAddr = "wss://api.hitbtc.com/api/2/ws"
	gdaxAPIAddr     = "wss://ws-feed.gdax.com"
)

// GdaxAPI wraps websocket connection for GdaxAPI
type GdaxAPI struct {
	conn  *websocket.Conn
	rates *StockRate

	log slf.StructuredLogger
}

//GDAXSocketEvent is a GDAX json parser structure
type GDAXSocketEvent struct {
	ProductID string `json:"product_id"`
	Price     string `json:"price"`
}

func (eChart *exchangeChart) newGdaxAPI(log slf.StructuredLogger) (*GdaxAPI, error) {
	gdaxAPI := &GdaxAPI{rates: eChart.rates.exchangeGdax, log: log.WithField("api", "gdax")}

	c, err := newWebSocketConn(gdaxAPIAddr)
	if err != nil {
		eChart.log.Errorf("new gdax connection: %s", err.Error())
		c, err = reconnectWebSocketConn(gdaxAPIAddr, log)
		if err != nil {
			eChart.log.Errorf("gdax reconnection: %s", err.Error())
			return nil, err
		}
	}

	gdaxAPI.conn = c

	return gdaxAPI, nil
}

func (gdax *GdaxAPI) subscribe() {
	subscribtion := `{"type":"subscribe","channels":[{"name":"ticker_1000","product_ids":["BTC-USD","BTC-EUR","ETH-BTC","ETH-USD","ETH-EUR"]}]}`
	gdax.conn.WriteMessage(websocket.TextMessage, []byte(subscribtion))
}

func (gdax *GdaxAPI) listen() {
	for {
		_, message, err := gdax.conn.ReadMessage()
		if err != nil {
			gdax.log.Errorf("read message: %s", err.Error())
			c, err := reconnectWebSocketConn(gdaxAPIAddr, gdax.log)
			if err != nil {
				gdax.log.Errorf("gdax reconnection: %s", err.Error())
				return
			}
			gdax.conn = c
			continue
		}

		rateRaw := &GDAXSocketEvent{}
		err = json.Unmarshal(message, rateRaw)
		if err != nil {
			gdax.log.Errorf("unmarshal error %s", err.Error())
			continue
		}

		gdax.updateRate(rateRaw)
	}
}

func (gdax *GdaxAPI) updateRate(rawRate *GDAXSocketEvent) {
	floatPrice, err := strconv.ParseFloat(rawRate.Price, 32)
	if err != nil {
		gdax.log.Errorf("parseFloat %s: %s", rawRate.Price, err.Error())
		return
	}

	gdax.rates.m.Lock()

	switch rawRate.ProductID {
	case "BTC-USD":
		gdax.rates.exchange.BTCtoUSD = floatPrice
		gdax.rates.exchange.USDtoBTC = 1 / floatPrice
	case "BTC-EUR":
		gdax.rates.exchange.EURtoBTC = 1 / floatPrice
	case "ETH-BTC":
		gdax.rates.exchange.ETHtoBTC = floatPrice
		//gdax.rates.exchange.BTCtoETH = 1 / floatPrice
	case "ETH-USD":
		gdax.rates.exchange.ETHtoUSD = floatPrice
	case "ETH-EUR":
		gdax.rates.exchange.ETHtoEUR = floatPrice
	default:
		gdax.log.Warnf("unknown rate: %+v", rawRate)
	}

	gdax.rates.m.Unlock()
}

// PoloniexAPI wraps websocket connection for PoloniexAPI
type PoloniexAPI struct {
	conn  *websocket.Conn
	rates *StockRate

	log slf.StructuredLogger
}

// PoloniexSocketEvent is a Poloniex json parser structure
type PoloniexSocketEvent struct {
	Data struct {
		Price     string `json:"last"`
		ProductID string `json:"symbol"`
	} `json:"params"`
}

func (eChart *exchangeChart) newPoloniexAPI(log slf.StructuredLogger) (*PoloniexAPI, error) {
	poloniexAPI := &PoloniexAPI{rates: eChart.rates.exchangePoloniex, log: log.WithField("api", "poloniex")}

	c, err := newWebSocketConn(poloniexAPIAddr)
	if err != nil {
		eChart.log.Errorf("new poloniex connection: %s", err.Error())
		c, err = reconnectWebSocketConn(gdaxAPIAddr, log)
		if err != nil {
			eChart.log.Errorf("poloniex connection: %s", err.Error())
			return nil, err
		}
	}

	poloniexAPI.conn = c

	return poloniexAPI, nil
}

func (poloniex *PoloniexAPI) subscribe() {
	subscribtionBTCUSD := `{"method":"subscribeTicker","params":{"symbol": "BTCUSD"},"id": 10000}`
	subscribtionETHBTC := `{"method":"subscribeTicker","params":{"symbol":"ETHBTC"},"id": 10000}`
	subscribtionETHUSD := `{"method":"subscribeTicker","params":{"symbol":"ETHUSD"},"id": 10000}`

	poloniex.conn.WriteMessage(websocket.TextMessage, []byte(subscribtionBTCUSD))
	poloniex.conn.WriteMessage(websocket.TextMessage, []byte(subscribtionETHBTC))
	poloniex.conn.WriteMessage(websocket.TextMessage, []byte(subscribtionETHUSD))
}

func (poloniex *PoloniexAPI) listen() {
	for {
		_, message, err := poloniex.conn.ReadMessage()
		if err != nil {
			poloniex.log.Errorf("read message: %s", err.Error())
			c, err := reconnectWebSocketConn(poloniexAPIAddr, poloniex.log)
			if err != nil {
				poloniex.log.Errorf("gdax reconnection: %s", err.Error())
				return
			}
			poloniex.conn = c
			continue
		}

		rateRaw := PoloniexSocketEvent{}
		err = json.Unmarshal(message, &rateRaw)
		if err != nil {
			log.Printf("error marshaling: %s/n", err.Error())
			continue
		}
		poloniex.updateRate(&rateRaw)
	}
}

func (poloniex *PoloniexAPI) updateRate(rateRaw *PoloniexSocketEvent) {
	floatPrice, err := strconv.ParseFloat(rateRaw.Data.Price, 32)
	if err != nil {
		// seems like here would be ttl messages, which is empty structures
		return
	}

	poloniex.rates.m.Lock()
	switch rateRaw.Data.ProductID {
	case "BTCUSD":
		poloniex.rates.exchange.BTCtoUSD = floatPrice
		poloniex.rates.exchange.USDtoBTC = 1 / floatPrice
	case "ETHBTC":
		poloniex.rates.exchange.ETHtoBTC = floatPrice
	case "ETHUSD":
		poloniex.rates.exchange.ETHtoUSD = floatPrice
	default:
		poloniex.log.Warnf("unknown rate: %+v", rateRaw)
	}
	poloniex.rates.m.Unlock()
}
