package statis

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/browser/client"
	"github.com/browser/common"
	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/browser/rlp"
	"github.com/browser/types"
	"go.uber.org/zap"
)

type analysis struct {
	tx          *sql.Tx
	CurBlock    *types.RpcBlock
	Receipts    []*types.Receipt
	DetailTxs   []*types.DetailTx
	nextDay     int64
	contractsST map[string]*db.ContrackStatistics
	tokensST    map[string]*db.TokenStatistics
}

var tokenAssetIDName map[uint64]string
var tokenShortName map[string]string

func NewAnalysis() *analysis {
	tokenAssetIDName = make(map[uint64]string, 0)
	tokenShortName = make(map[string]string, 0)
	return &analysis{
		contractsST: make(map[string]*db.ContrackStatistics, 0),
		tokensST:    make(map[string]*db.TokenStatistics, 0),
	}
}

func (a *analysis) process(data *types.BlockAndResult) error {
	a.tx = db.Mysql.Begin()
	a.prepare(data.Block, data.Receipts, data.DetailTxs)
	err := a.tx.Commit()
	if err != nil {
		ZapLog.Error("mysql tx commit failed")
		return err
	}
	a.tx = nil
	err = a.work()
	if err != nil {
		ZapLog.Panic("analysis block failed rollback", zap.Int64("Number", data.Block.Number.Int64()), zap.String("Hash", data.Block.Hash.String()))
	}

	return err
}

func (a *analysis) prepare(b *types.RpcBlock, r []*types.Receipt, d []*types.DetailTx) {
	a.CurBlock = b
	a.Receipts = r
	a.DetailTxs = d

	if a.CurBlock.Number.Int64() > 0 {
		if a.nextDay == 0 {
			a.prepareNextDayData()
			a.load()
		} else if int64(b.Time) > a.nextDay {
			a.getFeeData()
			a.processDayData()
			a.prepareNextDayData()
		}
	}
}

func (a *analysis) prepareNextDayData() {
	bt := time.Unix(int64(a.CurBlock.Time), 0)
	next := bt.Add(time.Hour)
	next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), 0, 0, 0, next.Location())
	a.nextDay = next.Unix()
}

func (a *analysis) load() {
	tokens, _ := db.LoadTokens()
	for _, token := range tokens {
		t := &db.TokenStatistics{}
		t.User = make(map[string]int, 0)
		t.Holder = make(map[string]int, 0)
		t.User_num = token.User_num
		t.Holder_num = token.Holder_num
		t.Call_num = token.Call_num
		t.FeeTotal = token.FeeTotal
		a.tokensST[token.Token_name] = t
	}

	contracts, _ := db.LoadContracts()
	for _, contract := range contracts {
		c := &db.ContrackStatistics{}
		c.User = make(map[string]int, 0)
		c.User_num = contract.User_num
		c.Call_num = contract.Call_num
		c.FeeTotal = contract.FeeTotal
		a.contractsST[contract.Contract_name] = c
	}

	datafile, err := os.Open("./data.txt")
	if err != nil {
		ZapLog.Info("Open ./data.txt fail", zap.Error(err))
		return
	}

	defer datafile.Close()

	r := bufio.NewReader(datafile)
	for {
		b, err := r.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		array := strings.Split(b, ",")
		tmp := strings.TrimSpace(array[2])
		if array[0] == "t" {
			if array[1] == "u" {
				for _, u := range array[3:] {
					a.tokensST[tmp].User[strings.TrimSpace(u)] = 1
				}
				if uint64(len(a.tokensST[tmp].User)) != a.tokensST[tmp].User_num {
					ZapLog.Panic(fmt.Sprintf("load token %s user:%d user_num:%d", tmp, len(a.tokensST[tmp].User), a.tokensST[tmp].User_num))
				}
			} else if array[1] == "h" {
				for _, u := range array[3:] {
					a.tokensST[tmp].Holder[strings.TrimSpace(u)] = 1
				}
				if uint64(len(a.tokensST[tmp].Holder)) != a.tokensST[tmp].Holder_num {
					ZapLog.Panic(fmt.Sprintf("load token %s holder:%d holder_num:%d", tmp, len(a.tokensST[tmp].Holder), a.tokensST[tmp].Holder_num))
				}
			} else {
				ZapLog.Panic("")
			}
		} else if array[0] == "c" {
			for _, u := range array[3:] {
				a.contractsST[tmp].User[strings.TrimSpace(u)] = 1
			}
			if uint64(len(a.contractsST[tmp].User)) != a.contractsST[tmp].User_num {
				ZapLog.Panic(fmt.Sprintf("load contract %s user:%d user_num:%d", tmp, len(a.contractsST[tmp].User), a.contractsST[tmp].User_num))
			}
		} else {
			ZapLog.Panic("")
		}

	}
}

