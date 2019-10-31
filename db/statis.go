package db

import (
	"database/sql"
	"fmt"
	"math/big"

	. "github.com/browser/log"
	"go.uber.org/zap"
)

type MysqlToken struct {
	Token_name  string
	User_num    uint64
	User_rank   int
	Call_num    uint64
	Call_rank   int
	Income_rank int
	FeeTotal    *big.Int
	Holder_num  uint64
	Holder_rank int
}

type MysqlContract struct {
	Contract_name string
	User_num      uint64
	User_rank     int
	Call_num      uint64
	Call_rank     int
	Income_rank   int
	FeeTotal      *big.Int
}

type MysqlFeeRank struct {
	Name string
	Type int64
	Fee  string
}

func GetBlockHeight() (int64, error) {
	var height int64
	sqlstr := fmt.Sprintf("SELECT height FROM statis_block_info")
	row := Mysql.db.QueryRow(sqlstr)
	err := row.Scan(&height)
	if err == sql.ErrNoRows {
		return 0, sql.ErrNoRows
	}
	if err != nil {
		ZapLog.Error("GetBlockHeight error", zap.Error(err), zap.String("sql", sqlstr))
		return 0, err
	}
	return height, nil
}

func LoadTokens() ([]*MysqlToken, error) {
	sqlstr := "SELECT token_name,user_num, user_rank,call_num, call_rank, income_rank,feeTotal" +
		",holder_num FROM statis_token"

	rows, err := Mysql.db.Query(sqlstr)
	defer rows.Close()
	if err != nil {
		ZapLog.Error("LoadTokens", zap.Error(err), zap.String("sql", sqlstr))
		return nil, err
	}

	datas := make([]*MysqlToken, 0)
	for rows.Next() {
		var feestr string
		data := &MysqlToken{}
		err := rows.Scan(&data.Token_name, &data.User_num, &data.User_rank, &data.Call_num, &data.Call_rank, &data.Income_rank, &feestr, &data.Holder_num)
		data.FeeTotal, _ = big.NewInt(0).SetString(feestr, 10)
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		if err != nil {
			ZapLog.Panic("LoadTokens error", zap.Error(err), zap.String("sql", sqlstr))
		}
		datas = append(datas, data)
	}

	return datas, nil
}

func LoadContracts() ([]*MysqlContract, error) {
	sqlstr := "SELECT contract_name,user_num, user_rank,call_num, call_rank, income_rank,feeTotal FROM statis_contract"

	rows, err := Mysql.db.Query(sqlstr)
	defer rows.Close()
	if err != nil {
		ZapLog.Error("LoadContracts", zap.Error(err), zap.String("sql", sqlstr))
		return nil, err
	}

	datas := make([]*MysqlContract, 0)
	for rows.Next() {
		var feestr string
		data := &MysqlContract{}
		err := rows.Scan(&data.Contract_name, &data.User_num, &data.User_rank, &data.Call_num, &data.Call_rank, &data.Income_rank, &feestr)
		data.FeeTotal, _ = big.NewInt(0).SetString(feestr, 10)
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		if err != nil {
			ZapLog.Panic("LoadContracts error", zap.Error(err), zap.String("sql", sqlstr))
		}
		datas = append(datas, data)
	}
	return datas, nil
}

type ContrackStatistics struct {
	User        map[string]int
	User_num    uint64
	User_rank   int
	Call_num    uint64
	Call_rank   int
	Income_rank int
	FeeTotal    *big.Int
}

type TokenStatistics struct {
	User        map[string]int
	User_num    uint64
	User_rank   int
	Call_num    uint64
	Call_rank   int
	Income_rank int
	FeeTotal    *big.Int
	Holder      map[string]int
	Holder_num  uint64
	Holder_rank int
}

func InsertToken(name string, data *TokenStatistics, tx *sql.Tx) {
	blocksql := fmt.Sprintf("REPLACE INTO statis_token(token_name, user_num, user_rank,"+
		" call_num, call_rank, holder_num,holder_rank,income_rank,feeTotal)"+
		" VALUES('%s', %d, %d, %d , %d, %d , %d,%d, '%s');",
		name, data.User_num, data.User_rank, data.Call_num, data.Call_rank,
		data.Holder_num, data.Holder_rank, data.Income_rank, data.FeeTotal.String())
	_, err := tx.Exec(blocksql)
	if err != nil {
		ZapLog.Panic("insertToken failed", zap.Error(err), zap.String("sql", blocksql))
	}
}

func InsertContract(name string, data *ContrackStatistics, tx *sql.Tx) {
	blocksql := fmt.Sprintf("REPLACE INTO statis_contract(contract_name, user_num, user_rank,"+
		" call_num, call_rank, income_rank,feeTotal)"+
		" VALUES('%s', %d, %d, %d , %d, %d, '%s');",
		name, data.User_num, data.User_rank, data.Call_num, data.Call_rank,
		data.Income_rank, data.FeeTotal.String())
	_, err := tx.Exec(blocksql)
	if err != nil {
		ZapLog.Panic("InsertContract failed", zap.Error(err), zap.String("sql", blocksql))
	}
}

func ReplaceTotalFee(name, nametype string, fee *big.Int, rank int, tx *sql.Tx) {
	blocksql := fmt.Sprintf("REPLACE INTO statis_fee_total(name,nametype,rank,fee) "+
		"VALUES('%s','%s',%d, '%s');", name, nametype, rank, fee.String())
	_, err := tx.Exec(blocksql)
	if err != nil {
		ZapLog.Panic("InsertContract failed", zap.Error(err), zap.String("sql", blocksql))
	}
}

func ReplaceBlockInfo(height int64) {
	blocksql := fmt.Sprintf("REPLACE INTO statis_block_info(id,height) VALUES(%d,%d);", 1, height)
	_, err := Mysql.db.Exec(blocksql)
	if err != nil {
		ZapLog.Panic("insertBlockInfo failed", zap.Error(err), zap.String("sql", blocksql))
	}
}

func InsertTokenInfo(name string, decimals uint64, assetid uint64, shortName string) {
	blocksql := fmt.Sprintf("REPLACE INTO statis_token_info(name,decimals,assetid,shortname) VALUES('%s',%d,%d,'%s') ", name, decimals, assetid, shortName)
	_, err := Mysql.db.Exec(blocksql)
	if err != nil {
		ZapLog.Panic("InsertTokenInfo failed", zap.Error(err), zap.String("sql", blocksql))
	}
}

func GetTokenInfoByAssetID(assetid uint64) (uint64, string, string, error) {
	var decimals uint64
	var name, shortname string
	sqlstr := fmt.Sprintf("SELECT decimals,name,shortname FROM statis_token_info where assetid = %d ", assetid)
	row := Mysql.db.QueryRow(sqlstr)
	err := row.Scan(&decimals, &name, &shortname)
	if err == sql.ErrNoRows {
		return 0, "", "", sql.ErrNoRows
	}
	if err != nil {
		ZapLog.Error("getTokenInfoByAssetID error", zap.Error(err), zap.String("sql", sqlstr))
		return 0, "", "", err
	}
	return decimals, name, shortname, nil
}
