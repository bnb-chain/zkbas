package svc

import (
	"github.com/bnb-chain/zkbas/common/model/account"
	asset "github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/model/l2TxEventMonitor"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/service/cronjob/mempoolMonitor/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config                config.Config
	L2TxEventMonitorModel l2TxEventMonitor.L2TxEventMonitorModel
	L2assetInfoModel      asset.AssetInfoModel
	AccountModel          account.AccountModel
	AccountHistoryModel   account.AccountHistoryModel
	MempoolModel          mempool.MempoolModel
	LiquidityModel        liquidity.LiquidityModel
	NftModel              nft.L2NftModel
	MempoolTxDetailModel  mempool.MempoolTxDetailModel
	RedisConnection       *redis.Redis
	DbEngine              *gorm.DB
}

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

func NewServiceContext(c config.Config) *ServiceContext {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))
	return &ServiceContext{
		Config:                c,
		L2TxEventMonitorModel: l2TxEventMonitor.NewL2TxEventMonitorModel(conn, c.CacheRedis, gormPointer),
		L2assetInfoModel:      asset.NewAssetInfoModel(conn, c.CacheRedis, gormPointer),
		AccountModel:          account.NewAccountModel(conn, c.CacheRedis, gormPointer),
		AccountHistoryModel:   account.NewAccountHistoryModel(conn, c.CacheRedis, gormPointer),
		MempoolModel:          mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),
		LiquidityModel:        liquidity.NewLiquidityModel(conn, c.CacheRedis, gormPointer),
		NftModel:              nft.NewL2NftModel(conn, c.CacheRedis, gormPointer),
		MempoolTxDetailModel:  mempool.NewMempoolDetailModel(conn, c.CacheRedis, gormPointer),
		RedisConnection:       redisConn,
	}
}
