package task

import (
	"database/sql"
	"fmt"
	"github.com/browser/client"
	"github.com/browser/config"
	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
	"math/big"
)

type BalanceTask struct {
	*Base
}

func subBalance(from string, assetId uint64, value *big.Int, h uint64, ut uint, dbTx *sql.Tx) error {

	balance, err := db.GetAccountBalance(from, assetId, dbTx)
	if err != nil {
		ZapLog.Error("GetAccountBalance error: ", zap.Error(err), zap.String("from", from))
		return err
	}
	amount := big.NewInt(0).Sub(balance, value)
	if amount.Cmp(big.NewInt(0)) < 0 {
		fmt.Println("---", balance.String(), value.String())
		ZapLog.Error("from balance not enough", zap.String("from", from))
		return BalanceNotEnough
	}
	err = db.UpdateAccountBalance(from, amount, assetId, h, ut, dbTx)
	if err != nil {
		ZapLog.Error("Transfer error", zap.Error(err), zap.String("from", from))
		return err
	}
	return nil
}

func addBalance(to string, assetId uint64, value *big.Int, h uint64, ut uint, dbTx *sql.Tx, careAsset bool) error {
	balance, err := db.GetAccountBalance(to, assetId, dbTx)
	if err != nil && err != sql.ErrNoRows {
		ZapLog.Error("Transfer error: ", zap.Error(err), zap.String("to", to))
		return err
	}
	if err == sql.ErrNoRows {
		if careAsset {
			return err
		}
		err = db.InsertAccountBalance(to, value, assetId, h, ut, dbTx)
		if err != nil {
			ZapLog.Error("Transfer error: ", zap.Error(err), zap.String("to", to))
			return err
		}
	} else {
		amount := big.NewInt(0).Add(balance, value)
		err = db.UpdateAccountBalance(to, amount, assetId, h, ut, dbTx)
		if err != nil {
			ZapLog.Error("Transfer error: ", zap.Error(err), zap.String("to", to))
			return err
		}
	}
	return nil
}

