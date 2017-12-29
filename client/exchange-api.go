package client

import (
	"encoding/json"
	"log"
	"net/url"
	"strconv"

	"github.com/KristinaEtc/slf"
	"github.com/gorilla/websocket"
)

const (
	exchangeDdax     = "Gdax"
	exchangePoloniex = "Poloniex"
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

func (eChart *exchangeChart) initGdaxAPI(log slf.StructuredLogger) (*GdaxAPI, error) {
	u := url.URL{Scheme: "wss", Host: "ws-feed.gdax.com", Path: ""}

	gdaxAPI := &GdaxAPI{rates: eChart.rates.exchangeGdax, log: log.WithField("api", "gdax")}
	gdaxAPI.log.Infof("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		gdaxAPI.log.Errorf("Dial: %s", err.Error())
		return nil, err
	}

	// TODO: Add Close() method
	// defer c.Close()
	// done := make(chan struct{})
	// defer ticker.Stop()

	// defer c.Close()
	// defer close(done)
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
		//gdax.log.Errorf("parseFloat %s: %s", rawRate.Price, err.Error())
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
	} `json:"data"`
}

func (eChart *exchangeChart) initPoloniexAPI(log slf.StructuredLogger) (*PoloniexAPI, error) {
	u := url.URL{Scheme: "ws", Host: "api.hitbtc.com", Path: "api/2/ws"}

	poloniexAPI := &PoloniexAPI{rates: eChart.rates.exchangePoloniex, log: log.WithField("api", "poloniex")}
	poloniexAPI.log.Infof("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		poloniexAPI.log.Errorf("Dial: %s", err.Error())
		return nil, err
	}

	poloniexAPI.conn = c
	//defer c.Close()
	//done := make(chan struct{})
	//defer cP.Close()
	//defer close(done)

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
			poloniex.log.Errorf("Read message: %s", err.Error())
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
