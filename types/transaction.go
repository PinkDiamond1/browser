package types

import (
	"math/big"
)

//type RPCTransaction struct {
//	RPCActions []*RPCAction `json:"actions"`
//	GasAssetID uint64       `json:"gasAssetID"`
//	GasPrice   *big.Int     `json:"gasPrice"`
//	GasCost    *big.Int     `json:"gasCost"`
//	Hash       Hash         `json:"txHash"`
//}

type RPCTransaction struct {
	BlockHash        Hash         `json:"blockHash"`
	BlockNumber      uint64       `json:"blockNumber"`
	Hash             Hash         `json:"txHash"`
	TransactionIndex uint64       `json:"transactionIndex"`
	RPCActions       []*RPCAction `json:"actions"`
	GasAssetID       uint64       `json:"gasAssetID"`
	GasPrice         *big.Int     `json:"gasPrice"`
	GasCost          *big.Int     `json:"gasCost"`
}
