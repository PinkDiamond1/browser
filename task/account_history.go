package task

import (
	"database/sql"
	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
)

type AccountHistoryTask struct {
	*Base
}

func (a *AccountHistoryTask) analysisAccountHistory(data *types.BlockAndResult, dbTx *sql.Tx) error {
	block := data.Block
	for _, tx := range block.Txs {
		for j, action := range tx.RPCActions {
			var aType int
			switch action.Type {
			case types.Transfer:
				aType = 0
			case types.CallContract:
				aType = 2
			default:
				aType = 6
			}
			if action.From.String() != "" {
				mAccountHistory := &db.MysqlAccountHistory{
					Account:     action.From.String(),
					TxHash:      tx.Hash.String(),
					ActionHash:  action.ActionHash.String(),
					ActionIndex: j,
					TxType:      aType,
					Height:      block.Head.Number.Uint64(),
				}
				err := db.InsertAccountHistory(mAccountHistory, dbTx)
				if err != nil {
					ZapLog.Error("InsertAccountHistory error: ", zap.Error(err), zap.String("account", action.From.String()))
					return err
				}
			}
			if action.To != action.From {
				if action.Type == types.CallContract {
					aType = 4
				}
				mAccountHistory := &db.MysqlAccountHistory{
					Account:     action.To.String(),
					TxHash:      tx.Hash.String(),
					ActionHash:  action.ActionHash.String(),
					ActionIndex: j,
					TxType:      aType,
					Height:      block.Head.Number.Uint64(),
				}
				err := db.InsertAccountHistory(mAccountHistory, dbTx)
				if err != nil {
					ZapLog.Error("InsertAccountHistory error: ", zap.Error(err), zap.String("account", action.To.String()))
					return err
				}
			}

		}
	}
	for i, iTxActions := range data.DetailTxs {
		tx := data.Block.Txs[i]
		for j, iActions := range iTxActions.InternalActions {
			for k, iAction := range iActions.InternalLogs {
				var aType int
				switch iAction.Action.Type {
				case types.Transfer:
					aType = 1
				case types.CallContract:
					//to不是合约
					aType = 3
				default:
					aType = 7
				}
				if iAction.Action.From.String() != "" {
					mAccountHistory := &db.MysqlAccountHistory{
						Account:     iAction.Action.From.String(),
						TxHash:      tx.Hash.String(),
						ActionHash:  iAction.Action.ActionHash.String(),
						ActionIndex: j,
						OtherIndex:  k,
						TxType:      aType,
						Height:      block.Head.Number.Uint64(),
					}
					err := db.InsertAccountHistory(mAccountHistory, dbTx)
					if err != nil {
						ZapLog.Error("InsertAccountHistory error: ", zap.Error(err), zap.String("account", iAction.Action.From.String()))
						return err
					}
				}
				if iAction.Action.To != iAction.Action.From {
					if iAction.Action.Type == types.CallContract {
						aType = 5
					}
					mAccountHistory := &db.MysqlAccountHistory{
						Account:     iAction.Action.To.String(),
						TxHash:      tx.Hash.String(),
						ActionHash:  iAction.Action.ActionHash.String(),
						ActionIndex: j,
						OtherIndex:  k,
						TxType:      aType,
						Height:      block.Head.Number.Uint64(),
					}
					err := db.InsertAccountHistory(mAccountHistory, dbTx)
					if err != nil {
						ZapLog.Error("InsertAccountHistory error: ", zap.Error(err), zap.String("account", iAction.Action.To.String()))
						return err
					}
				}
			}
		}
	}
	for i, receipt := range data.Receipts {
		tx := data.Block.Txs[i]
		for j, aResults := range receipt.ActionResults {
			for k, aR := range aResults.GasAllot {
				aType := 8
				mAccountHistory := &db.MysqlAccountHistory{
					Account:     aR.Account.String(),
					TxHash:      tx.Hash.String(),
					ActionHash:  tx.RPCActions[j].ActionHash.String(),
					ActionIndex: j,
					OtherIndex:  k,
					TxType:      aType,
					Height:      block.Head.Number.Uint64(),
				}
				err := db.InsertAccountHistory(mAccountHistory, dbTx)
				if err != nil {
					ZapLog.Error("InsertAccountHistory error: ", zap.Error(err), zap.String("account", aR.Account.String()))
					return err
				}
			}
		}
	}
	return nil
}

func (a *AccountHistoryTask) rollback(data *types.BlockAndResult, dbTx *sql.Tx) error {
	block := data.Block
	height := block.Head.Number.Uint64()
	for _, tx := range block.Txs {
		for _, action := range tx.RPCActions {
			if action.From.String() != "" {
				err := db.DeleteAccountHistoryByHeight(action.From.String(), height, dbTx)
				if err != nil {
					ZapLog.Error("DeleteAccountHistoryByHeight error: ", zap.Error(err), zap.String("from", action.From.String()), zap.Uint64("height", height))
					return err
				}
				if action.To != action.From {
					err := db.DeleteAccountHistoryByHeight(action.To.String(), height, dbTx)
					if err != nil {
						ZapLog.Error("DeleteAccountHistoryByHeight error: ", zap.Error(err), zap.String("to", action.To.String()), zap.Uint64("height", height))
						return err
					}
				}
			}
		}
	}
	for _, iTxActions := range data.DetailTxs {
		for _, iActions := range iTxActions.InternalActions {
			for _, iAction := range iActions.InternalLogs {
				err := db.DeleteAccountHistoryByHeight(iAction.Action.From.String(), height, dbTx)
				if err != nil {
					ZapLog.Error("DeleteAccountHistoryByHeight error: ", zap.Error(err), zap.String("from", iAction.Action.From.String()), zap.Uint64("height", height))
					return err
				}
				if iAction.Action.To != iAction.Action.From {
					err := db.DeleteAccountHistoryByHeight(iAction.Action.To.String(), height, dbTx)
					if err != nil {
						ZapLog.Error("DeleteAccountHistoryByHeight error: ", zap.Error(err), zap.String("to", iAction.Action.To.String()), zap.Uint64("height", height))
						return err
					}
				}
			}
		}
	}
	for _, receipt := range data.Receipts {
		for _, aResults := range receipt.ActionResults {
			for _, aR := range aResults.GasAllot {
				err := db.DeleteAccountHistoryByHeight(aR.Account.String(), height, dbTx)
				if err != nil {
					ZapLog.Error("DeleteAccountHistoryByHeight error: ", zap.Error(err), zap.String("from", aR.Account.String()), zap.Uint64("height", height))
					return err
				}
			}
		}
	}
	return nil
}

func (a *AccountHistoryTask) Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64) {
	a.startHeight = startHeight
	for {
		select {
		case d := <-data:
			if d.Block.Block.Head.Number.Uint64() >= a.startHeight {
				a.init()
				err := a.analysisAccountHistory(d.Block, a.Tx)
				if err != nil {
					ZapLog.Error("ActionTask analysisAction error: ", zap.Error(err), zap.Uint64("height", d.Block.Block.Head.Number.Uint64()))
					panic(err)
				}
				a.startHeight++
				a.commit()
			}
			result <- true
		case rd := <-rollbackData:
			if a.startHeight == rd.Block.Block.Head.Number.Uint64() {
				a.init()
				err := a.rollback(rd.Block, a.Tx)
				if err != nil {
					ZapLog.Error("ActionTask rollback error: ", zap.Error(err), zap.Uint64("height", rd.Block.Block.Head.Number.Uint64()))
					panic(err)
				}
				a.startHeight--
				a.commit()
			}
			result <- true
		}
	}
}
