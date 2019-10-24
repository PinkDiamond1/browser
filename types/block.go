package types

import (
	"fmt"
	"github.com/browser/common"
	"github.com/browser/rlp"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)

//// Header represents a block header in the block chain.
//type Header struct {
//	ParentHash Hash     `json:"parentHash"`
//	Coinbase   Name     `json:"miner"`
//	Difficulty *big.Int `json:"difficulty"`
//	Number     *big.Int `json:"number"`
//	GasLimit   uint64   `json:"gasLimit"`
//	GasUsed    uint64   `json:"gasUsed"`
//	Time       uint     `json:"timestamp"`
//	Extra      []byte   `json:"extraData"`
//}
//
//// Block represents an entire block in the block chain.
//type Block struct {
//	Head *Header
//	Txs  []*RPCTransaction
//}

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

type ForkID struct {
	Cur  uint64 `json:"cur"`
	Next uint64 `json:"next"`
}

const BloomByteLength = 256

type Bloom [BloomByteLength]byte

// MarshalText encodes b as a hex string with 0x prefix.
func (b Bloom) MarshalText() ([]byte, error) {
	return hexutil.Bytes(b[:]).MarshalText()
}

// UnmarshalText b as a hex string with 0x prefix.
func (b *Bloom) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Bloom", input, b[:])
}

type RpcBlock struct {
	Number               *big.Int          `json:"number"`
	Hash                 Hash              `json:"hash"`
	ProposedIrreversible uint64            `json:"proposedIrreversible"`
	ParentHash           Hash              `json:"parentHash"`
	Bloom                Bloom             `json:"logsBloom"`
	Root                 Hash              `json:"stateRoot"`
	CoinBase             Name              `json:"miner"`
	Difficulty           *big.Int          `json:"difficulty"`
	Extra                []byte            `json:"extraData"`
	Size                 uint64            `json:"size"`
	GasLimit             uint64            `json:"gasLimit"`
	GasUsed              uint64            `json:"gasUsed"`
	Time                 uint64            `json:"timestamp"`
	TxsRoot              Hash              `json:"transactionsRoot"`
	ReceiptsRoot         Hash              `json:"receiptsRoot"`
	ForkID               ForkID            `json:"forkID"`
	TotalDifficulty      *big.Int          `json:"totalDifficulty"`
	Txs                  []*RPCTransaction `json:"transactions"`
}
