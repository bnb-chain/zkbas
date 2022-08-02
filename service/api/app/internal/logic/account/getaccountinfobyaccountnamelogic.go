package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/checker"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAccountInfoByAccountNameLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
	account   account.Model
}

func NewGetAccountInfoByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByAccountNameLogic {
	return &GetAccountInfoByAccountNameLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx, ctx),
		account:   account.New(svcCtx),
	}
}

func (l *GetAccountInfoByAccountNameLogic) GetAccountInfoByAccountName(req *types.ReqGetAccountInfoByAccountName) (*types.RespGetAccountInfoByAccountName, error) {
	if checker.CheckAccountName(req.AccountName) {
		logx.Errorf("[CheckAccountName] req.AccountName: %s", req.AccountName)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountName")
	}
	accountName := checker.FormatSting(req.AccountName)
	if checker.CheckFormatAccountName(accountName) {
		logx.Errorf("[CheckFormatAccountName] accountName: %s", accountName)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountName")
	}
	info, err := l.account.GetAccountByAccountName(l.ctx, accountName)
	if err != nil {
		logx.Errorf("[GetAccountByAccountName] accountName: %s, err: %s", accountName, err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	account, err := l.globalRPC.GetLatestAccountInfoByAccountIndex(l.ctx, info.AccountIndex)
	if err != nil {
		logx.Errorf("[GetLatestAccountInfoByAccountIndex] err: %s", err.Error())
		if err == errorcode.RpcErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	resp := &types.RespGetAccountInfoByAccountName{
		AccountIndex: uint32(account.AccountIndex),
		AccountPk:    account.PublicKey,
		Nonce:        account.Nonce,
		Assets:       make([]*types.AccountAsset, 0),
	}
	for _, asset := range account.AccountAsset {
		resp.Assets = append(resp.Assets, &types.AccountAsset{
			AssetId:                  asset.AssetId,
			Balance:                  asset.Balance,
			LpAmount:                 asset.LpAmount,
			OfferCanceledOrFinalized: asset.OfferCanceledOrFinalized,
		})
	}
	return resp, nil
}
