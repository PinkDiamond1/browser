package task

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"github.com/browser/client"
	"github.com/browser/crypto"
	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/browser/types"
	"go.uber.org/zap"
	"strings"
)

type AccountTask struct {
	*Base
}

func (a *AccountTask) ActionToAccount(action *types.RPCAction, dbTx *sql.Tx, block *types.RpcBlock, oldAccounts map[string]struct{}) error {
	aType := action.Type
	payload, err := parsePayload(action)
	if err != nil {
		ZapLog.Error("parsePayload error:", zap.Error(err))
		return err
	}
	switch aType {
	case types.CreateAccount:
		arg := payload.(types.CreateAccountAction)
		founder := arg.Founder.String()
		if len(arg.Founder.String()) == 0 {
			founder = arg.AccountName.String()
		}
		parentName := ""
		if ok := strings.Contains(arg.AccountName.String(), "."); ok {
			parentName = action.From.String()
		}
		author := []*types.Author{
			&types.Author{
				Owner:  arg.PublicKey,
				Weight: uint64(1),
			},
		}
		dbAuthor := types.AccountAuthor{
			AuthorType: types.PubKeyType,
			Author:     arg.PublicKey.String(),
			Weight:     uint64(1),
		}
		accountAuthor := []*types.AccountAuthor{&dbAuthor}
		authorText, err := json.Marshal(accountAuthor)
		if err != nil {
			ZapLog.Error("accountAuthor Marshal error", zap.Error(err))
			return err
		}
		nAcct, err := client.GetAccountByName(arg.AccountName.String())
		if err != nil {
			ZapLog.Error("GetAccountByName error", zap.Error(err), zap.String("name", arg.AccountName.String()))
			return err
		}
		mAcct := &db.MysqlAccount{
			Name:                  arg.AccountName.String(),
			ParentName:            parentName,
			CreateUser:            action.From.String(),
			Founder:               founder,
			AccountID:             nAcct.AccountID,
			Number:                nAcct.Number,
			Nonce:                 0,
			Threshold:             1,
			UpdateAuthorThreshold: 1,
			Permissions:           string(authorText),
			Created:               block.Time,
			ContractCreated:       0,
			Description:           arg.Description,
		}
		authorVersion := types.RlpHash([]interface{}{
			author,
			mAcct.Threshold,
			mAcct.UpdateAuthorThreshold,
		})
		mAcct.AuthorVersion = authorVersion.String()
		mAcct.CodeHash = crypto.Keccak256Hash(nil).String()

		if block.Number.Uint64() == 0 {
			mAcct.ContractCreated = block.Time
		}
		err = db.InsertAccount(mAcct, dbTx)
		if err != nil {
			ZapLog.Error("InsertAccount error", zap.Error(err), zap.String("name", mAcct.Name))
			return err
		}
		oldAccounts[arg.AccountName.String()] = struct{}{}
	case types.UpdateAccount:
		arg := payload.(types.UpdateAccountAction)
		d := map[string]interface{}{
			"founder": arg.Founder.String(),
		}
		if _, ok := oldAccounts[action.From.String()]; !ok {
			account, err := db.GetAccountByName(action.From.String(), dbTx)
			if err != nil {
				ZapLog.Error("GetAccountByName error: ", zap.Error(err), zap.String("name", action.From.String()))
				return err
			}
			mOldAccount := &db.MysqlAccountRollback{
				Account: account,
				Height:  block.Number.Uint64() - 1,
			}
			err = db.InsertAccountRollback(mOldAccount, dbTx)
			if err != nil {
				ZapLog.Error("InsertAccountRollback error: ", zap.Error(err))
				return err
			}
		}
		oldAccounts[action.From.String()] = struct{}{}
		err := db.UpdateAccount(action.From.String(), d, dbTx)
		if err != nil {
			ZapLog.Error("UpdateAccount error", zap.Error(err), zap.String("name", action.From.String()))
			return err
		}
	case types.UpdateAccountAuthor:
		acct, err := db.GetAccountByName(action.From.String(), dbTx)
		if err != nil {
			ZapLog.Error("GetAccountByName error", zap.Error(err), zap.String("name", action.From.String()))
			return err
		}
		accountAuthors := make([]*types.AccountAuthor, 0)
		err = json.Unmarshal([]byte(acct.Permissions), &accountAuthors)
		if err != nil {
			ZapLog.Error("UpdateAccountAuthor Unmarshal error", zap.Error(err))
			return err
		}
		arg := payload.(types.AccountAuthorAction)
		f := func(owner types.Owner) types.AuthorType {
			var authorType types.AuthorType
			switch owner.(type) {
			case types.Name:
				authorType = types.AccountNameType
			case types.PubKey:
				authorType = types.PubKeyType
			case types.Address:
				authorType = types.AddressType
			}
			return authorType
		}
		for _, author := range arg.AuthorActions {
			switch author.ActionType {
			case types.AddAuthor:
				flag := true
				for _, v := range accountAuthors {
					if v.Author == author.Author.Owner.String() {
						flag = false
						break
					}
				}
				if flag {
					accountAuthor := &types.AccountAuthor{
						AuthorType: f(author.Author.Owner),
						Author:     author.Author.String(),
						Weight:     author.Author.Weight,
					}
					accountAuthors = append(accountAuthors, accountAuthor)
				}
			case types.DeleteAuthor:
				for i, v := range accountAuthors {
					if v.Author == author.Author.Owner.String() {
						accountAuthors = append(accountAuthors[0:i], accountAuthors[i+1:]...)
						break
					}
				}
			case types.UpdateAuthor:
				for i, v := range accountAuthors {
					if v.Author == author.Author.Owner.String() {
						accountAuthors[i].Weight = author.Author.Weight
						break
					}
				}
			}
		}
		authors := make([]*types.Author, 0)
		for _, author := range accountAuthors {
			at := &types.Author{
				Owner:  types.GenerateOwner(author.Author, author.AuthorType),
				Weight: author.Weight,
			}
			authors = append(authors, at)
		}
		values := make(map[string]interface{})
		if arg.Threshold != 0 {
			values["threshold"] = arg.Threshold
		}
		if arg.UpdateAuthorThreshold != 0 {
			values["update_author_threshold"] = arg.UpdateAuthorThreshold
		}
		authorVersion := types.RlpHash([]interface{}{
			authors,
			arg.Threshold,
			arg.UpdateAuthorThreshold,
		})
		values["author_version"] = authorVersion.String()
		data, err := json.Marshal(accountAuthors)
		if err != nil {
			ZapLog.Error("UpdateAccountAuthor Marshal error: ", zap.Error(err))
			return err
		}
		values["permissions"] = data
		if _, ok := oldAccounts[action.From.String()]; !ok {
			account, err := db.GetAccountByName(action.From.String(), dbTx)
			if err != nil {
				ZapLog.Error("GetAccountByName error: ", zap.Error(err), zap.String("name", action.From.String()))
				return err
			}
			mOldAccount := &db.MysqlAccountRollback{
				Account: account,
				Height:  block.Number.Uint64() - 1,
			}
			err = db.InsertAccountRollback(mOldAccount, dbTx)
			if err != nil {
				ZapLog.Error("InsertAccountRollback error: ", zap.Error(err))
				return err
			}
		}
		oldAccounts[action.From.String()] = struct{}{}
		err = db.UpdateAccount(action.From.String(), values, dbTx)
		if err != nil {
			ZapLog.Error("UpdateAccountAuthor error: ", zap.Error(err))
			return err
		}
	case types.CreateContract:
		if len(action.Payload) != 0 {
			values := map[string]interface{}{
				"contract_code":    hex.EncodeToString(action.Payload),
				"contract_created": block.Time,
			}
			code, err := client.GetCode(action.To.String())
			if err != nil {
				ZapLog.Error("GetCode failed", zap.String("name", action.To.String()), zap.Error(err))
				return err
			}
			values["code_hash"] = crypto.Keccak256Hash(code).String()
			err = db.UpdateAccount(action.To.String(), values, dbTx)
			if err != nil {
				ZapLog.Error("CreateContract failed", zap.Error(err))
				return err
			}
		}
	default:

	}
	return nil
}

