package db

import (
	"database/sql"
	"fmt"

	. "github.com/browser_service/log"
	"go.uber.org/zap"
)

type MysqlTokenFeeHistory struct {
	TokenId        uint64
	TxHash         string
	ActionIndex    int
	ActionHash     string
	FeeActionIndex int
	Height         uint64
}

func InsertTokenFeeHistory(data *MysqlTokenFeeHistory, dbTx *sql.Tx) error {
	tName := GetTableNameID2("token_fee_history_id", data.TokenId)
	insertSql := fmt.Sprintf("insert into %s (token_id, tx_hash, action_index, action_hash,fee_action_index,height) "+
		" values(%d,'%s',%d,'%s',%d,%d)", tName, data.TokenId, data.TxHash, data.ActionIndex, data.ActionHash, data.FeeActionIndex, data.Height)
	_, err := dbTx.Exec(insertSql)
	if err != nil {
		ZapLog.Error("InsertTokenFeeHistory error: ", zap.Error(err), zap.String("sql!!!", insertSql))
		return err
	}
	return nil
}

func DeleteTokenFeeHistoryByHeight(tokenId, height uint64, dbTx *sql.Tx) error {
	tName := GetTableNameID2("token_fee_history_id", tokenId)
	stmt, err := dbTx.Prepare(fmt.Sprintf("delete from %s where height = ?", tName))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("DeleteTokenFeeHistory error", zap.Error(err))
	}

	_, err = stmt.Exec(height)
	if err != nil {
		ZapLog.Panic("DeleteTokenFeeHistory error", zap.Error(err))
	}
	return nil
}
