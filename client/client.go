package client

import (
	"encoding/json"
	"errors"
	. "github.com/browser/log"
	"github.com/browser/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.uber.org/zap"
	"math/big"
)

const (
	methodCurrentBlock           = "ft_getCurrentBlock"
	methodBlockByNumber          = "ft_getBlockByNumber"
	methodChainConfig            = "ft_getChainConfig"
	methodBlockAndResultByNumber = "ft_getBlockAndResultByNumber"
	methodSendRawTransaction     = "ft_sendRawTransaction"
	methodAssetInfoByName        = "account_getAssetInfoByName"
	methodAssetInfoByID          = "account_getAssetInfoByID"
	methodGetAccountByName       = "account_getAccountByName"
	methodGetCode                = "account_getCode"
	methodDposCadidatesSize      = "dpos_candidatesSize"
	methodDposIrreversible       = "dpos_irreversible"
)

func GetCurrentBlockInfo() (int64, string, error) {
	request := NewRPCRequest("2.0", methodCurrentBlock, false)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("GetCurrentBlockInfo SendRPCRequest error", zap.Error(err))
		return 0, "", err
	}
	number, _ := big.NewInt(0).SetString(string(jsonParsed.Path("result.number").Data().(json.Number)), 10)
	miner := jsonParsed.Path("result.miner").Data().(string)
	return number.Int64(), miner, nil
}

func GetAssetInfoByName(assetName string) (*types.AssetObject, error) {
	request := NewRPCRequest("2.0", methodAssetInfoByName, assetName)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("GetAssetInfoByName SendRPCRequest error", zap.Error(err), zap.String("name", assetName))
		return nil, err
	}
	if result := jsonParsed.Path("result").Data(); result == nil {
		return nil, nil
	}
	mapAsset := jsonParsed.S("result").ChildrenMap()
	asset := &types.AssetObject{}
	assetId, _ := big.NewInt(0).SetString(string(mapAsset["assetId"].Data().(json.Number)), 10)
	asset.AssetId = assetId.Uint64()
	asset.AssetName = mapAsset["assetName"].Data().(string)
	asset.Symbol = mapAsset["symbol"].Data().(string)
	asset.Amount, _ = big.NewInt(0).SetString(string(mapAsset["amount"].Data().(json.Number)), 10)
	decimals, _ := big.NewInt(0).SetString(string(mapAsset["decimals"].Data().(json.Number)), 10)
	asset.Decimals = decimals.Uint64()
	asset.Founder = types.Name(mapAsset["founder"].Data().(string))
	asset.Owner = types.Name(mapAsset["owner"].Data().(string))
	asset.UpperLimit, _ = big.NewInt(0).SetString(string(mapAsset["upperLimit"].Data().(json.Number)), 10)
	return asset, nil
}

func GetAssetInfoById(assetid uint64) (*types.AssetObject, error) {
	request := NewRPCRequest("2.0", methodAssetInfoByID, assetid)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("GetAssetInfoById SendRPCRequest error", zap.Error(err), zap.Uint64("assetId", assetid))
		return nil, err
	}
	if result := jsonParsed.Path("result").Data(); result == nil {
		return nil, nil
	}
	mapAsset := jsonParsed.S("result").ChildrenMap()
	asset := &types.AssetObject{}
	assetId, _ := big.NewInt(0).SetString(string(mapAsset["assetId"].Data().(json.Number)), 10)
	asset.AssetId = assetId.Uint64()
	asset.AssetName = mapAsset["assetName"].Data().(string)
	asset.Symbol = mapAsset["symbol"].Data().(string)
	asset.Amount, _ = big.NewInt(0).SetString(string(mapAsset["amount"].Data().(json.Number)), 10)
	decimals, _ := big.NewInt(0).SetString(string(mapAsset["decimals"].Data().(json.Number)), 10)
	asset.Decimals = decimals.Uint64()
	asset.Founder = types.Name(mapAsset["founder"].Data().(string))
	asset.Owner = types.Name(mapAsset["owner"].Data().(string))
	asset.UpperLimit, _ = big.NewInt(0).SetString(string(mapAsset["upperLimit"].Data().(json.Number)), 10)
	return asset, nil
}

