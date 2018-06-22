package eth

import (
	"fmt"
)

func (c *Client) FactoryContract(hash string) {
	fmt.Println("Hash: ", hash)
	m, err := c.Rpc.TraceTransaction(hash)
	if err != nil {
		fmt.Printf("FactoryContract:c.Rpc.TraceTransaction %v", err.Error())
	}

	//TODO: send faled
	isFailed := m["failed"].(bool)
	//TODO: send to channel
	if !isFailed {
		rawTx, err := c.Rpc.EthGetTransactionByHash(hash)
		if err != nil {
			log.Errorf("FactoryContract:Get TX Err: %s", err.Error())
		}

		fi, err := parseFactoryInput(rawTx.Input)
		if err != nil {
			log.Errorf("FactoryContract:parseInput: %s", err.Error())
		}
		// returnValue is contract address
		fi.Contract = "0x" + m["returnValue"].(string)[24:]

		//TODO: send to multy-back
		fmt.Println("\nfi ", fi)
	}

}

func (c *Client) MultisigContract(hash string) {
	m, err := c.Rpc.TraceTransaction(hash)
	if err != nil {
		fmt.Printf("FactoryContract:c.Rpc.TraceTransaction %v", err.Error())
	}
	//TODO: send faled
	isFailed := m["failed"].(bool)
	if !isFailed {

		rawTx, err := c.Rpc.EthGetTransactionByHash(hash)
		if err != nil {
			log.Errorf("FactoryContract:Get TX Err: %s", err.Error())
		}

		switch rawTx.Input[:10] {
		case submitTransaction: // "c6427474": "submitTransaction(address,uint256,bytes)"
			// TODO: feth contract owners, send notfy to owners about transation. status: waiting for confirmations
		case executeTransaction: // "ee22610b": "executeTransaction(uint256)"
			// TODO: feth contract owners, send notfy to owners about transation. status: conformed transatcion
		case confirmTransaction: // "c01a8c84": "confirmTransaction(uint256)"
			// TODO: send notfy to owners about +1 confirmation
		case revokeConfirmation: // "20ea8d86": "revokeConfirmation(uint256)"
			// TODO: send notfy to owners about -1 confirmation
		case "": // simple transaction

		default:
			// wrong method
		}
		// n, _ := new(big.Int).SetString(m["returnValue"].(string), 10)
		// nonce = n.Int64()

	}
	// m["returnValue"],

}
