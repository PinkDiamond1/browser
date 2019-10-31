package mem

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	// . "github.com/browser/log"

	"github.com/abi"
	"github.com/browser/config"
	"github.com/browser/types"
)

var (
	abifile = "./reward.abi"
	from    = types.Name("dposcontract.reward")
)

type rpcRequest struct {
	ID      int         `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

//RPCResponse .
type RPCResponse struct {
	ID      int         `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
}

//FractalRequest .
type FractalRequest struct {
	//调用ft_call
	ChainID    int      `json:"chainID"`
	ActionType int      `json:"actionType"`
	GasAssetID int      `json:"gasAssetId"`
	From       string   `json:"from"`
	To         string   `json:"to"`
	Nonce      uint64   `json:"nonce"`
	AssetID    uint64   `json:"assetID"`
	Gas        uint64   `json:"gas"`
	GasPrice   *big.Int `json:"gasPrice"`
	Value      *big.Int `json:"value"`
	Data       string   `json:"data"`
	Password   string   `json:"password"`
}

// // func init() {
// // 	testcommon.SetDefultURL(config.Node.RpcUrl)
// // }

func newFractalRequest() *FractalRequest {
	var fr FractalRequest
	fr.ChainID = 1
	fr.ActionType = 0
	fr.GasAssetID = 0
	fr.From = from.String()
	fr.To = from.String()
	fr.Nonce = 1
	fr.AssetID = 0
	fr.Gas = 3000000000
	fr.GasPrice = big.NewInt(1)
	fr.Value = big.NewInt(0)
	return &fr
}

func doPost(jsonRequest []byte) ([]byte, error) {
	resp, err := http.Post(config.Node.RpcUrl, "application/json", bytes.NewReader(jsonRequest))
	if err != nil {
		return nil, fmt.Errorf("execute http post request error, error message: %s", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read message error, %s", err)
	}

	if resp.StatusCode == http.StatusOK {
		return body, nil
	} else {
		return nil, fmt.Errorf("execute http post request status error, error code: %d", resp.StatusCode)
	}
}

func transaction(request *FractalRequest) (string, error) {
	var interF []interface{}
	interF = append(interF, *request)
	interF = append(interF, "latest")
	req := rpcRequest{ID: 1, Jsonrpc: "2.0", Method: "ft_call", Params: interF}

	b, _ := json.Marshal(req)

	body, err := doPost(b)
	if err != nil {
		return "", err
	}

	resp := RPCResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return "", fmt.Errorf("json parse message error, %s", err)
	}

	return resp.Result.(string), nil
}

func input(abifile string, method string, params ...interface{}) (string, error) {
	var abicode string

	hexcode, err := ioutil.ReadFile(abifile)
	if err != nil {
		// ZapLog.Error("Could not load code from file: %v\n", zap.Error(err))
		return "", err
	}
	abicode = string(bytes.TrimRight(hexcode, "\n"))

	parsed, err := abi.JSON(strings.NewReader(abicode))
	if err != nil {
		// ZapLog.Error("abi.json error ", zap.Error(err))
		return "", err
	}

	input, err := parsed.Pack(method, params...)
	if err != nil {
		// ZapLog.Error("parsed.pack error ", zap.Error(err))
		return "", err
	}
	return types.Bytes2Hex(input), nil
}

// //RewardInfo .
// type RewardInfo struct {
// 	Time        *big.Int   `abi:"_rewardTime"`
// 	Num         *big.Int   `abi:"_cycleNum"`
// 	Amount      *big.Int   `abi:"_amount"`
// 	LockRatio   *big.Int   `abi:"_lockRatio"`
// 	SingleValue []*big.Int `abi:"_singleTicketValue"`
// 	Weights     []*big.Int `abi:"_weights"`
// }

// func getRewardInfo(index *big.Int) (*RewardInfo, error) {
// 	input, err := input(abifile, "getRewardInfo", index)
// 	if err != nil {
// 		return nil, err
// 	}
// 	fr := newFractalRequest()
// 	fr.Data = "0x" + input
// 	result, _ := transaction(fr)
// 	result = result[2:]
// 	abiFile, _ := os.Open(abifile)
// 	abi1, _ := abi.JSON(abiFile)
// 	r, err := hex.DecodeString(result)
// 	if err != nil {
// 		return nil, err
// 	}
// 	rewardInfo := RewardInfo{}
// 	err = abi1.Unpack(&rewardInfo, "getRewardInfo", r)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &rewardInfo, nil
// }

// //CycleInfo .
// type CycleInfo struct {
// 	Time      *big.Int   `abi:"_time"`
// 	Indexs    []*big.Int `abi:"_indexs"`
// 	Producers []*big.Int `abi:"_ranking"`
// }

// func getCycleInfo(epoch int64) (*CycleInfo, error) {
// 	input, err := input(abifile, "getCycleInfo", big.NewInt(epoch))
// 	if err != nil {
// 		return nil, err
// 	}
// 	fr := newFractalRequest()
// 	fr.Data = "0x" + input
// 	result, _ := transaction(fr)
// 	result = result[2:]

