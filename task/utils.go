package task

import (
	"bytes"
	. "github.com/browser_service/log"
	"github.com/browser_service/rlp"
	"github.com/browser_service/types"
	"go.uber.org/zap"
	"math/big"
)

func parsePayload(action *types.RPCAction) (interface{}, error) {
	if bytes.Compare(action.Payload, []byte{}) == 0 {
		return "", nil
	}
	aType := action.Type
	var parsedPayload interface{}
	switch aType {
	case types.CreateAccount:
		var arg types.CreateAccountAction
		err := rlp.DecodeBytes(action.Payload, &arg)
		if err != nil {
			ZapLog.Error("DecodeBytes CreateAccount payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	case types.UpdateAccount:
		var arg types.UpdateAccountAction
		err := rlp.DecodeBytes(action.Payload, &arg)
		if err != nil {
			ZapLog.Error("DecodeBytes UpdateAccount payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	case types.UpdateAccountAuthor:
		var arg types.AccountAuthorAction
		err := rlp.DecodeBytes(action.Payload, &arg)
		if err != nil {
			ZapLog.Error("DecodeBytes UpdateAccountAuthor payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	case types.IssueAsset:
		var arg types.IssueAssetObject
		err := rlp.DecodeBytes(action.Payload, &arg)
		if err != nil {
			ZapLog.Error("DecodeBytes IssueAsset payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	case types.IncreaseAsset:
		var arg types.IncAssetObject
		err := rlp.DecodeBytes(action.Payload, &arg)
		if err != nil {
			ZapLog.Error("DecodeBytes IncreaseAsset payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	case types.UpdateAsset:
		var arg types.UpdateAssetObject
		err := rlp.DecodeBytes(action.Payload, &arg)
		if err != nil {
			ZapLog.Error("DecodeBytes UpdateAsset payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	case types.UpdateAssetContract:
		var arg types.UpdateAssetContractObject
		err := rlp.DecodeBytes(action.Payload, &arg)
		if err != nil {
			ZapLog.Error("DecodeBytes UpdateAssetContract payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	case types.SetAssetOwner:
		var arg types.UpdateAssetOwnerObject
		err := rlp.DecodeBytes(action.Payload, &arg)
		if err != nil {
			ZapLog.Error("DecodeBytes SetAssetOwner payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	case types.WithdrawFee:
		var arg types.WithdrawInfo
		err := rlp.DecodeBytes(action.Payload, &arg)
		if err != nil {
			ZapLog.Error("DecodeBytes WithdrawFee payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	case types.RegCandidate:
		var arg types.DposRegisterCandidate
		err := rlp.DecodeBytes(action.Payload, &arg)
		if err != nil {
			ZapLog.Error("DecodeBytes RegCandidate payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	case types.UpdateCandidate:
		arg := types.DposUpdateCandidate{}
		if err := rlp.DecodeBytes(action.Payload, &arg); err != nil {
			ZapLog.Error("DecodeBytes UpdateCandidate payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	case types.VoteCandidate:
		arg := types.DposVoteCandidate{}
		if err := rlp.DecodeBytes(action.Payload, &arg); err != nil {
			ZapLog.Error("DecodeBytes VoteCandidate payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	case types.KickedCandidate:
		arg := types.DposKickedCandidate{}
		if err := rlp.DecodeBytes(action.Payload, &arg); err != nil {
			ZapLog.Error("DecodeBytes KickedCandidate payload failed, error: ", zap.Error(err))
			return nil, err
		}
		parsedPayload = arg
	default:
		return action.Payload, nil
	}
	if action.Type == types.IssueAsset {
		issueAssetPayload := parsedPayload.(types.IssueAssetObject)
		//if strings.Compare(issueAssetPayload.AssetName, "libra") == 0 || strings.Compare(issueAssetPayload.AssetName, "bitcoin") == 0 {
		tmpPayload := types.IssueAssetObject{
			//AssetName:   action.From.String() + ":" + issueAssetPayload.AssetName,
			AssetName:   issueAssetPayload.AssetName,
			Symbol:      issueAssetPayload.Symbol,
			Amount:      big.NewInt(0).Set(issueAssetPayload.Amount),
			Decimals:    issueAssetPayload.Decimals,
			Founder:     issueAssetPayload.Founder,
			Owner:       issueAssetPayload.Owner,
			UpperLimit:  big.NewInt(0).Set(issueAssetPayload.UpperLimit),
			Contract:    issueAssetPayload.Contract,
			Description: issueAssetPayload.Description,
		}
		parsedPayload = tmpPayload
		//}
	}
	return parsedPayload, nil
}
