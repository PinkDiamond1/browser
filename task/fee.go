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

type FeeTask struct {
	*Base
	TokenIncome    *big.Int
	ContractIncome *big.Int
	nodeTokens     map[string]string
}

func (f *FeeTask) getTokenName(dbTx *sql.Tx, name string) (string, error) {
	if tokenName, ok := f.nodeTokens[name]; ok {
		return tokenName, nil
	} else {
		asset, err := client.GetAssetInfoByName(name)
		if err != nil {
			ZapLog.Error("GetAssetInfoByName error", zap.Error(err), zap.String("name", name))
			return "", err
		}
		dbToken := db.QueryTokenById(dbTx, asset.AssetId)
		f.nodeTokens[name] = dbToken.AssetName
		return dbToken.AssetName, nil
	}
}

func (f *FeeTask) analysisFeeAction(data *types.BlockAndResult, dbTx *sql.Tx) error {
	receipts := data.Receipts
	txs := data.Block.Txs
	bigZero := big.NewInt(0)
	for i, receipt := range receipts {
		tx := txs[i]
		gasPrice := big.NewInt(0).Set(tx.GasPrice)
		for j, aRs := range receipt.ActionResults {
			at := tx.RPCActions[j]
			if at.Type == types.IssueAsset && aRs.Status == types.ReceiptStatusSuccessful {
				iActionAsset, err := parsePayload(at)
				if err != nil {
					ZapLog.Error("parsePayload error: ", zap.Error(err))
					return err
				}
				obj := iActionAsset.(types.IssueAssetObject)
				tokenName := obj.AssetName
				if idx := strings.Index(obj.AssetName, ":"); idx <= 0 {
					if len(at.From.String()) > 0 {
						tokenName = at.From.String() + ":" + obj.AssetName
					}
				}
				f.nodeTokens[obj.AssetName] = tokenName
			}
			feeFrom := at.From.String()
			for k, aR := range aRs.GasAllot {
				if tx.GasPrice.Cmp(bigZero) == 0 {
					if at.PayerGasPrice != nil {
						gasPrice.Set(at.PayerGasPrice)
						feeFrom = at.Payer.String()
					}
				}
				fee := big.NewInt(0).Mul(big.NewInt(int64(aR.Gas)), gasPrice)
				mFee := &db.MysqlFee{
					TxHash:      tx.Hash.String(),
					ActionHash:  at.ActionHash.String(),
					ActionIndex: j,
					FeeIndex:    k,
					Height:      data.Block.Number.Uint64(),
					Created:     data.Block.Time,
					AssetId:     tx.GasAssetID,
					From:        feeFrom,
					To:          aR.Account.String(),
					Amount:      fee,
					Reason:      aR.Reason,
				}
				if aR.Reason == 0 {
					assetName, err := f.getTokenName(dbTx, aR.Account.String())
					if err != nil {
						return err
					}
					mFee.To = assetName
				}
				err := db.InsertFee(mFee, dbTx)
				if err != nil {
					ZapLog.Error("InsertFee error: ", zap.Error(err))
					return err
				}
				//0 asset 1 contract 2 produce
				if aR.Reason == 0 {
					f.TokenIncome.Add(f.TokenIncome, fee)
				} else if aR.Reason == 1 {
					f.ContractIncome.Add(f.ContractIncome, fee)
				}
			}
		}
	}
	d := map[string]interface{}{
		"token_income":    f.TokenIncome.String(),
		"contract_income": f.ContractIncome.String(),
	}
	err := db.UpdateChainStatus(dbTx, d)
	if err != nil {
		ZapLog.Error("UpdateChainStatus error: ", zap.Error(err))
		return err
	}
	return nil
}

func (f *FeeTask) rollback(data *types.BlockAndResult, dbTx *sql.Tx) error {
	for i, tx := range data.Block.Txs {
		for j, at := range tx.RPCActions {
			err := db.DeleteFeeByActionHash(at.ActionHash, dbTx)
			if err != nil {
				ZapLog.Error("DeleteFeeByActionHash error:", zap.Error(err), zap.Uint64("height", data.Block.Number.Uint64()))
				return err
			}
			aRs := data.Receipts[i].ActionResults[j].GasAllot
			for _, aR := range aRs {
				fee := big.NewInt(0).Mul(big.NewInt(int64(aR.Gas)), tx.GasPrice)
				if aR.Reason == 0 {
					f.TokenIncome.Sub(f.TokenIncome, fee)
				} else if aR.Reason == 1 {
					f.ContractIncome.Sub(f.ContractIncome, fee)
				}
			}
		}
	}
	d := map[string]interface{}{
		"token_income":    f.TokenIncome.String(),
		"contract_income": f.ContractIncome.String(),
	}
	err := db.UpdateChainStatus(dbTx, d)
	if err != nil {
		ZapLog.Error("UpdateChainStatus error: ", zap.Error(err))
		return err
	}
	return nil
}

func (f *FeeTask) Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64) {
	f.nodeTokens = make(map[string]string)
	f.startHeight = startHeight
	chain, err := db.Mysql.GetChainStatus()
	if err != nil {
		ZapLog.Panic("GetChainStatus error", zap.Error(err))
	}
	f.TokenIncome = chain.TokenIncome
	f.ContractIncome = chain.ContractIncome
	for {
		select {
		case d := <-data:
			if d.Block.Block.Number.Uint64() >= f.startHeight {
				f.init()
				err := f.analysisFeeAction(d.Block, f.Tx)
				if err != nil {
					ZapLog.Error("FeeTask analysisFeeAction error: ", zap.Error(err), zap.Uint64("height", d.Block.Block.Number.Uint64()))
					panic(err)
				}
				f.startHeight++
				f.commit()
			}
			result <- true
		case rd := <-rollbackData:
			f.startHeight--
			if f.startHeight == rd.Block.Block.Number.Uint64() {
				f.init()
				err := f.rollback(rd.Block, f.Tx)
				if err != nil {
					ZapLog.Error("ActionTask rollback error: ", zap.Error(err), zap.Uint64("height", rd.Block.Block.Number.Uint64()))
					panic(err)
				}
				f.commit()
			}
			result <- true
		}
	}
}
