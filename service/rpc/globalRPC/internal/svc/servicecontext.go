package svc

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/account"
	asset "github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/pkg/multcache"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/config"
)

type ServiceContext struct {
	Config                config.Config
	MempoolModel          mempool.MempoolModel
	MempoolDetailModel    mempool.MempoolTxDetailModel
	AccountModel          account.AccountModel
	AccountHistoryModel   account.AccountHistoryModel
	TxModel               tx.TxModel
	TxDetailModel         tx.TxDetailModel
	FailTxModel           tx.FailTxModel
	LiquidityModel        liquidity.LiquidityModel
	LiquidityHistoryModel liquidity.LiquidityHistoryModel
	BlockModel            block.BlockModel

	NftModel        nft.L2NftModel
	CollectionModel nft.L2NftCollectionModel
	OfferModel      nft.OfferModel

	L2AssetModel asset.AssetInfoModel

	SysConfigModel sysconfig.SysconfigModel

	RedisConnection *redis.Redis

	GormPointer *gorm.DB

	Conn  sqlx.SqlConn
	Cache multcache.MultCache
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
		logx.Must(err)
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))
	return &ServiceContext{
		Config:                c,
		MempoolModel:          mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),
		MempoolDetailModel:    mempool.NewMempoolDetailModel(conn, c.CacheRedis, gormPointer),
		AccountModel:          account.NewAccountModel(conn, c.CacheRedis, gormPointer),
		AccountHistoryModel:   account.NewAccountHistoryModel(conn, c.CacheRedis, gormPointer),
		TxModel:               tx.NewTxModel(conn, c.CacheRedis, gormPointer, redisConn),
		TxDetailModel:         tx.NewTxDetailModel(conn, c.CacheRedis, gormPointer),
		FailTxModel:           tx.NewFailTxModel(conn, c.CacheRedis, gormPointer),
		LiquidityModel:        liquidity.NewLiquidityModel(conn, c.CacheRedis, gormPointer),
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(conn, c.CacheRedis, gormPointer),
		BlockModel:            block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		NftModel:              nft.NewL2NftModel(conn, c.CacheRedis, gormPointer),
		CollectionModel:       nft.NewL2NftCollectionModel(conn, c.CacheRedis, gormPointer),
		OfferModel:            nft.NewOfferModel(conn, c.CacheRedis, gormPointer),
		L2AssetModel:          asset.NewAssetInfoModel(conn, c.CacheRedis, gormPointer),
		SysConfigModel:        sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
		RedisConnection:       redisConn,
		GormPointer:           gormPointer,
		Conn:                  conn,
		Cache:                 multcache.NewRedisCache(c.CacheRedis[0].Host, c.CacheRedis[0].Pass, 10),
	}
}
