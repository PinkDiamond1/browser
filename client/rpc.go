package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/browser/config"
	. "github.com/browser/log"
	"go.uber.org/zap"
)

var (
	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
			//DisableKeepAlives:   true,
		},
		Timeout: time.Second * 500,
	}
)

type RPCRequest struct {
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

func NewRPCRequest(jsonRpc string, method string, param ...interface{}) *RPCRequest {
	r := new(RPCRequest)
	r.JsonRpc = jsonRpc
	r.Method = method
	r.Params = make([]interface{}, 0)
	r.Params = append(r.Params, param...)
	return r
}

func sendRPCRequest(rpcRequest *RPCRequest) (*gabs.Container, error) {
	host := config.Node.RpcUrl
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(rpcRequest); err != nil {
		ZapLog.Error("SendRPCRequest EncodeRequest error ", zap.Error(err))
		return nil, err
	}

	req, _ := http.NewRequest("POST", host, &buff)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		ZapLog.Error("SendRPCRequest error", zap.Error(err), zap.String("host", host), zap.String("request", buff.String()))
		return nil, err
	}
	defer resp.Body.Close()
	jsonParsed, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		ZapLog.Error("SendRPCRequest ParseJSONBuffer error", zap.Error(err))
		return nil, err
	}
	return jsonParsed, nil
}

func SendRPCRequest(rpcRequest *RPCRequest) (gc *gabs.Container, err error) {
	for i := 0; i < 10; i++ {
		gc, err := sendRPCRequest(rpcRequest)
		if err == nil {
			return gc, nil
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	return nil, err
}

func SendRPCRequstWithAuth(host string, username string, password string, rpcRequest *RPCRequest) (*gabs.Container, error) {
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(rpcRequest); err != nil {
		return nil, fmt.Errorf("SendRPCRequst EncodeRequest error --- %s", err)
	}
	req, _ := http.NewRequest("POST", host, &buff)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SendRPCRequst Post %s error --- %s(%s)", host, err, buff.String())
	}
	defer resp.Body.Close()
	jsonParsed, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("SendRPCRequst ParseJSONBuffer error --- %s(%s)", err, buff.String())
	}
	return jsonParsed, nil
}
