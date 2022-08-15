package sendrawtx

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/service/api/app/internal/fetcher/state"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

func SendAddLiquidityTx(ctx context.Context, svcCtx *svc.ServiceContext, stateFetcher state.Fetcher, rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseAddLiquidityTxInfo(rawTxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return "", errorcode.AppErrInvalidTx
	}

	if err := legendTxTypes.ValidateAddLiquidityTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError(err)
	}

	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, svcCtx.SysConfigModel); err != nil {
		return txId, err
	}

	liquidityInfo, err := stateFetcher.GetLatestLiquidityInfo(ctx, txInfo.PairIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid PairIndex")
		}
		logx.Errorf("fail to get liquidity info: %d, err: %s", txInfo.PairIndex, err.Error())
		return "", err
	}
	if liquidityInfo.AssetA == nil || liquidityInfo.AssetB == nil {
		logx.Errorf("invalid liquidity assets")
		return "", errorcode.AppErrInternal
	}
	if liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 {
		txInfo.LpAmount, err = util.ComputeEmptyLpAmount(txInfo.AssetAAmount, txInfo.AssetBAmount)
		if err != nil {
			logx.Errorf("cannot computer lp amount, err: %s", err.Error())
			return "", errorcode.AppErrInternal
		}
	} else {
		txInfo.LpAmount, err = util.ComputeLpAmount(liquidityInfo, txInfo.AssetAAmount)
		if err != nil {
			logx.Errorf("cannot computer lp amount, err: %s", err.Error())
			return "", errorcode.AppErrInternal
		}
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.FromAccountIndex], err = stateFetcher.GetLatestAccountInfo(ctx, txInfo.FromAccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.FromAccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = stateFetcher.GetBasicAccountInfo(ctx, txInfo.GasAccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return txId, errorcode.AppErrInvalidTxField.RefineError("invalid GasAccountIndex")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.GasAccountIndex, err.Error())
			return "", errorcode.AppErrInternal
		}
	}
	if accountInfoMap[liquidityInfo.TreasuryAccountIndex] == nil {
		accountInfoMap[liquidityInfo.TreasuryAccountIndex], err = stateFetcher.GetBasicAccountInfo(ctx, liquidityInfo.TreasuryAccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return txId, errorcode.AppErrInvalidTxField.RefineError("invalid liquidity")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", liquidityInfo.TreasuryAccountIndex, err.Error())
			return "", errorcode.AppErrInternal
		}
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify tx
	txDetails, err = txVerification.VerifyAddLiquidityTxInfo(
		accountInfoMap,
		liquidityInfo,
		txInfo)
	if err != nil {
		return "", errorcode.AppErrVerification.RefineError(err)
	}

	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return "", errorcode.AppErrInternal
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeAddLiquidity,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		txInfo.PairIndex,
		commonConstant.NilAssetId,
		txInfo.LpAmount.String(),
		"",
		string(txInfoBytes),
		"",
		txInfo.FromAccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	if err := svcCtx.MempoolModel.CreateBatchedMempoolTxs([]*mempool.MempoolTx{mempoolTx}); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		_ = CreateFailTx(svcCtx.FailTxModel, commonTx.TxTypeAddLiquidity, txInfo, err)
		return "", err
	}
	return txId, nil
}
