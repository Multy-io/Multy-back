package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/Appscrunch/Multy-back/store"
	"github.com/KristinaEtc/slf"
)

var (
	s1 = rand.NewSource(time.Now().UnixNano())
	r1 = rand.New(s1)
)

const (
	//secondsInDay   = 8640
	//numOfChartDots = 1200 //12 minutes

	saveToDBInterval       = time.Second * 60
	updateForExchangeChart = time.Hour

	defaultNSQAddr = "127.0.0.1:4150"
)

type Rates struct {
	BTCtoUSDDay    []RatesAPIBitstamp
	exchangeSingle *store.ExchangeRates

	m *sync.Mutex
}

type exchangeChart struct {
	rates    *Rates
	gdaxConn *GdaxAPI

	db store.UserStore

	log slf.StructuredLogger
}

func initExchangeChart(db store.UserStore) (*exchangeChart, error) {
	chart := &exchangeChart{
		rates: &Rates{
			exchangeSingle: &store.ExchangeRates{},
			BTCtoUSDDay:    make([]RatesAPIBitstamp, 0),
			m:              &sync.Mutex{},
		},
		db:  db,
		log: slf.WithContext("chart"),
	}
	chart.log.Debug("initExchangeChart")
	chart.getAllAPIBitstamp()

	gDaxConn, err := chart.initGdaxAPI(chart.log)
	if err != nil {
		return nil, fmt.Errorf("initGdaxAPI: %s", err)
	}
	chart.gdaxConn = gDaxConn

	go chart.run()

	return chart, nil
}

func (eChart *exchangeChart) saveToDB() {
	eChart.rates.m.Lock()
	eChart.rates.m.Unlock()

	eChart.db.InsertExchangeRate(*eChart.rates.exchangeSingle)
}

func (eChart *exchangeChart) run() error {

	tickerUpdateForChart := time.NewTicker(updateForExchangeChart)
	tickerSaveToDB := time.NewTicker(saveToDBInterval)

	go eChart.gdaxConn.listen()

	for {
		select {
		case _ = <-tickerSaveToDB.C:
			eChart.saveToDB()
		case _ = <-tickerUpdateForChart.C:
			eChart.getAllAPIBitstamp()
		}
	}
}

type RatesAPIBitstamp struct {
	Date  string `json:"date"`
	Price string `json:"price"`
}

func (eChart *exchangeChart) getAllAPIBitstamp() {
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

	var ratesAll = make([]RatesAPIBitstamp, 0)
	err = json.Unmarshal(bodyBytes, &ratesAll)
	if err != nil {
		eChart.log.Errorf("%s\n", err.Error())
		return
	}

	eChart.log.Debugf("getAllAPIBitstamp=[%v]", ratesAll)

	eChart.rates.m.Lock()
	eChart.rates.BTCtoUSDDay = ratesAll
	eChart.rates.m.Unlock()

	return
}

func (eChart *exchangeChart) getAllDay() []RatesAPIBitstamp {
	eChart.log.Debug("exchange chart: get all exchanges")

	eChart.rates.m.Lock()
	defer eChart.rates.m.Unlock()

	return eChart.rates.BTCtoUSDDay
}

func (eChart *exchangeChart) getLast() *store.ExchangeRates {
	eChart.rates.m.Lock()
	defer eChart.rates.m.Unlock()
	return eChart.rates.exchangeSingle
}
