package client

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/Appscrunch/Multy-back/btc"
	fcm "github.com/NaySoftware/go-fcm"
	"github.com/gin-gonic/gin"

	"github.com/nsqio/go-nsq"
)

type FirebaseConf struct {
	ServerKey string
}

type FirebaseConn struct {
	firebaseConf *FirebaseConf
	firebaseConn *fcm.FcmClient

	nsqConsumer *nsq.Consumer
	nsqConfig   *nsq.Config
}

func InitFirebaseConn(firebaseConf *FirebaseConf, c *gin.Engine) (*FirebaseConn, error) {
	log.Println("[INFO]  Firebase connection initialization")
	log.Println("[DEBUG] Server key=", firebaseConf.ServerKey)

	fConn := &FirebaseConn{
		firebaseConf: firebaseConf,
		firebaseConn: fcm.NewFcmClient(firebaseConf.ServerKey),
		nsqConfig:    nsq.NewConfig(),
	}

	nsqConsumer, err := nsq.NewConsumer(btc.TopicTransaction, "firebase", fConn.nsqConfig)
	if err != nil {
		log.Printf("[ERR] new nsq consumer: %s\n", err.Error())
		return nil, fmt.Errorf("new nsq consumer: %s", err.Error())
	}
	nsqConsumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		msgRaw := message.Body
		log.Printf("[DEBUG] firebase new transaction notify: %+v\n", string(msgRaw))
		log.Println("[DEBUG] nsqConsumer: Server key=", firebaseConf.ServerKey)

		msg := btc.BtcTransactionWithUserID{}
		err := json.Unmarshal(msgRaw, &msg)
		if err != nil {
			log.Printf("[ERR] firebase new transaction notify processing: %s\n", err.Error())
			return err
		}

		data := map[string]string{
			"txid":            msg.NotificationMsg.TxID,
			"transactionType": msg.NotificationMsg.TransactionType,
			"amount":          strconv.FormatFloat(msg.NotificationMsg.Amount, 'f', 3, 64),
			"address":         msg.NotificationMsg.Address,
		}
		log.Printf("DEBUG %+v\n", data)

		// TODO: add version /v1
		fConn.firebaseConn.NewFcmMsgTo("/topics/"+btc.TopicTransaction+"-"+msg.UserID, data)
		status, err := fConn.firebaseConn.Send()
		if err != nil {
			log.Printf("[ERR] firebase send transaction notify processing: %s\n", err.Error())
			return err
		}
		status.PrintResults()

		return nil
	}))

	if err = nsqConsumer.ConnectToNSQD("127.0.0.1:4150"); err != nil {
		log.Printf("[ERR] new nsq consumer: %s\n", err.Error())
		return nil, fmt.Errorf("connecting to nsq: %s", err.Error())
	}
	log.Println("[INFO]  Firebase connection initialization done")
	return fConn, nil
}
