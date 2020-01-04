package task

import (
	"database/sql"
	"github.com/browser/client"
	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
	"math/big"
	"strings"
)

type TokenTask struct {
	*Base
	//startHeight uint64
}

func (t *TokenTask) Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64) {
	t.startHeight = startHeight
	for {
		select {
		case block := <-data:
			if block.Block.Block.Number.Uint64() >= t.startHeight {
				t.init()
				block.Tx = t.Tx
				t.analysisToken(block)
				t.startHeight++
				t.commit()
			}
			result <- true
		case block := <-rollbackData:
			t.startHeight--
			if t.startHeight == block.Block.Block.Number.Uint64() {
				t.init()
				block.Tx = t.Tx
				t.rollback(block)
				t.commit()
			}
			result <- true
		}
	}
}

func (t *TokenTask) isTokenTxs(actionType types.ActionType) bool {
	result := false
	switch actionType {
	case types.IssueAsset:
		result = true
		break
	case types.IncreaseAsset:
		result = true
		break
	case types.DestroyAsset:
		result = true
		break
	case types.SetAssetOwner:
		result = true
		break
	case types.UpdateAsset:
		result = true
		break
	case types.UpdateAssetContract:
		result = true
		break
	default:
	}
	return result
}

func (t *TokenTask) analysisToken(block *TaskChanData) {
	txs := block.Block.Block.Txs
	receipts := block.Block.Receipts
	backupMap := make(map[uint64]struct{}, 0)
	for i, tx := range txs {
		for j, action := range tx.RPCActions {
			if receipts[i].ActionResults[j].Status == types.ReceiptStatusSuccessful {
				if t.isTokenTxs(action.Type) {
					isbackup := 0
					if types.IssueAsset != action.Type {
						_, ok := backupMap[action.AssetID]
						if !ok {
							backupMap[action.AssetID] = struct{}{}
							isbackup = 1
						}
					}
					saveToken(block.Tx, action, block.Block.Block.Time, block.Block.Block.Number.Uint64(), isbackup)
				}
			}
		}
	}

	//internal transaction
	internalTxs := block.Block.DetailTxs
	for _, tx := range internalTxs {
		for _, action := range tx.InternalActions {
			//if receipts[i].ActionResults[j].Status == types.ReceiptStatusSuccessful {
			for _, internalLog := range action.InternalLogs {
				if internalLog.Error == "" {
					if t.isTokenTxs(internalLog.Action.Type) {
						isbackup := 0
						if types.IssueAsset != internalLog.Action.Type {
							_, ok := backupMap[internalLog.Action.AssetID]
							if !ok {
								backupMap[internalLog.Action.AssetID] = struct{}{}
								isbackup = 1
							}
						}
						saveToken(block.Tx, internalLog.Action, block.Block.Block.Time, block.Block.Block.Number.Uint64(), isbackup)
					}
				}
			}
		}
	}

	//}
}

func (t *TokenTask) rollback(block *TaskChanData) {
	txs := block.Block.Block.Txs
	receipts := block.Block.Receipts
	for i, tx := range txs {
		for j, action := range tx.RPCActions {
			if receipts[i].ActionResults[j].Status == types.ReceiptStatusSuccessful {
				if t.isTokenTxs(action.Type) {
					rollbackToken(block.Tx, action, block.Block.Block.Time, block.Block.Block.Number.Uint64())
				}
			}
		}
	}

	//internal transaction
	internalTxs := block.Block.DetailTxs
	for i, tx := range internalTxs {
		for j, action := range tx.InternalActions {
			if receipts[i].ActionResults[j].Status == types.ReceiptStatusSuccessful {
				for _, internalLog := range action.InternalLogs {
					if internalLog.Error == "" {
						if t.isTokenTxs(internalLog.Action.Type) {
							rollbackToken(block.Tx, internalLog.Action, block.Block.Block.Time, block.Block.Block.Number.Uint64())
						}
					}
				}
			}
		}

	}
}

