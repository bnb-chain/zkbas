package block

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlocksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	block  block.Block
}

func NewGetBlocksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlocksLogic {
	return &GetBlocksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		block:  block.New(svcCtx),
	}
}

func (l *GetBlocksLogic) GetBlocks(req *types.ReqGetBlocks) (*types.RespGetBlocks, error) {
	blocks, err := l.block.GetBlocksList(int64(req.Limit), int64(req.Offset))
	if err != nil {
		logx.Errorf("[GetBlocksList] err:%v", err)
		return nil, err
	}
	total, err := l.block.GetBlocksTotalCount()
	if err != nil {
		logx.Errorf("[GetBlocksTotalCount] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetBlocks{
		Total: uint32(total),
	}
	for _, b := range blocks {
		block := &types.Block{
			BlockCommitment:                 b.BlockCommitment,
			BlockHeight:                     b.BlockHeight,
			StateRoot:                       b.StateRoot,
			PriorityOperations:              b.PriorityOperations,
			PendingOnChainOperationsHash:    b.PendingOnChainOperationsHash,
			PendingOnChainOperationsPubData: b.PendingOnChainOperationsPubData,
			CommittedTxHash:                 b.CommittedTxHash,
			CommittedAt:                     b.BlockHeight,
			VerifiedTxHash:                  b.VerifiedTxHash,
			VerifiedAt:                      b.BlockHeight,
			BlockStatus:                     b.BlockHeight,
		}
		resp.Blocks = append(resp.Blocks, block)
	}
	return resp, nil
}
