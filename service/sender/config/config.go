package config

import (
	"github.com/zeromicro/go-zero/core/logx"
)

type Config struct {
	Postgres struct {
		MasterDataSource string
		SlaveDataSource  string
	}
	ChainConfig struct {
		NetworkRPCSysConfigName string
		MaxWaitingTime          int64
		MaxBlockCount           int
		ConfirmBlocksCount      uint64
		CommitAddress           string
		VerifyAddress           string

		GasLimit uint64
		GasPrice uint64
		//((MaxGasPrice-GasPrice)/GasPrice)*100
		MaxGasPriceIncreasePercentage uint64 `json:",optional"`
	}
	KMSConfig struct {
		KMSKeyId string
		ChainId  int64
	}
	LogConf logx.LogConf
}
