package statis

import (
	"time"

	"github.com/browser/client"
	"github.com/browser/config"
	"github.com/browser/db"
	. "github.com/browser/log"
	"go.uber.org/zap"
)

var (
	syncDuration = time.Second
)

func Start() {
	a := NewAnalysis()
	fromNumber, _ := db.GetBlockHeight()
	for {
		block, err := client.GetBlockAndResult(fromNumber)
		if err != nil {
			ZapLog.Error("Sync getBlockByNumber", zap.Error(err))
			time.Sleep(syncDuration)
			continue
		}
		if block == nil {
			time.Sleep(syncDuration)
			continue
		}
		irreversible, _ := client.GetDposIrreversible()
		if block.Block.Number.Uint64() <= irreversible.BftIrreversible {
			if block.Block.Number.Int64()%config.Log.SyncBlockShowNumber == 0 {
				ZapLog.Info("statistical", zap.Int64("number", block.Block.Number.Int64()), zap.Uint64("height", irreversible.BftIrreversible), zap.Int("txs", len(block.Block.Txs)))
			}
			if err := a.process(block); err != nil {
				ZapLog.Error("statistical commitBlock ", zap.Error(err))
				panic(err)
			}
			fromNumber = block.Block.Number.Int64() + 1
		} else {
			time.Sleep(syncDuration)
		}
	}
}
