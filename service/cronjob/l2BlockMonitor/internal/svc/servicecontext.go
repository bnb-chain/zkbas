package svc

import (
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/l1TxSender"
	"github.com/bnb-chain/zkbas/common/model/l2BlockEventMonitor"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/service/cronjob/l2BlockMonitor/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config              config.Config
	Mempool             mempool.MempoolModel
	Block               block.BlockModel
	L2BlockEventMonitor l2BlockEventMonitor.L2BlockEventMonitorModel
	L1TxSender          l1TxSender.L1TxSenderModel
	SysConfig           sysconfig.SysconfigModel
	AccountModel        account.AccountModel
	AccountHistoryModel account.AccountHistoryModel
	NftModel            nft.L2NftModel
	NftHistoryModel     nft.L2NftHistoryModel
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
		Config:              c,
		Mempool:             mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),
		Block:               block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		L2BlockEventMonitor: l2BlockEventMonitor.NewL2BlockEventMonitorModel(conn, c.CacheRedis, gormPointer),
		L1TxSender:          l1TxSender.NewL1TxSenderModel(conn, c.CacheRedis, gormPointer),
		SysConfig:           sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
		AccountModel:        account.NewAccountModel(conn, c.CacheRedis, gormPointer),
		AccountHistoryModel: account.NewAccountHistoryModel(conn, c.CacheRedis, gormPointer),
		NftModel:            nft.NewL2NftModel(conn, c.CacheRedis, gormPointer),
		NftHistoryModel:     nft.NewL2NftHistoryModel(conn, c.CacheRedis, gormPointer),
	}
}
