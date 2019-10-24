package types

type DetailTx struct {
	TxHash          Hash              `json:"txhash"`
	InternalActions []*InternalAction `json:"actions"`
}

type InternalAction struct {
	InternalLogs []*InternalLog `json:"internalActions"`
}

type InternalLog struct {
	Action     *RPCAction `json:"action"`
	ActionType string     `json:"actionType"`
	GasUsed    uint64     `json:"gasUsed"`
	GasLimit   uint64     `json:"gasLimit"`
	Depth      uint64     `json:"depth"`
	Error      string     `json:"error"`
}

type BlockAndResult struct {
	Block     *RpcBlock   `json:"block"`
	Receipts  []*Receipt  `json:"receipts"`
	DetailTxs []*DetailTx `json:"detailTxs"`
}
