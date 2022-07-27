package logic

import (
	"context"
	"github.com/bnb-chain/zkbas/common/checker"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestPairInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetLatestPairInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestPairInfoLogic {
	return &GetLatestPairInfoLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetLatestPairInfoLogic) GetLatestPairInfo(in *globalRPCProto.ReqGetLatestPairInfo) (*globalRPCProto.RespGetLatestPairInfo, error) {
	if checker.CheckPairIndex(in.PairIndex) {
		logx.Errorf("[CheckPairIndex] param:%v", in.PairIndex)
		return nil, errorcode.GlobalRpcInvalidParam
	}
	liquidity, err := l.commglobalmap.GetLatestLiquidityInfoForReadWithCache(l.ctx, int64(in.PairIndex))
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForReadWithCache] err:%v", err)
		return nil, err
	}
	return &globalRPCProto.RespGetLatestPairInfo{
		AssetAAmount: liquidity.AssetA.String(),
		AssetAId:     uint32(liquidity.AssetAId),
		AssetBAmount: liquidity.AssetB.String(),
		AssetBId:     uint32(liquidity.AssetBId),
		LpAmount:     liquidity.LpAmount.String(),
	}, nil
}
