package mempool

import (
	mempoolModel "github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
)

type Mempool interface {
	GetMempoolTxs(offset int64, limit int64) (mempoolTx []*mempoolModel.MempoolTx, err error)
	GetMempoolTxsTotalCount() (count int64, err error)
	GetMempoolTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error)
	GetMempoolTxByTxHash(hash string) (mempoolTxs *mempoolModel.MempoolTx, err error)
}

func New(svcCtx *svc.ServiceContext) Mempool {
	return &mempool{
		table: `mempool_tx`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
