package db

import (
	"database/sql"
	"fmt"
	"math/big"

	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
)

func InsertBlockChain(tx *sql.Tx, hd *types.RpcBlock, bhash types.Hash, fee *big.Int, txCount int) error {
	tName := GetTableNameID1("block_id", hd.Number.Uint64())
	blocksql := fmt.Sprintf("INSERT INTO %s(hash, parent_hash, height, created, gas_limit, gas_used, producer, tx_count ,fee) VALUES('%s','%s',%d,%d,%d,%d,'%s',%d,'%s');",
		tName, bhash.String(), hd.ParentHash.String(), hd.Number.Int64(), hd.Time, hd.GasLimit, hd.GasUsed, hd.CoinBase.String(), txCount, fee.String())
	_, err := tx.Exec(blocksql)
	if err != nil {
		ZapLog.Error("insert block failed", zap.String("sql", blocksql), zap.Error(err))
		return err
	}
	return nil
}

func DeleteBlock(dbTx *sql.Tx, height uint64) error {
	tName := GetTableNameID1("block_id", height)
	stmt, err := dbTx.Prepare(fmt.Sprintf("delete from %s where height = %d", tName, height))
	defer stmt.Close()
	if err != nil {
		ZapLog.Error("DeleteBlock error", zap.Error(err), zap.Uint64("height", height))
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		ZapLog.Error("DeleteBlock error", zap.Error(err), zap.Uint64("height", height))
		return err
	}
	return nil
}

func InsertBlockTx(tx *sql.Tx, height uint64, txHash types.Hash) error {
	tName := GetTableNameID1("block_tx_rel_id", height)
	blockSql := fmt.Sprintf("INSERT INTO %s(height, tx_hash) VALUES(%d,'%s');",
		tName, height, txHash.String())
	_, err := tx.Exec(blockSql)
	if err != nil {
		ZapLog.Error("insert block tx failed", zap.String("sql", blockSql), zap.Error(err))
		return err
	}
	return nil
}

func DeleteBlockTx(dbTx *sql.Tx, height uint64) error {
	tName := GetTableNameID1("block_tx_rel_id", height)
	stmt, err := dbTx.Prepare(fmt.Sprintf("delete from %s where height = %d", tName, height))
	defer stmt.Close()
	if err != nil {
		ZapLog.Error("DeleteBlockTx error", zap.Error(err), zap.Uint64("height", height))
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		ZapLog.Error("DeleteBlockTx error", zap.Error(err), zap.Uint64("height", height))
		return err
	}
	return nil
}