func GetBlockTimeByNumber(number int64) (int64, error) {
	request := NewRPCRequest("2.0", methodBlockByNumber, number, true)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("GetBlockByNumber SendRPCRequest error", zap.Error(err), zap.Int64("height", number))
		return 0, err
	}
	value, _ := big.NewInt(0).SetString(string(jsonParsed.Path("result.timestamp").Data().(json.Number)), 10)
	return value.Int64() / 1000000000, nil
}

func GetCandidatesCount() (int64, error) {
	request := NewRPCRequest("2.0", methodDposCadidatesSize, 0)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("GetCandidatesCount SendRPCRequest error", zap.Error(err))
		return 0, err
	}
	registerCnt, _ := big.NewInt(0).SetString(string(jsonParsed.Path("result").Data().(json.Number)), 10)
	return registerCnt.Int64(), nil
}

//GetBlockAndResult get block and produce log
func GetBlockAndResult(number int64) (*types.BlockAndResult, error) {
	request := NewRPCRequest("2.0", methodBlockAndResultByNumber, number)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("GetBlockAndResult SendRPCRequest error", zap.Error(err), zap.Int64("height", number))
		return nil, err
	}
	if result := jsonParsed.Path("result").Data(); result == nil {
		return nil, nil
	}
	ret := &types.BlockAndResult{}
	//-----------------Block---------------
	HeadMap := jsonParsed.S("result", "block").ChildrenMap()
	block := &types.Block{}
	ret.Hash = types.HexToHash(HeadMap["hash"].Data().(string))
	head := &types.Header{}
	head.ParentHash = types.HexToHash(HeadMap["parentHash"].Data().(string))
	head.Coinbase = types.Name(HeadMap["miner"].Data().(string))
	head.Difficulty, _ = big.NewInt(0).SetString(string(HeadMap["difficulty"].Data().(json.Number)), 10)
	head.Number, _ = big.NewInt(0).SetString(string(HeadMap["number"].Data().(json.Number)), 10)
	gasLimit, _ := big.NewInt(0).SetString(string(HeadMap["gasLimit"].Data().(json.Number)), 10)
	head.GasLimit = gasLimit.Uint64()
	gasUsed, _ := big.NewInt(0).SetString(string(HeadMap["gasUsed"].Data().(json.Number)), 10)
	head.GasUsed = gasUsed.Uint64()
	time, _ := big.NewInt(0).SetString(string(HeadMap["timestamp"].Data().(json.Number)), 10)
	head.Time = uint(time.Uint64() / 1000000000)
	head.Extra, _ = hexutil.Decode(HeadMap["extraData"].Data().(string))

	txs := make([]*types.RPCTransaction, 0)
	txChildren := jsonParsed.S("result", "block", "transactions").Children()
	for _, txChild := range txChildren {
		tx := &types.RPCTransaction{}
		gasAssetId, _ := big.NewInt(0).SetString(string(txChild.Path("gasAssetID").Data().(json.Number)), 10)
		tx.GasAssetID = gasAssetId.Uint64()
		tx.GasPrice, _ = big.NewInt(0).SetString(string(txChild.Path("gasPrice").Data().(json.Number)), 10)
		tx.Hash = types.HexToHash(txChild.Path("txHash").Data().(string))
		tx.GasCost, _ = big.NewInt(0).SetString(string(txChild.Path("gasCost").Data().(json.Number)), 10)
		tx.RPCActions = make([]*types.RPCAction, 0)
		acChildren := txChild.S("actions").Children()
		for _, acChild := range acChildren {
			ac := &types.RPCAction{}
			tp, _ := big.NewInt(0).SetString(string(acChild.Path("type").Data().(json.Number)), 10)
			ac.Type = types.ActionType(tp.Uint64())
			nonce, _ := big.NewInt(0).SetString(string(acChild.Path("nonce").Data().(json.Number)), 10)
			ac.Nonce = nonce.Uint64()
			assetId, _ := big.NewInt(0).SetString(string(acChild.Path("assetID").Data().(json.Number)), 10)
			ac.AssetID = assetId.Uint64()
			ac.From = types.Name(acChild.Path("from").Data().(string))
			ac.To = types.Name(acChild.Path("to").Data().(string))
			gas, _ := big.NewInt(0).SetString(string(acChild.Path("gas").Data().(json.Number)), 10)
			ac.GasLimit = gas.Uint64()
			ac.Amount, _ = big.NewInt(0).SetString(string(acChild.Path("value").Data().(json.Number)), 10)
			ac.Remark, _ = hexutil.Decode(acChild.Path("remark").Data().(string))
			ac.Payload, _ = hexutil.Decode(acChild.Path("payload").Data().(string))
			ac.ActionHash = types.HexToHash(acChild.Path("actionHash").Data().(string))
			tx.RPCActions = append(tx.RPCActions, ac)
		}
		txs = append(txs, tx)
	}
	block.Head = head
	block.Txs = txs
	ret.Block = block
	//-----------------Receipt---------------
	receipts := make([]*types.Receipt, 0)
	receiptsChildren := jsonParsed.S("result", "receipts").Children()
	for _, receiptsChild := range receiptsChildren {
		receipt := &types.Receipt{}
		CumulativeGasUsed, _ := big.NewInt(0).SetString(string(receiptsChild.Path("CumulativeGasUsed").Data().(json.Number)), 10)
		receipt.CumulativeGasUsed = CumulativeGasUsed.Uint64()
		receipt.TxHash = types.HexToHash(receiptsChild.Path("TxHash").Data().(string))
		TotalGasUsed, _ := big.NewInt(0).SetString(string(receiptsChild.Path("TotalGasUsed").Data().(json.Number)), 10)
		receipt.TotalGasUsed = TotalGasUsed.Uint64()
		receipt.ActionResults = make([]*types.ActionResult, 0)
		reChildren := receiptsChild.S("ActionResults").Children()
		for _, reChild := range reChildren {
			actionResult := &types.ActionResult{}
			Status, _ := big.NewInt(0).SetString(string(reChild.Path("Status").Data().(json.Number)), 10)
			actionResult.Status = Status.Uint64()
			GasUsed, _ := big.NewInt(0).SetString(string(reChild.Path("GasUsed").Data().(json.Number)), 10)
			actionResult.GasUsed = GasUsed.Uint64()
			actionResult.Error = reChild.Path("Error").Data().(string)
			actionResult.GasAllot = make([]*types.GasDistribution, 0)
			gasAllotChildren := reChild.S("GasAllot").Children()
			for _, gasAllotChild := range gasAllotChildren {
				gasDistribution := &types.GasDistribution{}
				gasDistribution.Account = types.Name(gasAllotChild.Path("name").Data().(string))
				gas, _ := big.NewInt(0).SetString(string(gasAllotChild.Path("gas").Data().(json.Number)), 10)
				gasDistribution.Gas = gas.Uint64()
				typeId, _ := big.NewInt(0).SetString(string(gasAllotChild.Path("typeId").Data().(json.Number)), 10)
				gasDistribution.Reason = typeId.Uint64()
				actionResult.GasAllot = append(actionResult.GasAllot, gasDistribution)
			}
			receipt.ActionResults = append(receipt.ActionResults, actionResult)
		}
		receipts = append(receipts, receipt)
	}
	ret.Receipts = receipts
	//-----------------DetailTx---------------
	detailTxs := make([]*types.DetailTx, 0)
	detailTxsChildren := jsonParsed.S("result", "detailTxs").Children()
	for _, detailTxsChild := range detailTxsChildren {
		detailTx := &types.DetailTx{}
		detailTx.TxHash = types.HexToHash(detailTxsChild.Path("txhash").Data().(string))
		detailTx.InternalActions = make([]*types.InternalAction, 0)
		iTxsChildren := detailTxsChild.S("actions").Children()
		for _, iTxsChild := range iTxsChildren {
			internalTxs := &types.InternalAction{}
			internalTxs.InternalLogs = make([]*types.InternalLog, 0)
			iTxChildren := iTxsChild.S("internalActions").Children()
			for _, iTxChild := range iTxChildren {
				itx := &types.InternalLog{}
				itx.ActionType = iTxChild.Path("actionType").Data().(string)
				gasUsed, _ := big.NewInt(0).SetString(string(iTxChild.Path("gasUsed").Data().(json.Number)), 10)
				itx.GasUsed = gasUsed.Uint64()
				gasLimit, _ := big.NewInt(0).SetString(string(iTxChild.Path("gasLimit").Data().(json.Number)), 10)
				itx.GasLimit = gasLimit.Uint64()
				depth, _ := big.NewInt(0).SetString(string(iTxChild.Path("depth").Data().(json.Number)), 10)
				itx.Depth = depth.Uint64()
				itx.Error = iTxChild.Path("error").Data().(string)
				actionMap := iTxChild.S("action").ChildrenMap()
				rpcAction := &types.RPCAction{}
				tp, _ := big.NewInt(0).SetString(string(actionMap["type"].Data().(json.Number)), 10)
				rpcAction.Type = types.ActionType(tp.Uint64())
				nonce, _ := big.NewInt(0).SetString(string(actionMap["nonce"].Data().(json.Number)), 10)
				rpcAction.Nonce = nonce.Uint64()
				rpcAction.From = types.Name(actionMap["from"].Data().(string))
				rpcAction.To = types.Name(actionMap["to"].Data().(string))
				assetID, _ := big.NewInt(0).SetString(string(actionMap["assetID"].Data().(json.Number)), 10)
				rpcAction.AssetID = assetID.Uint64()
				gas, _ := big.NewInt(0).SetString(string(actionMap["gas"].Data().(json.Number)), 10)
				rpcAction.GasLimit = gas.Uint64()
				rpcAction.Amount, _ = big.NewInt(0).SetString(string(actionMap["value"].Data().(json.Number)), 10)
				rpcAction.Payload, _ = hexutil.Decode(actionMap["payload"].Data().(string))
				itx.Action = rpcAction
				internalTxs.InternalLogs = append(internalTxs.InternalLogs, itx)
			}
			detailTx.InternalActions = append(detailTx.InternalActions, internalTxs)
		}
		detailTxs = append(detailTxs, detailTx)
	}
	ret.DetailTxs = detailTxs
	return ret, nil
}