func (a *AccountTask) analysisAccount(data *types.BlockAndResult, dbTx *sql.Tx) error {
	oldAccounts := make(map[string]struct{})
	txs := data.Block.Txs
	receipts := data.Receipts
	detailTxs := data.DetailTxs

	if txs == nil {
		return nil
	}

	var containsInternalTxs = true
	if detailTxs == nil || len(detailTxs) == 0 {
		containsInternalTxs = false
	}

	for i, tx := range txs {
		receipt := receipts[i]
		for j, at := range tx.RPCActions {
			actionReceipt := receipt.ActionResults[j]
			if actionReceipt.Status == types.ReceiptStatusSuccessful {
				err := a.ActionToAccount(at, dbTx, data.Block, oldAccounts)
				if err != nil {
					ZapLog.Error("ActionToAccount error: ", zap.Error(err))
					return err
				}
				if containsInternalTxs {
					internalActions := detailTxs[i].InternalActions[j]
					for _, iat := range internalActions.InternalLogs {
						err := a.ActionToAccount(iat.Action, dbTx, data.Block, oldAccounts)
						if err != nil {
							ZapLog.Error("ActionToAccount error: ", zap.Error(err))
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func (a *AccountTask) rollbackAccount(action *types.RPCAction, dbTx *sql.Tx, deleteAccounts map[string]uint64) error {
	switch action.Type {
	case types.CreateAccount:
		payload, err := parsePayload(action)
		if err != nil {
			ZapLog.Error("parsePayload error:", zap.Error(err))
			return err
		}
		arg := payload.(types.CreateAccountAction)
		err = db.DeleteAccountByName(arg.AccountName.String(), dbTx)
		if err != nil {
			ZapLog.Error("DeleteAccountByName error:", zap.Error(err), zap.String("name", arg.AccountName.String()))
			return err
		}
	case types.UpdateAccount:
		oldAccount, err := db.GetOldAccountByName(action.From.String(), dbTx)
		if err != nil {
			ZapLog.Error("GetOldAccountByName error:", zap.Error(err), zap.String("name", action.From.String()))
			return err
		}
		data := map[string]interface{}{
			"founder": oldAccount.Account.Founder,
		}
		err = db.UpdateAccount(action.From.String(), data, dbTx)
		if err != nil {
			ZapLog.Error("UpdateAccount error:", zap.Error(err), zap.String("name", action.From.String()))
			return err
		}
		deleteAccounts[action.From.String()] = oldAccount.Height
	case types.UpdateAccountAuthor:
		oldAccount, err := db.GetOldAccountByName(action.From.String(), dbTx)
		if err != nil {
			ZapLog.Error("GetOldAccountByName error:", zap.Error(err), zap.String("name", action.From.String()))
			return err
		}
		data := map[string]interface{}{
			"founder": oldAccount.Account.Founder,
		}
		err = db.UpdateAccount(action.From.String(), data, dbTx)
		if err != nil {
			ZapLog.Error("UpdateAccount error:", zap.Error(err), zap.String("name", action.From.String()))
			return err
		}
		deleteAccounts[action.From.String()] = oldAccount.Height
	case types.CreateContract:
		data := map[string]interface{}{
			"contract_code":    "",
			"contract_created": 0,
			"code_hash":        "",
		}
		err := db.UpdateAccount(action.From.String(), data, dbTx)
		if err != nil {
			ZapLog.Error("UpdateAccount error:", zap.Error(err), zap.String("name", action.From.String()))
			return err
		}
	}
	return nil
}

func (a *AccountTask) rollback(data *types.BlockAndResult, dbTx *sql.Tx) error {
	txs := data.Block.Txs
	receipts := data.Receipts
	detailTxs := data.DetailTxs
	deleteAccounts := make(map[string]uint64)
	for i, tx := range txs {
		receipt := receipts[i]
		for j, at := range tx.RPCActions {
			actionReceipt := receipt.ActionResults[j]
			if actionReceipt.Status == types.ReceiptStatusSuccessful {
				err := a.rollbackAccount(at, dbTx, deleteAccounts)
				if err != nil {
					ZapLog.Error("rollbackAccount error:", zap.Error(err))
					return err
				}
				if len(detailTxs) != 0 {
					internalActions := detailTxs[i].InternalActions[j]
					for _, iat := range internalActions.InternalLogs {
						err := a.rollbackAccount(iat.Action, dbTx, deleteAccounts)
						if err != nil {
							ZapLog.Error("rollbackAccount error:", zap.Error(err))
							return err
						}
					}
				}
			}
		}
	}
	for name, h := range deleteAccounts {
		err := db.DeleteRollbackAccountByNameAndHeight(name, h, dbTx)
		if err != nil {
			ZapLog.Error("DeleteRollbackAccountByNameAndHeight error:", zap.Error(err), zap.String("name", name))
			return err
		}
	}
	return nil
}

func (a *AccountTask) Start(data chan *TaskChanData, rollbackData chan *TaskChanData, result chan bool, startHeight uint64) {
	a.startHeight = startHeight
	for {
		select {
		case d := <-data:
			if d.Block.Block.Number.Uint64() >= a.startHeight {
				a.init()
				err := a.analysisAccount(d.Block, a.Tx)
				if err != nil {
					ZapLog.Error("AccountTask analysisAccount error: ", zap.Error(err), zap.Uint64("height", d.Block.Block.Number.Uint64()))
					panic(err)
				}
				a.startHeight++
				a.commit()
			}
			result <- true
		case rd := <-rollbackData:
			a.startHeight--
			if a.startHeight == rd.Block.Block.Number.Uint64() {
				a.init()
				err := a.rollback(rd.Block, a.Tx)
				if err != nil {
					ZapLog.Error("AccountTask rollback error: ", zap.Error(err), zap.Uint64("height", rd.Block.Block.Number.Uint64()))
					panic(err)
				}
				a.commit()
			}
			result <- true
		}
	}
}
