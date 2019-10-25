package types

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)

type AccountAuthor struct {
	AuthorType AuthorType
	Author     string
	Weight     uint64
}

type CreateAccountAction struct {
	AccountName Name   `json:"accountName,omitempty"`
	Founder     Name   `json:"founder,omitempty"`
	PublicKey   PubKey `json:"publicKey,omitempty"`
	Description string `json:"description,omitempty"`
}

type UpdateAccountAction struct {
	Founder Name `json:"founder,omitempty"`
}

type AssetBalance struct {
	AssetID uint64   `json:"assetID"`
	Balance *big.Int `json:"balance"`
}

type Account struct {
	AcctName              Name            `json:"accountName"`
	Founder               Name            `json:"founder"`
	AccountID             uint64          `json:"accountID"`
	Number                uint64          `json:"number"`
	Nonce                 uint64          `json:"nonce"`
	Code                  hexutil.Bytes   `json:"code"`
	CodeHash              Hash            `json:"codeHash"`
	CodeSize              uint64          `json:"codeSize"`
	Threshold             uint64          `json:"threshold"`
	UpdateAuthorThreshold uint64          `json:"updateAuthorThreshold"`
	AuthorVersion         Hash            `json:"authorVersion"`
	Balances              []*AssetBalance `json:"balances"`
	Authors               []*Author       `json:"authors"`
	Suicide               bool            `json:"suicide"`
	Destroy               bool            `json:"destroy"`
	Description           string          `json:"description"`
}