func GetChainConfig() (*types.ChainConfig, error) {
	request := NewRPCRequest("2.0", methodChainConfig, false)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("GetChainConfig SendRPCRequest error", zap.Error(err))
		return nil, err
	}
	if result := jsonParsed.Path("result").Data(); result == nil {
		return nil, err
	}
	chainConfig := &types.ChainConfig{}
	chainConfig.BootNodes = make([]string, 0)
	chainConfigMap := jsonParsed.S("result").ChildrenMap()
	chainConfig.ChainID, _ = big.NewInt(0).SetString(string(chainConfigMap["chainId"].Data().(json.Number)), 10)
	chainConfig.ChainName = chainConfigMap["chainName"].Data().(string)
	chainConfig.AssetName = chainConfigMap["assetName"].Data().(string)
	chainConfig.SysName = chainConfigMap["systemName"].Data().(string)
	chainConfig.AccountName = chainConfigMap["accountName"].Data().(string)
	chainConfig.DposName = chainConfigMap["dposName"].Data().(string)
	chainConfig.FeeName = chainConfigMap["feeName"].Data().(string)
	chainConfig.SysToken = chainConfigMap["systemToken"].Data().(string)
	dPosCfg := &types.DposConfig{}
	dPosConfigMap := jsonParsed.S("result", "dposParams").ChildrenMap()
	tmp, _ := big.NewInt(0).SetString(string(dPosConfigMap["blockInterval"].Data().(json.Number)), 10)
	dPosCfg.BlockInterval = tmp.Uint64()
	tmp, _ = tmp.SetString(string(dPosConfigMap["blockFrequency"].Data().(json.Number)), 10)
	dPosCfg.BlockFrequency = tmp.Uint64()
	tmp, _ = tmp.SetString(string(dPosConfigMap["candidateScheduleSize"].Data().(json.Number)), 10)
	dPosCfg.CandidateScheduleSize = tmp.Uint64()
	chainConfig.DposCfg = dPosCfg
	return chainConfig, nil
}

