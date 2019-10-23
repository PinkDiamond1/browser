package db

import (
	"database/sql"
	"fmt"
	"math/big"

	. "github.com/browser/log"
	"go.uber.org/zap"
)

type MysqlBalance struct {
	Name         string
	AssetId      uint64
	Amount       *big.Int
	UpdateHeight uint64
	UpdateTime   uint64
}

func GetAccountBalance(name string, assetId uint64, dbTx *sql.Tx) (*big.Int, error) {
	var balance string
	tName := GetTableNameHash("balance_hash", name)
	querySql := fmt.Sprintf("select amount from %s where account_name = '%s' and asset_id = %d",
		tName, name, assetId)
	row := dbTx.QueryRow(querySql)
	err := row.Scan(&balance)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		ZapLog.Error("getAccountAssetBalance failed", zap.Error(err), zap.String("sql", querySql))
		return nil, err
	}
	data := big.NewInt(0)
	data.SetString(balance, 10)
	return data, nil
}

func UpdateAccountBalance(name string, amount *big.Int, assetId uint64, updateHeight uint64, updateTime uint, dbTx *sql.Tx) error {
	tName := GetTableNameHash("balance_hash", name)
	updateSql := fmt.Sprintf("update %s set amount = '%s', update_height = %d, update_time = %d where account_name = '%s' and asset_id = %d",
		tName, amount.String(), updateHeight, updateTime, name, assetId)
	_, err := dbTx.Exec(updateSql)
	if err != nil {
		ZapLog.Error("UpdateAccountBalance error: ", zap.Error(err), zap.String("sql", updateSql))
		return err
	}
	return nil
}

func InsertAccountBalance(name string, amount *big.Int, assetId uint64, height uint64, updateTime uint, dbTx *sql.Tx) error {
	tName := GetTableNameHash("balance_hash", name)
	insertSql := fmt.Sprintf("insert %s set account_name = '%s', asset_id = %d, amount = '%s', update_height = %d, update_time = %d",
		tName, name, assetId, amount.String(), height, updateTime)
	_, err := dbTx.Exec(insertSql)
	if err != nil {
		ZapLog.Error("UpdateAccountBalance error: ", zap.Error(err), zap.String("sql", insertSql))
		return err
	}
	return nil
}
