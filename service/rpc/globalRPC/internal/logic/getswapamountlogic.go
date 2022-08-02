package logic

import (
	"context"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/checker"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

type GetSwapAmountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetSwapAmountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSwapAmountLogic {
	return &GetSwapAmountLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetSwapAmountLogic) GetSwapAmount(in *globalRPCProto.ReqGetSwapAmount) (*globalRPCProto.RespGetSwapAmount, error) {
	if checker.CheckPairIndex(in.PairIndex) {
		logx.Errorf("[CheckPairIndex] Parameter mismatch: %d", in.PairIndex)
		return nil, errorcode.RpcErrInvalidParam
	}
	deltaAmount, isTure := new(big.Int).SetString(in.AssetAmount, 10)
	if !isTure {
		logx.Errorf("[SetString] err, AssetAmount: %s", in.AssetAmount)
		return nil, errorcode.RpcErrInvalidParam
	}

	liquidity, err := l.commglobalmap.GetLatestLiquidityInfoForReadWithCache(l.ctx, int64(in.PairIndex))
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForReadWithCache] err: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.RpcErrNotFound
		}
		return nil, errorcode.RpcErrInternal
	}
	if liquidity.AssetA == nil || liquidity.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidity.AssetB == nil || liquidity.AssetB.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("liquidity: %v, err: %s", liquidity, errorcode.RpcErrLiquidityInvalidAssetAmount.Error())
		return &globalRPCProto.RespGetSwapAmount{}, errorcode.RpcErrLiquidityInvalidAssetAmount
	}

	if int64(in.AssetId) != liquidity.AssetAId && int64(in.AssetId) != liquidity.AssetBId {
		logx.Errorf("input:%v,liquidity: %v, err: %s", in, liquidity, errorcode.RpcErrLiquidityInvalidAssetAmount.Error())
		return &globalRPCProto.RespGetSwapAmount{}, errorcode.RpcErrLiquidityInvalidAssetID
	}
	logx.Errorf("[ComputeDelta] liquidity: %v", liquidity)
	logx.Errorf("[ComputeDelta] in: %v", in)
	logx.Errorf("[ComputeDelta] deltaAmount: %v", deltaAmount)

	var assetAmount *big.Int
	var toAssetId int64
	assetAmount, toAssetId, err = util.ComputeDelta(liquidity.AssetA, liquidity.AssetB, liquidity.AssetAId, liquidity.AssetBId,
		int64(in.AssetId), in.IsFrom, deltaAmount, liquidity.FeeRate)
	if err != nil {
		logx.Errorf("[ComputeDelta] err: %s", err.Error())
		return nil, errorcode.RpcErrInternal
	}
	logx.Errorf("[ComputeDelta] assetAmount:%v", assetAmount)
	return &globalRPCProto.RespGetSwapAmount{
		SwapAssetAmount: assetAmount.String(),
		SwapAssetId:     uint32(toAssetId),
	}, nil
}
