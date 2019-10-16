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
