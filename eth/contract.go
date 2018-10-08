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
	m, err := c.Rpc.TraceTransaction(hash)
	if err != nil {
		fmt.Printf("FactoryContract:c.Rpc.TraceTransaction %v", err.Error())
	}

	rawTx, err := c.Rpc.EthGetTransactionByHash(hash)
	if err != nil {
		log.Errorf("FactoryContract:Get TX Err: %s", err.Error())
	}

	fi, err := parseFactoryInput(rawTx.Input)
	if err != nil {
		log.Errorf("FactoryContract:parseInput: %s", err.Error())
	}

	if len(m["returnValue"].(string)) > 25 {
		fi.Contract = "0x" + m["returnValue"].(string)[24:]
	}

	log.Warnf("rtn v %v ", m["returnValue"].(string))

	fi.TxOfCreation = hash
	fi.FactoryAddress = c.Multisig.FactoryAddress
	deployStatus := int64(store.MultisigStatusRejected)
	if !m["failed"].(bool) {
		deployStatus = store.MultisigStatusDeployed
	}
	fi.DeployStatus = deployStatus

	c.Multisig.UsersContracts.Store(fi.Contract, fi.FactoryAddress)

	c.NewMultisig <- *fi

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

func (c *Client) GetInvocationStatus(hash string) (bool, string, error) {
	m, err := c.Rpc.TraceTransaction(hash)
	if err != nil {
		log.Errorf("FactoryContract:c.Rpc.TraceTransaction %v", err.Error())
		return false, "", err
	}

	isFailed, ok := m["failed"].(bool)
	if !ok {
		return false, "", errors.New("Trace failed unavalible ")
	}

	returnValue, ok := m["returnValue"].(string)
	if !ok {
		return false, "", errors.New("Trace returnValue unavalible ")
	}

	return !isFailed, returnValue, nil
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
