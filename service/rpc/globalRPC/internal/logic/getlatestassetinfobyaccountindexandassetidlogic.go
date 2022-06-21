package logic

import (
	"context"

	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/logic/errcode"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/l2asset"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/mempool"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"

	"github.com/bnb-chain/zkbas/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestAssetInfoByAccountIndexAndAssetIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	account       account.AccountModel
	l2asset       l2asset.L2asset
	mempool       mempool.Mempool
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetLatestAssetInfoByAccountIndexAndAssetIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestAssetInfoByAccountIndexAndAssetIdLogic {

	return &GetLatestAssetInfoByAccountIndexAndAssetIdLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		account:       account.New(svcCtx),
		l2asset:       l2asset.New(svcCtx),
		mempool:       mempool.New(svcCtx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetLatestAssetInfoByAccountIndexAndAssetIdLogic) GetLatestAssetInfoByAccountIndexAndAssetId(in *globalRPCProto.ReqGetLatestAssetInfoByAccountIndexAndAssetId) (*globalRPCProto.RespGetLatestAssetInfoByAccountIndexAndAssetId, error) {
	if utils.CheckAccountIndex(in.AccountIndex) {
		logx.Errorf("[CheckAccountIndex] param:%v", in.AccountIndex)
		return nil, errcode.ErrInvalidParam
	}
	accountInfo, err := l.commglobalmap.GetLatestAccountInfo(int64(in.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfo] err:%v", err)
		return nil, err
	}
	return &globalRPCProto.RespGetLatestAssetInfoByAccountIndexAndAssetId{
		Balance: accountInfo.AssetInfo[int64(in.AssetId)].Balance.String(),
	}, nil
}
