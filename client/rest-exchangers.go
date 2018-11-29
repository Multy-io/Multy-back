/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details

Currency implementation is abstractly support multiple exchangers functionality, so API user can compare
available exchanger services and choose any.
For not we support only Changelly exchanger 3rd party.
 */
package client

import (
	"github.com/Multy-io/Multy-back/exchanger"
	"github.com/gin-gonic/gin"
	"net/http"
)


func (restClient *RestClient) GetExchangerSupportedCurrencies() gin.HandlerFunc {
	return func(c *gin.Context) {
		changellyExchanger, _ := restClient.
			ExchangerFactory.
			GetExchanger(exchanger.ExchangeChangellyCanonicalName)
		supportedCurrencies, err := changellyExchanger.GetSupportedCurrencies()

		if err != nil {
			restClient.log.Errorf("An error occurred on GetSupportedCurrencies request: [%s]", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": msgErrHeaderError})
			return
		}

		var currencyNames []string
		for _, currency := range supportedCurrencies {
			currencyNames = append(currencyNames, currency.Name)
		}

		c.JSON(http.StatusOK, gin.H{
			"code": http.StatusOK,
			"message": http.StatusText(http.StatusOK),
			"currencies": currencyNames,
		})
	}
}

func (restClient *RestClient) GetExchangerAmountExchange() gin.HandlerFunc {
	return func(c *gin.Context) {
		type RequestGetExchangeAmount struct {
			From string			`json:"from"`
			To string			`json:"to"`
			Amount float64		`json:"amount"`
		}

		var requestData RequestGetExchangeAmount
		err := decodeBody(c, &requestData)
		if err != nil {
			restClient.log.Errorf("GetExchangerAmountExchange: Failed to decode body: %s\t[addr=%s]",
				err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": msgErrRequestBodyError})
			return
		}
		changellyExchanger, _ := restClient.
			ExchangerFactory.
			GetExchanger(exchanger.ExchangeChangellyCanonicalName)
		exchangeAmount, err := changellyExchanger.GetExchangeAmount(
			exchanger.CurrencyExchanger{ Name: requestData.From},
			exchanger.CurrencyExchanger{ Name: requestData.To},
			requestData.Amount,
		)

		if err != nil {
			restClient.log.Errorf("An error occurred on GetExchangerAmountExchange request: [%s]", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": msgErrServerError})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"message": http.StatusText(http.StatusOK),
				"amount": exchangeAmount,
			})
		}
	}
}

func (restClient *RestClient) CreateExchangerTransaction() gin.HandlerFunc {
	return func(c *gin.Context) {
		type RequestCreateTransaction struct {
			From string			`json:"from"`
			To string			`json:"to"`
			Amount float64		`json:"amount"`
			Address string		`json:"address"`
		}

		var requestData RequestCreateTransaction
		err := decodeBody(c, &requestData)
		if err != nil {
			restClient.log.Errorf("CreateExchangerTransaction: Failed to decode body: %s\t[addr=%s]",
				err.Error(), c.Request.RemoteAddr)
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": msgErrRequestBodyError})
			return
		}

		changellyExchanger, _ := restClient.
			ExchangerFactory.
			GetExchanger(exchanger.ExchangeChangellyCanonicalName)
		transaction, err := changellyExchanger.CreateTransaction(
			exchanger.CurrencyExchanger{ Name: requestData.From},
			exchanger.CurrencyExchanger{ Name: requestData.To},
			requestData.Amount,
			requestData.Address,
		)

		if err != nil {
			restClient.log.Errorf("An error occurred on CreateExchangerTransaction request: [%s]", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": msgErrServerError})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"message": http.StatusText(http.StatusOK),
				"transactionId": transaction.Id,
				"payinAddress": transaction.PayInAddress,
				"payoutAddress": transaction.PayOutAddress,
			})
		}
	}
}
