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

type MysqlAction struct {
	TxHash            string
	ActionHash        string
	ActionIndex       int
	Nonce             uint64
	Height            uint64
	Created           uint64
	GasAssetId        uint64
	TransferAssetId   uint64
	ActionType        uint64
	From              string
	To                string
	Amount            *big.Int
	GasLimit          uint64
	GasUsed           uint64
	State             uint64
	ErrorMsg          string
	Remark            []byte
	Payload           []byte
	PayloadSize       int
	InternalCount     int
	Payer             string
	PayerGasPrice     *big.Int
	ParentSigner      string
	PayerParentSigner string
}

func InsertAction(data *MysqlAction, dbTx *sql.Tx) error {
	tName := GetTableNameHash("actions_hash", data.TxHash)
	insertSql := fmt.Sprintf("insert into %s (tx_hash, action_hash, action_index,nonce,height,created,gas_asset_id,transfer_asset_id,"+
		"action_type,from_account,to_account,amount,gas_limit,gas_used,state,error_msg,remark,payload,payload_size,internal_action_count, payer, payer_gas_price, parent_signer, payer_parent_signer) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", tName)
	stmt, err := dbTx.Prepare(insertSql)
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("InsertAction error", zap.Error(err), zap.String("txHash", data.TxHash), zap.String("sql", insertSql))
	}
	state := strconv.FormatUint(data.State, 10)
	_, err = stmt.Exec(data.TxHash, data.ActionHash, data.ActionIndex, data.Nonce, data.Height, data.Created, data.GasAssetId, data.TransferAssetId,
		data.ActionType, data.From, data.To, data.Amount.String(), data.GasLimit, data.GasUsed, state, data.ErrorMsg, data.Remark, data.Payload, data.PayloadSize, data.InternalCount, data.Payer, data.PayerGasPrice.String(), data.ParentSigner, data.PayerParentSigner)
	if err != nil {
		ZapLog.Panic("InsertAction error", zap.Error(err), zap.String("txHash", data.TxHash))
	}
	return nil
}

func DeleteActionByTxHash(hash types.Hash, dbTx *sql.Tx) error {
	tName := GetTableNameHash("actions_hash", hash.String())
	stmt, err := dbTx.Prepare(fmt.Sprintf("delete from %s where tx_hash = '%s'", tName, hash.String()))
	defer stmt.Close()
	if err != nil {
		ZapLog.Panic("DeleteActionByTxHash error", zap.Error(err), zap.String("txHash", hash.String()))
	}

	_, err = stmt.Exec()
	if err != nil {
		ZapLog.Panic("DeleteActionByTxHash error", zap.Error(err), zap.String("txHash", hash.String()))
	}
	return nil
}
