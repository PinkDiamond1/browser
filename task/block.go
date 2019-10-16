package task

import (
	"github.com/browser/client"
	"math/big"

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
	height := d.Block.Head.Number.Uint64()
	blockFee := big.NewInt(0)
	txs := d.Block.Txs
	for i := 0; i < len(txs); i++ {
		receipt := d.Receipts[i]
		err := db.InsertBlockTx(b.Tx, height, receipt.TxHash)
		if err != nil {
			ZapLog.Panic("InsertBlockTx error: ", zap.Error(err), zap.Uint64("height", height))
			return err
		}
		fee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(receipt.TotalGasUsed), big.NewInt(0).SetUint64(txs[i].GasPrice.Uint64()))
		blockFee = blockFee.Add(blockFee, fee)
	}
	err := db.InsertBlockChain(b.Tx, d.Block.Head, d.Hash, blockFee, len(txs))
	if err != nil {
		ZapLog.Panic("InsertBlockChain error: ", zap.Error(err), zap.Uint64("height", height))
		return err
	}
	h, _, err := client.GetCurrentBlockInfo()
	if err != nil {
		ZapLog.Error("GetCurrentBlockInfo error", zap.Error(err))
		return err
	}
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
	height := d.Block.Head.Number.Uint64()
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
	err = db.UpdateBlockStatus(b.Tx, height)
	if err != nil {
		ZapLog.Panic("UpdateBlockStatus error: ", zap.Error(err), zap.Uint64("height", height))
		return err
	}
	return nil
}

func (b *BlockTask) Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64) {
	b.startHeight = startHeight
	for {
		select {
		case d := <-data:
			if d.Block.Block.Head.Number.Uint64() >= b.startHeight {
				b.init()
				err := b.analysisBlock(d.Block, b.Tx)
				if err != nil {
					ZapLog.Error("AccountTask analysisAccount error: ", zap.Error(err), zap.Uint64("height", d.Block.Block.Head.Number.Uint64()))
					panic(err)
				}
				b.startHeight++
				b.commit()
			}
			result <- true
		case rd := <-rollbackData:
			if b.startHeight == rd.Block.Block.Head.Number.Uint64() {
				b.init()
				err := b.rollback(rd.Block, b.Tx)
				if err != nil {
					ZapLog.Error("AccountTask rollback error: ", zap.Error(err), zap.Uint64("height", rd.Block.Block.Head.Number.Uint64()))
					panic(err)
				}
				b.startHeight--
				b.commit()
			}
			result <- true
		}
	}
}
