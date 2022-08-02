package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type SendWithdrawNftTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendWithdrawNftTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendWithdrawNftTxLogic {
	return &SendWithdrawNftTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendWithdrawNftTxLogic) SendWithdrawNftTx(req *types.ReqSendWithdrawNftTx) (*types.RespSendWithdrawNftTx, error) {
	txIndex, err := l.globalRpc.SendWithdrawNftTx(l.ctx, req.TxInfo)
	if err != nil {
		logx.Errorf("[transaction.SendWithdrawNftTx] err: %s", err.Error())
		return nil, err
	}
	return &types.RespSendWithdrawNftTx{TxId: txIndex}, nil
}
