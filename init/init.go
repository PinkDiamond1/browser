package init

import (
	"fmt"
	"strconv"

	"github.com/browser/client"
	"github.com/browser/config"
	"github.com/browser/db"
	. "github.com/browser/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	readConfig()
	watchConfig()
	InitLog()
	initChainConfig()
	db.InitDb()
}

func readConfig() {
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	config.Mysql = &config.MysqlConfig{
		Username: Get("browser.mysql.user"),
		Password: Get("browser.mysql.password"),
		Ip:       Get("browser.mysql.ip"),
		Port:     Get("browser.mysql.port"),
		Database: Get("browser.mysql.database"),
	}

	config.Node = &config.NodeConfig{
		RpcUrl: Get("browser.node.rpcHost"),
	}

	logFile := &config.LogFileConfig{
		Enable:       viper.GetBool("browser.log.file.enable"),
		Path:         viper.GetString("browser.log.file.path"),
		RotationTime: viper.GetInt("browser.log.file.rotationTime"),
		MaxAge:       viper.GetInt64("browser.log.file.maxAge"),
	}

	config.Log = &config.LogConfig{
		Level:               viper.GetString("browser.log.level"),
		FileConfig:          logFile,
		Console:             viper.GetBool("browser.log.console"),
		SyncBlockShowNumber: viper.GetInt64("browser.log.syncBlockShowNumber"),
	}

	tasks := viper.Get("browser.tasks").([]interface{})
	for _, task := range tasks {
		config.Tasks = append(config.Tasks, task.(string))
	}

	config.BlockDataChanBufferSize = viper.GetInt("browser.blockDataChanBufferSize")
}

func Get(key string) string {
	if value, ok := viper.Get(key).(int); ok {
		return strconv.Itoa(value)
	} else if value, ok := viper.Get(key).(string); ok {
		return value
	} else {
		return ""
	}
}

func watchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		ZapLog.Info("config file changed", zap.String("name", in.Name))
		readConfig()
	})
}

func initChainConfig() {
	chainConfig, err := client.GetChainConfig()
	if err != nil {
		ZapLog.Error("GetChainConfig error: ", zap.Error(err))
		panic(err)
	}
	block, err := client.GetBlockByNumber(1)
	if err != nil {
		ZapLog.Error("GetBlockTimeByNumber error: ", zap.Error(err))
		panic(err)
	}
	sTime := block.Time / 1000000000
	config.Chain = &config.ChainConfig{
		FeeAssetId:            uint64(0),
		ChainName:             chainConfig.ChainName,
		SysName:               chainConfig.SysName,
		ChainAssetName:        chainConfig.AssetName,
		ChainFeeName:          chainConfig.FeeName,
		ChainDposName:         chainConfig.DposName,
		ChainAccountName:      chainConfig.AccountName,
		CandidateScheduleSize: chainConfig.DposCfg.CandidateScheduleSize,
		BlockFrequency:        chainConfig.DposCfg.BlockFrequency,
		ChainId:               chainConfig.ChainID.Uint64(),
		StartTime:             sTime,
	}
}
