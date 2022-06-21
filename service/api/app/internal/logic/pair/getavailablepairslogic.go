package pair

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/liquidity"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAvailablePairsLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	liquidity liquidity.Liquidity
}

func NewGetAvailablePairsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAvailablePairsLogic {
	return &GetAvailablePairsLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		liquidity: liquidity.New(svcCtx),
	}
}

func (l *GetAvailablePairsLogic) GetAvailablePairs(req *types.ReqGetAvailablePairs) (*types.RespGetAvailablePairs, error) {
	liquidityAssets, err := l.liquidity.GetAllLiquidityAssets()
	if err != nil {
		logx.Error("[GetAllLiquidityAssets] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetAvailablePairs{}
	for _, asset := range liquidityAssets {
		resp.Pairs = append(resp.Pairs, &types.Pair{
			PairIndex:    uint32(asset.PairIndex),
			AssetAId:     uint32(asset.AssetAId),
			AssetAName:   asset.AssetA,
			AssetBId:     uint32(asset.AssetBId),
			AssetBName:   asset.AssetB,
			FeeRate:      asset.FeeRate,
			TreasuryRate: asset.TreasuryRate,
		})
	}
	return resp, nil
}
