package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type SendTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTxLogic {
	return &SendTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendTxLogic) SendTx(req *types.ReqSendTx) (resp *types.RespSendTx, err error) {
	//err := utils.CheckRequestParam(utils.TypeTxType, reflect.ValueOf(req.TxType))
	txId, err := l.globalRpc.SendTx(l.ctx, req.TxType, req.TxInfo)
	if err != nil {
		logx.Errorf("[transaction.SendTx] err: %s", err.Error())
		return nil, err
	}
	return &types.RespSendTx{TxId: txId}, nil
}
