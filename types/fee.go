package types

import (
	"math/big"
)

//WithdrawInfo record withdraw info
type WithdrawInfo struct {
	ObjectName string
	ObjectType uint64
	Founder    Name
	AssetInfo  []*WithdrawAsset
}

//WithdrawAsset  withdraw asset info
type WithdrawAsset struct {
	AssetID uint64
	Amount  *big.Int
}

//AssetFee asset fee
type AssetFee struct {
	AssetID  uint64   `json:"assetID”`  //资产ID
	TotalFee *big.Int `json:"totalFee”` //收到过的手续费数量
	// RemainFee *big.Int `json:"remainFee”` //未提取的手续费数量
}

type ObjectFee struct {
	ObjectFeeID uint64      `json:"objectFeeID”` // 为每个手续费对象分配的ID
	ObjectType  uint64      `json:"objectType”`  //手续费类型，0：资产 1：合约 2：矿工
	ObjectName  string      `json:"objectName”`  //对象名称：账号名或资产名
	AssetFees   []*AssetFee `json:"assetFee”`    //对象收到的各种资产类型的手续费数量，目前只有FT
}

type ObjectFeeResult struct {
	Continue   bool         `json:"continue"`
	ObjectFees []*ObjectFee `json:"objectFees"`
}
