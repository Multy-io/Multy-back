package client

import (
	"log"
	"testing"
)

func Test_processGetExchangeEvent(t *testing.T) {
	ech, err := initExchangeChart()
	if err != nil {
		t.Fatal(err.Error())
	}

	resp, err := ech.getUpdatedRated()
	if err != nil {
		t.Fatal(err.Error())
	}

	log.Printf("%+v", resp)
	//t.Fatal("")
}