// const (
// 	AssetGas    = 0
// 	ContractGas = 1
// 	CoinbaseGas = 2
// )
func (a *analysis) getFeeData() {
	start := uint64(1)
	count := uint64(1000)

	tokenFee2 := make(map[string]*big.Int, 0)
	contractFee2 := make(map[string]*big.Int, 0)
	for {
		fees, err := client.GetFeeResultByTime(uint64(a.nextDay), start, count)
		if err != nil {
			ZapLog.Panic("getFeeData rpc error")
		}

		for _, fee := range fees.ObjectFees {
			switch fee.ObjectType {
			case 0:
				if strings.Compare(fee.ObjectName, "libra") == 0 || strings.Compare(fee.ObjectName, "bitcoin") == 0 {
					fee.ObjectName = getFullName(fee.ObjectName)
				}
				tokenFee2[fee.ObjectName] = fee.AssetFees[0].TotalFee
			case 1:
				contractFee2[fee.ObjectName] = fee.AssetFees[0].TotalFee
			}
		}
		if !fees.Continue {
			break
		}
		start += uint64(len(fees.ObjectFees))
	}

	for name, data := range a.tokensST {
		feeTotal2, ok := tokenFee2[name]
		if !ok {
			continue
		}
		data.FeeTotal = big.NewInt(0).Set(feeTotal2)
	}

	for name, data := range a.contractsST {
		feeTotal2, ok := contractFee2[name]
		if !ok {
			continue
		}
		data.FeeTotal = big.NewInt(0).Set(feeTotal2)
	}
}

