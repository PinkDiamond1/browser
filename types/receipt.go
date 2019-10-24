package types

import "math/big"

const (
	// ReceiptStatusFailed is the status code of a action if execution failed.
	ReceiptStatusFailed = 0

	// ReceiptStatusSuccessful is the status code of a action if execution succeeded.
	ReceiptStatusSuccessful = 1
)

//Reason 0 asset 1 contract 2 produce
type GasDistribution struct {
	Account Name   `json:"name"`
	Gas     uint64 `json:"gas"`
	Reason  uint64 `json:"typeId"`
}

// ActionResult represents the results the transaction action.
type ActionResult struct {
	GasAllot []*GasDistribution `json:"GasAllot"`
	Status   uint64             `json:"Status"`
	GasUsed  uint64             `json:"GasUsed"`
	Error    string             `json:"Error"`
	// Index    uint64
}

// Receipt represents the results of a transaction.
type Receipt struct {
	PostState         []byte          `json:"PostState"`
	ActionResults     []*ActionResult `json:"ActionResults"`
	CumulativeGasUsed uint64          `json:"CumulativeGasUsed"`
	TxHash            Hash            `json:"TxHash"`
	TotalGasUsed      uint64          `json:"TotalGasUsed"`
	//Logs              []*Log
}

func (g *GasDistribution) NewRpcAction() *RPCAction {
	action := RPCAction{
		To:     g.Account,
		Amount: big.NewInt(int64(g.Gas)),
		//From:   Name(g.Fromaccount),
	}
	return &action
}
