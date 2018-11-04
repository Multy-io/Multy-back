package eth

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"

	pb "github.com/Multy-io/Multy-ETH-node-service/node-streamer"
	"github.com/Multy-io/Multy-back/store"
)

func (c *Client) FactoryContract(hash string) {
	log.Debugf("FactoryContract")
	m, err := c.Rpc.TransactionReceipt(hash)
	if err != nil {
		fmt.Printf("FactoryContract:c.Rpc.TransactionReceipt %v", err.Error())
	}

	receipt := m["logs"].([]interface{})
	multisigAddress := ""
	deployed := false
	if len(receipt) > 0 && len(receipt[0].(map[string]interface{})["data"].(string)) > 90 {
		multisigAddress = "0x" + receipt[0].(map[string]interface{})["data"].(string)[90:]
		if m["status"] == "0x1" {
			deployed = true
		}
	}

	rawTx, err := c.Rpc.EthGetTransactionByHash(hash)
	if err != nil {
		log.Errorf("FactoryContract:Get TX Err: %s", err.Error())
	}

	fi, err := parseFactoryInput(rawTx.Input)
	if err != nil {
		log.Errorf("FactoryContract:parseInput: %s", err.Error())
	}

	// if len(m["returnValue"].(string)) > 25 {
	// 	fi.Contract = "0x" + m["returnValue"].(string)[24:]
	// }

	fi.Contract = multisigAddress

	fi.TxOfCreation = hash
	fi.FactoryAddress = c.Multisig.FactoryAddress
	deployStatus := int64(store.MultisigStatusRejected)
	if deployed {
		deployStatus = store.MultisigStatusDeployed
	}
	fi.DeployStatus = deployStatus

	log.Warnf("DeployStatus  %v ", deployed)

	c.Multisig.UsersContracts.Store(fi.Contract, fi.FactoryAddress)

	c.NewMultisigStream <- *fi

}

func parseFactoryInput(in string) (*pb.Multisig, error) {
	// fetch method id by hash
	log.Debugf("parseFactoryInput")
	fi := &pb.Multisig{}
	if in[:10] == MultiSigFactory {
		log.Debugf("parseFactoryInput:", MultiSigFactory)
		in := in[10:]

		c := in[64:128]
		confirmations, _ := new(big.Int).SetString(c, 10)
		fi.Confirmations = confirmations.Int64()

		in = in[192:]

		contractAddresses := []string{}
		re := regexp.MustCompile(`.{64}`) // Every 64 chars
		parts := re.FindAllString(in, -1) // Split the string into 64 chars blocks.

		for _, address := range parts {
			contractAddresses = append(contractAddresses, "0x"+address[24:])
		}
		fi.Addresses = contractAddresses

		return fi, nil
	}

	return fi, errors.New("Wrong method name")
}

func (c *Client) GetInvocationStatus(hash, method string) (bool, string, error) {
	m, err := c.Rpc.TransactionReceipt(hash)
	if err != nil {
		log.Errorf("FactoryContract:c.Rpc.TransactionReceipt %v", err.Error())
		return false, "", err
	}
	receipt := m["logs"].([]interface{})

	returnValue := ""
	deployed := false
	if len(receipt) > 0 {
		returnValue = receipt[0].(map[string]interface{})["data"].(string)
		if returnValue == "0x" {
			returnValue = ""
		}
		if method == store.SubmitTransaction {
			returnValue = receipt[0].(map[string]interface{})["topics"].([]interface{})[1].(string)[2:]
		}
		if m["status"].(string) == "0x1" {
			deployed = true
		}
	}
	log.Warnf("GetInvocationStatus: deployed: %v returnValue: %v", deployed, returnValue)

	return deployed, returnValue, nil
	// if !isFailed {

	// 	rawTx, err := c.Rpc.EthGetTransactionByHash(hash)
	// 	if err != nil {
	// 		log.Errorf("FactoryContract:Get TX Err: %s", err.Error())
	// 	}
	// 	// returnValue := m["returnValue"].(string)

	// 	switch rawTx.Input[:10] {
	// 	case submitTransaction: // "c6427474": "submitTransaction(address,uint256,bytes)"
	// 		// TODO: feth contract owners, send notfy to owners about transation. status: waiting for confirmations
	// 	case confirmTransaction: // "c01a8c84": "confirmTransaction(uint256)"
	// 		// TODO: send notfy to owners about +1 confirmation. store confiramtions id db
	// 	case revokeConfirmation: // "20ea8d86": "revokeConfirmation(uint256)"
	// 		// TODO: send notfy to owners about -1 confirmation. store confirmations in db
	// 	case executeTransaction: // "ee22610b": "executeTransaction(uint256)"
	// 		// TODO: feth contract owners, send notfy to owners about transation. status: conformed transatcion
	// 	case "": // simple transaction
	// 		// TODO: notify owners about new transation
	// 	default:
	// 		// wrong method
	// 	}
	// 	// n, _ := new(big.Int).SetString(m["returnValue"].(string), 10)
	// 	// nonce = n.Int64()

	// }
	// m["returnValue"],
}

// func (c *Client) MultisigContract(hash string) *pb.Multisig {

// }
