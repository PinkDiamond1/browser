package task

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
)

type InternalTask struct {
	*Base
}

const maxuint = 2147483647

func (i *InternalTask) analysisInternalAction(data *types.BlockAndResult, dbTx *sql.Tx) error {
	for i, itx := range data.DetailTxs {
		tx := data.Block.Txs[i]
		receipt := data.Receipts[i]
		for j, ias := range itx.InternalActions {
			a := tx.RPCActions[j]
			ar := receipt.ActionResults[j]
			for k, ia := range ias.InternalLogs {
				if ia.Action.Type == types.CallContract {
					if bytes.Equal(ia.Action.Payload, []byte{}) {
						ia.Action.Type = types.Transfer
					} else {
						acct, err := db.GetAccountByName(ia.Action.To.String(), dbTx)
						if err != nil {
							ZapLog.Error("GetAccountByName error", zap.String("name", ia.Action.To.String()), zap.Error(err))
							return err
						}
						if acct.ContractCreated <= 0 {
							ia.Action.Type = types.Transfer
						}
					}
				}

				if uint64(maxuint) < ia.Action.AssetID {
					ia.Action.AssetID = uint64(maxuint)
				}
				mInternal := &db.MysqlInternal{
					TxHash:        tx.Hash.String(),
					ActionHash:    a.ActionHash.String(),
					ActionIndex:   j,
					InternalIndex: k,
					Height:        data.Block.Number.Uint64(),
					Created:       data.Block.Time,
					AssetId:       ia.Action.AssetID,
					ActionType:    uint64(ia.Action.Type),
					From:          ia.Action.From.String(),
					To:            ia.Action.To.String(),
					Amount:        ia.Action.Amount,
					GasLimit:      ia.Action.GasLimit,
					GasUsed:       ia.GasUsed,
					Depth:         ia.Depth,
					State:         ar.Status,
					ErrorMsg:      ia.Error,
				}
				parsedPayload, err := parsePayload(ia.Action)
				if err != nil {
					ZapLog.Warn("parsePayload error: ", zap.Error(err), zap.Uint64("actionType", uint64(ia.Action.Type)), zap.Binary("payload", ia.Action.Payload))
				}
				if ar.Status == types.ReceiptStatusSuccessful {
					if ia.Action.Type == types.IssueAsset {
						issueAssetPayload := parsedPayload.(types.IssueAssetObject)
						if idx := strings.Index(issueAssetPayload.AssetName, ":"); idx <= 0 {
							if len(ia.Action.From.String()) > 0 {
								issueAssetPayload.AssetName = ia.Action.From.String() + ":" + issueAssetPayload.AssetName
							}
						}
						parsedPayload = issueAssetPayload
					}
				}
				jsonPayload, err := json.Marshal(parsedPayload)
				if err != nil {
					ZapLog.Error("Marshal error: ", zap.Error(err), zap.Uint64("actionType", uint64(ia.Action.Type)))
					return err
				}
				mInternal.Payload = jsonPayload
				err = db.InsertInternalAction(mInternal, dbTx)
				if err != nil {
					ZapLog.Error("InsertInternalAction error: ", zap.Error(err))
					return err
				}
			}
		}

	}
	return nil
}

func (i *InternalTask) rollback(data *types.BlockAndResult, dbTx *sql.Tx) error {
	for _, tx := range data.Block.Txs {
		for _, at := range tx.RPCActions {
			err := db.DeleteInternalByActionHash(at.ActionHash, dbTx)
			if err != nil {
				ZapLog.Error("DeleteInternalByActionHash error:", zap.Error(err), zap.Uint64("height", data.Block.Number.Uint64()))
				return err
			}
		}

	}
	return nil
}

func (i *InternalTask) Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64) {
	i.startHeight = startHeight
	for {
		select {
		case d := <-data:
			if d.Block.Block.Number.Uint64() >= i.startHeight {
				i.init()
				err := i.analysisInternalAction(d.Block, i.Tx)
				if err != nil {
					ZapLog.Error("InternalTask analysisInternalAction error: ", zap.Error(err), zap.Uint64("height", d.Block.Block.Number.Uint64()))
					panic(err)
				}
				i.startHeight++
				i.commit()
			}
			result <- true
		case rd := <-rollbackData:
			i.startHeight--
			if rd.Block.Block.Number.Uint64() >= i.startHeight {
				i.init()
				err := i.rollback(rd.Block, i.Tx)
				if err != nil {
					ZapLog.Error("InternalTask rollback error: ", zap.Error(err), zap.Uint64("height", rd.Block.Block.Number.Uint64()))
					panic(err)
				}
				i.commit()
			}
			result <- true
		}
	}
}
