package dispatch

import (
	"github.com/browser_service/client"
	"github.com/browser_service/config"
	"github.com/browser_service/db"
	. "github.com/browser_service/log"
	"github.com/browser_service/rlp"
	"github.com/browser_service/task"
	"github.com/browser_service/types"
	"go.uber.org/zap"
	"time"
)

var (
	syncDuration = 100 * time.Millisecond
)

func NewDispatch() *Dispatch {
	blockDataChan := make(chan *types.BlockAndResult, config.BlockDataChanBufferSize)
	taskStatusMap, startHeight := getTaskStatus()
	taskCount := len(config.Tasks)
	taskDataChan := make([]chan *task.TaskChanData, taskCount)
	taskRollbackDataChan := make([]chan *task.TaskChanData, taskCount)
	taskResultChan := make(chan bool, taskCount)
	for i := 0; i < taskCount; i++ {
		taskDataChan[i] = make(chan *task.TaskChanData, 1)
		taskRollbackDataChan[i] = make(chan *task.TaskChanData, 1)
	}
	isRollbackChan := make(chan bool)
	return &Dispatch{
		blockDataChan:        blockDataChan,
		taskStatusMap:        taskStatusMap,
		startHeight:          startHeight,
		taskDataChan:         taskDataChan,
		taskResultChan:       taskResultChan,
		taskCount:            taskCount,
		taskRollbackDataChan: taskRollbackDataChan,
		isRollbackChan:       isRollbackChan,
	}
}

type Dispatch struct {
	blockDataChan        chan *types.BlockAndResult
	taskStatusMap        map[string]*db.TaskStatus
	startHeight          uint64
	batchTo              uint64
	taskDataChan         []chan *task.TaskChanData
	taskRollbackDataChan []chan *task.TaskChanData
	taskResultChan       chan bool
	taskCount            int
	isRollbackChan       chan bool
	currentBlock         *types.BlockAndResult
}

func (d *Dispatch) Start() {
	//async start task
	d.startTasks()
	//async start send block data task
	d.sendBlockToTask()

	//batch pull block
	d.batchPullIrreversibleBlock()

	//start a single block pull task
	isRollback := false
	startHeight := d.batchTo
	for {
		select {
		case isRollback = <-d.isRollbackChan:
			if !isRollback {
				//clear block
				for _, ok := <-d.blockDataChan; ok; {
				}
				startHeight = d.startHeight
			} else {
				time.Sleep(time.Duration(1) * time.Second)
			}
		default:
		}
		if !isRollback {
			block, err := client.GetBlockAndResult(int64(startHeight))
			if err != nil {
				ZapLog.Error("sync getBlockByNumber", zap.Error(err))
				time.Sleep(syncDuration)
				continue
			}
			if block == nil {
				time.Sleep(syncDuration)
				continue
			}
			d.blockDataChan <- block
			startHeight++
		}
	}
}

func getTaskStatus() (map[string]*db.TaskStatus, uint64) {
	taskStatusMap := db.Mysql.GetTaskStatus(config.Tasks)
	startHeight := uint64(0)
	if len(taskStatusMap) != len(config.Tasks) {
		for _, taskType := range config.Tasks {
			if _, ok := taskStatusMap[taskType]; !ok {
				db.Mysql.InitTaskStatus(taskType)
			}
		}
		taskStatusMap = db.Mysql.GetTaskStatus(config.Tasks)
	}

	startHeightInit := true
	for _, taskStatus := range taskStatusMap {
		if startHeightInit {
			startHeight = taskStatus.Height
			startHeightInit = false
		} else {
			if taskStatus.Height < startHeight {
				startHeight = taskStatus.Height
			}
		}

		if startHeight == 0 {
			break
		}
	}
	return taskStatusMap, startHeight
}

func (d *Dispatch) batchPullIrreversibleBlock() {
	irreversible, err := client.GetDposIrreversible()
	d.batchTo = irreversible.BftIrreversible
	if err != nil {
		ZapLog.Panic("get dpos irreversible error", zap.Error(err))
	}

	if d.startHeight < irreversible.BftIrreversible {
		scanning(int64(d.startHeight), int64(irreversible.BftIrreversible), d.blockDataChan)
	} else {
		d.batchTo = d.startHeight
	}
}