func GetAccountByName(name string) (*types.Account, error) {
	request := NewRPCRequest("2.0", methodGetAccountByName, name)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("GetAccountByName SendRPCRequest error", zap.Error(err), zap.String("name", name))
		return nil, err
	}
	if result := jsonParsed.Path("result").Data(); result == nil {
		return nil, err
	}
	account := &types.Account{
		AcctName: types.Name(name),
	}
	accountInfoMap := jsonParsed.S("result").ChildrenMap()
	founder := accountInfoMap["founder"].Data().(string)
	account.Founder = types.Name(founder)
	accountId, _ := big.NewInt(0).SetString(string(accountInfoMap["accountID"].Data().(json.Number)), 10)
	account.AccountID = accountId.Uint64()
	number, _ := big.NewInt(0).SetString(string(accountInfoMap["number"].Data().(json.Number)), 10)
	account.Number = number.Uint64()
	nonce, _ := big.NewInt(0).SetString(string(accountInfoMap["nonce"].Data().(json.Number)), 10)
	account.Nonce = nonce.Uint64()
	account.Code = make([]byte, 0)
	code := accountInfoMap["code"].Data().(string)
	account.Code = append(account.Code, []byte(code)...)
	account.CodeHash = types.HexToHash(accountInfoMap["codeHash"].Data().(string))
	authorVersion := accountInfoMap["authorVersion"].Data().(string)
	account.AuthorVersion.SetBytes([]byte(authorVersion))
	codeSize, _ := big.NewInt(0).SetString(string(accountInfoMap["codeSize"].Data().(json.Number)), 10)
	account.CodeSize = codeSize.Uint64()
	threshold, _ := big.NewInt(0).SetString(string(accountInfoMap["threshold"].Data().(json.Number)), 10)
	account.Threshold = threshold.Uint64()
	return account, nil
}

