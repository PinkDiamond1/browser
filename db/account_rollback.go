package db

import (
	"database/sql"
	"fmt"
	. "github.com/browser/log"
	"go.uber.org/zap"
)

type MysqlAccountRollback struct {
	Account *MysqlAccount
	Height  uint64
}

func InsertAccountRollback(data *MysqlAccountRollback, dbTx *sql.Tx) error {
	stmt, err := dbTx.Prepare(fmt.Sprintf("insert into account_rollback (s_name, parent_name, create_user, founder, account_id, account_number, nonce,author_version,threshold,update_author_threshold,permissions,created," +
		"contract_code,code_hash, contract_created,description,suicide,destroy, height) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"))
	defer stmt.Close()

	if err != nil {
		ZapLog.Panic("InsertAccount error", zap.Error(err), zap.String("name", data.Account.Name))
	}

	_, err = stmt.Exec(data.Account.Name, data.Account.ParentName, data.Account.CreateUser, data.Account.Founder, data.Account.AccountID, data.Account.Number, data.Account.Nonce, data.Account.AuthorVersion, data.Account.Threshold, data.Account.UpdateAuthorThreshold, data.Account.Permissions, data.Account.Created,
		data.Account.ContractCode, data.Account.CodeHash, data.Account.ContractCreated, data.Account.Description, data.Account.Suicide, data.Account.Destroy, data.Height)
	if err != nil {
		ZapLog.Panic("InsertAccount error", zap.Error(err), zap.String("name", data.Account.Name))
	}
	return nil
}

func GetOldAccountByName(name string, dbTx *sql.Tx) (*MysqlAccountRollback, error) {
	sqlstr := fmt.Sprintf("insert into account_rollback (s_name, parent_name, create_user, founder, account_id, account_number, nonce,author_version,threshold,update_author_threshold,permissions,created,"+
		"contract_code,code_hash, contract_created,description,suicide,destroy, height FROM account where s_name = '%s order by height desc' ", name)
	row := dbTx.QueryRow(sqlstr)
	a := &MysqlAccount{}
	oa := &MysqlAccountRollback{
		Account: a,
	}
	err := row.Scan(&a.Name, &a.ParentName, &a.CreateUser, &a.Founder, &a.AccountID, &a.Number, &a.Nonce, &a.AuthorVersion, &a.Threshold, &a.UpdateAuthorThreshold, &a.Permissions, &a.Created,
		&a.ContractCode, &a.CodeHash, &a.ContractCreated, &a.Description, &a.Suicide, &a.Destroy, &oa.Height)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		ZapLog.Error("GetAccount error", zap.String("sql", sqlstr))
		return nil, err
	}
	return oa, nil

}

func DeleteRollbackAccountByNameAndHeight(name string, height uint64, dbTx *sql.Tx) error {
	stmt, err := dbTx.Prepare(fmt.Sprintf("delete from account_rollback where s_name = %s and height = %d", name, height))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("DeleteRollbackAccountByNameAndHeight error", zap.Error(err), zap.String("name", name))
	}

	_, err = stmt.Exec()
	if err != nil {
		ZapLog.Panic("DeleteRollbackAccountByNameAndHeight error", zap.Error(err), zap.String("name", name))
	}
	return nil
}
