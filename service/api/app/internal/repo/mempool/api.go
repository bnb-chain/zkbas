package mempool

import (
	"context"

	mempoolModel "github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type Mempool interface {
	GetMempoolTxs(offset int, limit int) (mempoolTx []*mempoolModel.MempoolTx, err error)
	GetMempoolTxsTotalCount() (count int64, err error)
	GetMempoolTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error)
	GetMempoolTxByTxHash(hash string) (mempoolTxs *mempoolModel.MempoolTx, err error)
	GetMempoolTxByTxId(ctx context.Context, txId int64) (mempoolTxs *mempoolModel.MempoolTx, err error)
}

func New(svcCtx *svc.ServiceContext) Mempool {
	return &model{
		table: `mempool_tx`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
