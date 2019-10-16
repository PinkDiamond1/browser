package task

import (
	"database/sql"
	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
)

const (
	TaskTypeBlock          = "block"
	TaskTypeTxs            = "txs"
	TaskTypeAction         = "action"
	TaskTypeInternalAction = "internalAction"
	TaskTypeFeeAction      = "feeAction"
	TaskTypeAccount        = "account"
	TaskTypeAccountBalance = "accountBalance"
	TaskTypeToken          = "token"
	TaskTypeAccountHistory = "accountHistory"
	TaskTypeTokenHistory   = "tokenHistory"
	TaskTypeFeeHistory     = "feeHistory"
)

type Base struct {
	taskType    string
	startHeight uint64
	Tx          *sql.Tx
}

func (b *Base) init() {
	b.Tx = db.Mysql.Begin()

}

func (b *Base) commit() {
	db.UpdateTaskStatus(b.Tx, b.taskType, b.startHeight)
	err := b.Tx.Commit()
	if err != nil {
		_ = b.Tx.Rollback()
		ZapLog.Panic("commit transaction err", zap.Error(err), zap.String("taskType", b.taskType), zap.Uint64("height", b.startHeight))
	}
}

type TaskChanData struct {
	Block *types.BlockAndResult
	Tx    *sql.Tx
}

type Task interface {
	init()
	commit()
	Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64)
}

var TaskFunc = map[string]Task{
	TaskTypeBlock:          &BlockTask{Base: &Base{taskType: TaskTypeBlock}},
	TaskTypeTxs:            &TransactionTask{Base: &Base{taskType: TaskTypeTxs}},
	TaskTypeAction:         &ActionTask{Base: &Base{taskType: TaskTypeAction}},
	TaskTypeInternalAction: &InternalTask{Base: &Base{taskType: TaskTypeInternalAction}},
	TaskTypeFeeAction:      &FeeTask{Base: &Base{taskType: TaskTypeFeeAction}},
	TaskTypeAccount:        &AccountTask{Base: &Base{taskType: TaskTypeAccount}},
	TaskTypeAccountBalance: &BalanceTask{Base: &Base{taskType: TaskTypeAccountBalance}},
	TaskTypeToken:          &TokenTask{Base: &Base{taskType: TaskTypeToken}},
	TaskTypeAccountHistory: &AccountHistoryTask{Base: &Base{taskType: TaskTypeAccountHistory}},
	TaskTypeTokenHistory:   &TokeHistoryTask{Base: &Base{taskType: TaskTypeTokenHistory}},
	TaskTypeFeeHistory:     &TokenFeeHistoryTask{Base: &Base{taskType: TaskTypeFeeHistory}},
}
