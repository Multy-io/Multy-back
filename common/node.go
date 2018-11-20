package common

import (
	"github.com/Multy-io/Multy-back/store"
	"github.com/graarh/golang-socketio"
)

type NodeServiceConfiguration interface {
	InitHandlers(dbConf *store.Conf, coinTypes []store.CoinType, nsqAddr string) (interface{}, error)
	SetUserData(userStore store.UserStore, ct []store.CoinType) ([]store.ServiceInfo, error)
	RegisterWebSocketEvents(socketServer *gosocketio.Server) error
}

