package client

import (
	"log"
	"testing"
)

func Test_processGetExchangeEvent(t *testing.T) {
	req := EventGetExchangeReq{
		From: "RUB",
		To:   "BTC",
	}
	resp := processGetExchangeEvent(req)
	if resp.Error != "" {
		t.Fatal(resp.Error)
	}
	log.Printf("%+v", resp)
	//t.Fatal("")
}
