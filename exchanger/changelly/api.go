package changelly

import "time"

const (
	ExchangeChangellyCanonicalName  = "changelly"
	RpcGetCurrencies = "getCurrencies"
	RpcGetTransactionMinimumAmount = "getMinAmount"
)


type ExchangerChangelly struct {
	name string
	config InitConfig
}

type InitConfig struct {
	apiUrl string
	apiKey string
	apiSecret string
}

type rpcPacket struct {
	id string					`json:"id"`
	jsonrpc string				`json:"jsonrpc"`
	method string				`json:"method""`
	params map[string]string	`json:"params"`
}

func (ec *ExchangerChangelly) Init(config interface{}) error {
	ec.name = ExchangeChangellyCanonicalName
	ec.config = config.(InitConfig)

	return nil
}

func (ec *ExchangerChangelly) GetName() string {
	return ec.name
}

func (ec *ExchangerChangelly) GetSupportedCurrencies() {

}

func (ec *ExchangerChangelly) sendRequest(methodName string, params map[string]string) {
	var request = rpcPacket{
		id: string(int32(time.Now().Unix())),
		jsonrpc: "2.0",
		method: methodName,
		params: params,
	}

	requestHash, err := ec.getRequestHash(request)

	var headers = map[string]string{
		"api-key": ec.config.apiKey,
		"sign": requestHash,
		"Content-type": "application/json",
	}

}

func (ec *ExchangerChangelly) getRequestHash(request rpcPacket) (string, error) {
	return "", nil
}