func (d *Dispatch) startTasks() {
	for i, taskType := range config.Tasks {
		if taskFunc, ok := task.TaskFunc[taskType]; ok {
			go taskFunc.Start(d.taskDataChan[i], d.taskRollbackDataChan[i], d.taskResultChan, d.taskStatusMap[taskType].Height)
		} else {
			ZapLog.Panic("task type or func not existing")
		}
	}
}

func (d *Dispatch) sendBlockToTask() {
	go func() {
		for {
			block := <-d.blockDataChan

			if block.Block.Head.Number.Uint64() > d.batchTo { //the inverse calculation block
				if d.currentBlock == nil {
					d.currentBlock = block
				} else {
					if d.currentBlock.Hash.String() != block.Block.Head.ParentHash.String() {
						d.currentBlock = block
						d.rollback()
						continue
					}
					d.currentBlock = block
				}
			}

			taskData := &task.TaskChanData{
				Block: block,
				Tx:    nil,
			}
			d.cacheBlock(taskData) //cache reversible block
			for _, taskDataChan := range d.taskDataChan {
				taskDataChan <- taskData
			}
			d.checkTaskResult()
			ZapLog.Info("commit success", zap.String("height", block.Block.Head.Number.String()))
		}
	}()
}

func (d *Dispatch) checkTaskResult() {
	returnCount := 0
	for {
		select {
		case <-d.taskResultChan:
			returnCount += 1
		default:

		}
		if returnCount == d.taskCount {
			return
		}
	}
}

func (d *Dispatch) rollback() {
	d.isRollbackChan <- true
	endHeight := d.currentBlock.Block.Head.Number.Uint64() - 1

	for ; ; endHeight-- {
		dbBlock := db.Mysql.GetBlockOriginalByHeight(endHeight)
		chainBlock, err := client.GetBlockAndResult(int64(endHeight))
		if err != nil {
			ZapLog.Panic("rpc get block error", zap.Error(err))
		}
		if dbBlock.BlockHash != chainBlock.Hash.String() {
			rollbackData := &task.TaskChanData{
				Block: BlobToBlock(dbBlock),
				Tx:    nil,
			}
			for _, rollbackChan := range d.taskRollbackDataChan {
				rollbackChan <- rollbackData
			}
			d.checkTaskResult()
		} else {
			d.startHeight = endHeight + 1
			break
		}
	}

	d.isRollbackChan <- false
}

func (d *Dispatch) cacheBlock(blockData *task.TaskChanData) {
	height := blockData.Block.Block.Head.Number.Uint64()
	if height > d.batchTo {
		irreversible, err := client.GetDposIrreversible()
		if err != nil {
			ZapLog.Panic("cache block data error", zap.Error(err))
		}

		if blockData.Block.Block.Head.Number.Uint64() > irreversible.BftIrreversible {
			db.AddReversibleBlockCache(blockData.Tx, BlockToBlob(blockData.Block))
			db.DeleteIrreversibleCache(blockData.Tx, height)
		}
	}
}

func BlobToBlock(block *db.BlockOriginal) *types.BlockAndResult {
	blockAndResult := &types.BlockAndResult{}
	err := rlp.DecodeBytes(block.BlockData, blockAndResult)
	if err != nil {
		ZapLog.Panic("decode block byte data error", zap.Error(err))
	}
	return blockAndResult
}

func BlockToBlob(block *types.BlockAndResult) *db.BlockOriginal {
	data, err := rlp.EncodeToBytes(block)
	if err != nil {
		ZapLog.Panic("encode block error", zap.Error(err))
	}
	result := &db.BlockOriginal{
		BlockData:  data,
		Height:     block.Block.Head.Number.Uint64(),
		BlockHash:  block.Hash.String(),
		ParentHash: block.Block.Head.ParentHash.String(),
	}
	return result
}
