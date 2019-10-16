package types

import (
	"math/big"
)

const (
	// ReceiptStatusFailed is the status code of a action if execution failed.
	ReceiptStatusFailed = 0

	// ReceiptStatusSuccessful is the status code of a action if execution succeeded.
	ReceiptStatusSuccessful = 1
)

//Reason 0 asset 1 contract 2 produce
type GasDistribution struct {
	Account     Name   `json:"account"`
	Gas         uint64 `json:"gas"`
	Reason      uint64
	//Fromaccount string
}

// ActionResult represents the results the transaction action.
type ActionResult struct {
	GasAllot []*GasDistribution
	Status   uint64
	// Index    uint64
	GasUsed uint64
	Error   string
}

// Receipt represents the results of a transaction.
type Receipt struct {
	// PostState         []byte
	ActionResults     []*ActionResult
	CumulativeGasUsed uint64
	// Logs              []*Log
	TxHash       Hash
	TotalGasUsed uint64
}

func (g *GasDistribution) NewRpcAction() *RPCAction {
	action := RPCAction{
		To:     g.Account,
		Amount: big.NewInt(int64(g.Gas)),
		//From:   Name(g.Fromaccount),
	}
	return &action
}
