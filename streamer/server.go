package streamer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	bind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/parnurzeal/gorequest"
	"google.golang.org/grpc"

	"github.com/Multy-io/Multy-ETH-node-service/eth"
	pb "github.com/Multy-io/Multy-ETH-node-service/node-streamer"
	"github.com/Multy-io/Multy-back/store"
	"github.com/jekabolt/slf"
	_ "github.com/jekabolt/slflog"
)

var log = slf.WithContext("streamer")

// Server implements streamer interface and is a gRPC server
type Server struct {
	UsersData       *sync.Map
	Multisig        *eth.Multisig
	EthCli          *eth.Client
	Info            *store.ServiceInfo
	NetworkID       int
	ResyncUrl       string
	EtherscanAPIKey string
	EtherscanAPIURL string
	ABIcli          *ethclient.Client
	GRPCserver      *grpc.Server
	Listener        net.Listener
	ReloadChan      chan struct{}
}

func (s *Server) ServiceInfo(c context.Context, in *pb.Empty) (*pb.ServiceVersion, error) {
	return &pb.ServiceVersion{
		Branch:    s.Info.Branch,
		Commit:    s.Info.Commit,
		Buildtime: s.Info.Buildtime,
		Lasttag:   "",
	}, nil
}

func (s *Server) EventGetGasPrice(ctx context.Context, in *pb.Empty) (*pb.GasPrice, error) {
	gp, err := s.EthCli.GetGasPrice()
	if err != nil {
		return &pb.GasPrice{}, err
	}
	return &pb.GasPrice{
		Gas: gp.String(),
	}, nil
}

func (s *Server) GetERC20Info(ctx context.Context, in *pb.ERC20Address) (*pb.ERC20Info, error) {
	addressInfo := &pb.ERC20Info{}

	// erc token tx hisotry
	url := s.EtherscanAPIURL + "/api?module=account&action=tokentx&address=" + in.GetAddress() + "&startblock=0&endblock=999999999&sort=asc&apikey=" + s.EtherscanAPIKey
	request := gorequest.New()
	resp, _, errs := request.Get(url).Retry(10, 3*time.Second, http.StatusForbidden, http.StatusBadRequest, http.StatusInternalServerError).End()
	if len(errs) > 0 {

	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(`return errors.Wrap(err, "failed to read response body")`)
	}

	tokenresp := store.EtherscanResp{}
	if err := json.Unmarshal(respBody, &tokenresp); err != nil {
		fmt.Println(`fmt.Println("err Unmarshal ", err)`)
	}
	for _, tx := range tokenresp.Result {
		addressInfo.History = append(addressInfo.History, &tx)
	}

	// erc token balances
	tokens := map[string]string{}
	for _, token := range tokenresp.Result {
		tokens[token.ContractAddress] = ""
	}
	for contract := range tokens {
		token, err := NewToken(common.HexToAddress(contract), s.ABIcli)
		if err != nil {
			log.Errorf("GetMultisigInfo - %v", err)
			return nil, err
		}
		balance, err := token.BalanceOf(&bind.CallOpts{}, common.HexToAddress(in.GetAddress()))
		if err != nil {
			log.Errorf("GetERC20Info:token.BalanceOf %v", err.Error())
		}
		addressInfo.Balances = append(addressInfo.Balances, &pb.ERC20Balances{
			Address: contract,
			Balance: balance.String(),
		})
	}

	if in.OnlyBalances {
		addressInfo.History = nil
	}

	return addressInfo, nil
}

func (s *Server) GetMultisigInfo(ctx context.Context, in *pb.AddressToResync) (*pb.ContractInfo, error) {

	contract, err := NewMultiSigWallet(common.HexToAddress(in.GetAddress()), s.ABIcli)
	if err != nil {
		log.Errorf("GetMultisigInfo - %v", err)
		return nil, err
	}
	contractOwners, err := contract.GetOwners(&bind.CallOpts{})
	if err != nil {
		log.Errorf("GetMultisigInfo contract.GetOwners - %v", err)
		return nil, err
	}
	owners := []string{}
	for _, owner := range contractOwners {
		owners = append(owners, strings.ToLower(owner.String()))
	}

	required, err := contract.Required(&bind.CallOpts{})
	if err != nil {
		log.Errorf("GetMultisigInfo contract.Required - %v", err)
		return nil, err
	}

	return &pb.ContractInfo{
		ConfirmationsRequired: required.Int64(),
		ContractOwners:        owners,
	}, err

}

