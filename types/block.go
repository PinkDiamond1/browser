package types

import (
	"fmt"
	"github.com/browser_service/common"
	"github.com/browser_service/rlp"
	"math/big"
)

// Header represents a block header in the block chain.
type Header struct {
	ParentHash Hash     `json:"parentHash"`
	Coinbase   Name     `json:"miner"`
	Difficulty *big.Int `json:"difficulty"`
	Number     *big.Int `json:"number"`
	GasLimit   uint64   `json:"gasLimit"`
	GasUsed    uint64   `json:"gasUsed"`
	Time       uint     `json:"timestamp"`
	Extra      []byte   `json:"extraData"`
}

// Block represents an entire block in the block chain.
type Block struct {
	Head *Header
	Txs  []*RPCTransaction
}

func RlpHash(x interface{}) (h Hash) {
	hw := common.Get256()
	defer common.Put256(hw)
	err := rlp.Encode(hw, x)
	if err != nil {
		panic(fmt.Sprintf("rlp hash encode err: %v", err))
	}
	hw.Sum(h[:0])
	return h
}
