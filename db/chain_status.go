package db

import (
	"database/sql"
	"fmt"
	. "github.com/browser/log"
	"go.uber.org/zap"
	"math/big"
	"strings"
)

type MysqlChainStatus struct {
	Height         uint64
	TxCount        uint64
	ProducerNumber uint64
	FeeIncome      *big.Int
	TokenIncome    *big.Int
	ContractIncome *big.Int
}

func initChainStatus(db *sql.DB) error {
	insertSql := "insert into chain_status(height,tx_count,producer_number,fee_income,token_income,contract_income) " +
		"values(0, 0, 0, '0', '0', '0')"
	_, err := db.Exec(insertSql)
	if err != nil {
		ZapLog.Panic("InitChainStatus error", zap.Error(err))
	}
	return nil
}

func UpdateChainStatus(dbTx *sql.Tx, values map[string]interface{}) error {
	var fields []string
	var params []interface{}
	for k, v := range values {
		tmp := fmt.Sprintf(" %s=?", k)
		fields = append(fields, tmp)
		params = append(params, v)
	}
	updateSql := fmt.Sprintf("update chain_status set %s",
		strings.Join(fields, ","))
	stmt, err := dbTx.Prepare(updateSql)
	defer stmt.Close()

	if err != nil {
		ZapLog.Error("UpdateChainStatus error: ", zap.Error(err), zap.String("sql", updateSql))
		return err
	}
	_, err = stmt.Exec(params...)
	if err != nil {
		ZapLog.Error("UpdateChainStatus error: ", zap.Error(err), zap.String("sql", updateSql))
		return err
	}
	return nil
}


