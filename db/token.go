package db

import (
	"database/sql"
	"fmt"
	"math/big"

	. "github.com/browser_service/log"
	"go.uber.org/zap"
)

type Token struct {
	Id                    uint64
	AssetName             string
	AssetSymbol           string
	Decimals              uint64
	AssetId               uint64
	ContractName          string
	Description           string
	CreateUser            string
	CreateTime            uint
	AssetOwner            string
	Founder               string
	UpperLimit            *big.Int
	Liquidity             *big.Int
	CumulativeIssue       *big.Int
	CumulativeDestruction *big.Int
	UpdateTime            uint
}

func AddToken(tx *sql.Tx, token *Token) {
	stmt, err := tx.Prepare("insert into token (asset_name, asset_symbol, decimals, asset_id, contract_name, description, create_user, create_time, asset_owner, founder, upper_limit, liquidity, cumulative_issue, cumulative_destruction, update_time) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare insert token error", zap.Error(err), zap.String("token", fmt.Sprint(token)))
	}

	_, err = stmt.Exec(token.AssetName, token.AssetSymbol, token.Decimals, token.AssetId, token.ContractName, token.Description, token.CreateUser, token.CreateTime, token.AssetOwner, token.Founder, token.UpperLimit.String(), token.Liquidity.String(), token.CumulativeIssue.String(), token.CumulativeDestruction.String(), token.UpdateTime)
	if err != nil {
		ZapLog.Panic("insert token error", zap.Error(err), zap.String("tokenName", token.AssetName), zap.String("token", fmt.Sprint(token)))
	}
}

func ReplaceToken(tx *sql.Tx, token *Token) {
	stmt, err := tx.Prepare("replace into token (asset_name, asset_symbol, decimals, asset_id, contract_name, description, create_user, create_time, asset_owner, founder, upper_limit, liquidity, cumulative_issue, cumulative_destruction, update_time) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare replace token error", zap.Error(err), zap.String("token", fmt.Sprint(token)))
	}

	_, err = stmt.Exec(token.AssetName, token.AssetSymbol, token.Decimals, token.AssetId, token.ContractName, token.Description, token.CreateUser, token.CreateTime, token.AssetOwner, token.Founder, token.UpperLimit, token.Liquidity, token.CumulativeIssue, token.CumulativeDestruction, token.UpdateTime)
	if err != nil {
		ZapLog.Panic("replace token error", zap.Error(err), zap.String("token", fmt.Sprint(token)))
	}
}

func QueryTokenByName(tx *sql.Tx, assetName string) *Token {
	stmt, err := tx.Prepare("select id, asset_name, asset_symbol, decimals, asset_id, contract_name, description, create_user, create_time, asset_owner, founder, upper_limit, liquidity, cumulative_issue, cumulative_destruction, update_time from token where asset_name = ?")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare select token by asset name error", zap.Error(err), zap.String("assetName", assetName))
	}

	row := stmt.QueryRow(assetName)
	token := &Token{}
	err = row.Scan(&token.Id, &token.AssetName, &token.AssetSymbol, &token.Decimals, &token.AssetId, &token.ContractName, &token.Description, &token.CreateUser, &token.CreateTime, &token.AssetOwner, &token.Founder, &token.UpperLimit, &token.Liquidity, &token.CumulativeIssue, &token.CumulativeDestruction, &token.UpdateTime)
	if err != nil {
		ZapLog.Panic("select token by asset name error", zap.Error(err), zap.String("assetName", assetName))
	}
	return token
}

func QueryTokenById(tx *sql.Tx, assetId uint64) *Token {
	stmt, err := tx.Prepare("select id, asset_name, asset_symbol, decimals, asset_id, contract_name, description, create_user, create_time, asset_owner, founder, upper_limit, liquidity, cumulative_issue, cumulative_destruction, update_time from token where asset_id = ?")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare select token by asset name error", zap.Error(err), zap.Uint64("assetId", assetId))
	}

	row := stmt.QueryRow(assetId)
	token := &Token{}
	var UpperLimit, Liquidity, CumulativeIssue, CumulativeDestruction string
	err = row.Scan(&token.Id, &token.AssetName, &token.AssetSymbol, &token.Decimals, &token.AssetId, &token.ContractName, &token.Description, &token.CreateUser, &token.CreateTime, &token.AssetOwner, &token.Founder, &UpperLimit, &Liquidity, &CumulativeIssue, &CumulativeDestruction, &token.UpdateTime)
	token.UpperLimit, _ = big.NewInt(0).SetString(UpperLimit, 10)
	token.Liquidity, _ = big.NewInt(0).SetString(Liquidity, 10)
	token.CumulativeIssue, _ = big.NewInt(0).SetString(CumulativeIssue, 10)
	token.CumulativeDestruction, _ = big.NewInt(0).SetString(CumulativeDestruction, 10)
	if err != nil {
		ZapLog.Panic("select token by asset id error", zap.Error(err), zap.Uint64("assetId", assetId))
	}
	return token
}

