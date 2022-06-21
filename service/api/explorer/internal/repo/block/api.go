package block

import (
	table "github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
)

type Block interface {
	GetCommitedBlocksCount() (count int64, err error)
	GetExecutedBlocksCount() (count int64, err error)
	GetBlockWithTxsByCommitment(BlockCommitment string) (block *table.Block, err error)
	GetBlockByBlockHeight(blockHeight int64) (block *table.Block, err error)
	GetBlockWithTxsByBlockHeight(blockHeight int64) (block *table.Block, err error)
	GetBlocksList(limit int64, offset int64) (blocks []*table.Block, err error)
	GetBlocksTotalCount() (count int64, err error)
}

func New(svcCtx *svc.ServiceContext) Block {
	return &block{
		table:     `block`,
		db:        svcCtx.GormPointer,
		cache:     svcCtx.Cache,
		redisConn: svcCtx.RedisConn,
	}
}
