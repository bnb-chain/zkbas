package sendrawtx

import (
	"context"
	"time"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type addLiquidityTxSender struct {
	txType          int
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	gasChecker      GasChecker
	nonceChecker    NonceChecker
	mempoolTxSender MempoolTxSender
}

func NewAddLiquidityTxSender(ctx context.Context, svcCtx *svc.ServiceContext,
	gasChecker *gasChecker, nonceChecker *nonceChecker, sender *mempoolTxSender) *addLiquidityTxSender {
	return &addLiquidityTxSender{
		txType:          commonTx.TxTypeAddLiquidity,
		ctx:             ctx,
		svcCtx:          svcCtx,
		gasChecker:      gasChecker,
		nonceChecker:    nonceChecker,
		mempoolTxSender: sender,
	}
}

func (s *addLiquidityTxSender) SendTx(rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseAddLiquidityTxInfo(rawTxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return "", errorcode.AppErrInvalidTx
	}

	if err := legendTxTypes.ValidateAddLiquidityTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError(err)
	}

	//check expire time
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("invalid ExpiredAt")
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid ExpiredAt")
	}

	//TODO: check signature

	liquidity, err := s.svcCtx.StateFetcher.GetLatestLiquidity(s.ctx, txInfo.PairIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid PairIndex")
		}
		logx.Errorf("fail to get liquidity info: %d, err: %s", txInfo.PairIndex, err.Error())
		return "", err
	}
	if liquidity.AssetA == nil || liquidity.AssetB == nil {
		logx.Errorf("invalid liquidity assets")
		return "", errorcode.AppErrInternal
	}

	//check from account
	fromAccount, err := s.svcCtx.StateFetcher.GetLatestAccount(s.ctx, txInfo.FromAccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.FromAccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}

	assetA, ok := fromAccount.AssetInfo[liquidity.AssetAId]
	if !ok || assetA.Balance.Cmp(txInfo.AssetAAmount) < 0 {
		logx.Errorf("not enough assetA in balance: %d, err: %s", fromAccount.AccountIndex, err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError("not enough assetA in balance")
	}

	assetB, ok := fromAccount.AssetInfo[liquidity.AssetBId]
	if !ok || assetB.Balance.Cmp(txInfo.AssetBAmount) < 0 {
		logx.Errorf("not enough assetB in balance: %d, err: %s", fromAccount.AccountIndex, err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError("not enough assetB in balance")
	}

	//check gas
	if err := s.gasChecker.CheckGas(fromAccount, txInfo.GasAccountIndex, txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount); err != nil {
		return "", errorcode.AppErrInvalidTxField.RefineError(err.Error())
	}

	//check nonce
	if err := s.nonceChecker.CheckNonce(fromAccount, txInfo.Nonce); err != nil {
		return "", errorcode.AppErrInvalidTxField.RefineError(err.Error())
	}

	//send mempool tx
	mempoolTx := &mempool.MempoolTx{
		TxType:        int64(s.txType),
		GasFeeAssetId: txInfo.GasFeeAssetId,
		GasFee:        txInfo.GasFeeAssetAmount.String(),
		NftIndex:      commonConstant.NilTxNftIndex,
		PairIndex:     txInfo.PairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      txInfo.LpAmount.String(),
		Memo:          "",
		AccountIndex:  txInfo.FromAccountIndex,
		Nonce:         txInfo.Nonce,
		ExpiredAt:     txInfo.ExpiredAt,
	}
	txId, err = s.mempoolTxSender.SendMempoolTx(func(txInfo interface{}) ([]byte, error) {
		return legendTxTypes.ComputeAddLiquidityMsgHash(txInfo.(*legendTxTypes.AddLiquidityTxInfo), mimc.NewMiMC())
	}, txInfo, mempoolTx)

	return txId, err
}
