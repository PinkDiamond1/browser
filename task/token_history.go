package task

import (
	"database/sql"
	"github.com/browser/client"
	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
)

type TokeHistoryTask struct {
	*Base
}

func (t *TokeHistoryTask) analysisTokenHistory(data *types.BlockAndResult, dbTx *sql.Tx) error {
	block := data.Block
	detailTxs := data.DetailTxs
	receipts := data.Receipts
	for i, tx := range block.Txs {
		receipt := receipts[i]
		for j, aT := range tx.RPCActions {
			aTR := receipt.ActionResults[j]
			if aTR.Status == types.ReceiptStatusFailed {
				continue
			}
			if aT.Type == types.IssueAsset || aT.Type == types.IncreaseAsset || aT.Type == types.DestroyAsset {
				assetId := aT.AssetID
				payload, err := parsePayload(aT)
				if err != nil {
					ZapLog.Error("parsePayload error: ", zap.Error(err))
					return err
				}
				if aT.Type == types.IssueAsset {
					arg := payload.(types.IssueAssetObject)
					tokenInfo, err := client.GetAssetInfoByName(arg.AssetName)
					if err != nil {
						ZapLog.Error("GetAssetInfoByName error: ", zap.Error(err))
						return err
					}
					assetId = tokenInfo.AssetId
				}
				if aT.Type == types.IncreaseAsset {
					arg := payload.(types.IncAssetObject)
					assetId = arg.AssetId
				}
				mTH := &db.MysqlTokenHistory{
					TokenId:     assetId,
					TxHash:      tx.Hash.String(),
					ActionIndex: j,
					ActionHash:  aT.ActionHash.String(),
					TxType:      0,
					ActionType:  uint64(aT.Type),
				}
				err = db.InsertTokenHistory(mTH, dbTx)
				if err != nil {
					ZapLog.Error("InsertTokenHistory", zap.Error(err))
					return err
				}
			}
			if len(detailTxs) != 0 {
				iAts := detailTxs[i].InternalActions[j]
				for k, iAt := range iAts.InternalLogs {
					if iAt.Action.Type == types.IssueAsset || iAt.Action.Type == types.IncreaseAsset || iAt.Action.Type == types.DestroyAsset {
						assetId := iAt.Action.AssetID
						payload, err := parsePayload(iAt.Action)
						if err != nil {
							ZapLog.Error("parsePayload error: ", zap.Error(err))
							return err
						}
						if iAt.Action.Type == types.IssueAsset {
							arg := payload.(types.IssueAssetObject)
							tokenInfo, err := client.GetAssetInfoByName(arg.AssetName)
							if err != nil {
								ZapLog.Error("GetAssetInfoByName error: ", zap.Error(err))
								return err
							}
							assetId = tokenInfo.AssetId
						}
						if iAt.Action.Type == types.IncreaseAsset {
							arg := payload.(types.IncAssetObject)
							assetId = arg.AssetId
						}
						mTH := &db.MysqlTokenHistory{
							TokenId:       assetId,
							TxHash:        tx.Hash.String(),
							ActionIndex:   j,
							ActionHash:    aT.ActionHash.String(),
							InternalIndex: k,
							TxType:        1,
							ActionType:    uint64(iAt.Action.Type),
						}
						err = db.InsertTokenHistory(mTH, dbTx)
						if err != nil {
							ZapLog.Error("InsertTokenHistory", zap.Error(err))
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func (t *TokeHistoryTask) rollback(data *types.BlockAndResult, dbTx *sql.Tx) error {
	return db.DeleteTokenHistoryByHeight(data.Block.Head.Number.Uint64(), dbTx)
}

func (t *TokeHistoryTask) Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64) {
	t.startHeight = startHeight
	for {
		select {
		case d := <-data:
			if d.Block.Block.Head.Number.Uint64() >= t.startHeight {
				t.init()
				err := t.analysisTokenHistory(d.Block, t.Tx)
				if err != nil {
					ZapLog.Error("ActionTask analysisAction error: ", zap.Error(err), zap.Uint64("height", d.Block.Block.Head.Number.Uint64()))
					panic(err)
				}
				t.startHeight++
				t.commit()
			}
			result <- true
		case rd := <-rollbackData:
			t.startHeight--
			if t.startHeight == rd.Block.Block.Head.Number.Uint64() {
				t.init()
				err := t.rollback(rd.Block, t.Tx)
				if err != nil {
					ZapLog.Error("ActionTask rollback error: ", zap.Error(err), zap.Uint64("height", rd.Block.Block.Head.Number.Uint64()))
					panic(err)
				}
				t.commit()
			}
			result <- true
		}
	}
}