func (a *analysis) processDayData() {
	//token
	var tokenUserRank []common.Uint64Sort
	var holderCallRank []common.Uint64Sort
	var tokenCallRank []common.Uint64Sort
	var totalRank []common.BigIntSort
	for name, token := range a.tokensST {
		element1 := common.Uint64Sort{
			Name:  name,
			Value: token.User_num,
		}
		tokenUserRank = append(tokenUserRank, element1)

		element4 := common.Uint64Sort{
			Name:  name,
			Value: token.Holder_num,
		}
		holderCallRank = append(holderCallRank, element4)

		element2 := common.Uint64Sort{Name: name,
			Value: token.Call_num,
		}
		tokenCallRank = append(tokenCallRank, element2)

		element3 := common.BigIntSort{
			Name:  "0" + name,
			Value: token.FeeTotal,
		}
		totalRank = append(totalRank, element3)
	}
	common.Uint64SorterProcess(tokenUserRank)
	common.Uint64SorterProcess(tokenCallRank)
	common.Uint64SorterProcess(holderCallRank)

	var rank = 1
	var socre uint64
	var first = true
	for _, e := range tokenUserRank {
		if first {
			socre = e.Value
			first = false
		} else {
			if socre == e.Value {

			} else {
				socre = e.Value
				rank++
			}
		}
		a.tokensST[e.Name].User_rank = rank
	}
	first = true
	rank = 1
	for _, e := range tokenCallRank {
		if first {
			socre = e.Value
			first = false
		} else {
			if socre == e.Value {

			} else {
				socre = e.Value
				rank++
			}
		}
		a.tokensST[e.Name].Call_rank = rank
	}
	first = true
	rank = 1
	for _, e := range holderCallRank {
		if first {
			socre = e.Value
			first = false
		} else {
			if socre == e.Value {

			} else {
				socre = e.Value
				rank++
			}
		}
		a.tokensST[e.Name].Holder_rank = rank
	}
	// contract
	var contractUserRank []common.Uint64Sort
	var contractCallRank []common.Uint64Sort
	for name, contrack := range a.contractsST {
		element1 := common.Uint64Sort{
			Name:  name,
			Value: contrack.User_num,
		}
		contractUserRank = append(contractUserRank, element1)

		element2 := common.Uint64Sort{
			Name:  name,
			Value: contrack.Call_num,
		}
		contractCallRank = append(contractCallRank, element2)

		element3 := common.BigIntSort{
			Name:  "1" + name,
			Value: contrack.FeeTotal,
		}
		totalRank = append(totalRank, element3)
	}
	common.Uint64SorterProcess(contractUserRank)
	common.Uint64SorterProcess(contractCallRank)
	common.BigIntSorterProcess(totalRank)

	first = true
	rank = 1
	for _, e := range contractUserRank {
		if first {
			socre = e.Value
			first = false
		} else {
			if socre == e.Value {

			} else {
				socre = e.Value
				rank++
			}
		}
		a.contractsST[e.Name].User_rank = rank
	}

	first = true
	rank = 1
	for _, e := range contractCallRank {
		if first {
			socre = e.Value
			first = false
		} else {
			if socre == e.Value {

			} else {
				socre = e.Value
				rank++
			}
		}
		a.contractsST[e.Name].Call_rank = rank
	}
	var bigSocre *big.Int
	first = true
	rank = 1
	for _, e := range totalRank {
		if first {
			bigSocre = e.Value
			first = false
		} else {
			if bigSocre.Cmp(e.Value) == 0 {
			} else {
				bigSocre = e.Value
				rank++
			}
		}
		nT := e.Name[0:1]
		name := e.Name[1:len(e.Name)]
		switch nT {
		case "0":
			a.tokensST[name].Income_rank = rank
		case "1":
			a.contractsST[name].Income_rank = rank
		}

	}
	//insert db
	//删除文件
	os.Remove("./data.txt")

	datafile, error := os.Create("./data.txt")
	if error != nil {
		panic("Create data.txt")
	}
	defer datafile.Close()
	for name, t := range a.tokensST {
		if uint64(len(t.User)) != t.User_num {
			panic(fmt.Sprintf("t u ,%s %d %d", name, len(t.User), t.User_num))
		}
		if uint64(len(t.Holder)) != t.Holder_num {
			panic(fmt.Sprintf("t h ,%s %d %d", name, len(t.Holder), t.Holder_num))
		}
		db.InsertToken(name, t, a.tx)

		var users bytes.Buffer
		users.WriteString("t,")
		users.WriteString("u,")
		users.WriteString(name + ",")
		for u := range t.User {
			users.WriteString(u)
			users.WriteString(",")
		}
		if users.Len() > 0 {
			users.Truncate(users.Len() - 1)
			users.WriteString("\n")
		}
		datafile.WriteString(users.String())
		var holders bytes.Buffer
		holders.WriteString("t,")
		holders.WriteString("h,")
		holders.WriteString(name + ",")
		for h := range t.Holder {
			holders.WriteString(h)
			holders.WriteString(",")
		}
		if holders.Len() > 0 {
			holders.Truncate(holders.Len() - 1)
			holders.WriteString("\n")
		}
		datafile.WriteString(holders.String())
		if t.FeeTotal.Int64() != 0 {
			db.ReplaceTotalFee(name, "0", t.FeeTotal, t.Income_rank, a.tx)
		}
	}

	for name, c := range a.contractsST {
		if uint64(len(c.User)) != c.User_num {
			panic(fmt.Sprintf("c h ,%s %d %d", name, len(c.User), c.User_num))
		}
		db.InsertContract(name, c, a.tx)

		var users bytes.Buffer
		users.WriteString("c,")
		users.WriteString("u,")
		users.WriteString(name + ",")
		for u := range c.User {
			users.WriteString(u)
			users.WriteString(",")
		}
		if users.Len() > 0 {
			users.Truncate(users.Len() - 1)
			users.WriteString("\n")
		}
		datafile.WriteString(users.String())
		if c.FeeTotal.Int64() != 0 {
			db.ReplaceTotalFee(name, "1", c.FeeTotal, c.Income_rank, a.tx)
		}
	}
	db.ReplaceBlockInfo(a.CurBlock.Number.Int64() + 1)
}

func (a *analysis) work() error {
	err := a.analysisTxs()
	if err != nil {
		ZapLog.Error("analysisTxs failed", zap.Error(err))
		return err
	}
	return nil
}

func (a *analysis) analysisTxs() error {
	txs := a.CurBlock.Txs
	for i := 0; i < len(txs); i++ {
		receipt := a.Receipts[i]
		state := 1
		for j := 0; j < len(receipt.ActionResults); j++ {
			if receipt.ActionResults[j].Status != uint64(types.ReceiptStatusSuccessful) {
				state = 0
			}
		}

		err := a.analysisActions(i, state)
		if err != nil {
			ZapLog.Error("analysisActions", zap.Error(err))
			return err
		}
	}
	return nil
}

func (a *analysis) analysisActions(txindex int, state int) error {
	tx := a.CurBlock.Txs[txindex]
	actions := tx.RPCActions
	var internalActions *types.DetailTx
	if len(a.DetailTxs) > 0 {
		internalActions = a.DetailTxs[txindex]
	}
	for i := 0; i < len(actions); i++ {
		var internalLog *types.InternalAction
		if internalActions != nil {
			internalLog = internalActions.InternalActions[i]
		}
		action := actions[i]

		err := a.analysisAction(action, state)
		if err != nil {
			ZapLog.Error("analysisActions", zap.Error(err))
			return err
		}
		if internalLog != nil {
			for j := 0; j < len(internalLog.InternalLogs); j++ {
				iLog := internalLog.InternalLogs[j]
				if iLog.Error != "" {
					state = 0
				}
				err = a.analysisInternalAction(iLog, state)
				if err != nil {
					ZapLog.Error("analysisInternalAction", zap.Error(err))
					return err
				}
			}
		}
	}
	return nil
}

