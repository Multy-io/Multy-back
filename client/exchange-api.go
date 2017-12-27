package client

import (
	"encoding/json"
	"log"
	"net/url"
	"strconv"

	"github.com/KristinaEtc/slf"
	"github.com/gorilla/websocket"
)

// GdaxAPI wraps websocket connection for GdaxAPI.
type GdaxAPI struct {
	conn  *websocket.Conn
	rates *Rates

	log slf.StructuredLogger
}

//GDAXSocketEvent is a GDAX json parser structure
type GDAXSocketEvent struct {
	ProductID string `json:"product_id"`
	Price     string `json:"price"`
}

func (eChart *exchangeChart) initGdaxAPI(log slf.StructuredLogger) (*GdaxAPI, error) {
	u := url.URL{Scheme: "wss", Host: "ws-feed.gdax.com", Path: ""}

	gdaxAPI := &GdaxAPI{rates: eChart.rates, log: log.WithField("api", "gdax")}
	gdaxAPI.log.Infof("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		gdaxAPI.log.Errorf("dial: %s", err.Error())
		return nil, err
	}

	// TODO: Add Close() method
	// defer c.Close()
	// done := make(chan struct{})
	// defer ticker.Stop()

	// defer c.Close()
	// defer close(done)

	subscribtion := `{"type":"subscribe","channels":[{"name":"ticker_1000","product_ids":["BTC-USD","BTC-EUR","ETH-BTC","ETH-USD","ETH-EUR"]}]}`
	c.WriteMessage(websocket.TextMessage, []byte(subscribtion))

	gdaxAPI.conn = c

	return gdaxAPI, nil
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
		gdax.log.Errorf("parseFloat %s: %s", rawRate.Price, err.Error())
		return
	}

	gdax.rates.m.Lock()

	switch rawRate.ProductID {
	case "BTC-USD":
		gdax.rates.exchangeSingle.BTCtoUSD = floatPrice
		gdax.rates.exchangeSingle.USDtoBTC = 1 / floatPrice
	case "BTC-EUR":
		gdax.rates.exchangeSingle.EURtoBTC = 1 / floatPrice
	case "ETH-BTC":
		gdax.rates.exchangeSingle.ETHtoBTC = floatPrice
		//gdax.rates.exchangeSingle.BTCtoETH = 1 / floatPrice
	case "ETH-USD":
		gdax.rates.exchangeSingle.ETHtoUSD = floatPrice
	case "ETH-EUR":
		gdax.rates.exchangeSingle.ETHtoEUR = floatPrice
	default:
		log.Printf("unknown rate: %+v\n", rawRate)
	}

	gdax.rates.m.Unlock()
}
