package types

import (
	"encoding/json"
	"math/big"

	. "github.com/browser/log"
	"go.uber.org/zap"
)

// ChainConfig is the core config which determines the blockchain settings.
type ChainConfig struct {
	BootNodes        []string      `json:"bootnodes,omitempty"` // enode URLs of the P2P bootstrap nodes
	ChainID          *big.Int      `json:"chainId,omitempty"`   // chainId identifies the current chain and is used for replay protection
	ChainName        string        `json:"chainName,omitempty"` // chain name
	ChainURL         string        `json:"chainUrl,omitempty"`  // chain url
	AccountNameCfg   *NameConfig   `json:"accountParams,omitempty"`
	AssetNameCfg     *NameConfig   `json:"assetParams,omitempty"`
	ChargeCfg        *ChargeConfig `json:"chargeParams,omitempty"`
	ForkedCfg        *FrokedConfig `json:"upgradeParams,omitempty"`
	DposCfg          *DposConfig   `json:"dposParams,omitempty"`
	SysName          string        `json:"systemName,omitempty"`  // system name
	AccountName      string        `json:"accountName,omitempty"` // account name
	AssetName        string        `json:"assetName,omitempty"`   // asset name
	DposName         string        `json:"dposName,omitempty"`    // system name
	SnapshotInterval uint64        `json:"snapshotInterval,omitempty"`
	FeeName          string        `json:"feeName,omitempty"`     //fee name
	SysToken         string        `json:"systemToken,omitempty"` // system token
	SysTokenID       uint64        `json:"sysTokenID,omitempty"`
	SysTokenDecimals uint64        `json:"sysTokenDecimal,omitempty"`
	ReferenceTime    uint64        `json:"referenceTime,omitempty"`
}

type NameConfig struct {
	Level     uint64 `json:"level,omitempty"`
	Length    uint64 `json:"length,omitempty"`
	SubLength uint64 `json:"subLength,omitempty"`
}

type ChargeConfig struct {
	AssetRatio    uint64 `json:"assetRatio,omitempty"`
	ContractRatio uint64 `json:"contractRatio,omitempty"`
}

type FrokedConfig struct {
	ForkBlockNum   uint64 `json:"blockCnt,omitempty"`
	Forkpercentage uint64 `json:"upgradeRatio,omitempty"`
}

type DposConfig struct {
	MaxURLLen             uint64   `json:"maxURLLen,omitempty"`            // url length
	UnitStake             *big.Int `json:"unitStake,omitempty"`            // state unit
	CandidateMinQuantity  *big.Int `json:"candidateMinQuantity,omitempty"` // min quantity
	VoterMinQuantity      *big.Int `json:"voterMinQuantity,omitempty"`     // min quantity
	ActivatedMinQuantity  *big.Int `json:"activatedMinQuantity,omitempty"` // min active quantity
	BlockInterval         uint64   `json:"blockInterval,omitempty"`
	BlockFrequency        uint64   `json:"blockFrequency,omitempty"`
	CandidateScheduleSize uint64   `json:"candidateScheduleSize,omitempty"`
	BackupScheduleSize    uint64   `json:"backupScheduleSize,omitempty"`
	EpchoInterval         uint64   `json:"epchoInterval,omitempty"`
	FreezeEpchoSize       uint64   `json:"freezeEpchoSize,omitempty"`
	ExtraBlockReward      *big.Int `json:"extraBlockReward,omitempty"`
	BlockReward           *big.Int `json:"blockReward,omitempty"`
}

// Genesis specifies the header fields, state of a genesis block.
type Genesis struct {
	Config *ChainConfig `json:"config"`
	//Dpos          *DposConfig       `json:"dpos"`
	Timestamp uint64 `json:"timestamp"`
	//ExtraData     []byte            `json:"extraData"`
	GasLimit   uint64   `json:"gasLimit" `
	Difficulty *big.Int `json:"difficulty" `
	//Coinbase      Name              `json:"coinbase"`
	//AllocAccounts []*GenesisAccount `json:"allocAccounts"`
	//AllocAssets   []*AssetObject    `json:"allocAssets"`
	AllocAccounts   []*GenesisAccount   `json:"allocAccounts,omitempty"`
	AllocCandidates []*GenesisCandidate `json:"allocCandidates,omitempty"`
	AllocAssets     []*GenesisAsset     `json:"allocAssets,omitempty"`
}

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Name   Name   `json:"name,omitempty"`
	PubKey PubKey `json:"pubKey,omitempty"`
}

// GenesisCandidate is an cadicate in the state of the genesis block.
type GenesisCandidate struct {
	Name  string   `json:"name,omitempty"`
	URL   string   `json:"url,omitempty"`
	Stake *big.Int `json:"stake,omitempty"`
}

// GenesisAsset is an asset in the state of the genesis block.
type GenesisAsset struct {
	Name       string   `json:"name,omitempty"`
	Symbol     string   `json:"symbol,omitempty"`
	Amount     *big.Int `json:"amount,omitempty"`
	Decimals   uint64   `json:"decimals,omitempty"`
	Founder    string   `json:"founder,omitempty"`
	Owner      string   `json:"owner,omitempty"`
	UpperLimit *big.Int `json:"upperLimit,omitempty"`
}

func (g *Genesis) UnmarshalJSON(input []byte) {
	var dec Genesis
	if err := json.Unmarshal(input, &dec); err != nil {
		ZapLog.Panic("Genesis Unmarshal failed", zap.Error(err))
	}
	if dec.Config != nil {
		g.Config = dec.Config
	}
	if len(dec.AllocAccounts) > 0 {
		g.AllocAccounts = dec.AllocAccounts
	}

	if len(dec.AllocAssets) > 0 {
		g.AllocAssets = dec.AllocAssets
	}
}