func (s *Server) EventInitialAdd(c context.Context, ud *pb.UsersData) (*pb.ReplyInfo, error) {
	log.Debugf("EventInitialAdd len - %v", len(ud.Map))

	udMap := sync.Map{}
	for addr, ex := range ud.GetMap() {
		udMap.Store(strings.ToLower(addr), store.AddressExtended{
			UserID:       ex.GetUserID(),
			WalletIndex:  int(ex.GetWalletIndex()),
			AddressIndex: int(ex.GetAddressIndex()),
		})
	}

	*s.UsersData = udMap

	for key, value := range ud.GetUsersContracts() {
		s.Multisig.UsersContracts.Store(key, value)
	}

	return &pb.ReplyInfo{
		Message: "ok",
	}, nil
}

// EventAddNewAddress us used to add new watch address to existing pairs
func (s *Server) EventAddNewAddress(c context.Context, wa *pb.WatchAddress) (*pb.ReplyInfo, error) {
	newMap := *s.UsersData
	// if newMap == nil {
	// 	newMap = sync.Map{}
	// }
	_, ok := newMap.Load(wa.Address)
	if ok {
		return &pb.ReplyInfo{
			Message: "err: Address already binded",
		}, nil
	}
	newMap.Store(strings.ToLower(wa.Address), store.AddressExtended{
		UserID:       wa.UserID,
		WalletIndex:  int(wa.WalletIndex),
		AddressIndex: int(wa.AddressIndex),
	})

	*s.UsersData = newMap

	log.Debugf("EventAddNewAddress - %v", newMap)

	return &pb.ReplyInfo{
		Message: "ok",
	}, nil

}

func (s *Server) EventAddNewMultisig(ctx context.Context, address *pb.WatchAddress) (*pb.ReplyInfo, error) {
	log.Debugf("EventAddNewMultisig")

	// store multisig adddress in map
	s.Multisig.UsersContracts.Store(address.GetAddress(), "")

	// resync multisig transacttions
	addr := address.GetAddress()
	url := s.ResyncUrl + addr + "&action=txlist&module=account"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &pb.ReplyInfo{
			Message: fmt.Sprintf("EventResyncAddress: http.NewRequest = %s", err.Error()),
		}, nil
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &pb.ReplyInfo{
			Message: fmt.Sprintf("EventResyncAddress: http.DefaultClient.Do = %s", err.Error()),
		}, nil
	}
	defer res.Body.Close()

	reTx := resyncTx{}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &pb.ReplyInfo{
			Message: fmt.Sprintf("EventResyncAddress: ioutil.ReadAll = %s", err.Error()),
		}, nil
	}

	if err := json.Unmarshal(body, &reTx); err != nil {
		return &pb.ReplyInfo{
			Message: fmt.Sprintf("EventResyncAddress: json.Unmarshal = %s", err.Error()),
		}, nil
	}

	if !strings.Contains(reTx.Message, "OK") {
		return &pb.ReplyInfo{
			Message: fmt.Sprintf("EventResyncAddress: !strings.Contains OK a.k.a. bad response form 3-party"),
		}, nil
	}

	log.Debugf("EventResyncAddress %d", len(reTx.Result))

	for _, hash := range reTx.Result {
		s.EthCli.ResyncMultisig(hash.Hash)
	}

	return &pb.ReplyInfo{
		Message: "ok",
	}, nil

}

func (s *Server) EventGetBlockHeight(ctx context.Context, in *pb.Empty) (*pb.BlockHeight, error) {
	h, err := s.EthCli.GetBlockHeight()
	if err != nil {
		return &pb.BlockHeight{}, err
	}
	return &pb.BlockHeight{
		Height: int64(h),
	}, nil
}

func (s *Server) EventGetCode(ctx context.Context, in *pb.AddressToResync) (*pb.ReplyInfo, error) {
	code, err := s.EthCli.GetCode(in.Address)
	if err != nil {
		return &pb.ReplyInfo{}, err
	}
	return &pb.ReplyInfo{
		Message: code,
	}, nil
}

func (s *Server) EventGetAdressNonce(c context.Context, in *pb.AddressToResync) (*pb.Nonce, error) {
	n, err := s.EthCli.GetAddressNonce(in.GetAddress())
	if err != nil {
		return &pb.Nonce{}, err
	}
	return &pb.Nonce{
		Nonce: int64(n),
	}, nil
}

