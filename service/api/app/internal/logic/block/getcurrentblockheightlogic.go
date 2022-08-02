package block

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetCurrentBlockHeightLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	block  block.Block
}

func NewGetCurrentBlockHeightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentBlockHeightLogic {
	return &GetCurrentBlockHeightLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		block:  block.New(svcCtx),
	}
}

func (l *GetCurrentBlockHeightLogic) GetCurrentBlockHeight() (resp *types.RespCurrentBlockHeight, err error) {
	height, err := l.block.GetCurrentBlockHeight(l.ctx)
	if err != nil {
		logx.Errorf("[GetBlockWithTxsByBlockHeight] err: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	return &types.RespCurrentBlockHeight{
		Height: height,
	}, nil
}
