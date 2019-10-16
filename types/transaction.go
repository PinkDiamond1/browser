package types

import (
	"math/big"
)

type RPCTransaction struct {
	RPCActions []*RPCAction `json:"actions"`
	GasAssetID uint64       `json:"gasAssetID"`
	GasPrice   *big.Int     `json:"gasPrice"`
	GasCost    *big.Int     `json:"gasCost"`
	Hash       Hash         `json:"txhash"`
}
