package task

import (
	"math/big"

	"github.com/browser/client"

	"database/sql"

	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
)

type BlockTask struct {
	*Base
}

func (b *BlockTask) analysisBlock(d *types.BlockAndResult, dbTx *sql.Tx) error {
	height := d.Block.Number.Uint64()
	blockFee := big.NewInt(0)
	txs := d.Block.Txs
	bigZero := big.NewInt(0)
	for i := 0; i < len(txs); i++ {
		receipt := d.Receipts[i]
		err := db.InsertBlockTx(b.Tx, height, receipt.TxHash)
		if err != nil {
			ZapLog.Panic("InsertBlockTx error: ", zap.Error(err), zap.Uint64("height", height))
			return err
		}
		gasPrice := big.NewInt(0).Set(txs[i].GasPrice)
		if gasPrice.Cmp(bigZero) == 0 {
			if txs[i].RPCActions[0].PayerGasPrice != nil {
				gasPrice.Set(txs[i].RPCActions[0].PayerGasPrice)
			}
		}
		fee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(receipt.TotalGasUsed), gasPrice)
		blockFee = blockFee.Add(blockFee, fee)
	}
	err := db.InsertBlockChain(b.Tx, d.Block, d.Block.Hash, blockFee, len(txs))
	if err != nil {
		ZapLog.Panic("InsertBlockChain error: ", zap.Error(err), zap.Uint64("height", height))
		return err
	}
	cb, err := client.GetCurrentBlockInfo()
	if err != nil {
		ZapLog.Error("GetCurrentBlockInfo error", zap.Error(err))
		return err
	}
	h := cb.Number.Uint64()
	candidatesCount, err := client.GetCandidatesCount()
	if err != nil {
		ZapLog.Error("GetCandidatesCount error", zap.Error(err))
		return err
	}
	data := map[string]interface{}{
		"height":          h,
		"producer_number": candidatesCount,
	}
	err = db.UpdateChainStatus(dbTx, data)
	if err != nil {
		ZapLog.Error("UpdateChainStatus error: ", zap.Error(err))
		return err
	}
	return nil
}

func (b *BlockTask) rollback(d *types.BlockAndResult, dbTx *sql.Tx) error {
	height := d.Block.Number.Uint64()
	err := db.DeleteBlock(b.Tx, height)
	if err != nil {
		ZapLog.Panic("DeleteBlock error: ", zap.Error(err), zap.Uint64("height", height))
		return err
	}
	err = db.DeleteBlockTx(b.Tx, height)
	if err != nil {
		ZapLog.Panic("DeleteBlockTx error: ", zap.Error(err), zap.Uint64("height", height))
		return err
	}
	data := map[string]interface{}{
		"height": height,
	}
	err = db.UpdateChainStatus(dbTx, data)
	if err != nil {
		ZapLog.Error("UpdateChainStatus error: ", zap.Error(err))
		return err
	}
	return nil
}

func (b *BlockTask) Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64) {
	b.startHeight = startHeight
	for {
		select {
		case d := <-data:
			if d.Block.Block.Number.Uint64() >= b.startHeight {
				b.init()
				err := b.analysisBlock(d.Block, b.Tx)
				if err != nil {
					ZapLog.Error("AccountTask analysisAccount error: ", zap.Error(err), zap.Uint64("height", d.Block.Block.Number.Uint64()))
					panic(err)
				}
				b.startHeight++
				b.commit()
			}
			result <- true
		case rd := <-rollbackData:
			b.startHeight--
			if b.startHeight == rd.Block.Block.Number.Uint64() {
				b.init()
				err := b.rollback(rd.Block, b.Tx)
				if err != nil {
					ZapLog.Error("AccountTask rollback error: ", zap.Error(err), zap.Uint64("height", rd.Block.Block.Number.Uint64()))
					panic(err)
				}
				b.commit()
			}
			result <- true
		}
	}
}