func UpdateTokenById(tx *sql.Tx, token *Token) {
	stmt, err := tx.Prepare("update token set liquidity = ?, cumulative_issue = ?, cumulative_destruction = ?, update_time = ?, asset_owner = ?, founder = ? where id = ?")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare update token error", zap.Error(err), zap.String("token", fmt.Sprint(token)))
	}

	_, err = stmt.Exec(token.Liquidity.String(), token.CumulativeIssue.String(), token.CumulativeDestruction.String(), token.UpdateTime, token.AssetOwner, token.Founder, token.Id)
	if err != nil {
		ZapLog.Panic("update token error", zap.Error(err), zap.String("token", fmt.Sprint(token)))
	}
}

func DeleteTokenByName(tx *sql.Tx, assetName string) {
	stmt, err := tx.Prepare("delete from token where asset_name = ?")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare update token error", zap.Error(err), zap.String("asset_name", assetName))
	}

	_, err = stmt.Exec(assetName)
	if err != nil {
		ZapLog.Panic("prepare update token error", zap.Error(err), zap.String("asset_name", assetName))
	}
}

func AddBackupToken(tx *sql.Tx, token *Token, height uint64) {
	stmt, err := tx.Prepare("insert into token_backup (height, asset_name, asset_symbol, decimals, asset_id, contract_name, description, create_user, create_time, asset_owner, founder, upper_limit, liquidity, cumulative_issue, cumulative_destruction, update_time) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare insert token_backup error", zap.Error(err), zap.String("token", fmt.Sprint(token)))
	}

	_, err = stmt.Exec(height, token.AssetName, token.AssetSymbol, token.Decimals, token.AssetId, token.ContractName, token.Description, token.CreateUser, token.CreateTime, token.AssetOwner, token.Founder, token.UpperLimit.String(), token.Liquidity.String(), token.CumulativeIssue.String(), token.CumulativeDestruction.String(), token.UpdateTime)
	if err != nil {
		ZapLog.Panic("insert token_backup error", zap.Error(err), zap.String("token", fmt.Sprint(token)))
	}
}

func DeleteTokenBackupByHeightName(tx *sql.Tx, assetName string, height uint64) {
	stmt, err := tx.Prepare("delete from token_backup where height = ? and asset_name = ? ")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare delete token_backup error", zap.Error(err), zap.String("asset_name", assetName))
	}

	_, err = stmt.Exec(height, assetName)
	if err != nil {
		ZapLog.Panic("prepare delete token_backup error", zap.Error(err), zap.String("asset_name", assetName))
	}
}

func QueryTokenBackupById(tx *sql.Tx, assetId uint64, height uint64) *Token {
	stmt, err := tx.Prepare("select id, asset_name, asset_symbol, decimals, asset_id, contract_name, description, create_user, create_time, asset_owner, founder, upper_limit, liquidity, cumulative_issue, cumulative_destruction, update_time from token_backup where asset_id = ? and height = ? ")
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("prepare select token_backup by asset name error", zap.Error(err), zap.Uint64("assetId", assetId))
	}

	row := stmt.QueryRow(assetId, height)
	token := &Token{}
	err = row.Scan(&token.Id, &token.AssetName, &token.AssetSymbol, &token.Decimals, &token.AssetId, &token.ContractName, &token.Description, &token.CreateUser, &token.CreateTime, &token.AssetOwner, &token.Founder, &token.UpperLimit, &token.Liquidity, &token.CumulativeIssue, &token.CumulativeDestruction, &token.UpdateTime)
	if err != nil {
		ZapLog.Panic("select token_backup by asset id error", zap.Error(err), zap.Uint64("assetName", assetId))
	}
	return token
}
