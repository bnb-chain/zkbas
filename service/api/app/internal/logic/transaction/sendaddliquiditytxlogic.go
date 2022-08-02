package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type SendAddLiquidityTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendAddLiquidityTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendAddLiquidityTxLogic {
	return &SendAddLiquidityTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendAddLiquidityTxLogic) SendAddLiquidityTx(req *types.ReqSendAddLiquidityTx) (*types.RespSendAddLiquidityTx, error) {
	txIndex, err := l.globalRpc.SendAddLiquidityTx(l.ctx, req.TxInfo)
	if err != nil {
		logx.Errorf("[transaction.SendAddLiquidityTx] err: %s", err.Error())
		return nil, err
	}
	return &types.RespSendAddLiquidityTx{TxId: txIndex}, nil
}
