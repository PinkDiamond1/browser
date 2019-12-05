package task

import (
	"database/sql"
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

func subBalance(from string, assetId uint64, value *big.Int, h uint64, ut uint64, dbTx *sql.Tx) error {
	balance, err := db.GetAccountBalance(from, assetId, dbTx)
	if err != nil {
		ZapLog.Error("GetAccountBalance error: ", zap.Error(err), zap.String("from", from))
		return err
	}
	amount := big.NewInt(0).Sub(balance, value)
	if amount.Cmp(big.NewInt(0)) < 0 {
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

func addBalance(to string, assetId uint64, value *big.Int, h uint64, ut uint64, dbTx *sql.Tx) error {
	balance, err := db.GetAccountBalance(to, assetId, dbTx)
	if err != nil && err != sql.ErrNoRows {
		ZapLog.Error("Transfer error: ", zap.Error(err), zap.String("to", to))
		return err
	}
	if err == sql.ErrNoRows {
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

func (b *BalanceTask) analysisBalance(data *types.BlockAndResult, dbTx *sql.Tx) error {
	txs := data.Block.Txs
	receipts := data.Receipts
	detailTxs := data.DetailTxs
	balanceChangedMap := make(map[string]map[uint64]*big.Int)
	zeroBig := big.NewInt(0)
	for i, tx := range txs {
		receipt := receipts[i]
		for j, at := range tx.RPCActions {
			actionReceipt := receipt.ActionResults[j]
			gasPrice := big.NewInt(0).Set(tx.GasPrice)
			gasFrom := at.From.String()
			if gasPrice.Cmp(zeroBig) == 0 {
				if at.PayerGasPrice != nil {
					gasPrice.Set(at.PayerGasPrice)
					gasFrom = at.Payer.String()
				}
			}
			fee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(actionReceipt.GasUsed), gasPrice)
			if data.Block.Number.Uint64() > 0 {
				if gasFrom != "" {
					changeBalance(balanceChangedMap, gasFrom, tx.GasAssetID, fee, false)
				}
				changeBalance(balanceChangedMap, config.Chain.ChainFeeName, tx.GasAssetID, fee, true)
			}
			if actionReceipt.Status == types.ReceiptStatusSuccessful {
				if at.Amount.Cmp(big.NewInt(0)) > 0 {
					if at.From.String() != "" {
						changeBalance(balanceChangedMap, at.From.String(), at.AssetID, at.Amount, false)
					}
					if at.To.String() != "" {
						changeBalance(balanceChangedMap, at.To.String(), at.AssetID, at.Amount, true)
					}
				}
				if data.Block.Number.Uint64() == 0 && at.Type == types.IssueAsset {
					payload, err := parsePayload(at)
					if err != nil {
						ZapLog.Error("parse payload error: ", zap.Error(err))
						return err
					}
					arg := payload.(types.IssueAssetObject)
					assetInfo, err := client.GetAssetInfoByName(arg.AssetName)
					if err != nil {
						ZapLog.Error("GetAssetInfoByName error", zap.Error(err))
						return err
					}
					err = db.InsertAccountBalance(arg.Owner.String(), arg.Amount, assetInfo.AssetId, data.Block.Number.Uint64(), data.Block.Time, dbTx)
					if err != nil {
						ZapLog.Error("InsertAccountBalance error: ", zap.Error(err), zap.String("owner", arg.Owner.String()))
						return err
					}
				}
				if len(detailTxs) != 0 {
					internalActions := detailTxs[i].InternalActions[j]
					for _, iat := range internalActions.InternalLogs {
						if iat.Action.From.String() != "" {
							changeBalance(balanceChangedMap, iat.Action.From.String(), iat.Action.AssetID, iat.Action.Amount, false)
						}
						if iat.Action.To.String() != "" {
							changeBalance(balanceChangedMap, iat.Action.To.String(), iat.Action.AssetID, iat.Action.Amount, true)
						}
					}
				}
			}
		}
	}
	bigZero := big.NewInt(0)
	h := data.Block.Number.Uint64()
	ut := data.Block.Time
	for name, bs := range balanceChangedMap {
		for assetId, v := range bs {
			rs := v.Cmp(bigZero)
			if rs > 0 {
				err := addBalance(name, assetId, v, h, ut, dbTx)
				if err != nil {
					ZapLog.Error("addBalance error", zap.Error(err), zap.String("name", name), zap.Uint64("assetId", assetId))
					return err
				}
			} else if rs < 0 {
				absv := v.Abs(v)
				err := subBalance(name, assetId, absv, h, ut, dbTx)
				if err != nil {
					ZapLog.Error("subBalance error", zap.Error(err), zap.String("name", name), zap.Uint64("assetId", assetId))
					return err
				}

			}
		}
	}
	return nil
}

func changeBalance(balancesMap map[string]map[uint64]*big.Int, name string, assetId uint64, value *big.Int, add bool) {
	if add {
		if bs, ok := balancesMap[name]; ok {
			if b, ok := bs[assetId]; ok {
				b = b.Add(b, value)
			} else {
				balancesMap[name][assetId] = big.NewInt(0).Set(value)
			}
		} else {
			balancesMap[name] = make(map[uint64]*big.Int)
			balancesMap[name][assetId] = big.NewInt(0).Set(value)
		}
	} else {
		if bs, ok := balancesMap[name]; ok {
			if b, ok := bs[assetId]; ok {
				b = b.Sub(b, value)
			} else {
				b = big.NewInt(0)
				b = b.Sub(b, value)
				bs[assetId] = b
			}
		} else {
			b := big.NewInt(0)
			b = b.Sub(b, value)
			balancesMap[name] = make(map[uint64]*big.Int)
			balancesMap[name][assetId] = b
		}
	}
}

func (b *BalanceTask) rollback(data *types.BlockAndResult, dbTx *sql.Tx) error {
	txs := data.Block.Txs
	receipts := data.Receipts
	detailTxs := data.DetailTxs
	zeroBig := big.NewInt(0)
	balanceChangedMap := make(map[string]map[uint64]*big.Int)
	for i, tx := range txs {
		receipt := receipts[i]
		for j, at := range tx.RPCActions {
			actionReceipt := receipt.ActionResults[j]
			gasPrice := big.NewInt(0).Set(tx.GasPrice)
			gasFrom := at.From.String()
			if gasPrice.Cmp(zeroBig) == 0 {
				if at.PayerGasPrice != nil {
					gasPrice.Set(at.PayerGasPrice)
					gasFrom = at.Payer.String()
				}
			}
			fee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(actionReceipt.GasUsed), gasPrice)
			if gasFrom != "" {
				err := addBalance(gasFrom, tx.GasAssetID, fee, data.Block.Number.Uint64(), data.Block.Time, dbTx)
				if err != nil {
					ZapLog.Error("add fee error: ", zap.Error(err), zap.String("fee from", gasFrom))
					return err
				}
			}
			if actionReceipt.Status == types.ReceiptStatusSuccessful {
				if at.Amount.Cmp(big.NewInt(0)) > 0 {
					if at.To.String() != "" {
						changeBalance(balanceChangedMap, at.To.String(), at.AssetID, at.Amount, false)
					}
					if at.From.String() != "" {
						changeBalance(balanceChangedMap, at.From.String(), at.AssetID, at.Amount, true)
					}
				}
				if len(detailTxs) != 0 {
					internalActions := detailTxs[i].InternalActions[j]
					for _, iat := range internalActions.InternalLogs {
						if iat.Action.To.String() != "" {
							changeBalance(balanceChangedMap, iat.Action.To.String(), iat.Action.AssetID, iat.Action.Amount, false)
						}
						if iat.Action.From.String() != "" {
							changeBalance(balanceChangedMap, iat.Action.From.String(), iat.Action.AssetID, iat.Action.Amount, true)
						}
					}
				}
			}
		}
	}
	bigZero := big.NewInt(0)
	h := data.Block.Number.Uint64()
	ut := data.Block.Time
	for name, bs := range balanceChangedMap {
		for assetId, v := range bs {
			rs := v.Cmp(bigZero)
			if rs > 0 {
				err := addBalance(name, assetId, v, h, ut, dbTx)
				if err != nil {
					ZapLog.Error("addBalance error", zap.Error(err), zap.String("name", name), zap.Uint64("assetId", assetId))
					return err
				}
			} else if rs < 0 {
				absv := v.Abs(v)
				err := subBalance(name, assetId, absv, h, ut, dbTx)
				if err != nil {
					ZapLog.Error("subBalance error", zap.Error(err), zap.String("name", name), zap.Uint64("assetId", assetId))
					return err
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
			if d.Block.Block.Number.Uint64() >= b.startHeight {
				b.init()
				err := b.analysisBalance(d.Block, b.Tx)
				if err != nil {
					ZapLog.Error("BalanceTask analysisBalance error: ", zap.Error(err), zap.Uint64("height", d.Block.Block.Number.Uint64()))
					panic(err)
				}
				b.startHeight++
				b.commit()
			}
			result <- true
		case rd := <-rollbackData:
			b.startHeight--
			if rd.Block.Block.Number.Uint64() == b.startHeight {
				b.init()
				err := b.rollback(rd.Block, b.Tx)
				if err != nil {
					ZapLog.Error("BalanceTask rollback error: ", zap.Error(err), zap.Uint64("height", rd.Block.Block.Number.Uint64()))
					panic(err)
				}
				b.commit()
			}
			result <- true
		}
	}
}
