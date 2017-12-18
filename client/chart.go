package client

import (
	"math/rand"
	"sync"
	"time"

	"github.com/ventu-io/slf"
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
	BTCtoUSDDay    map[string]float64
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
			BTCtoUSDDay:    make(map[string]float64),
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
	eChart.updateAll()
	eChart.ticker = time.NewTicker(time.Duration(eChart.interval) * time.Second)
	eChart.log.Debugf("updateExchange: ticker=%ds", eChart.interval)

	for {
		select {
		case _ = <-eChart.ticker.C:
			eChart.update()
		}
	}
}

func (eChart *exchangeChart) update() {
	eChart.log.Debug("updateExchange; mock implementation")

	eChart.rates.m.Lock()
	defer eChart.rates.m.Unlock()

	eChart.rates.exchangeSingle.ETHtoBTC = r1.Float64()*5 + 5
	eChart.rates.exchangeSingle.USDtoBTC = r1.Float64()*5 + 5
	eChart.rates.exchangeSingle.ETHtoBTC = r1.Float64()*5 + 5

	eChart.rates.exchangeSingle.ETHtoUSD = r1.Float64()*5 + 5
	eChart.rates.exchangeSingle.ETHtoEUR = r1.Float64()*5 + 5

	// TODO: do it gracefullcy
	theOldest, theNewest := eChart.getExtremRates(eChart.rates.BTCtoUSDDay)
	delete(eChart.rates.BTCtoUSDDay, theOldest.Format(time.RFC3339))
	eChart.rates.BTCtoUSDDay[theNewest.Add(time.Duration(eChart.interval)*time.Second).Format(time.RFC3339)] = r1.Float64()*5 + 5

	eChart.rates.exchangeSingle.BTCtoUSD = eChart.rates.BTCtoUSDDay[theNewest.Add(time.Duration(eChart.interval)*time.Second).Format(time.RFC3339)]

	return
}

func (eChart *exchangeChart) updateAll() {
	eChart.log.Debug("updateExchange; mock implementation")

	aDayAgoTime := time.Now()
	aDayAgoTime.AddDate(0, 0, -1)

	for i := 0; i < numOfChartDots; i += eChart.interval {
		timeInString := aDayAgoTime.Add(-time.Second * time.Duration(i)).Format(time.RFC3339)
		eChart.rates.BTCtoUSDDay[timeInString] = r1.Float64()*5 + 5
	}

	eChart.log.Debugf("updateRateAll: BTCtoUSDDay=%+v", eChart.rates.BTCtoUSDDay)
	return
}

func (eChart *exchangeChart) getAll() map[string]float64 {
	eChart.log.Debug("exchange chart: get all exchanges")

	eChart.rates.m.Lock()
	defer eChart.rates.m.Unlock()
	return eChart.rates.BTCtoUSDDay
}

func (eChart *exchangeChart) getLast() *EventExchangeChart {
	eChart.log.Debug("exchange chart: get last exchanges")

	eChart.rates.m.Lock()
	defer eChart.rates.m.Unlock()
	return eChart.rates.exchangeSingle
}

func (eChart *exchangeChart) getExtremRates(rates map[string]float64) (time.Time, time.Time) {
	var min, max time.Time
	for rt := range rates {
		t, err := time.Parse(time.RFC3339, rt)
		if err != nil {
			eChart.log.Errorf("parse string time to time: %s", err.Error())
			return time.Now(), time.Now()
		}
		if t.Unix() <= min.Unix() {
			min = t
		}
		if t.Unix() > max.Unix() {
			max = t
		}
	}
	return min, max
}