func saveToken(tx *sql.Tx, action *types.RPCAction, blockTime uint64, height uint64, backup int) {
	iActionAsset, _ := parsePayload(action)
	if action.Type == types.IssueAsset {
		obj := iActionAsset.(types.IssueAssetObject)
		tokenName := obj.AssetName
		if idx := strings.Index(obj.AssetName, ":"); idx <= 0 {
			if len(action.From.String()) > 0 {
				tokenName = action.From.String() + ":" + obj.AssetName
			}
		}
		tokenInfo, err := client.GetAssetInfoByName(obj.AssetName)
		if err != nil || tokenInfo == nil {
			ZapLog.Panic("get asset info by name error", zap.Error(err), zap.String("assetName", obj.AssetName))
		}
		dbToken := &db.Token{}
		dbToken.AssetName = tokenName
		dbToken.AssetSymbol = obj.Symbol
		dbToken.Decimals = tokenInfo.Decimals
		dbToken.AssetId = tokenInfo.AssetId
		dbToken.ContractName = tokenInfo.Contract.String()
		dbToken.Description = obj.Description
		dbToken.CreateUser = action.From.String()
		dbToken.CreateTime = blockTime
		dbToken.AssetOwner = obj.Owner.String()
		dbToken.Founder = obj.Founder.String()
		dbToken.UpperLimit = obj.UpperLimit
		dbToken.Liquidity = obj.Amount
		dbToken.CumulativeIssue = obj.Amount
		dbToken.CumulativeDestruction = big.NewInt(0)
		dbToken.UpdateTime = blockTime

		if dbToken.Founder == "" {
			dbToken.Founder = dbToken.AssetOwner
		}
		db.AddToken(tx, dbToken)
	} else if action.Type == types.IncreaseAsset {
		obj := iActionAsset.(types.IncAssetObject)
		dbToken := db.QueryTokenById(tx, obj.AssetId)
		if backup == 1 {
			db.AddBackupToken(tx, dbToken, height)
		}
		dbToken.Liquidity = big.NewInt(0).Add(dbToken.Liquidity, obj.Amount)
		dbToken.CumulativeIssue = big.NewInt(0).Add(dbToken.CumulativeIssue, obj.Amount)
		dbToken.UpdateTime = blockTime
		db.UpdateTokenById(tx, dbToken)
	} else if action.Type == types.DestroyAsset {
		dbToken := db.QueryTokenById(tx, action.AssetID)
		if backup == 1 {
			db.AddBackupToken(tx, dbToken, height)
		}
		dbToken.Liquidity = big.NewInt(0).Sub(dbToken.Liquidity, action.Amount)
		dbToken.CumulativeDestruction = big.NewInt(0).Add(dbToken.CumulativeDestruction, action.Amount)
		dbToken.UpdateTime = blockTime
		db.UpdateTokenById(tx, dbToken)
	} else if action.Type == types.SetAssetOwner {
		obj := iActionAsset.(types.UpdateAssetOwnerObject)
		dbToken := db.QueryTokenById(tx, obj.AssetID)
		if backup == 1 {
			db.AddBackupToken(tx, dbToken, height)
		}
		dbToken.AssetOwner = obj.Owner.String()
		dbToken.UpdateTime = blockTime
		db.UpdateTokenById(tx, dbToken)
	} else if action.Type == types.UpdateAsset {
		obj := iActionAsset.(types.UpdateAssetObject)
		dbToken := db.QueryTokenById(tx, obj.AssetID)
		if backup == 1 {
			db.AddBackupToken(tx, dbToken, height)
		}
		dbToken.Founder = obj.Founder.String()
		dbToken.UpdateTime = blockTime
		db.UpdateTokenById(tx, dbToken)
	} else if action.Type == types.UpdateAssetContract {
		obj := iActionAsset.(types.UpdateAssetContractObject)
		dbToken := db.QueryTokenById(tx, obj.AssetID)
		if backup == 1 {
			db.AddBackupToken(tx, dbToken, height)
		}
		dbToken.ContractName = obj.Contract.String()
		dbToken.UpdateTime = blockTime
		db.UpdateTokenById(tx, dbToken)
		db.AddBackupToken(tx, dbToken, height)
	}
}

func rollbackToken(tx *sql.Tx, action *types.RPCAction, blockTime uint64, height uint64) {
	iActionAsset, _ := parsePayload(action)
	if action.Type == types.IssueAsset {
		obj := iActionAsset.(types.IssueAssetObject)
		tokenName := obj.AssetName
		if idx := strings.Index(obj.AssetName, ":"); idx <= 0 {
			if len(action.From.String()) > 0 {
				tokenName = action.From.String() + ":" + obj.AssetName
			}
		}
		db.DeleteTokenByName(tx, tokenName)
	} else if action.Type == types.IncreaseAsset {
		obj := iActionAsset.(types.IncAssetObject)
		dbToken := db.QueryTokenBackupById(tx, obj.AssetId, height)
		db.ReplaceToken(tx, dbToken)
	} else if action.Type == types.DestroyAsset {
		dbToken := db.QueryTokenBackupById(tx, action.AssetID, height)
		db.ReplaceToken(tx, dbToken)
	} else if action.Type == types.SetAssetOwner {
		obj := iActionAsset.(types.UpdateAssetOwnerObject)
		dbToken := db.QueryTokenBackupById(tx, obj.AssetID, height)
		db.ReplaceToken(tx, dbToken)
	} else if action.Type == types.UpdateAsset {
		obj := iActionAsset.(types.UpdateAssetObject)
		dbToken := db.QueryTokenBackupById(tx, obj.AssetID, height)
		db.ReplaceToken(tx, dbToken)
	} else if action.Type == types.UpdateAssetContract {
		obj := iActionAsset.(types.UpdateAssetContractObject)
		dbToken := db.QueryTokenBackupById(tx, obj.AssetID, height)
		db.ReplaceToken(tx, dbToken)
	}
}
