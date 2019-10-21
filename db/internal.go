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

type MysqlInternal struct {
	TxHash        string
	ActionHash    string
	ActionIndex   int
	InternalIndex int
	Height        uint64
	Created       uint
	AssetId       uint64
	ActionType    uint64
	From          string
	To            string
	Amount        *big.Int
	GasLimit      uint64
	GasUsed       uint64
	Depth         uint64
	State         uint64
	ErrorMsg      string
	Payload       []byte
}

func InsertInternalAction(data *MysqlInternal, dbTx *sql.Tx) error {
	tName := GetTableNameHash("internal_actions_hash", data.ActionHash)
	stmt, err := dbTx.Prepare(fmt.Sprintf("insert into %s (tx_hash, action_hash,height,created,action_index,internal_index,asset_id,"+
		"action_type,from_account,to_account,amount,gas_limit,gas_used,depth,state,error_msg,payload) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", tName))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("InsertInternalAction error", zap.Error(err), zap.String("txHash", data.TxHash))
	}

	_, err = stmt.Exec(data.TxHash, data.ActionHash, data.Height, data.Created, data.ActionIndex, data.InternalIndex, data.AssetId,
		data.ActionType, data.From, data.To, data.Amount.String(), data.GasLimit, data.GasUsed, data.Depth, strconv.FormatUint(data.State, 10), data.ErrorMsg, data.Payload)
	if err != nil {
		ZapLog.Panic("InsertInternalAction error", zap.Error(err), zap.String("txHash", data.TxHash))
	}
	return nil
}

func DeleteInternalByActionHash(actionHash types.Hash, dbTx *sql.Tx) error {
	tName := GetTableNameHash("internal_actions_hash", actionHash.String())
	stmt, err := dbTx.Prepare(fmt.Sprintf("delete from %s where action_hash = %s", tName, actionHash.String()))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("DeleteInternalByActionHash error", zap.Error(err), zap.String("actionHash", actionHash.String()))
	}

	_, err = stmt.Exec()
	if err != nil {
		ZapLog.Panic("DeleteInternalByActionHash error", zap.Error(err), zap.String("actionHash", actionHash.String()))
	}
	return nil
}
