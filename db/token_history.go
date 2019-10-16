package db

import (
	"database/sql"
	"fmt"

	. "github.com/browser_service/log"
	"go.uber.org/zap"
)

type MysqlTokenHistory struct {
	TokenId       uint64
	TxHash        string
	ActionIndex   int
	ActionHash    string
	InternalIndex int
	TxType        int
	ActionType    uint64
	Height        uint64
}

func InsertTokenHistory(data *MysqlTokenHistory, dbTx *sql.Tx) error {
	stmt, err := dbTx.Prepare(fmt.Sprintf("insert into token_history (token_id, tx_hash,action_index,action_hash,internal_index,tx_type,height,action_type) values (?,?,?,?,?,?,?,?)"))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("InsertTokenHistory error", zap.Error(err), zap.Uint64("tokenId", data.TokenId))
	}

	_, err = stmt.Exec(data.TokenId, data.TxHash, data.ActionIndex, data.ActionHash, data.InternalIndex, data.TxType, data.Height, data.ActionType)
	if err != nil {
		ZapLog.Panic("InsertTokenHistory error", zap.Error(err), zap.Uint64("tokenId", data.TokenId))
	}
	return nil
}

func DeleteTokenHistoryByHeight(height uint64, dbTx *sql.Tx) error {
	stmt, err := dbTx.Prepare(fmt.Sprintf("delete from token_history where height = ?"))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("DeleteTokenHistoryByHeight error", zap.Error(err), zap.Uint64("height", height))
	}

	_, err = stmt.Exec(height)
	if err != nil {
		ZapLog.Panic("DeleteTokenHistoryByHeight error", zap.Error(err), zap.Uint64("height", height))
	}
	return nil
}