func (a *analysis) analysisAction(action *types.RPCAction, state int) error {
	if state == types.ReceiptStatusSuccessful {
		a.analysisAccount(action)
	}
	return nil
}

func (a *analysis) analysisInternalAction(internalLog *types.InternalLog, state int) error {
	if state == types.ReceiptStatusSuccessful {
		a.analysisAccount(internalLog.Action)
	}
	return nil
}

func (a *analysis) tokenProcess(action *types.RPCAction) {
	t := a.getTokenName(action.AssetID)
	token, _ := a.tokensST[t]
	user := action.From.String()
	holder := action.To.String()

	if user == "" || holder == "" {
		return
	}

	if _, ok := token.User[user]; !ok {
		token.User[user] = 1
		token.User_num++
	}

	if _, ok := token.Holder[holder]; !ok {
		token.Holder[holder] = 1
		token.Holder_num++
	}

	token.Call_num++
}

func (a *analysis) contractProcess(action *types.RPCAction) {
	c := action.To.String()
	contract, ok := a.contractsST[c]
	if !ok {
		contract = &db.ContrackStatistics{}
		contract.User = make(map[string]int, 0)
		contract.FeeTotal = big.NewInt(0)
		a.contractsST[c] = contract
	}

	user := action.From.String()
	if _, ok := contract.User[user]; !ok {
		contract.User[user] = 1
		contract.User_num++
	}

	contract.Call_num++
}

var account = "fractal.account"
var asset = "fractal.asset"
var dpos = "fractal.dpos"
var fee = "fractal.fee"
var admin = "fractal.admin"

func (a *analysis) analysisAccount(action *types.RPCAction) {
	if action.Amount.Cmp(types.Big0) > 0 {
		a.tokenProcess(action)
	}
	if account == action.To.String() || asset == action.To.String() || dpos == action.To.String() || fee == action.To.String() || action.To.String() == admin {
		a.contractProcess(action)
	}
	aType := action.Type
	switch aType {
	case types.CallContract:
		if len(action.Payload) != 0 {
			a.contractProcess(action)
		}
	case types.CreateContract:
		a.contractProcess(action)
	case types.IssueAsset:
		var asset types.IssueAssetObject
		err := rlp.DecodeBytes(action.Payload, &asset)
		if err != nil {
			panic(fmt.Sprintf("IssueAsset rlp decode failed"))
		}
		assetInfo, err := client.GetAssetInfoByName(asset.AssetName)
		if err != nil {
			panic(fmt.Sprintf("GetAssetInfoByName err:%s", err))
		}

		assetName := asset.AssetName
		if action.From.String() != "" {
			if strings.Compare(asset.AssetName, "libra") == 0 || strings.Compare(asset.AssetName, "bitcoin") == 0 {
				assetName = action.From.String() + ":" + asset.AssetName
				tokenShortName[asset.AssetName] = assetName
			}
		}
		db.InsertTokenInfo(assetName, assetInfo.Decimals, assetInfo.AssetId, asset.AssetName)
		tokenAssetIDName[assetInfo.AssetId] = assetName

		token, ok := a.tokensST[assetName]
		if !ok {
			token = &db.TokenStatistics{}
			token.User = make(map[string]int, 0)
			token.Holder = make(map[string]int, 0)
			token.FeeTotal = big.NewInt(0)
			a.tokensST[assetName] = token
		}

		holder := asset.Owner.String()
		token.Holder[holder] = 1
		token.Holder_num++
	default:
	}
}

func (a *analysis) getTokenName(assetId uint64) string {
	data, ok := tokenAssetIDName[assetId]
	if !ok {
		_, name, shortName, err := db.GetTokenInfoByAssetID(assetId)
		if err == sql.ErrNoRows {
			ZapLog.Panic("getTokenName null", zap.Uint64("assetId", assetId))
		} else if err != nil {
			ZapLog.Panic(fmt.Sprintf("getTokenName err:%s", err))
		}
		data = name
		tokenAssetIDName[assetId] = name
		if !strings.Contains(shortName, ":") {
			tokenShortName[shortName] = name
		}
	}
	return data
}

func getFullName(shortName string) string {
	name, _ := tokenShortName[shortName]
	return name
}
