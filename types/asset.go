package types

import (
	"math/big"
)

type AssetObject struct {
	AssetId    uint64   `json:"assetId,omitempty"`
	Number     uint64   `json:"number,omitempty"`
	AssetName  string   `json:"assetName"`
	Symbol     string   `json:"symbol"`
	Amount     *big.Int `json:"amount"`
	Decimals   uint64   `json:"decimals"`
	Founder    Name     `json:"founder"`
	Owner      Name     `json:"owner"`
	AddIssue   *big.Int `json:"addIssue"`
	UpperLimit *big.Int `json:"upperLimit"`
	Contract   Name     `json:"contract"`
	Detail     string   `json:"detail"`
}

type IssueAssetObject struct {
	AssetName   string   `json:"assetName"`
	Symbol      string   `json:"symbol"`
	Amount      *big.Int `json:"amount"`
	Decimals    uint64   `json:"decimals"`
	Founder     Name     `json:"founder"`
	Owner       Name     `json:"owner"`
	UpperLimit  *big.Int `json:"upperLimit"`
	Contract    Name     `json:"contract"`
	Description string   `json:"description"`
}

type IncAssetObject struct {
	AssetId uint64   `json:"assetId,omitempty"`
	Amount  *big.Int `json:"amount,omitempty"`
	To      Name     `json:"acceptor,omitempty"`
}

type UpdateAssetObject struct {
	AssetID uint64 `json:"assetId,omitempty"`
	Founder Name   `json:"founder"`
}

type UpdateAssetOwnerObject struct {
	AssetID uint64 `json:"assetId,omitempty"`
	Owner   Name   `json:"owner"`
}

type UpdateAssetContractObject struct {
	AssetID  uint64 `json:"assetId,omitempty"`
	Contract Name   `json:"contract"`
}
