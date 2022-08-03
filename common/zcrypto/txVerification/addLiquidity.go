/*
 * Copyright © 2021 Zkbas Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package txVerification

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/util"
)

func VerifyAddLiquidityTxInfo(
	accountInfoMap map[int64]*AccountInfo,
	liquidityInfo *LiquidityInfo,
	txInfo *AddLiquidityTxInfo,
) (txDetails []*MempoolTxDetail, err error) {
	// verify params
	if accountInfoMap[txInfo.FromAccountIndex] == nil ||
		accountInfoMap[liquidityInfo.TreasuryAccountIndex] == nil ||
		accountInfoMap[txInfo.GasAccountIndex] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId].Balance.Cmp(ZeroBigInt) <= 0 ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetBId] == nil ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetBId].Balance.Cmp(ZeroBigInt) <= 0 ||
		accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId] == nil ||
		liquidityInfo == nil ||
		liquidityInfo.AssetAId != txInfo.AssetAId ||
		liquidityInfo.AssetBId != txInfo.AssetBId ||
		txInfo.AssetAAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.AssetBAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.LpAmount.Cmp(ZeroBigInt) < 0 ||
		txInfo.GasFeeAssetAmount.Cmp(ZeroBigInt) < 0 {
		logx.Errorf("invalid params")
		return nil, errors.New("invalid params")
	}
	// verify nonce
	if txInfo.Nonce != accountInfoMap[txInfo.FromAccountIndex].Nonce {
		logx.Errorf("invalid nonce, actual: %d, expected: %d",
			txInfo.Nonce, accountInfoMap[txInfo.FromAccountIndex].Nonce)
		return nil, fmt.Errorf("invalid nonce, actual: %d, expected: %d",
			txInfo.Nonce, accountInfoMap[txInfo.FromAccountIndex].Nonce)
	}
	// add tx info
	var (
		assetDeltaMap             = make(map[int64]map[int64]*big.Int)
		poolDeltaForToAccount     *LiquidityInfo
		lpDeltaForFromAccount     *big.Int
		lpDeltaForTreasuryAccount *big.Int
	)
	// init delta map
	assetDeltaMap[txInfo.FromAccountIndex] = make(map[int64]*big.Int)
	if assetDeltaMap[txInfo.GasAccountIndex] == nil {
		assetDeltaMap[txInfo.GasAccountIndex] = make(map[int64]*big.Int)
	}
	// from account asset A
	assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetAId] = ffmath.Neg(txInfo.AssetAAmount)
	// from account asset B
	assetDeltaMap[txInfo.FromAccountIndex][txInfo.AssetBId] = ffmath.Neg(txInfo.AssetBAmount)
	// from account asset Gas
	if assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] == nil {
		assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] = ffmath.Neg(txInfo.GasFeeAssetAmount)
	} else {
		assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId] = ffmath.Sub(
			assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId],
			txInfo.GasFeeAssetAmount,
		)
	}
	poolAssetADelta := txInfo.AssetAAmount
	poolAssetBDelta := txInfo.AssetBAmount
	// from account lp
	lpDeltaForTreasuryAccount, err = util.ComputeSLp(liquidityInfo.AssetA, liquidityInfo.AssetB, liquidityInfo.KLast, liquidityInfo.FeeRate, liquidityInfo.TreasuryRate)
	if err != nil {
		logx.Errorf("fail to compute liquidity, err: %s", err.Error())
		return nil, errors.New("internal error")
	}
	poolLp := ffmath.Sub(liquidityInfo.LpAmount, lpDeltaForTreasuryAccount)
	// lp = \Delta{x}/x * poolLp
	if liquidityInfo.AssetA.Cmp(ZeroBigInt) == 0 {
		lpDeltaForFromAccount, err = util.CleanPackedAmount(new(big.Int).Sqrt(ffmath.Multiply(txInfo.AssetAAmount, txInfo.AssetBAmount)))
		if err != nil {
			logx.Errorf("unable to compute lp delta: %s", err.Error())
			return nil, errors.New("internal error")
		}
	} else {
		lpDeltaForFromAccount, err = util.CleanPackedAmount(ffmath.Div(ffmath.Multiply(poolAssetADelta, poolLp), liquidityInfo.AssetA))
		if err != nil {
			logx.Errorf("unable to compute lp delta: %s", err.Error())
			return nil, errors.New("internal error")
		}
	}
	// pool account pool info
	finalPoolA := ffmath.Add(liquidityInfo.AssetA, poolAssetADelta)
	finalPoolB := ffmath.Add(liquidityInfo.AssetB, poolAssetBDelta)
	poolDeltaForToAccount = &LiquidityInfo{
		PairIndex:            txInfo.PairIndex,
		AssetAId:             txInfo.AssetAId,
		AssetA:               poolAssetADelta,
		AssetBId:             txInfo.AssetBId,
		AssetB:               poolAssetBDelta,
		LpAmount:             lpDeltaForFromAccount,
		KLast:                ffmath.Multiply(finalPoolA, finalPoolB),
		FeeRate:              liquidityInfo.FeeRate,
		TreasuryAccountIndex: liquidityInfo.TreasuryAccountIndex,
		TreasuryRate:         liquidityInfo.TreasuryRate,
	}
	// set tx info
	txInfo.KLast, err = util.CleanPackedAmount(ffmath.Multiply(finalPoolA, finalPoolB))
	if err != nil {
		return nil, errors.New("internal error")
	}
	txInfo.TreasuryAmount = lpDeltaForTreasuryAccount
	// gas account asset Gas
	if assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] == nil {
		assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] = txInfo.GasFeeAssetAmount
	} else {
		assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId] = ffmath.Add(
			assetDeltaMap[txInfo.GasAccountIndex][txInfo.GasFeeAssetId],
			txInfo.GasFeeAssetAmount,
		)
	}
	// check balance
	// check asset A
	if accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetAId].Balance.Cmp(txInfo.AssetAAmount) < 0 {
		logx.Errorf("not enough balance of asset %d", txInfo.AssetAId)
		return nil, fmt.Errorf("not enough balance of asset %d", txInfo.AssetAId)
	}
	// check asset B
	if accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.AssetBId].Balance.Cmp(txInfo.AssetBAmount) < 0 {
		logx.Errorf("not enough balance of asset %d", txInfo.AssetBId)
		return nil, fmt.Errorf("not enough balance of asset %d", txInfo.AssetAId)
	}
	// asset Gas
	if accountInfoMap[txInfo.FromAccountIndex].AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(
		new(big.Int).Abs(assetDeltaMap[txInfo.FromAccountIndex][txInfo.GasFeeAssetId])) < 0 {
		logx.Errorf("not enough balance of gas")
		return nil, errors.New("not enough balance of gas")
	}
	// check lp amount
	if lpDeltaForFromAccount.Cmp(txInfo.LpAmount) < 0 {
		logx.Errorf("invalid lp amount")
		return nil, errors.New("invalid lp amount")
	}
	// compute hash
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeAddLiquidityMsgHash(txInfo, hFunc)
	if err != nil {
		logx.Errorf("unable to compute tx hash: %s", err.Error())
		return nil, errors.New("internal error")
	}
	// verify signature
	if err := VerifySignature(txInfo.Sig, msgHash, accountInfoMap[txInfo.FromAccountIndex].PublicKey); err != nil {
		return nil, err
	}
	// compute tx details
	// from account asset A
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetAId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetAId, ffmath.Neg(txInfo.AssetAAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// from account asset B
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.AssetBId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.AssetBId, ffmath.Neg(txInfo.AssetBAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// from account asset Gas
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, ffmath.Neg(txInfo.GasFeeAssetAmount), ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// from account lp
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  accountInfoMap[txInfo.FromAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.PairIndex, ZeroBigInt, lpDeltaForFromAccount, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// pool info
	order++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    LiquidityAssetType,
		AccountIndex: commonConstant.NilTxAccountIndex,
		AccountName:  commonConstant.NilAccountName,
		BalanceDelta: poolDeltaForToAccount.String(),
		Order:        order,
		AccountOrder: commonConstant.NilAccountOrder,
	})
	// treasury account
	order++
	accountOrder++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    GeneralAssetType,
		AccountIndex: liquidityInfo.TreasuryAccountIndex,
		AccountName:  accountInfoMap[liquidityInfo.TreasuryAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.PairIndex, ZeroBigInt, lpDeltaForTreasuryAccount, ZeroBigInt,
		).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	// gas account asset Gas
	order++
	accountOrder++
	txDetails = append(txDetails, &MempoolTxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  accountInfoMap[txInfo.GasAccountIndex].AccountName,
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount, ZeroBigInt, ZeroBigInt).String(),
		Order:        order,
		AccountOrder: accountOrder,
	})
	return txDetails, nil
}
