package task

import (
	"database/sql"
	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
	"math/big"
)

type TransactionTask struct {
	*Base
	TxCount   uint64
	FeeIncome *big.Int
}

func (b *TransactionTask) analysisTransaction(data *types.BlockAndResult, dbTx *sql.Tx) error {
	block := data.Block
	receipts := data.Receipts
	for i, tx := range block.Txs {
		receipt := receipts[i]
		mTx := &db.MysqlTx{
			Hash:        tx.Hash.String(),
			Height:      block.Head.Number.Uint64(),
			GasUsed:     receipt.TotalGasUsed,
			GasCost:     tx.GasCost,
			GasPrice:    tx.GasPrice,
			GasAssetId:  tx.GasAssetID,
			BlockHash:   data.Hash.String(),
			TxIndex:     i,
			ActionCount: len(tx.RPCActions),
		}
		state := 1
		for j := 0; j < len(receipt.ActionResults); j++ {
			if receipt.ActionResults[j].Status != uint64(types.ReceiptStatusSuccessful) {
				state = 0
			}
		}
		mTx.State = state
		err := db.InsertTransaction(mTx, dbTx)
		if err != nil {
			ZapLog.Error("InsertTransaction error: ", zap.Error(err), zap.Uint64("height", block.Head.Number.Uint64()), zap.Int("txIndex", i))
			return err
		}
		fee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(mTx.GasUsed), mTx.GasPrice)
		b.FeeIncome = b.FeeIncome.Add(b.FeeIncome, fee)
	}
	b.TxCount += uint64(len(block.Txs))
	d := map[string]interface{}{
		"tx_count":   b.TxCount,
		"fee_income": b.FeeIncome.String(),
	}
	err := db.UpdateChainStatus(dbTx, d)
	if err != nil {
		ZapLog.Error("UpdateChainStatus error: ", zap.Error(err))
		return err
	}
	return nil
}

func (b *TransactionTask) rollback(data *types.BlockAndResult, dbTx *sql.Tx) error {
	for i, tx := range data.Block.Txs {
		receipt := data.Receipts[i]
		err := db.DeleteTransactionByHash(tx.Hash, dbTx)
		if err != nil {
			ZapLog.Error("DeleteTransactionByHash error: ", zap.Error(err), zap.Uint64("height", data.Block.Head.Number.Uint64()), zap.String("txHash", tx.Hash.String()))
			return err
		}
		fee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(receipt.TotalGasUsed), tx.GasPrice)
		b.FeeIncome = b.FeeIncome.Sub(b.FeeIncome, fee)
	}
	b.TxCount -= uint64(len(data.Block.Txs))
	d := map[string]interface{}{
		"tx_count":   b.TxCount,
		"fee_income": b.FeeIncome.String(),
	}
	err := db.UpdateChainStatus(dbTx, d)
	if err != nil {
		ZapLog.Error("UpdateChainStatus error: ", zap.Error(err))
		return err
	}
	return nil
}

func (b *TransactionTask) Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64) {
	b.startHeight = startHeight
	chain, err := db.Mysql.GetChainStatus()
	if err != nil {
		ZapLog.Panic("GetChainStatus error", zap.Error(err))
	}
	b.TxCount = chain.TxCount
	b.FeeIncome = chain.FeeIncome
	for {
		select {
		case d := <-data:
			if d.Block.Block.Head.Number.Uint64() >= b.startHeight {
				b.init()
				err := b.analysisTransaction(d.Block, b.Tx)
				if err != nil {
					ZapLog.Error("TransactionTask analysisBlock error: ", zap.Error(err), zap.Uint64("height", d.Block.Block.Head.Number.Uint64()))
					panic(err)
				}
				b.startHeight++
				b.commit()
			}
			result <- true
		case rd := <-rollbackData:
			b.startHeight--
			if b.startHeight == rd.Block.Block.Head.Number.Uint64() {
				b.init()
				err := b.rollback(rd.Block, b.Tx)
				if err != nil {
					ZapLog.Error("TransactionTask analysisBlock error: ", zap.Error(err), zap.Uint64("height", rd.Block.Block.Head.Number.Uint64()))
					panic(err)
				}
				b.commit()
			}
			result <- true
		}
	}
}
