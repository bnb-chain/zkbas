package block

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
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
	blocks, err := l.block.GetBlocksList(l.ctx, int64(req.Limit), int64(req.Offset))
	if err != nil {
		logx.Errorf("[GetBlocksList] err: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	total, err := l.block.GetBlocksTotalCount(l.ctx)
	if err != nil {
		return nil, errorcode.AppErrInternal
	}
	resp := &types.RespGetBlocks{
		Total:  uint32(total),
		Blocks: make([]*types.Block, 0),
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
			CommittedAt:                     b.CommittedAt,
			VerifiedTxHash:                  b.VerifiedTxHash,
			VerifiedAt:                      b.VerifiedAt,
			BlockStatus:                     b.BlockStatus,
		}
		for _, t := range b.Txs {
			tx := utils.GormTx2Tx(t)
			block.Txs = append(block.Txs, tx)
		}
		resp.Blocks = append(resp.Blocks, block)
	}
	return resp, nil
}
