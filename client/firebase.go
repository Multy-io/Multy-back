package client

import (
	"log"

	"github.com/Appscrunch/Multy-back/btc"
	fcm "github.com/NaySoftware/go-fcm"
	"github.com/gin-gonic/gin"
)

type FirebaseConf struct {
	ServerKey string
	topicName string
}

type FirebaseConn struct {
	conf  *FirebaseConf
	conn  *fcm.FcmClient
	btcCh chan btc.BtcTransactionWithUserID
}

func InitFirebaseConn(conf *FirebaseConf, c *gin.Engine) *FirebaseConn {
	fConn := &FirebaseConn{
		conf:  conf,
		conn:  fcm.NewFcmClient(conf.ServerKey),
		btcCh: make(chan btc.BtcTransactionWithUserID, 0),
	}
	go fConn.listenBTC()
	return fConn
}

func (fConn *FirebaseConn) listenBTC() {
	var newTransactionWithUserID btc.BtcTransactionWithUserID

	topic := fConn.conf.topicName
	for {
		select {
		case newTransactionWithUserID = <-fConn.btcCh:
			id := newTransactionWithUserID.UserID
			c := fConn.conn.NewFcmMsgTo(topic+id, newTransactionWithUserID)
			if status, err := c.Send(); err != nil {
				status.PrintResults()
			} else {
				log.Printf("[ERR] (fConn *FilebaseConn) listenBTC: %s\n", err.Error())
			}
		}
	}
}
