package client

import (
	"encoding/json"
	"errors"
	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
)

const (
	methodCurrentBlock           = "ft_getCurrentBlock"
	methodBlockByNumber          = "ft_getBlockByNumber"
	methodChainConfig            = "ft_getChainConfig"
	methodBlockAndResultByNumber = "ft_getBlockAndResultByNumber"
	methodSendRawTransaction     = "ft_sendRawTransaction"
	methodAssetInfoByName        = "account_getAssetInfoByName"
	methodAssetInfoByID          = "account_getAssetInfoByID"
	methodGetAccountByName       = "account_getAccountByName"
	methodGetCode                = "account_getCode"
	methodDposCadidatesSize      = "dpos_candidatesSize"
	methodDposIrreversible       = "dpos_irreversible"
)

func GetCurrentBlockInfo() (*types.RpcBlock, error) {
	data := &types.RpcBlock{}
	err := GetData(methodCurrentBlock, data, true)
	if err != nil {
		ZapLog.Error("methodCurrentBlock error", zap.Error(err))
		return nil, err
	}
	data.Time = data.Time / 1000000000
	return data, nil
}

func GetBlockByNumber(height uint64) (*types.RpcBlock, error) {
	data := &types.RpcBlock{}
	err := GetData(methodBlockByNumber, data, height, true)
	if err != nil {
		ZapLog.Error("GetBlockByNumber error", zap.Error(err))
		return nil, err
	}
	data.Time = data.Time / 1000000000
	return data, nil
}

func GetAssetInfoByName(assetName string) (*types.AssetObject, error) {
	data := &types.AssetObject{}
	err := GetData(methodAssetInfoByName, data, assetName)
	if err != nil {
		ZapLog.Error("GetAssetInfoByName error", zap.Error(err), zap.String("assetName", assetName))
		return nil, err
	}
	return data, nil
}

func GetAssetInfoById(assetid uint64) (*types.AssetObject, error) {
	data := &types.AssetObject{}
	err := GetData(methodAssetInfoByID, data, assetid)
	if err != nil {
		ZapLog.Error("GetAssetInfoById error", zap.Error(err), zap.Uint64("assetid", assetid))
		return nil, err
	}
	return data, nil
}

func GetCandidatesCount() (uint64, error) {
	data := uint64(0)
	err := GetData(methodDposCadidatesSize, &data, 0)
	if err != nil {
		ZapLog.Error("GetCandidatesCount error", zap.Error(err))
		return data, err
	}
	return data, nil
}

func GetBlockAndResult(number int64) (*types.BlockAndResult, error) {
	data := &types.BlockAndResult{}
	err := GetData(methodBlockAndResultByNumber, &data, number)
	if err == ErrNull {
		return nil, err
	}
	if err != nil {
		ZapLog.Error("GetBlockAndResult error", zap.Error(err), zap.Int64("height", number))
		return data, err
	}
	data.Block.Time = data.Block.Time / 1000000000
	return data, nil
}

func GetChainConfig() (*types.ChainConfig, error) {
	data := &types.ChainConfig{}
	err := GetData(methodChainConfig, &data)
	if err != nil {
		ZapLog.Error("GetChainConfig error", zap.Error(err))
		return data, err
	}
	return data, nil
}

func GetAccountByName(name string) (*types.Account, error) {
	data := &types.Account{}
	err := GetData(methodGetAccountByName, &data, name)
	if err != nil {
		ZapLog.Error("GetAccountByName error", zap.Error(err), zap.String("name", name))
		return data, err
	}
	return data, nil
}

func GetDposIrreversible() (*types.DposIrreversible, error) {
	data := &types.DposIrreversible{}
	err := GetData(methodDposIrreversible, &data)
	if err != nil {
		ZapLog.Error("GetDposIrreversible error", zap.Error(err))
		return data, err
	}
	return data, nil
}

func GetCode(name string) (*types.Code, error) {
	data := &types.Code{}
	err := GetData(methodGetCode, &data)
	if err != nil {
		ZapLog.Error("GetCode error", zap.Error(err), zap.String("name", name))
		return data, err
	}
	return data, nil
}

func GetData(method string, outData interface{}, params ...interface{}) error {
	request := NewRPCRequest("2.0", method, params...)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("GetData error", zap.Error(err), zap.String("method", method))
		return err
	}
	result := jsonParsed.Path("result")
	if result.Data() == nil {
		return ErrNull
	}
	err = json.Unmarshal([]byte(result.String()), outData)
	if err != nil {
		ZapLog.Error("Unmarshal error", zap.Error(err), zap.String("method", method))
		return err
	}
	return nil

}

func SendRawTransaction(rawData string) (types.Hash, error) {
	nullHash := types.Hash{}
	request := NewRPCRequest("2.0", methodSendRawTransaction, rawData)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("SendRawTransaction SendRPCRequest error", zap.Error(err), zap.String("rawData", rawData))
		return nullHash, err
	}
	var resErr error
	if result := jsonParsed.Path("error").Data(); result != nil {
		errMap := jsonParsed.S("error").ChildrenMap()
		errMsg := errMap["message"].Data().(string)
		resErr = errors.New(errMsg)
		return nullHash, resErr
	}
	if result := jsonParsed.Path("result").Data(); result == nil {
		return nullHash, errors.New("get hash failed")
	} else {
		hash := types.HexToHash(result.(string))
		return hash, nil
	}
}
