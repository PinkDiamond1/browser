package types

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)

// ActionType type of Action.
type ActionType uint64

const (
	// CallContract represents the call contract action.
	CallContract ActionType = iota
	// CreateContract repesents the create contract action.
	CreateContract
)

const (
	//CreateAccount repesents the create account.
	CreateAccount ActionType = 0x100 + iota
	//UpdateAccount repesents update account.
	UpdateAccount
	// DeleteAccount repesents the delete account action.
	DeleteAccount
	//UpdateAccountAuthor represents the update account author.
	UpdateAccountAuthor
)

const (
	// IncreaseAsset Asset operation
	IncreaseAsset ActionType = 0x200 + iota
	// IssueAsset repesents Issue asset action.
	IssueAsset
	//DestroyAsset destroy asset
	DestroyAsset
	// SetAssetOwner repesents set asset new owner action.
	SetAssetOwner
	//SetAssetFounder set asset founder
	//SetAssetFounder
	UpdateAsset
	//Transfer repesents transfer asset action.
	Transfer
	UpdateAssetContract
)

const (
	// RegCandidate repesents register candidate action.
	RegCandidate ActionType = 0x300 + iota
	// UpdateCandidate repesents update candidate action.
	UpdateCandidate
	// UnregCandidate repesents unregister candidate action.
	UnregCandidate
	// RefundCandidate repesents unregister candidate action.
	RefundCandidate
	// VoteCandidate repesents voter vote candidate action.
	VoteCandidate
)

const (
	// KickedCandidate kicked
	KickedCandidate ActionType = 0x400 + iota
	// ExitTakeOver exit
	ExitTakeOver
)

const (
	// WithdrawFee
	WithdrawFee ActionType = 0x500 + iota
)

var ActionTypeToString map[ActionType]string = map[ActionType]string{
	CallContract:        "CallContract",
	CreateContract:      "CreateContract",
	CreateAccount:       "CreateAccount",
	UpdateAccount:       "UpdateAccount",
	IncreaseAsset:       "IncreaseAsset",
	IssueAsset:          "IssueAsset",
	DestroyAsset:        "DestroyAsset",
	SetAssetOwner:       "SetAssetOwner",
	Transfer:            "Transfer",
	UpdateAccountAuthor: "UpdateAccountAuthor",
	UpdateCandidate:     "UpdateCandidate",
	UnregCandidate:      "UnregCandidate",
	RefundCandidate:     "RefundCandidate",
	VoteCandidate:       "VoteCandidate",
	WithdrawFee:         "WithdrawFee",
	UpdateAsset:         "UpdateAsset",
	RegCandidate:        "RegCandidate",
	KickedCandidate:     "KickedCandidate",
	ExitTakeOver:        "ExitTakeOver",
	UpdateAssetContract: "UpdateAssetContract",
}

type RPCAction struct {
	Type          ActionType    `json:"type"`
	Nonce         uint64        `json:"nonce"`
	From          Name          `json:"from"`
	To            Name          `json:"to"`
	AssetID       uint64        `json:"assetID"`
	GasLimit      uint64        `json:"gas"`
	Amount        *big.Int      `json:"value"`
	Remark        hexutil.Bytes `json:"remark"`
	Payload       hexutil.Bytes `json:"payload"`
	//PayloadParsed interface{}   `json:"payload_parsed"`
	ActionHash    Hash          `json:"action_hash"`
}
