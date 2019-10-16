package config

var Mysql *MysqlConfig

type MysqlConfig struct {
	Username string
	Password string
	Ip       string
	Port     string
	Database string
}

var Node *NodeConfig

type NodeConfig struct {
	RpcUrl string
}

var Log *LogConfig

type LogConfig struct {
	Level               string
	FileConfig          *LogFileConfig
	Console             bool
	SyncBlockShowNumber int64
}

type LogFileConfig struct {
	Enable       bool
	Path         string
	RotationTime int
	MaxAge       int64
}

var Tasks []string
var BlockDataChanBufferSize int

var Chain *ChainConfig

type ChainConfig struct {
	FeeAssetId            uint64
	ChainName             string
	SysName               string
	ChainAssetName        string
	ChainFeeName          string
	ChainDposName         string
	ChainAccountName      string
	CandidateScheduleSize uint64
	BlockFrequency        uint64
	ChainId               uint64
	StartTime             int64
}
