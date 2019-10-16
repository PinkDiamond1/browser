package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/browser_service/db"
	. "github.com/browser_service/log"
	"github.com/browser_service/types"
	"go.uber.org/zap"
	"strings"
)

type ActionTask struct {
	*Base
}

func (a *ActionTask) analysisAction(data *types.BlockAndResult, dbTx *sql.Tx) error {
	block := data.Block
	receipts := data.Receipts
	internalTxs := data.DetailTxs
	for i, tx := range block.Txs {
		receipt := receipts[i]
		for j, action := range tx.RPCActions {
			actionResult := receipt.ActionResults[j]
			var internalCount int
			if len(internalTxs) != 0 {
				if len(internalTxs[i].InternalActions) != 0 {
					internalCount = len(internalTxs[i].InternalActions[j].InternalLogs)
				}
			}
			mAction := &db.MysqlAction{
				TxHash:          tx.Hash.String(),
				ActionHash:      action.ActionHash.String(),
				ActionIndex:     j,
				Nonce:           action.Nonce,
				Height:          block.Head.Number.Uint64(),
				Created:         block.Head.Time,
				GasAssetId:      tx.GasAssetID,
				TransferAssetId: action.AssetID,
				ActionType:      uint64(action.Type),
				From:            action.From.String(),
				To:              action.To.String(),
				Amount:          action.Amount,
				GasLimit:        action.GasLimit,
				GasUsed:         actionResult.GasUsed,
				State:           actionResult.Status,
				ErrorMsg:        actionResult.Error,
				Remark:          []byte(fmt.Sprintf("%s", []byte(action.Remark))),
				InternalCount:   internalCount,
			}
			parsedPayload, err := parsePayload(action)
			if err != nil {
				ZapLog.Warn("parsePayload error: ", zap.Error(err), zap.Binary("payload", action.Payload))
			}
			if actionResult.Status == types.ReceiptStatusSuccessful {
				if action.Type == types.IssueAsset {
					issueAssetPayload := parsedPayload.(types.IssueAssetObject)
					if strings.Compare(issueAssetPayload.AssetName, "libra") == 0 || strings.Compare(issueAssetPayload.AssetName, "bitcoin") == 0 {
						issueAssetPayload.AssetName = action.From.String() + ":" + issueAssetPayload.AssetName
					}
					parsedPayload = issueAssetPayload
				}
			}
			jsonPayload, err := json.Marshal(parsedPayload)
			if err != nil {
				ZapLog.Error("json marshal error: ", zap.Error(err), zap.Binary("payload", parsedPayload.([]byte)))
				return err
			}
			mAction.Payload = jsonPayload
			err = db.InsertAction(mAction, dbTx)
			if err != nil {
				ZapLog.Error("InsertAction error: ", zap.Error(err))
				return err
			}
		}
	}

	return nil
}

func (a *ActionTask) rollback(data *types.BlockAndResult, dbTx *sql.Tx) error {
	for _, tx := range data.Block.Txs {
		err := db.DeleteActionByTxHash(tx.Hash, dbTx)
		if err != nil {
			ZapLog.Error("DeleteActionByTxHash error: ", zap.Error(err), zap.Uint64("height", data.Block.Head.Number.Uint64()), zap.String("txHash", tx.Hash.String()))
			return err
		}
	}
	return nil
}

func (a *ActionTask) Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64) {
	a.startHeight = startHeight
	for {
		select {
		case d := <-data:
			if d.Block.Block.Head.Number.Uint64() >= a.startHeight {
				a.init()
				err := a.analysisAction(d.Block, a.Tx)
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
				err := a.rollback(rd.Block, rd.Tx)
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
