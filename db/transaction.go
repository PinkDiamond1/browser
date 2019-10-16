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

type MysqlTx struct {
	Hash        string
	Height      uint64
	GasUsed     uint64
	GasCost     *big.Int
	GasPrice    *big.Int
	GasAssetId  uint64
	State       int
	BlockHash   string
	TxIndex     int
	ActionCount int
}

func InsertTransaction(data *MysqlTx, tx *sql.Tx) error {
	tName := GetTableNameHash("txs_hash", data.Hash)
	stmt, err := tx.Prepare(fmt.Sprintf("insert into %s (hash, height, gas_used, gas_cost, gas_price, gas_asset_id, state, block_hash, tx_index, action_count) values (?,?,?,?,?,?,?,?,?,?)", tName))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("InsertTransaction error", zap.Error(err), zap.String("hash", data.Hash))
	}
	state := strconv.Itoa(data.State)
	_, err = stmt.Exec(data.Hash, data.Height, data.GasUsed, data.GasCost.String(), data.GasPrice.Uint64(), data.GasAssetId, state, data.BlockHash, data.TxIndex, data.ActionCount)
	if err != nil {
		ZapLog.Panic("InsertTransaction error", zap.Error(err), zap.String("hash", data.Hash))
	}
	return nil
}

func DeleteTransactionByHash(hash types.Hash, tx *sql.Tx) error {
	tName := GetTableNameHash("txs_hash", hash.String())
	stmt, err := tx.Prepare(fmt.Sprintf("delete from %s where hash = %s", tName, hash.String()))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("DeleteTransactionByHash error", zap.Error(err), zap.String("hash", hash.String()))
	}

	_, err = stmt.Exec()
	if err != nil {
		ZapLog.Panic("DeleteTransactionByHash error", zap.Error(err), zap.String("hash", hash.String()))
	}
	return nil
}