func SendRawTransaction(rawData string) (types.Hash, error) {
	nullHash := types.Hash{}
	request := NewRPCRequest("2.0", methodSendRawTransaction, rawData)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("SendRawTransaction SendRPCRequest error", zap.Error(err), zap.String("rawData", rawData))
		return nullHash, err
	}
	var resErr error
	if result := jsonParsed.Path("error").Data(); result != nil {
		errMap := jsonParsed.S("error").ChildrenMap()
		errMsg := errMap["message"].Data().(string)
		resErr = errors.New(errMsg)
		return nullHash, resErr
	}
	if result := jsonParsed.Path("result").Data(); result == nil {
		return nullHash, errors.New("get hash failed")
	} else {
		hash := types.HexToHash(result.(string))
		return hash, nil
	}
}

func GetDposIrreversible() (*types.DposIrreversible, error) {
	request := NewRPCRequest("2.0", methodDposIrreversible)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("GetDposIrreversible SendRPCRequest error", zap.Error(err))
		return nil, err
	}
	if result := jsonParsed.Path("result").Data(); result == nil {
		ZapLog.Error("GetDposIrreversible response to json error", zap.Error(err))
		return nil, err
	}
	response := &types.DposIrreversible{}
	dPosIrreversibleMap := jsonParsed.S("result").ChildrenMap()
	reversible, _ := big.NewInt(0).SetString(string(dPosIrreversibleMap["reversible"].Data().(json.Number)), 10)
	response.Reversible = reversible.Uint64()
	proposedIrreversible, _ := big.NewInt(0).SetString(string(dPosIrreversibleMap["proposedIrreversible"].Data().(json.Number)), 10)
	response.ProposedIrreversible = proposedIrreversible.Uint64()
	bftIrreversible, _ := big.NewInt(0).SetString(string(dPosIrreversibleMap["bftIrreversible"].Data().(json.Number)), 10)
	response.BftIrreversible = bftIrreversible.Uint64()
	return response, nil
}

func GetCode(name string) (string, error) {
	request := NewRPCRequest("2.0", methodGetCode, name)
	jsonParsed, err := SendRPCRequest(request)
	if err != nil {
		ZapLog.Error("GetDposIrreversible SendRPCRequest error", zap.Error(err), zap.String("name", name))
		return "", err
	}
	if result := jsonParsed.Path("result").Data(); result == nil {
		return "", err
	}
	code := jsonParsed.S("result").Data().(string)
	return code, nil
}