func transfer(from, to string, assetId uint64, value *big.Int, h uint64, ut uint, dbTx *sql.Tx) error {
	if from != "" && from != config.Chain.ChainFeeName {
		err := subBalance(from, assetId, value, h, ut, dbTx)
		if err != nil {
			return err
		}
	}
	if to != "" && to != config.Chain.ChainFeeName {
		err := addBalance(to, assetId, value, h, ut, dbTx, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BalanceTask) analysisBalance(data *types.BlockAndResult, dbTx *sql.Tx) error {
	txs := data.Block.Txs
	receipts := data.Receipts
	detailTxs := data.DetailTxs
	for i, tx := range txs {
		receipt := receipts[i]
		for j, at := range tx.RPCActions {
			actionReceipt := receipt.ActionResults[j]
			fee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(actionReceipt.GasUsed), big.NewInt(0).SetUint64(tx.GasPrice.Uint64()))
			if data.Block.Head.Number.Uint64() > 0 {
				if at.From.String() != "" && at.From.String() != config.Chain.ChainFeeName {
					err := subBalance(at.From.String(), tx.GasAssetID, fee, data.Block.Head.Number.Uint64(), data.Block.Head.Time, dbTx)
					if err != nil {
						ZapLog.Error("sub fee error: ", zap.Error(err), zap.String("fee from", at.From.String()))
						return err
					}
				}
			}
			if actionReceipt.Status == types.ReceiptStatusSuccessful {
				if at.Amount.Cmp(big.NewInt(0)) > 0 {
					err := transfer(at.From.String(), at.To.String(), at.AssetID, at.Amount, data.Block.Head.Number.Uint64(), data.Block.Head.Time, dbTx)
					if err != nil {
						ZapLog.Error("transfer error: ", zap.Error(err))
						return err
					}
				}
				payload, err := parsePayload(at)
				if err != nil {
					ZapLog.Error("parse payload error: ", zap.Error(err))
					return err
				}
				if data.Block.Head.Number.Uint64() == 0 && at.Type == types.IssueAsset {
					arg := payload.(types.IssueAssetObject)
					assetInfo, err := client.GetAssetInfoByName(arg.AssetName)
					if err != nil {
						ZapLog.Error("GetAssetInfoByName error", zap.Error(err))
						return err
					}
					err = db.InsertAccountBalance(arg.Owner.String(), arg.Amount, assetInfo.AssetId, data.Block.Head.Number.Uint64(), data.Block.Head.Time, dbTx)
					if err != nil {
						ZapLog.Error("InsertAccountBalance error: ", zap.Error(err), zap.String("owner", arg.Owner.String()))
						return err
					}
				}
				if len(detailTxs) != 0 {
					internalActions := detailTxs[i].InternalActions[j]
					for _, iat := range internalActions.InternalLogs {
						err := transfer(iat.Action.From.String(), iat.Action.To.String(), iat.Action.AssetID, iat.Action.Amount, data.Block.Head.Number.Uint64(), data.Block.Head.Time, dbTx)
						if err != nil {
							ZapLog.Error("transfer error: ", zap.Error(err))
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func (b *BalanceTask) rollback(data *types.BlockAndResult, dbTx *sql.Tx) error {
	txs := data.Block.Txs
	receipts := data.Receipts
	detailTxs := data.DetailTxs
	for i, tx := range txs {
		receipt := receipts[i]
		for j, at := range tx.RPCActions {
			actionReceipt := receipt.ActionResults[j]
			fee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(actionReceipt.GasUsed), big.NewInt(0).SetUint64(tx.GasPrice.Uint64()))
			if at.From.String() != "" && at.From.String() != config.Chain.ChainFeeName {
				err := addBalance(at.From.String(), tx.GasAssetID, fee, data.Block.Head.Number.Uint64(), data.Block.Head.Time, dbTx, false)
				if err != nil {
					ZapLog.Error("sub fee error: ", zap.Error(err), zap.String("fee from", at.From.String()))
					return err
				}
			}
			if actionReceipt.Status == types.ReceiptStatusSuccessful {
				if at.Amount.Cmp(big.NewInt(0)) > 0 {
					err := transfer(at.To.String(), at.From.String(), at.AssetID, at.Amount, data.Block.Head.Number.Uint64(), data.Block.Head.Time, dbTx)
					if err != nil {
						ZapLog.Error("transfer error: ", zap.Error(err))
						return err
					}
				}
				if len(detailTxs) != 0 {
					internalActions := detailTxs[i].InternalActions[j]
					for _, iat := range internalActions.InternalLogs {
						err := transfer(iat.Action.To.String(), iat.Action.From.String(), iat.Action.AssetID, iat.Action.Amount, data.Block.Head.Number.Uint64(), data.Block.Head.Time, dbTx)
						if err != nil {
							ZapLog.Error("transfer error: ", zap.Error(err))
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func (b *BalanceTask) Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64) {
	b.startHeight = startHeight
	for {
		select {
		case d := <-data:
			if d.Block.Block.Head.Number.Uint64() >= b.startHeight {
				b.init()
				err := b.analysisBalance(d.Block, b.Tx)
				if err != nil {
					ZapLog.Error("BalanceTask analysisAction error: ", zap.Error(err), zap.Uint64("height", d.Block.Block.Head.Number.Uint64()))
					panic(err)
				}
				b.startHeight++
				b.commit()
			}
			result <- true
		case rd := <-rollbackData:
			b.startHeight--
			if rd.Block.Block.Head.Number.Uint64() == b.startHeight {
				fmt.Println("----", rd.Block.Block.Head.Number.Uint64(), b.startHeight)
				b.init()
				err := b.rollback(rd.Block, b.Tx)
				if err != nil {
					ZapLog.Error("ActionTask rollback error: ", zap.Error(err), zap.Uint64("height", rd.Block.Block.Head.Number.Uint64()))
					panic(err)
				}
				b.commit()
			}
			result <- true
		}
	}
}
