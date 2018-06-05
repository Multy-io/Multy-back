/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package client

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Appscrunch/Multy-back/store"
	fcm "github.com/NaySoftware/go-fcm"
	"github.com/gin-gonic/gin"
	"github.com/jekabolt/slf"

	"github.com/nsqio/go-nsq"
)

type FirebaseConf struct {
	ServerKey string
}

type FirebaseClient struct {
	conf   *FirebaseConf
	client *fcm.FcmClient

	nsqConsumer *nsq.Consumer
	nsqConfig   *nsq.Config

	log slf.StructuredLogger
}

func InitFirebaseConn(conf *FirebaseConf, c *gin.Engine, nsqAddr string) (*FirebaseClient, error) {
	fClient := &FirebaseClient{
		conf:      conf,
		client:    fcm.NewFcmClient(conf.ServerKey),
		nsqConfig: nsq.NewConfig(),

		log: slf.WithContext("firebase"),
	}
	fClient.log.Info("Firebase connection initialization")
	fClient.log.Debugf("Firebase cert len =%d", len(fClient.conf.ServerKey))

	nsqConsumer, err := nsq.NewConsumer(store.TopicTransaction, "firebase", fClient.nsqConfig)
	if err != nil {
		return nil, fmt.Errorf("new nsq consumer: %s", err.Error())
	}
	nsqConsumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		msgRaw := message.Body
		// fClient.log.Debugf("firebase new transaction notify: %+v", string(msgRaw))

		msg := store.TransactionWithUserID{}
		err := json.Unmarshal(msgRaw, &msg)
		if err != nil {
			return err
		}

		data := map[string]string{
			"txid":            msg.NotificationMsg.TxID,
			"transactionType": strconv.Itoa(msg.NotificationMsg.TransactionType),
			// "amount":          strconv.FormatFloat(msg.NotificationMsg.Amount, 'f', 3, 64),
			"amount":  msg.NotificationMsg.Amount,
			"address": msg.NotificationMsg.Address,
		}
		// fClient.log.Debugf("data=%+v", data)

		// TODO: add version /v1
		fClient.client.NewFcmMsgTo("/topics/"+store.TopicTransaction+"-"+msg.UserID, data) //
		status, err := fClient.client.Send()
		if err != nil {
			return err
		}
		status.PrintResults()
		// TODO: add version /v1
		fClient.client.NewFcmMsgTo(store.TopicTransaction+"-"+msg.UserID, data) //
		status, err = fClient.client.Send()
		if err != nil {
			return err
		}
		status.PrintResults()

		return nil
	}))

	if err = nsqConsumer.ConnectToNSQD(nsqAddr); err != nil {
		return nil, fmt.Errorf("connecting to nsq: %s", err.Error())
	}
	fClient.log.Debugf("Firebase connection initialization done")
	return fClient, nil
}
