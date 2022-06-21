package logic

import (
	"context"

	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/logic/errcode"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/liquidity"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/mempool"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
	"github.com/bnb-chain/zkbas/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestAccountLpLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	liquidity     liquidity.Liquidity
	mempool       mempool.Mempool
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetLatestAccountLpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestAccountLpLogic {
	return &GetLatestAccountLpLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		liquidity:     liquidity.New(svcCtx),
		mempool:       mempool.New(svcCtx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetLatestAccountLpLogic) GetLatestAccountLp(in *globalRPCProto.ReqGetLatestAccountLp) (*globalRPCProto.RespGetLatestAccountLp, error) {
	if utils.CheckAccountIndex(in.AccountIndex) {
		logx.Errorf("[CheckAccountIndex] param:%v", in.AccountIndex)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckPairIndex(in.PairIndex) {
		logx.Errorf("[CheckPairIndex] param:%v", in.PairIndex)
		return nil, errcode.ErrInvalidParam
	}
	accountInfo, err := l.commglobalmap.GetLatestAccountInfo(int64(in.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfo] err:%v", err)
		return nil, err
	}
	return &globalRPCProto.RespGetLatestAccountLp{
		LpAmount: accountInfo.AssetInfo[int64(in.PairIndex)].LpAmount.String(),
	}, nil
}