func (s *Server) EventGetAdressBalance(ctx context.Context, in *pb.AddressToResync) (*pb.Balance, error) {
	b, err := s.EthCli.GetAddressBalance(in.GetAddress())
	if err != nil {
		return &pb.Balance{}, err
	}
	p, err := s.EthCli.GetAddressPendingBalance(in.GetAddress())
	if err != nil {
		return &pb.Balance{}, err
	}
	return &pb.Balance{
		Balance:        b.String(),
		PendingBalance: p.String(),
	}, nil
}

func (s *Server) CheckRejectTxs(ctx context.Context, txs *pb.TxsToCheck) (*pb.RejectedTxs, error) {
	reTxs := &pb.RejectedTxs{}
	for _, tx := range txs.Hash {
		rtx, _ := s.EthCli.Rpc.EthGetTransactionByHash(tx)
		if len(rtx.Hash) == 0 {
			reTxs.RejectedTxs = append(reTxs.RejectedTxs, tx)
		}
	}
	return reTxs, nil
}

func (s *Server) SyncState(ctx context.Context, in *pb.BlockHeight) (*pb.ReplyInfo, error) {

	currentH, err := s.EthCli.GetBlockHeight()
	if err != nil {
		log.Errorf("s.BtcCli.RpcClient.GetBlockCount: %v ", err.Error())
	}
	if in.GetHeight() <= 0 {
		return &pb.ReplyInfo{
			Message: "bad",
		}, nil
	}
	log.Warnf("currentH %v lastH %v difference %v ", currentH, in.GetHeight(), int64(currentH)-in.GetHeight())

	for lastH := int(in.GetHeight()); lastH < currentH; lastH++ {
		b, err := s.EthCli.Rpc.EthGetBlockByNumber(lastH, false)
		if err != nil {
			log.Errorf("s.BtcCli.RpcClient.GetBlockHash: %v", err.Error())
			continue
		}
		go s.EthCli.BlockTransaction(b.Hash)
	}

	return &pb.ReplyInfo{
		Message: "ok",
	}, nil
}

func (s *Server) EventGetAllMempool(_ *pb.Empty, stream pb.NodeCommuunications_EventGetAllMempoolServer) error {
	mp, err := s.EthCli.GetAllTxPool()
	if err != nil {
		return err
	}

	for _, txs := range mp["pending"].(map[string]interface{}) {
		for _, tx := range txs.(map[string]interface{}) {
			gas, err := strconv.ParseInt(tx.(map[string]interface{})["gas"].(string), 0, 64)
			if err != nil {
				log.Errorf("EventGetAllMempool:strconv.ParseInt")
			}
			hash := tx.(map[string]interface{})["hash"].(string)
			stream.Send(&pb.MempoolRecord{
				Category: int32(gas),
				HashTX:   hash,
			})
		}
	}
	return nil
}

type resyncTx struct {
	Message string `json:"message"`
	Result  []struct {
		Hash string `json:"hash"`
	} `json:"result"`
}

func (s *Server) EventResyncAddress(c context.Context, address *pb.AddressToResync) (*pb.ReplyInfo, error) {
	log.Debugf("EventResyncAddress")
	addr := address.GetAddress()
	url := s.ResyncUrl + addr + "&action=txlist&module=account"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &pb.ReplyInfo{
			Message: fmt.Sprintf("EventResyncAddress: http.NewRequest = %s", err.Error()),
		}, nil
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return &pb.ReplyInfo{
			Message: fmt.Sprintf("EventResyncAddress: http.DefaultClient.Do = %s", err.Error()),
		}, nil
	}
	defer res.Body.Close()

	reTx := resyncTx{}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &pb.ReplyInfo{
			Message: fmt.Sprintf("EventResyncAddress: ioutil.ReadAll = %s", err.Error()),
		}, nil
	}

	if err := json.Unmarshal(body, &reTx); err != nil {
		return &pb.ReplyInfo{
			Message: fmt.Sprintf("EventResyncAddress: json.Unmarshal = %s", err.Error()),
		}, nil
	}

	if !strings.Contains(reTx.Message, "OK") {
		return &pb.ReplyInfo{
			Message: fmt.Sprintf("EventResyncAddress: bad resp form 3-party"),
		}, nil
	}

	log.Debugf("EventResyncAddress %d", len(reTx.Result))

	for _, hash := range reTx.Result {
		s.EthCli.ResyncAddress(hash.Hash)
	}

	return &pb.ReplyInfo{
		Message: "ok",
	}, nil

}
func (s *Server) EventSendRawTx(c context.Context, tx *pb.RawTx) (*pb.ReplyInfo, error) {
	hash, err := s.EthCli.SendRawTransaction(tx.GetTransaction())
	if err != nil {
		return &pb.ReplyInfo{
			Message: "err: wrong raw tx",
		}, fmt.Errorf("err: wrong raw tx %s", err.Error())
	}

	return &pb.ReplyInfo{
		Message: hash,
	}, nil

}

