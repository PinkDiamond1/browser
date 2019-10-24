package task

import (
	"database/sql"

	"github.com/browser/client"
	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
)

type TokenFeeHistoryTask struct {
	*Base
}

func (f *TokenFeeHistoryTask) analysis(data *types.BlockAndResult, dbTx *sql.Tx) error {
	receipts := data.Receipts
	txs := data.Block.Txs
	for i, receipt := range receipts {
		tx := txs[i]
		for j, aRs := range receipt.ActionResults {
			at := tx.RPCActions[j]
			for k, aR := range aRs.GasAllot {
				if aR.Reason == 0 {
					tokenInfo, err := client.GetAssetInfoByName(aR.Account.String())
					if err != nil {
						ZapLog.Error("GetShortTokenByName", zap.Error(err), zap.String("token name", aR.Account.String()))
						return err
					}
					mTFH := &db.MysqlTokenFeeHistory{
						TokenId:        tokenInfo.AssetId,
						TxHash:         tx.Hash.String(),
						ActionIndex:    j,
						ActionHash:     at.ActionHash.String(),
						FeeActionIndex: k,
						Height:         data.Block.Number.Uint64(),
					}
					err = db.InsertTokenFeeHistory(mTFH, dbTx)
					if err != nil {
						ZapLog.Error("analysis TokenFeeHistoryTask err: ", zap.Error(err))
						return err
					}
				}
			}
		}
	}
	return nil
}

func (a *TokenFeeHistoryTask) rollback(data *types.BlockAndResult, dbTx *sql.Tx) error {
	receipts := data.Receipts
	for _, receipt := range receipts {
		for _, aRs := range receipt.ActionResults {
			for _, aR := range aRs.GasAllot {
				if aR.Reason == 0 {
					tokenInfo := db.QueryTokenByName(dbTx, aR.Account.String())
					err := db.DeleteTokenFeeHistoryByHeight(tokenInfo.AssetId, data.Block.Number.Uint64(), dbTx)
					if err != nil {
						ZapLog.Error("DeleteTokenFeeHistoryByHeight", zap.Error(err), zap.Uint64("height", data.Block.Number.Uint64()))
						return err
					}
				}
			}
		}
	}
	return nil
}

func (a *TokenFeeHistoryTask) Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64) {
	a.startHeight = startHeight
	for {
		select {
		case d := <-data:
			if d.Block.Block.Number.Uint64() >= a.startHeight {
				a.init()
				err := a.analysis(d.Block, a.Tx)
				if err != nil {
					ZapLog.Error("TokenFeeHistoryTask analysis error: ", zap.Error(err), zap.Uint64("height", d.Block.Block.Number.Uint64()))
					panic(err)
				}
				a.startHeight++
				a.commit()
			}
			result <- true
		case rd := <-rollbackData:
			a.startHeight--
			if rd.Block.Block.Number.Uint64() == a.startHeight {
				a.init()
				err := a.rollback(rd.Block, a.Tx)
				if err != nil {
					ZapLog.Error("TokenFeeHistoryTask rollback error: ", zap.Error(err), zap.Uint64("height", rd.Block.Block.Number.Uint64()))
					panic(err)
				}
				a.commit()
			}
			result <- true
		}
	}
}
