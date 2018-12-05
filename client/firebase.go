/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/Multy-io/Multy-back/currencies"
	"github.com/Multy-io/Multy-back/store"
	"github.com/gin-gonic/gin"
	"github.com/jekabolt/slf"
	"google.golang.org/api/option"

	"github.com/nsqio/go-nsq"
)

type FirebaseConf struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

type FirebaseClient struct {
	conf *FirebaseConf
	// client *fcm.FcmClient
	app *firebase.App

	nsqConsumer *nsq.Consumer
	nsqConfig   *nsq.Config

	log slf.StructuredLogger
}

func InitFirebaseConn(conf *FirebaseConf, c *gin.Engine, nsqAddr string) (*FirebaseClient, error) {
	fClient := &FirebaseClient{
		conf: conf,
		// client:    fcm.NewFcmClient(conf.ServerKey),
		nsqConfig: nsq.NewConfig(),

		log: slf.WithContext("firebase").WithCaller(slf.CallerShort),
	}
	fClient.log.Info("Firebase connection initialization")

	service, err := NewPushService("./multy.config")
	if err != nil {
		return nil, fmt.Errorf("NewPushService: %s", err.Error())
	}

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
		txType := msg.NotificationMsg.TransactionType
		// if txType == store.TxStatusAppearedInMempoolIncoming || txType == store.TxStatusAppearedInBlockIncoming || txType == store.TxStatusInBlockConfirmedIncoming {
		if txType == store.TxStatusAppearedInMempoolIncoming {
			topic := store.TopicTransaction + "-" + msg.UserID
			// topic := "btcTransactionUpdate-" + msg.UserID
			// topic := "btcTransactionUpdate-003b1e5227ce5f45b22676dc4b55ea00e1410c5f3cf8ae972724fa5d93ecc4585e"

			messageKeys := map[string]string{
				"score":           "1",
				"time":            time.Now().Format(time.Kitchen),
				"amount":          msg.NotificationMsg.Amount,
				"transactionType": strconv.Itoa(msg.NotificationMsg.TransactionType),
				"currencyid":      strconv.Itoa(msg.NotificationMsg.CurrencyID),
				"networkid":       strconv.Itoa(msg.NotificationMsg.NetworkID),
				"walletindex":     strconv.Itoa(msg.NotificationMsg.WalletIndex),
				"txid":            msg.NotificationMsg.TxID,
			}

			messageToSend := &messaging.Message{
				Data: messageKeys,
				APNS: &messaging.APNSConfig{
					Payload: &messaging.APNSPayload{
						Aps: &messaging.Aps{
							Alert: &messaging.ApsAlert{
								Title: "",
								// Body:  msg.NotificationMsg.Amount + " " + currencies.CurrencyNames[msg.NotificationMsg.CurrencyID],
								LocKey:  store.TopicNewIncoming,
								LocArgs: []string{convertToHuman(msg.NotificationMsg.Amount, currencies.Dividers[msg.NotificationMsg.CurrencyID]), currencies.CurrencyNames[msg.NotificationMsg.CurrencyID]},
							},
						},
					},
				},
				Topic: topic,
			}

			fClient.log.Debugf("\n\n MessageToSend : %v\n\n", messageToSend)

			ctx := context.Background()
			client, err := service.Messaging(ctx)
			if err != nil {
				fClient.log.Errorf("service.Messaging: %v", err.Error())
			}

			response, err := client.Send(ctx, messageToSend)
			if err != nil {
				fClient.log.Errorf("client.Send : %v", err.Error())
			}

			fClient.log.Errorf("\n\nFirebase push resp : %v\n\n", response)
		}

		return nil
	}))

	if err = nsqConsumer.ConnectToNSQD(nsqAddr); err != nil {
		return nil, fmt.Errorf("connecting to nsq: %s", err.Error())
	}
	fClient.log.Debugf("Firebase connection initialization done")
	return fClient, nil
}

func NewPushService(withCredentialsFile string) (*firebase.App, error) {
	opt := option.WithCredentialsFile(withCredentialsFile)
	return firebase.NewApp(context.Background(), nil, opt)
}
