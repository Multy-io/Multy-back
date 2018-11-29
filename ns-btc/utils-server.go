package nsbtc

import "github.com/Multy-io/Multy-back/store"

var (
	ErrGrpcTransport = "rpc error: code = Unavailable desc = transport is closing"
)

func reverseResyncTx(ss []store.ResyncTx) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}
