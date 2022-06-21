package block

import (
	table "github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type Block interface {
	GetBlockByBlockHeight(blockHeight int64) (block *table.Block, err error)
}

func New(svcCtx *svc.ServiceContext) Block {
	return &block{
		table: `block`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
