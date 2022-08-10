package nft

import (
	"context"

	nftModel "github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type Nft interface {
	GetNftListByAccountIndex(ctx context.Context, accountIndex, limit, offset int64) (nfts []*nftModel.L2Nft, err error)
	GetAccountNftTotalCount(ctx context.Context, accountIndex int64) (int64, error)
}

func New(svcCtx *svc.ServiceContext) Nft {
	return &nft{
		table: nftModel.L2NftTableName,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}