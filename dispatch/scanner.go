package dispatch

import (
	"context"
	"github.com/browser_service/client"
	"github.com/browser_service/config"
	. "github.com/browser_service/log"
	"github.com/browser_service/types"
	"go.uber.org/zap"
	"sync"
)

var (
	safetyConfirmations = int64(0)
	rpcWorkers          = 10
)

func scanning(fromHeight, toHeight int64, blockDataChan chan *types.BlockAndResult) {
	//init
	scannerCtx, _ := context.WithCancel(context.Background())
	ZapLog.Info("batch pull block start", zap.Int64("fromHeight", fromHeight), zap.Int64("toHeight", toHeight))

	startNumber := fromHeight
	chParse := make(chan int64, 10)
	chSave := make(chan *types.BlockAndResult, 100)

	wg := &sync.WaitGroup{}
	cctx, cancel := context.WithCancel(scannerCtx)

	//rpc goroutine
	wg.Add(1)
	go func(ctx context.Context, chParse chan int64) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if fromHeight+safetyConfirmations <= toHeight {
					chParse <- fromHeight
					fromHeight = fromHeight + 1
				}
			}
		}
	}(cctx, chParse)

	//rpc goroutine
	wg.Add(1)
	go func(ctx context.Context, chParse chan int64, chSave chan *types.BlockAndResult) {
		defer wg.Done()
		defer close(chParse)

		workers := rpcWorkers
		chQueue := make([]chan *types.BlockAndResult, workers)
		for i := 0; i < workers; i++ {
			chQueue[i] = make(chan *types.BlockAndResult)
		}
		defer func() {
			for _, ch := range chQueue {
				close(ch)
			}
		}()

		twg := &sync.WaitGroup{}
		twg.Add(workers)
		for i := 0; i < workers; i++ {
			go func(chParse chan int64, chQueue []chan *types.BlockAndResult) {
				defer twg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case number := <-chParse:
						block, err := client.GetBlockAndResult(number)
						if err != nil {
							ZapLog.Panic("getBlockByNumber failed", zap.Error(err))
						}
						index := (number - startNumber) % int64(rpcWorkers)

						select {
						case chQueue[index] <- block:
							break
						case <-ctx.Done():
							break
						}
					}
				}
			}(chParse, chQueue)
		}

		twg.Add(1)
		go func(chQueue []chan *types.BlockAndResult) {
			defer twg.Done()
			i := 0
			ch := chQueue[i]
			for {
				select {
				case <-ctx.Done():
					return
				case ret := <-ch:
					chSave <- ret
					i++
					if i == len(chQueue) {
						i = 0
					}
					ch = chQueue[i]
				}
			}
		}(chQueue)
		twg.Wait()
	}(cctx, chParse, chSave)

	//db goroutine
	wg.Add(1)
	go func(ctx context.Context, chSave chan *types.BlockAndResult, cancel context.CancelFunc) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case block := <-chSave:
				currentHeight := block.Block.Head.Number.Int64()
				ZapLog.Debug("get block data finished", zap.Int64("height", currentHeight))
				if currentHeight%config.Log.SyncBlockShowNumber == 0 {
					ZapLog.Info("get block data finished", zap.Int64("height", currentHeight))
				}
				blockDataChan <- block
				if currentHeight == toHeight {
					ZapLog.Info("Cancel")
					cancel()
				}
			}
		}
	}(cctx, chSave, cancel)
	wg.Wait()
	close(chSave)
	ZapLog.Info("End", zap.Int64("fromNumber", fromHeight), zap.Int64("toNumber", toHeight))
	ZapLog.Info("Sync OK")
}
