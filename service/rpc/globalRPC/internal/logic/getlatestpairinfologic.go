package logic

import (
	"context"

	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/logic/errcode"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
	"github.com/bnb-chain/zkbas/utils"

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
	if utils.CheckPairIndex(in.PairIndex) {
		logx.Errorf("[CheckPairIndex] param:%v", in.PairIndex)
		return nil, errcode.ErrInvalidParam
	}
	liquidity, err := l.commglobalmap.GetLatestLiquidityInfoForRead(int64(in.PairIndex))
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForRead] err:%v", err)
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
