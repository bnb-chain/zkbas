package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/core/executor"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type SendTxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTxLogic {
	return &SendTxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (s *SendTxLogic) SendTx(req *types.ReqSendTx) (resp *types.TxHash, err error) {
	resp = &types.TxHash{}
	bc := core.NewBlockChainForDryRun(s.svcCtx.AccountModel, s.svcCtx.LiquidityModel, s.svcCtx.NftModel, s.svcCtx.MempoolModel,
		s.svcCtx.RedisCache)
	newTx := &tx.Tx{
		TxHash: types2.EmptyTxHash, // Would be computed in prepare method of executors.
		TxType: int64(req.TxType),
		TxInfo: req.TxInfo,

		GasFeeAssetId: types2.NilAssetId,
		GasFee:        types2.NilAssetAmount,
		PairIndex:     types2.NilPairIndex,
		NftIndex:      types2.NilNftIndex,
		CollectionId:  types2.NilCollectionNonce,
		AssetId:       types2.NilAssetId,
		TxAmount:      types2.NilAssetAmount,
		NativeAddress: types2.EmptyL1Address,

		BlockHeight: types2.NilBlockHeight,
		TxStatus:    tx.StatusPending,
	}

	err = bc.ApplyTransaction(newTx)
	if err != nil {
		return resp, err
	}
	if err := s.svcCtx.MempoolModel.CreateMempoolTxs([]*tx.Tx{newTx}); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", newTx, err.Error())
		return resp, types2.AppErrInternal
	}

	resp.TxHash = newTx.TxHash
	return resp, nil
}

func (s *SendTxLogic) getExecutor(txType int, txInfo string) (executor.TxExecutor, error) {
	bc := core.NewBlockChainForDryRun(s.svcCtx.AccountModel, s.svcCtx.LiquidityModel, s.svcCtx.NftModel,
		s.svcCtx.MempoolModel, s.svcCtx.AssetModel, s.svcCtx.SysConfigModel, s.svcCtx.RedisCache)
	t := &tx.Tx{TxType: int64(txType), TxInfo: txInfo}

	switch txType {
	case types2.TxTypeTransfer:
		return executor.NewTransferExecutor(bc, t)
	case types2.TxTypeSwap:
		return executor.NewSwapExecutor(bc, t)
	case types2.TxTypeAddLiquidity:
		return executor.NewAddLiquidityExecutor(bc, t)
	case types2.TxTypeRemoveLiquidity:
		return executor.NewRemoveLiquidityExecutor(bc, t)
	case types2.TxTypeWithdraw:
		return executor.NewWithdrawExecutor(bc, t)
	case types2.TxTypeTransferNft:
		return executor.NewTransferNftExecutor(bc, t)
	case types2.TxTypeAtomicMatch:
		return executor.NewAtomicMatchExecutor(bc, t)
	case types2.TxTypeCancelOffer:
		return executor.NewCancelOfferExecutor(bc, t)
	case types2.TxTypeWithdrawNft:
		return executor.NewWithdrawNftExecutor(bc, t)
	case types2.TxTypeCreateCollection:
		return executor.NewCreateCollectionExecutor(bc, t)
	case types2.TxTypeMintNft:
		return executor.NewMintNftExecutor(bc, t)
	default:
		logx.Errorf("invalid tx type: %v", txType)
		return nil, types2.AppErrInvalidTxType
	}
}
