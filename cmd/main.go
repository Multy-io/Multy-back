/*
Copyright 2018 Idealnaya rabota LLC
Licensed under Multy.io license.
See LICENSE for details
*/
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/jekabolt/config"
	_ "github.com/jekabolt/slflog"

	multy "github.com/Appscrunch/Multy-back"
	"github.com/Appscrunch/Multy-back/store"
	"github.com/jekabolt/slf"
)

var (
	log = slf.WithContext("main")

	branch    string
	commit    string
	buildtime string
	lasttag   string
	// TODO: add all default params
	globalOpt = multy.Configuration{
		Name: "my-test-back",
		Database: store.Conf{
			Address:             "localhost:27017",
			DBUsers:             "userDB-test",
			DBFeeRates:          "BTCMempool-test",
			DBTx:                "DBTx-test",
			DBStockExchangeRate: "dev-DBStockExchangeRate",
			Username:            "Username",
			Password:            "Password",
		},
		RestAddress:    "localhost:7778",
		SocketioAddr:   "localhost:7780",
		NSQAddress:     "nsq:4150",
		BTCNodeAddress: "localhost:18334",
		// Etherium: ethereum.Conf{
		// 	Address: "88.198.47.112",
		// 	RpcPort: ":18545",
		// 	WsPort:  ":8545",
		// },
	}
)

func main() {
	config.ReadGlobalConfig(&globalOpt, "multy configuration")

	log.Error("--------------------------------new multy back server session")
	log.Infof("CONFIGURATION=%+v", globalOpt)

	log.Infof("branch: %s", branch)
	log.Infof("commit: %s", commit)
	log.Infof("build time: %s", buildtime)
	log.Infof("tag: %s", lasttag)

	var gracefulStop = make(chan os.Signal)

	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	sc := store.ServerConfig{
		BranchName: branch,
		CommitHash: commit,
		Build:      buildtime,
		Tag:        lasttag,
	}

	globalOpt.MultyVerison = sc

	mu, err := multy.Init(&globalOpt)
	if err != nil {
		log.Fatalf("Server initialization: %s\n", err.Error())
	}

	// TODO: Last state
	// go func() {
	// var ls store.LastState

	// raw, err := ioutil.ReadFile("./state.json")
	// if err != nil {
	// 	log.Errorf("ioutil.ReadFile: %v", err.Error())
	// }
	// log.Errorf("\n\n\nLAT STATE\n\n\n %v", string(raw))
	// json.Unmarshal(raw, &ls)

	// rp, err := mu.BTC.CliMain.SyncState(context.Background(), &btcpb.BlockHeight{
	// 	Height: ls.BTCMainBlock,
	// })

	// log.Errorf("\n\n\nmu.BTC.CliMain\n\n\n %v", ls.BTCMainBlock)

	// if err != nil {
	// 	log.Errorf("mu.BTC.CliMain.SyncState %v", err.Error())
	// }
	// log.Debugf("mu.BTC.CliMainSyncState: Reply: %v", rp)

	// rp, err = mu.BTC.CliTest.SyncState(context.Background(), &btcpb.BlockHeight{
	// 	Height: ls.BTCTestBlock,
	// })

	// log.Errorf("\n\n\nmu.BTC.CliTest\n\n\n %v", ls.BTCTestBlock)
	// if err != nil {
	// 	log.Errorf("mu.BTC.CliTest.SyncState %v", err.Error())
	// }
	// log.Debugf("mu.BTC.CliMainSyncState: Reply: %v", rp)
	// // }()

	// go func() {
	// 	sig := <-gracefulStop
	// 	fmt.Printf("caught sig: %+v", sig)
	// 	fmt.Println("Gracefull stop")
	// 	time.Sleep(time.Second)
	// 	main, err := mu.BTC.CliMain.EventGetBlockHeight(context.Background(), &btcpb.Empty{})
	// 	if err != nil {
	// 		log.Errorf("mu.BTC.CliMain.EventGetBlockHeight %v", err.Error())
	// 	}
	// 	test, _ := mu.BTC.CliTest.EventGetBlockHeight(context.Background(), &btcpb.Empty{})
	// 	if err != nil {
	// 		log.Errorf("mu.BTC.CliTest.EventGetBlockHeight %v", err.Error())
	// 	}

	// 	ls := store.LastState{
	// 		BTCMainBlock: main.GetHeight(),
	// 		BTCTestBlock: test.GetHeight(),
	// 		OffTime:      time.Now().String(),
	// 	}

	// 	fileState, err := json.Marshal(&ls)
	// 	if err != nil {
	// 		log.Errorf("json.Marshal %v", err.Error())
	// 	}
	// 	err = ioutil.WriteFile("state.json", fileState, 0644)
	// 	if err != nil {
	// 		log.Errorf("json.Marshal %v", err.Error())
	// 	}

	// 	log.Infof("\n\noff state main : %v test: %v\n\n", main, test)

	// 	os.Exit(0)
	// }()

	if err = mu.Run(); err != nil {
		log.Fatalf("Server running: %s\n", err.Error())
	}

}