// 	abiFile, _ := os.Open(abifile)
// 	abi1, _ := abi.JSON(abiFile)
// 	cycleInfo := CycleInfo{}
// 	r, err := hex.DecodeString(result)
// 	if err != nil {
// 		return &cycleInfo, err
// 	}

// 	err = abi1.Unpack(&cycleInfo, "getCycleInfo", r)
// 	if err != nil {
// 		return &cycleInfo, err
// 	}
// 	return &cycleInfo, nil
// }

// //VoterRange .
// type VoterRange struct {
// 	Start *big.Int `abi:"_start"`
// 	End   *big.Int `abi:"_end"`
// }

// func GetVoterGetRange(accountID int64) (*VoterRange, error) {
// 	input, err := input(abifile, "getVoterGetRange", common.BigToAddress(big.NewInt(accountID)))
// 	if err != nil {
// 		return nil, err
// 	}
// 	fr := newFractalRequest()
// 	fr.Data = "0x" + input
// 	result, _ := transaction(fr)
// 	result = result[2:]

// 	abiFile, _ := os.Open(abifile)
// 	abi1, _ := abi.JSON(abiFile)
// 	voterRange := VoterRange{}
// 	r, err := hex.DecodeString(result)
// 	if err != nil {
// 		return &voterRange, err
// 	}

// 	err = abi1.Unpack(&voterRange, "getVoterGetRange", r)
// 	if err != nil {
// 		return &voterRange, err
// 	}
// 	return &voterRange, nil
// }

// //VoterRewardAmount .
// type VoterRewardAmount struct {
// 	Num       *big.Int `abi:"_cycleNum"`
// 	LockRatio *big.Int `abi:"_lockRatio"`
// 	Amount    *big.Int `abi:"_amount"`
// }

// func GetVoterRewardAmount(accountID int64, index int64) (*VoterRewardAmount, error) {
// 	input, err := input(abifile, "getVoterRewardAmount", common.BigToAddress(big.NewInt(accountID)), big.NewInt(index))
// 	if err != nil {
// 		return nil, err
// 	}
// 	fr := newFractalRequest()
// 	fr.Data = "0x" + input
// 	result, _ := transaction(fr)
// 	result = result[2:]

// 	abiFile, _ := os.Open(abifile)
// 	abi1, _ := abi.JSON(abiFile)
// 	voterRewardAmount := VoterRewardAmount{}
// 	r, err := hex.DecodeString(result)
// 	if err != nil {
// 		return &voterRewardAmount, err
// 	}

// 	err = abi1.Unpack(&voterRewardAmount, "getVoterRewardAmount", r)
// 	if err != nil {
// 		return &voterRewardAmount, err
// 	}
// 	return &voterRewardAmount, nil
// }

// //ProducerRange .
// type ProducerRange struct {
// 	Start *big.Int `abi:"_start"`
// 	End   *big.Int `abi:"_end"`
// }

// func GetProducerGetRange(accountID int64) (*ProducerRange, error) {
// 	input, err := input(abifile, "getProducerGetRange", common.BigToAddress(big.NewInt(accountID)))
// 	if err != nil {
// 		return nil, err
// 	}
// 	fr := newFractalRequest()
// 	fr.Data = "0x" + input
// 	result, _ := transaction(fr)
// 	result = result[2:]

// 	abiFile, _ := os.Open(abifile)
// 	abi1, _ := abi.JSON(abiFile)
// 	producerRange := ProducerRange{}
// 	r, err := hex.DecodeString(result)
// 	if err != nil {
// 		return &producerRange, err
// 	}

// 	err = abi1.Unpack(&producerRange, "getProducerGetRange", r)
// 	if err != nil {
// 		return &producerRange, err
// 	}
// 	return &producerRange, nil
// }

// //ProducerRewardAmount .
// type ProducerRewardAmount struct {
// 	Num       *big.Int `abi:"_cycleNum"`
// 	LockRatio *big.Int `abi:"_lockRatio"`
// 	Amount    *big.Int `abi:"_amount"`
// }

// func GetProducerRewardAmount(accountID int64, index int64) (*ProducerRewardAmount, error) {
// 	input, err := input(abifile, "getProducerRewardAmount", common.BigToAddress(big.NewInt(accountID)), big.NewInt(index))
// 	if err != nil {
// 		return nil, err
// 	}
// 	fr := newFractalRequest()
// 	fr.Data = "0x" + input
// 	result, _ := transaction(fr)
// 	result = result[2:]

// 	abiFile, _ := os.Open(abifile)
// 	abi1, _ := abi.JSON(abiFile)
// 	producerRewardAmount := ProducerRewardAmount{}
// 	r, err := hex.DecodeString(result)
// 	if err != nil {
// 		return &producerRewardAmount, err
// 	}

// 	err = abi1.Unpack(&producerRewardAmount, "getProducerRewardAmount", r)
// 	if err != nil {
// 		return &producerRewardAmount, err
// 	}
// 	return &producerRewardAmount, nil
// }
