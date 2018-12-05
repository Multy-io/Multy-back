/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package client

import (
	"fmt"
	"math/rand"
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
	saveToDBInterval = time.Second * 10
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
		log: slf.WithContext("chart").WithCaller(slf.CallerShort),
	}
	chart.log.Debug("new exchange chart")

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
	eChart.db.InsertExchangeRate(*eChart.rates.exchangeGdax.exchange, exchangeDdax)
	eChart.rates.exchangeGdax.m.Unlock()

	eChart.rates.exchangePoloniex.m.Lock()
	eChart.db.InsertExchangeRate(*eChart.rates.exchangePoloniex.exchange, exchangePoloniex)
	eChart.rates.exchangePoloniex.m.Unlock()
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
