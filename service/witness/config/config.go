package config

import (
	"encoding/json"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/tree"
)

const (
	WitnessAppId    = "zkbnb-witness"
	SystemConfigKey = "SystemConfig"
	Namespace       = "application"
)

type TreeDB struct {
	Driver tree.Driver
	//nolint:staticcheck
	LevelDBOption tree.LevelDBOption `json:",optional"`
	//nolint:staticcheck
	RedisDBOption tree.RedisDBOption `json:",optional"`
	//nolint:staticcheck
	RoutinePoolSize    int `json:",optional"`
	AssetTreeCacheSize int
}

type Config struct {
	Postgres       apollo.Postgres
	TreeDB         TreeDB
	LogConf        logx.LogConf
	EnableRollback bool
	DbRoutineSize  int `json:",optional"`
	DbBatchSize    int
}

func InitSystemConfiguration(config *Config, configFile string) error {
	if err := InitSystemConfigFromEnvironment(config); err != nil {
		logx.Errorf("Init system configuration from environment raise error: %v", err)
	} else {
		logx.Infof("Init system configuration from environment Successfully")
		return nil
	}
	if err := InitSystemConfigFromConfigFile(config, configFile); err != nil {
		logx.Errorf("Init system configuration from config file raise error: %v", err)
		panic("Init system configuration from config file raise error:" + err.Error())
	} else {
		logx.Infof("Init system configuration from config file Successfully")
		return nil
	}
}

func InitSystemConfigFromEnvironment(c *Config) error {
	commonConfig, err := apollo.InitCommonConfig(WitnessAppId)
	if err != nil {
		return err
	}
	c.Postgres = commonConfig.Postgres
	c.EnableRollback = commonConfig.EnableRollback

	systemConfigString, err := apollo.LoadApolloConfigFromEnvironment(WitnessAppId, Namespace, SystemConfigKey)
	if err != nil {
		return err
	}

	systemConfig := &Config{}
	err = json.Unmarshal([]byte(systemConfigString), systemConfig)
	if err != nil {
		return err
	}

	c.TreeDB = systemConfig.TreeDB
	c.LogConf = systemConfig.LogConf
	c.DbBatchSize = systemConfig.DbBatchSize
	c.DbRoutineSize = systemConfig.DbRoutineSize
	return nil
}

func InitSystemConfigFromConfigFile(c *Config, configFile string) error {
	conf.Load(configFile, c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	return nil
}