func (s *Server) EventDeleteMempool(_ *pb.Empty, stream pb.NodeCommuunications_EventDeleteMempoolServer) error {
	defer close(s.EthCli.DeleteMempoolStream)
	for range s.EthCli.DeleteMempoolStream {
		select {
		case del := <-s.EthCli.DeleteMempoolStream:
			err := stream.Send(&del)
			if err != nil && err.Error() == ErrGrpcTransport {
				log.Errorf("EventDeleteMempoolStream:stream.Send() %v ", err.Error())
				s.ReloadChan <- struct{}{}
			}
		case <-s.EthCli.Stop:
			log.Debugf("EventDeleteMempoolStream close")
			return nil
		}
	}
	return nil
}

func (s *Server) EventAddMempoolRecord(_ *pb.Empty, stream pb.NodeCommuunications_EventAddMempoolRecordServer) error {
	defer close(s.EthCli.AddToMempoolStream)
	for range s.EthCli.AddToMempoolStream {
		select {
		case add := <-s.EthCli.AddToMempoolStream:
			err := stream.Send(&add)
			if err != nil && err.Error() == ErrGrpcTransport {
				log.Errorf("EventAddMempoolRecord:stream.Send() %v ", err.Error())
				s.ReloadChan <- struct{}{}
			}
		case <-s.EthCli.Stop:
			log.Debugf("EventAddMempoolRecord close")
			return nil
		}
	}
	return nil
}

func (s *Server) NewTx(_ *pb.Empty, stream pb.NodeCommuunications_NewTxServer) error {
	defer close(s.EthCli.TransactionsStream)
	for range s.EthCli.TransactionsStream {
		select {
		case tx := <-s.EthCli.TransactionsStream:
			log.Infof("NewTx history - %v", tx.String())
			err := stream.Send(&tx)
			if err != nil && err.Error() == ErrGrpcTransport {
				log.Errorf("NewTx:stream.Send() %v ", err.Error())
				s.ReloadChan <- struct{}{}
			}
		case <-s.EthCli.Stop:
			log.Debugf("NewTx close")
			return nil
		}
	}
	return nil
}

func (s *Server) EventNewBlock(_ *pb.Empty, stream pb.NodeCommuunications_EventNewBlockServer) error {
	defer close(s.EthCli.BlockStream)
	for range s.EthCli.BlockStream {
		select {
		case h := <-s.EthCli.BlockStream:
			log.Infof("New block height - %v", h.GetHeight())
			err := stream.Send(&h)
			if err != nil && err.Error() == ErrGrpcTransport {
				log.Errorf("EventNewBlock:stream.Send() %v ", err.Error())
				s.ReloadChan <- struct{}{}
			}
		case <-s.EthCli.Stop:
			log.Debugf("EventNewBlock close")
			return nil
		}
	}
	return nil
}

func (s *Server) AddMultisig(_ *pb.Empty, stream pb.NodeCommuunications_AddMultisigServer) error {
	defer close(s.EthCli.NewMultisigStream)
	for range s.EthCli.NewMultisigStream {
		select {
		case m := <-s.EthCli.NewMultisigStream:
			log.Infof("AddMultisig new contract address - %v", m.GetContract())
			err := stream.Send(&m)
			log.Debugf("Multisig sent on address contract %v", m.Contract)
			if err != nil && err.Error() == ErrGrpcTransport {
				log.Errorf("AddMultisig:stream.Send() %v ", err.Error())
				s.ReloadChan <- struct{}{}
			}
		case <-s.EthCli.Stop:
			log.Debugf("AddMultisig close")
			return nil
		}
	}
	return nil
}
