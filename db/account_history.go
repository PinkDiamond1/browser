package db

import (
	"database/sql"
	"fmt"
	"strconv"

	. "github.com/browser/log"
	"go.uber.org/zap"
)

type MysqlAccountHistory struct {
	Account     string
	TxHash      string
	ActionHash  string
	ActionIndex int
	OtherIndex  int
	TxType      int
	Height      uint64
}

func InsertAccountHistory(data *MysqlAccountHistory, dbTx *sql.Tx) error {
	tName := GetTableNameHash("account_action_history_hash", data.Account)
	stmt, err := dbTx.Prepare(fmt.Sprintf("insert into %s (account_name, tx_hash,action_hash, action_index,other_index,tx_type,height) values (?,?,?,?,?,?,?)", tName))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("InsertAccountHistory error", zap.Error(err), zap.String("table", tName), zap.String("account history", fmt.Sprint(data)))
	}

	_, err = stmt.Exec(data.Account, data.TxHash, data.ActionHash, data.ActionIndex, data.OtherIndex, strconv.Itoa(data.TxType), data.Height)
	if err != nil {
		ZapLog.Panic("InsertAccountHistory error", zap.Error(err), zap.String("table", tName), zap.String("account history", fmt.Sprint(data)))
	}
	return nil
}

func DeleteAccountHistoryByHeight(account string, height uint64, dbTx *sql.Tx) error {
	tName := GetTableNameHash("account_action_history_hash", account)
	stmt, err := dbTx.Prepare(fmt.Sprintf("delete from %s where account_name = '%s' and height = %d", tName, account, height))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("DeleteAccountHistoryByHeight error", zap.Error(err), zap.String("account", account), zap.Uint64("height", height))
	}

	_, err = stmt.Exec()
	if err != nil {
		ZapLog.Panic("DeleteAccountHistoryByHeight error", zap.Error(err), zap.String("account", account), zap.Uint64("height", height))
	}
	return nil
}
