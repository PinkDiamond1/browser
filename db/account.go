package db

import (
	"database/sql"
	"fmt"
	. "github.com/browser/log"
	"go.uber.org/zap"
	"strings"
)

type MysqlAccount struct {
	Name                  string
	ParentName            string
	CreateUser            string
	Founder               string
	AccountID             uint64
	Number                uint64
	Nonce                 uint64
	AuthorVersion         string
	Threshold             uint64
	UpdateAuthorThreshold uint64
	Permissions           string
	ContractCode          string
	CodeHash              string
	Created               uint
	ContractCreated       uint
	Suicide               bool
	Destroy               bool
	Description           string
}

func InsertAccount(data *MysqlAccount, dbTx *sql.Tx) error {
	stmt, err := dbTx.Prepare(fmt.Sprintf("insert into account (s_name, parent_name, create_user, founder, account_id, account_number, nonce,author_version,threshold,update_author_threshold,permissions,created," +
		"contract_code,code_hash, contract_created,description,suicide,destroy) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("InsertAccount error", zap.Error(err), zap.String("name", data.Name))
	}

	_, err = stmt.Exec(data.Name, data.ParentName, data.CreateUser, data.Founder, data.AccountID, data.Number, data.Nonce, data.AuthorVersion, data.Threshold, data.UpdateAuthorThreshold, data.Permissions, data.Created,
		data.ContractCode, data.CodeHash, data.ContractCreated, data.Description, data.Suicide, data.Destroy)
	if err != nil {
		ZapLog.Panic("InsertAccount error", zap.Error(err), zap.String("name", data.Name))
	}
	return nil
}

func UpdateAccount(name string, values map[string]interface{}, dbTx *sql.Tx) error {
	var fields []string
	var params []interface{}
	for k, v := range values {
		tmp := fmt.Sprintf(" %s=?", k)
		fields = append(fields, tmp)
		params = append(params, v)
	}
	updateSql := fmt.Sprintf("update account set %s where s_name = '%s'",
		strings.Join(fields, ","), name)
	stmt, err := dbTx.Prepare(updateSql)
	defer stmt.Close()

	if err != nil {
		ZapLog.Error("updateAccount error: ", zap.Error(err), zap.String("sql", updateSql))
		return err
	}
	_, err = stmt.Exec(params...)
	if err != nil {
		ZapLog.Error("updateAccount error: ", zap.Error(err), zap.String("sql", updateSql))
		return err
	}
	return nil
}

func GetAccountByName(name string, dbTx *sql.Tx) (*MysqlAccount, error) {
	sqlStr := "select s_name, parent_name, create_user, founder, account_id, account_number, nonce,author_version,threshold,update_author_threshold,permissions,created," +
		"contract_code,code_hash, contract_created,description,suicide,destroy FROM account where s_name = ?"
	row := dbTx.QueryRow(sqlStr, name)
	a := &MysqlAccount{}
	err := row.Scan(&a.Name, &a.ParentName, &a.CreateUser, &a.Founder, &a.AccountID, &a.Number, &a.Nonce, &a.AuthorVersion, &a.Threshold, &a.UpdateAuthorThreshold, &a.Permissions, &a.Created,
		&a.ContractCode, &a.CodeHash, &a.ContractCreated, &a.Description, &a.Suicide, &a.Description)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		ZapLog.Error("GetAccount error", zap.String("sql", sqlStr))
		return nil, err
	}
	return a, nil
}

func DeleteAccountByName(name string, dbTx *sql.Tx) error {
	stmt, err := dbTx.Prepare(fmt.Sprintf("delete from acount where s_name = %s", name))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("DeleteAccountByName error", zap.Error(err), zap.String("name", name))
	}

	_, err = stmt.Exec()
	if err != nil {
		ZapLog.Panic("DeleteAccountByName error", zap.Error(err), zap.String("name", name))
	}
	return nil
}
