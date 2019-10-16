package types

type DetailTx struct {
	TxHash          Hash
	InternalActions []*InternalAction
}

type InternalAction struct {
	InternalLogs []*InternalLog
}

type InternalLog struct {
	Action     *RPCAction
	ActionType string
	GasUsed    uint64
	GasLimit   uint64
	Depth      uint64
	Error      string
}

type BlockAndResult struct {
	Hash      Hash
	Block     *Block
	Receipts  []*Receipt
	DetailTxs []*DetailTx
}
