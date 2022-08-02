package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type SendMintNftTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendMintNftTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMintNftTxLogic {
	return &SendMintNftTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendMintNftTxLogic) SendMintNftTx(req *types.ReqSendMintNftTx) (*types.RespSendMintNftTx, error) {
	nftIndex, err := l.globalRpc.SendMintNftTx(l.ctx, req.TxInfo)
	if err != nil {
		logx.Errorf("[transaction.SendMintNftTx] err: %s", err.Error())
		return nil, err
	}
	return &types.RespSendMintNftTx{NftIndex: nftIndex}, nil
}
