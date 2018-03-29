package streamer

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/Appscrunch/Multy-BTC-node-service/btc"
	pb "github.com/Appscrunch/Multy-back/node-streamer/btc"
	"github.com/blockcypher/gobcy"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Server implements streamer interface and is a gRPC server
type Server struct {
	UsersData *map[string]string
	BtcAPI    *gobcy.API
	M         *sync.Mutex
}

// EventInitialAdd us used to add initial pairs of watch addresses
func (s *Server) EventInitialAdd(c context.Context, ud *pb.UsersData) (*pb.ReplyInfo, error) {
	fmt.Println("[DEBUG] EventInitialAdd - ", ud.Map)
	s.M.Lock()
	defer s.M.Unlock()
	*s.UsersData = ud.Map
	return &pb.ReplyInfo{
		Message: "ok",
	}, nil
}

// EventAddNewAddress us used to add new watch address to existing pairs
func (s *Server) EventAddNewAddress(c context.Context, wa *pb.WatchAddress) (*pb.ReplyInfo, error) {
	s.M.Lock()
	defer s.M.Unlock()
	newMap := *s.UsersData
	if newMap == nil {
		newMap = map[string]string{}
	}
	fmt.Println("map----", newMap)
	//TODO: binded address fix
	// _, ok := newMap[wa.Address]
	// if ok {
	// 	return nil, errors.New("Address already binded")
	// }
	newMap[wa.Address] = wa.UserID

	*s.UsersData = newMap
	return &pb.ReplyInfo{
		Message: "ok",
	}, nil

}

func (s *Server) EventGetBlockHeight(ctx context.Context, in *pb.Empty) (*pb.BlockHeight, error) {
	h, err := btc.GetBlockHeight()
	if err != nil {
		return &pb.BlockHeight{}, err
	}
	return &pb.BlockHeight{
		Height: h,
	}, nil
}

// EventAddNewAddress us used to add new watch address to existing pairs
func (s *Server) EventGetAllMempool(_ *pb.Empty, stream pb.NodeCommuunications_EventGetAllMempoolServer) error {
	mp, err := btc.GetAllMempool()
	if err != nil {
		return err
	}

	for _, rec := range mp {
		stream.Send(&pb.MempoolRecord{
			Category: int32(rec.Category),
			HashTX:   rec.HashTX,
		})
	}

	return nil
}
func (s *Server) EventResyncAddress(c context.Context, address *pb.AddressToResync) (*pb.ReplyInfo, error) {
	allResync := []resyncTx{}
	requestTimes := 0
	addrInfo, err := s.BtcAPI.GetAddrFull(address.Address, map[string]string{"limit": "50"})
	if err != nil {
		return nil, fmt.Errorf("[ERR] EventResyncAddress: s.BtcAPI.GetAddrFull : %s", err.Error())
	}

	if addrInfo.FinalNumTX > 50 {
		requestTimes = int(float64(addrInfo.FinalNumTX) / 50.0)
	}

	for _, tx := range addrInfo.TXs {
		allResync = append(allResync, resyncTx{
			hash:        tx.Hash,
			blockHeight: tx.BlockHeight,
		})
	}

	for i := 0; i < requestTimes; i++ {
		addrInfo, err := s.BtcAPI.GetAddrFull(address.Address, map[string]string{"limit": "50", "before": strconv.Itoa(allResync[len(allResync)-1].blockHeight)})
		if err != nil {
			return nil, fmt.Errorf("[ERR] EventResyncAddress: s.BtcAPI.GetAddrFull : %s", err.Error())
		}
		for _, tx := range addrInfo.TXs {
			allResync = append(allResync, resyncTx{
				hash:        tx.Hash,
				blockHeight: tx.BlockHeight,
			})
		}
	}

	reverseResyncTx(allResync)

	for _, reTx := range allResync {
		txHash, err := chainhash.NewHashFromStr(reTx.hash)
		if err != nil {
			return nil, fmt.Errorf("[ERR] resyncAddress: chainhash.NewHashFromStr = %s", err.Error())
		}
		fmt.Println(btc.RpcClient)
		rawTx, err := btc.RpcClient.GetRawTransactionVerbose(txHash)
		if err != nil {
			return nil, fmt.Errorf("[ERR] resyncAddress: RpcClient.GetRawTransactionVerbose = %s", err.Error())
		}

		btc.ProcessTransaction(int64(reTx.blockHeight), rawTx, true, s.UsersData)
	}

	return &pb.ReplyInfo{
		Message: "ok",
	}, nil
}

func (s *Server) EventSendRawTx(c context.Context, tx *pb.RawTx) (*pb.ReplyInfo, error) {
	hash, err := btc.RpcClient.SendCyberRawTransaction(tx.Transaction, true)
	if err != nil {
		fmt.Println("kek")
		return &pb.ReplyInfo{
			Message: "err: wrong raw tx",
		}, fmt.Errorf("err: wrong raw tx %s", err.Error())

	}

	return &pb.ReplyInfo{
		Message: hash.String(),
	}, nil

}

func (s *Server) EventDeleteMempool(_ *pb.Empty, stream pb.NodeCommuunications_EventDeleteMempoolServer) error {
	for {
		select {
		case del := <-btc.DeleteMempool:
			stream.Send(&del)
		}
	}
	return nil
}

func (s *Server) EventAddMempoolRecord(_ *pb.Empty, stream pb.NodeCommuunications_EventAddMempoolRecordServer) error {
	for {
		select {
		case add := <-btc.AddToMempool:
			stream.Send(&add)
		}
	}
	return nil
}

func (s *Server) EventDeleteSpendableOut(_ *pb.Empty, stream pb.NodeCommuunications_EventDeleteSpendableOutServer) error {
	for {
		select {
		case delSp := <-btc.DelSpOut:
			stream.Send(&delSp)
		}
	}
	return nil
}
func (s *Server) EventAddSpendableOut(_ *pb.Empty, stream pb.NodeCommuunications_EventAddSpendableOutServer) error {
	for {
		select {
		case addSp := <-btc.AddSpOut:
			stream.Send(&addSp)
		}
	}
	return nil
}
func (s *Server) NewTx(_ *pb.Empty, stream pb.NodeCommuunications_NewTxServer) error {
	for {
		select {
		case tx := <-btc.TransactionsCh:
			fmt.Printf("NewTx %v", tx.String())
			stream.Send(&tx)
		}
	}
	return nil
}
