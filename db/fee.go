package db

import (
	"database/sql"
	"fmt"
	"math/big"
	"strconv"

	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
)

type MysqlFee struct {
	TxHash      string
	ActionHash  string
	ActionIndex int
	FeeIndex    int
	Height      uint64
	Created     uint
	AssetId     uint64
	From        string
	To          string
	Amount      *big.Int
	Reason      uint64
}

func InsertFee(data *MysqlFee, dbTx *sql.Tx) error {
	tName := GetTableNameHash("fee_actions_hash", data.TxHash)
	stmt, err := dbTx.Prepare(fmt.Sprintf("insert into %s (tx_hash, action_hash,action_index,fee_index,height,created,asset_id,"+
		"from_account, to_account, amount, reason) values (?,?,?,?,?,?,?,?,?,?,?)", tName))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("InsertFee error", zap.Error(err), zap.String("ActionHash", data.ActionHash))
	}
	reason := strconv.FormatUint(data.Reason, 10)
	_, err = stmt.Exec(data.TxHash, data.ActionHash, data.ActionIndex, data.FeeIndex, data.Height, data.Created, data.AssetId,
		data.From, data.To, data.Amount.String(), reason)
	if err != nil {
		ZapLog.Panic("InsertFee error", zap.Error(err), zap.String("ActionHash", data.ActionHash))
	}
	return nil
}

func DeleteFeeByActionHash(actionHash types.Hash, dbTx *sql.Tx) error {
	tName := GetTableNameHash("fee_actions_hash", actionHash.String())
	stmt, err := dbTx.Prepare(fmt.Sprintf("delete from %s where action_hash = %s", tName, actionHash.String()))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("DeleteFeeByActionHash error", zap.Error(err), zap.String("actionHash", actionHash.String()))
	}

	_, err = stmt.Exec()
	if err != nil {
		ZapLog.Panic("DeleteFeeByActionHash error", zap.Error(err), zap.String("actionHash", actionHash.String()))
	}
	return nil
}
