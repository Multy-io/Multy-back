package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/KristinaEtc/slf"
)

var (
	s1 = rand.NewSource(time.Now().UnixNano())
	r1 = rand.New(s1)
)

const (
	secondsInDay   = 8640
	numOfChartDots = 1200 //12 minutes

	defaultNSQAddr = "127.0.0.1:4150"
)

type Rates struct {
	BTCtoUSDDay    []RatesFromApi
	exchangeSingle *EventExchangeChart

	m *sync.Mutex
}

type exchangeChart struct {
	rates *Rates

	ticker   *time.Ticker
	interval int
	log      slf.StructuredLogger
}

type EventExchangeChart struct {
	EURtoBTC float64
	USDtoBTC float64
	ETHtoBTC float64

	ETHtoUSD float64
	ETHtoEUR float64

	BTCtoUSD float64
}

func initExchangeChart() (*exchangeChart, error) {
	chart := &exchangeChart{
		rates: &Rates{
			exchangeSingle: &EventExchangeChart{},
			BTCtoUSDDay:    make([]RatesFromApi, 0),
			m:              &sync.Mutex{},
		},
		log:      slf.WithContext("chart"),
		interval: secondsInDay / numOfChartDots,
	}
	chart.log.Debug("initExchangeChart")

	go chart.run()

	return chart, nil

}

func (eChart *exchangeChart) run() error {
	eChart.log.Debug("exchange chart: run")
	eChart.getsAllFromAPI()
	eChart.ticker = time.NewTicker(time.Duration(eChart.interval) * time.Second)
	eChart.log.Debugf("updateExchange: ticker=%ds", eChart.interval)

	tickerHour := time.NewTicker(time.Hour)

	for {
		select {
		case _ = <-eChart.ticker.C:
			eChart.update()

		case _ = <-tickerHour.C:
			eChart.updateAll()
		}
	}
}

func (eChart *exchangeChart) updateAll() {
	eChart.rates.m.Lock()
	defer eChart.rates.m.Unlock()

	log.Println("not implemented for json array")
	/*
		min, max := eChart.getExtremRates(eChart.rates.BTCtoUSDDay)
		delete(eChart.rates.BTCtoUSDDay, min)
		eChart.rates.BTCtoUSDDay[max] = strconv.FormatFloat(eChart.rates.exchangeSingle.USDtoBTC, 'f', 2, 64)
	*/
	return
}

func (eChart *exchangeChart) update() {
	eChart.rates.m.Lock()
	defer eChart.rates.m.Unlock()

	exchangeSingle, err := eChart.getUpdatedRated()
	if err != nil {
		eChart.log.Errorf("coult not update rates: %s", err.Error())
		return
	}

	eChart.rates.exchangeSingle = exchangeSingle
	return
}

func (eChart *exchangeChart) getUpdatedRated() (*EventExchangeChart, error) {
	//log.Println("getUpdatedRated")

	reqURI := fmt.Sprintf("https://min-api.cryptocompare.com/data/price?fsym=%s&tsyms=%s", "BTC", "EUR,USD,ETH")
	_, err := url.ParseRequestURI(reqURI)
	if err != nil {
		log.Printf("[ERR] processGetExchangeEvent: wrong reqURI: [%s], %s\n", reqURI, err.Error())
		return nil, err
	}

	//log.Printf("[DEBUG] processGetExchangeEvent: reqURI=%s", reqURI)
	resp, err := http.Get(reqURI)
	if err != nil {
		log.Printf("[ERR] processGetExchangeEvent: get exchange: [%s], err=%s\n", reqURI, err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERR] processGetExchangeEvent: get exchange: response status code=%d\n", resp.StatusCode)
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERR] processGetExchangeEvent: get exchange: get response body: %s\n", err.Error())
		return nil, err
	}

	//log.Printf("[DEBUG] processGetExchangeEvent: resp=[%s]\n", string(bodyBytes))

	var ratesRaw map[string]float64
	if err := json.Unmarshal(bodyBytes, &ratesRaw); err != nil {
		log.Printf("[ERR] processGetExchangeEvent: parse responce=%s\n", err.Error())
		return nil, err
	}

	rates := &EventExchangeChart{
		EURtoBTC: 1 / ratesRaw["EUR"],
		USDtoBTC: 1 / ratesRaw["USD"],
		ETHtoBTC: 1 / ratesRaw["ETH"],
	}
	rates.ETHtoUSD = rates.ETHtoBTC / ratesRaw["USD"]
	rates.ETHtoEUR = rates.ETHtoBTC / ratesRaw["EUR"]

	rates.BTCtoUSD = ratesRaw["USD"]
	//	log.Printf("[DEBUG] rates=[%+v]\n", rates)

	return rates, nil
}

type RatesFromApi struct {
	Date  string `json:"date"`
	Price string `json:"price"`
}

func (eChart *exchangeChart) getsAllFromAPI() {
	eChart.log.Debug("updateAll")

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

	var ratesAll = make([]RatesFromApi, 0)
	err = json.Unmarshal(bodyBytes, &ratesAll)
	if err != nil {
		eChart.log.Errorf("%s\n", err.Error())
		return
	}

	eChart.log.Debugf("getsAllFromAPI=[%v]", ratesAll)
	eChart.saveRates(ratesAll)

	return
}

func (eChart *exchangeChart) saveRates(allRates []RatesFromApi) {
	eChart.rates.m.Lock()
	defer eChart.rates.m.Unlock()

	/*for _, rate := range allRates {
		eChart.rates.BTCtoUSDDay[rate.Date] = rate.Price
	}*/
	eChart.rates.BTCtoUSDDay = allRates
}

func (eChart *exchangeChart) getAll() []RatesFromApi {
	eChart.log.Debug("exchange chart: get all exchanges")

	eChart.rates.m.Lock()
	defer eChart.rates.m.Unlock()

	return eChart.rates.BTCtoUSDDay
}

func (eChart *exchangeChart) getLast() *EventExchangeChart {
	//eChart.log.Debug("exchange chart: get last exchanges")

	eChart.rates.m.Lock()
	defer eChart.rates.m.Unlock()
	return eChart.rates.exchangeSingle
}

func (eChart *exchangeChart) getExtremRates(rates map[string]string) (string, string) {
	var min, max time.Time
	for rt := range rates {
		t, err := time.Parse(time.RFC3339, rt)
		if err != nil {
			eChart.log.Errorf("parse string time to time: %s", err.Error())
			return "", ""
		}
		if t.Unix() <= min.Unix() {
			min = t
		}
		if t.Unix() > max.Unix() {
			max = t
		}
	}

	return min.String(), max.String()
}
