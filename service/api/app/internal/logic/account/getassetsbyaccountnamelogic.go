package account

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/errcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
	"github.com/bnb-chain/zkbas/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssetsByAccountNameLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	account   account.AccountModel
	globalRPC globalrpc.GlobalRPC
}

func NewGetAssetsByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetsByAccountNameLogic {
	return &GetAssetsByAccountNameLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		account:   account.New(svcCtx),
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *GetAssetsByAccountNameLogic) GetAssetsByAccountName(req *types.ReqGetAssetsByAccountName) (*types.RespGetAssetsByAccountName, error) {
	resp := &types.RespGetAssetsByAccountName{
		Assets: make([]*types.Asset, 0),
	}
	if utils.CheckAccountName(req.AccountName) {
		logx.Errorf("[CheckAccountName] param:%v", req.AccountName)
		return nil, errcode.ErrInvalidParam
	}
	account, err := l.account.GetAccountByAccountName(l.ctx, req.AccountName)
	if err != nil {
		logx.Errorf("[GetAccountByAccountName] err:%v", err)
		return nil, err
	}
	assets, err := l.globalRPC.GetLatestAccountInfoByAccountIndex(uint32(account.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfoByAccountIndex] err:%v", err)
		return nil, err
	}
	for _, asset := range assets {
		v := &types.Asset{
			AssetId: asset.AssetId,
			Balance: asset.Balance,
		}
		resp.Assets = append(resp.Assets, v)
	}
	return resp, nil
}
