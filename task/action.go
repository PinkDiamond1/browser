package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/browser/types"
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
			getSigner := func(account, substr string, index uint64) string {
				var singer string
				if index <= 0 {
					return singer
				}
				cnt := strings.IndexAny(account, substr)
				if index > uint64(cnt) {
					return singer
				}
				for i := uint64(0); i < index; i++ {
					lastIndex := strings.LastIndex(account, ".")
					account = string([]byte(account)[:lastIndex])
				}
				singer = account
				return singer
			}
			var parentSinger, payerParentSigner string
			if action.ParentIndex > 0 {
				parentSinger = getSigner(action.From.String(), ".", action.ParentIndex)
			}
			if action.Payer != "" && action.PayerParentIndex > 0 {
				payerParentSigner = getSigner(action.Payer.String(), ".", action.PayerParentIndex)
			}
			mAction := &db.MysqlAction{
				TxHash:            tx.Hash.String(),
				ActionHash:        action.ActionHash.String(),
				ActionIndex:       j,
				Nonce:             action.Nonce,
				Height:            block.Number.Uint64(),
				Created:           block.Time,
				GasAssetId:        tx.GasAssetID,
				TransferAssetId:   action.AssetID,
				ActionType:        uint64(action.Type),
				From:              action.From.String(),
				To:                action.To.String(),
				Amount:            action.Amount,
				GasLimit:          action.GasLimit,
				GasUsed:           actionResult.GasUsed,
				State:             actionResult.Status,
				ErrorMsg:          actionResult.Error,
				Remark:            []byte(fmt.Sprintf("%s", []byte(action.Remark))),
				InternalCount:     internalCount,
				Payer:             action.Payer.String(),
				PayerGasPrice:     action.PayerGasPrice,
				ParentSigner:      parentSinger,
				PayerParentSigner: payerParentSigner,
			}
			parsedPayload, err := parsePayload(action)
			if err != nil {
				ZapLog.Warn("parsePayload error: ", zap.Error(err), zap.Binary("payload", action.Payload))
			}
			if actionResult.Status == types.ReceiptStatusSuccessful {
				if action.Type == types.IssueAsset {
					issueAssetPayload := parsedPayload.(types.IssueAssetObject)
					if idx := strings.Index(issueAssetPayload.AssetName, ":"); idx <= 0 {
						if len(action.From.String()) > 0 {
							issueAssetPayload.AssetName = action.From.String() + ":" + issueAssetPayload.AssetName
						}
					}
					parsedPayload = issueAssetPayload
				}
			}
			if parsedPayload == nil {
				parsedPayload = action.Payload
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
			ZapLog.Error("DeleteActionByTxHash error: ", zap.Error(err), zap.Uint64("height", data.Block.Number.Uint64()), zap.String("txHash", tx.Hash.String()))
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
			if d.Block.Block.Number.Uint64() >= a.startHeight {
				a.init()
				err := a.analysisAction(d.Block, a.Tx)
				if err != nil {
					ZapLog.Error("ActionTask analysisAction error: ", zap.Error(err), zap.Uint64("height", d.Block.Block.Number.Uint64()))
					panic(err)
				}
				a.startHeight++
				a.commit()
			}
			result <- true
		case rd := <-rollbackData:
			a.startHeight--
			if a.startHeight == rd.Block.Block.Number.Uint64() {
				a.init()
				err := a.rollback(rd.Block, a.Tx)
				if err != nil {
					ZapLog.Error("ActionTask rollback error: ", zap.Error(err), zap.Uint64("height", rd.Block.Block.Number.Uint64()))
					panic(err)
				}
				a.commit()
			}
			result <- true
		}
	}
}
