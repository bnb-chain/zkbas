package svc

import (
	"time"

	"github.com/bnb-chain/zkbas/service/api/app/internal/fetcher/price"
	"github.com/bnb-chain/zkbas/service/api/app/internal/fetcher/state"

	gocache "github.com/patrickmn/go-cache"
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
	"github.com/bnb-chain/zkbas/service/api/app/internal/config"
)

const cacheDefaultExpiration = time.Millisecond * 500
const cacheDefaultPurgeInterval = time.Second * 60

type ServiceContext struct {
	Config        config.Config
	Conn          sqlx.SqlConn
	GormPointer   *gorm.DB
	RedisConn     *redis.Redis
	Cache         *gocache.Cache
	CodeVersion   string
	GitCommitHash string

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
	NftModel              nft.L2NftModel
	CollectionModel       nft.L2NftCollectionModel
	OfferModel            nft.OfferModel
	L2AssetModel          asset.AssetInfoModel
	SysConfigModel        sysconfig.SysconfigModel

	PriceFetcher price.Fetcher
	StateFetcher state.Fetcher
}

func NewServiceContext(c config.Config) *ServiceContext {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Must(err)
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
	redisConn := redis.New(c.CacheRedis[0].Host, func(p *redis.Redis) {
		p.Type = c.CacheRedis[0].Type
		p.Pass = c.CacheRedis[0].Pass
	})
	cache := gocache.New(cacheDefaultExpiration, cacheDefaultPurgeInterval)
	mempoolModel := mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer)
	mempoolDetailModel := mempool.NewMempoolDetailModel(conn, c.CacheRedis, gormPointer)
	accountModel := account.NewAccountModel(conn, c.CacheRedis, gormPointer)
	liquidityModel := liquidity.NewLiquidityModel(conn, c.CacheRedis, gormPointer)
	nftModel := nft.NewL2NftModel(conn, c.CacheRedis, gormPointer)
	offerModel := nft.NewOfferModel(conn, c.CacheRedis, gormPointer)
	return &ServiceContext{
		Config:                c,
		Conn:                  conn,
		GormPointer:           gormPointer,
		RedisConn:             redisConn,
		Cache:                 cache,
		MempoolModel:          mempoolModel,
		MempoolDetailModel:    mempoolDetailModel,
		AccountModel:          accountModel,
		AccountHistoryModel:   account.NewAccountHistoryModel(conn, c.CacheRedis, gormPointer),
		TxModel:               tx.NewTxModel(conn, c.CacheRedis, gormPointer),
		TxDetailModel:         tx.NewTxDetailModel(conn, c.CacheRedis, gormPointer),
		FailTxModel:           tx.NewFailTxModel(conn, c.CacheRedis, gormPointer),
		LiquidityModel:        liquidityModel,
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(conn, c.CacheRedis, gormPointer),
		BlockModel:            block.NewBlockModel(conn, c.CacheRedis, gormPointer),
		NftModel:              nftModel,
		CollectionModel:       nft.NewL2NftCollectionModel(conn, c.CacheRedis, gormPointer),
		OfferModel:            offerModel,
		L2AssetModel:          asset.NewAssetInfoModel(conn, c.CacheRedis, gormPointer),
		SysConfigModel:        sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),

		PriceFetcher: price.NewFetcher(cache),
		StateFetcher: state.NewFetcher(redisConn, mempoolModel, mempoolDetailModel, accountModel,
			liquidityModel, nftModel, offerModel),
	}
}
