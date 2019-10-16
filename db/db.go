package db

import (
	"database/sql"
	"fmt"
	"github.com/browser/config"
	. "github.com/browser/log"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"math/big"
	"net/url"
	"strings"
	"time"
)

var Mysql *mysql

type mysql struct {
	db *sql.DB
}

func (mysql *mysql) Close() {
	_ = mysql.db.Close()
}

func (mysql *mysql) Begin() *sql.Tx {
	tx, err := Mysql.db.Begin()
	if err != nil {
		ZapLog.Panic("mysql start transaction error", zap.Error(err))
	}
	return tx
}

func InitDb() {
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=%s&parseTime=true",
		config.Mysql.Username, config.Mysql.Password, config.Mysql.Ip, config.Mysql.Port, config.Mysql.Database, url.QueryEscape("Asia/Shanghai"))

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		ZapLog.Error("connected mysql server error", zap.Error(err), zap.String("str", connStr))
		panic(err)
	}

	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	db.SetConnMaxLifetime(60 * time.Second)
	err = initChainStatus(db)
	if err != nil {
		ZapLog.Error("initChainStatus error", zap.Error(err))
		panic(err)
	}
	Mysql = &mysql{db: db}
}

type TaskStatus struct {
	Id       uint64
	TaskType string
	Height   uint64
}

func (mysql *mysql) InitTaskStatus(taskType string) {
	stmt, err := mysql.db.Prepare("insert into task_status (task_type, height) values (?, 0)")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("InitTaskStatus error", zap.Error(err), zap.String("task_type", taskType))
	}

	_, err = stmt.Exec(taskType)
	if err != nil {
		ZapLog.Panic("InitTaskStatus error", zap.Error(err), zap.String("task_type", taskType))
	}
}

func (mysql *mysql) GetTaskStatus(taskTypes []string) map[string]*TaskStatus {
	var result = make(map[string]*TaskStatus)
	inCondition := strings.Join(taskTypes, `","`)
	inCondition = fmt.Sprintf(`"%s"`, inCondition)
	sql := fmt.Sprintf(`select id, task_type, height from task_status where task_type in (%s)`, inCondition)
	rows, err := mysql.db.Query(sql)
	defer rows.Close()
	if err != nil {
		ZapLog.Panic("GetTaskStatus error", zap.Error(err), zap.String("task_types", fmt.Sprint(taskTypes)))
	}

	for rows.Next() {
		row := &TaskStatus{}
		err = rows.Scan(&row.Id, &row.TaskType, &row.Height)
		if err != nil {
			ZapLog.Panic("GetTaskStatus error", zap.Error(err), zap.String("task_types", fmt.Sprint(taskTypes)))
		}
		result[row.TaskType] = row
	}
	return result
}

func (m *mysql) GetChainStatus() (*MysqlChainStatus, error) {
	sqlStr := "select height,tx_count,producer_number,fee_income,token_income,contract_income from chain_status"
	row := m.db.QueryRow(sqlStr)
	a := &MysqlChainStatus{}
	var feeIncome, tokenIncome, contractIncome string
	err := row.Scan(&a.Height, &a.TxCount, &a.ProducerNumber, &feeIncome, &tokenIncome, &contractIncome)
	if err != nil {
		ZapLog.Error("GetChainStatus error", zap.String("sql", sqlStr))
		return nil, err
	}
	a.FeeIncome, _ = big.NewInt(0).SetString(feeIncome, 10)
	a.TokenIncome, _ = big.NewInt(0).SetString(tokenIncome, 10)
	a.ContractIncome, _ = big.NewInt(0).SetString(contractIncome, 10)
	return a, nil
}

type BlockOriginal struct {
	Id         uint64
	BlockData  []byte
	Height     uint64
	BlockHash  string
	ParentHash string
}

func (mysql *mysql) GetBlockOriginalByHeight(height uint64) *BlockOriginal {
	stmt, err := mysql.db.Prepare("select id, block_data, height, block_hash, parent_hash from block_original where height = ?")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare query block_original error", zap.Error(err))
	}

	row := stmt.QueryRow(height)
	result := &BlockOriginal{}
	err = row.Scan(&result.Id, &result.BlockData, &result.Height, &result.BlockHash, &result.ParentHash)
	if err != nil {
		ZapLog.Panic("query block_original error", zap.Error(err))
	}
	return result
}

func AddReversibleBlockCache(tx *sql.Tx, block *BlockOriginal) {
	stmt, err := tx.Prepare("insert into block_original (block_data, height, block_hash, parent_hash) values (?, ?, ?, ?)")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare insert block_original error")
	}

	_, err = stmt.Exec(block.BlockData, block.Height, block.BlockHash, block.ParentHash)
	if err != nil {
		ZapLog.Panic("insert block_original error", zap.Error(err))
	}

}

func DeleteIrreversibleCache(tx *sql.Tx, height uint64) {
	stmt, err := tx.Prepare("delete from block_original where height <= ?")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare delete irreversible cache error", zap.Error(err))
	}

	_, err = stmt.Exec(height)
	if err != nil {
		ZapLog.Panic("delete irreversible cache error", zap.Error(err))
	}

}

func UpdateTaskStatus(tx *sql.Tx, taskType string, height uint64) {
	stmt, err := tx.Prepare("update task_status set height = ? where task_type = ?")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare update task_status error", zap.Error(err))
	}

	_, err = stmt.Exec(height, taskType)
	if err != nil {
		ZapLog.Panic("update task_status error", zap.Error(err))
	}
}
