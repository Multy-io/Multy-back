/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/Multy-io/Multy-back/store"
	"github.com/jekabolt/slf"
)

var (
	s1 = rand.NewSource(time.Now().UnixNano())
	r1 = rand.New(s1)
)

const (
	saveToDBInterval       = time.Second * 10
	updateForExchangeChart = time.Hour
)

// StockRate stores rates from specific stock and protected with mutex
type StockRate struct {
	exchange *store.ExchangeRates
	m        *sync.Mutex
}

// Rates stores rates from all supported stocks
type Rates struct {
	BTCtoUSDDay []store.RatesAPIBitstamp
	mDay        *sync.Mutex

	exchangeGdax     *StockRate
	exchangePoloniex *StockRate
}

type exchangeChart struct {
	rates        *Rates
	gdaxConn     *GdaxAPI
	poloniexConn *PoloniexAPI

	db  store.UserStore
	log slf.StructuredLogger
}

func newExchangeChart(db store.UserStore) (*exchangeChart, error) {
	chart := &exchangeChart{
		rates: &Rates{
			exchangeGdax: &StockRate{
				exchange: &store.ExchangeRates{},
				m:        &sync.Mutex{},
			},
			exchangePoloniex: &StockRate{
				exchange: &store.ExchangeRates{},
				m:        &sync.Mutex{},
			},

			BTCtoUSDDay: make([]store.RatesAPIBitstamp, 0),
			mDay:        &sync.Mutex{},
		},
		db:  db,
		log: slf.WithContext("chart"),
	}
	chart.log.Debug("new exchange chart")

	//moved to next release
	//chart.getDayAPIBitstamp()

	gDaxConn, err := chart.newGdaxAPI(chart.log)
	if err != nil {
		return nil, fmt.Errorf("initGdaxAPI: %s", err)
	}
	chart.gdaxConn = gDaxConn

	poloniexConn, err := chart.newPoloniexAPI(chart.log)
	if err != nil {
		return nil, fmt.Errorf("initPoloniexAPI: %s", err)
	}
	chart.poloniexConn = poloniexConn

	go chart.run()

	return chart, nil
}

func (eChart *exchangeChart) run() error {
	tickerSaveToDB := time.NewTicker(saveToDBInterval)

	go eChart.gdaxConn.listen()
	//TODO: fix poloniex
	go eChart.poloniexConn.listen()

	eChart.gdaxConn.subscribe()
	eChart.poloniexConn.subscribe()

	for {
		select {
		case _ = <-tickerSaveToDB.C:
			eChart.saveToDB()
		}
	}
}

func (eChart *exchangeChart) saveToDB() {
	eChart.rates.exchangeGdax.m.Lock()
	//eChart.log.Debugf("gdax=%+v", eChart.rates.exchangeGdax.exchange)
	eChart.db.InsertExchangeRate(*eChart.rates.exchangeGdax.exchange, exchangeDdax)
	eChart.rates.exchangeGdax.m.Unlock()

	eChart.rates.exchangePoloniex.m.Lock()
	eChart.db.InsertExchangeRate(*eChart.rates.exchangePoloniex.exchange, exchangePoloniex)
	eChart.rates.exchangePoloniex.m.Unlock()
}

func (eChart *exchangeChart) updateDayRates() {
	eChart.rates.mDay.Lock()
	_, err := eChart.db.GetExchangeRatesDay()
	if err != nil {
		eChart.log.Errorf("GetExchangeRatesDay: %s", err.Error())
		return
	}
	eChart.rates.mDay.Unlock()
}

func (eChart *exchangeChart) getDayAPIBitstamp() {
	eChart.log.Debug("update rates day")

	reqURI := "https://www.bitstamp.net/api/transactions/?date=hour"

	eChart.log.Debugf("reqURI=%s", reqURI)
	resp, err := http.Get(reqURI)
	if err != nil {
		eChart.log.Errorf("[%s], err=%s", reqURI, err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		eChart.log.Errorf("get exchange: response status code=%d", resp.StatusCode)
		return
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		eChart.log.Errorf("get exchange: get response body: %s", err.Error())
		return
	}

	var ratesAll = make([]store.RatesAPIBitstamp, 0)
	err = json.Unmarshal(bodyBytes, &ratesAll)
	if err != nil {
		eChart.log.Errorf("%s\n", err.Error())
		return
	}

	eChart.log.Debugf("rates 24h=[%v]", ratesAll)

	eChart.rates.mDay.Lock()
	eChart.rates.BTCtoUSDDay = ratesAll
	eChart.rates.mDay.Unlock()
}

func (eChart *exchangeChart) getExchangeDay() []store.RatesAPIBitstamp {
	eChart.log.Debug("exchange chart: get exchanges for 24 hours")

	eChart.rates.mDay.Lock()
	defer eChart.rates.mDay.Unlock()

	/*for _, k := range eChart.rates.BTCtoUSDDay {
		i, _ := strconv.Atoi(k.Date)
		log.Println(time.Unix(int64(i), 0).Format(time.RFC3339), "=", k.Price)
	}*/
	return eChart.rates.BTCtoUSDDay
}

func (eChart *exchangeChart) getExchangeGdax() *store.ExchangeRates {
	eChart.rates.exchangeGdax.m.Lock()
	defer eChart.rates.exchangeGdax.m.Unlock()

	return eChart.rates.exchangeGdax.exchange
}

func (eChart *exchangeChart) getExchangePoloniex() *store.ExchangeRates {
	eChart.rates.exchangePoloniex.m.Lock()
	defer eChart.rates.exchangePoloniex.m.Unlock()

	return eChart.rates.exchangePoloniex.exchange
}
