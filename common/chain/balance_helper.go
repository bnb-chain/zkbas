/*
 * Copyright © 2021 ZkBAS Protocol
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

package chain

import (
	"errors"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas/types"
)

func ComputeNewBalance(assetType int64, balance string, balanceDelta string) (newBalance string, err error) {
	switch assetType {
	case types.FungibleAssetType:
		assetInfo, err := types.ParseAccountAsset(balance)
		if err != nil {
			return "", err
		}
		assetDelta, err := types.ParseAccountAsset(balanceDelta)
		if err != nil {
			return "", err
		}
		assetInfo.Balance = ffmath.Add(assetInfo.Balance, assetDelta.Balance)
		assetInfo.LpAmount = ffmath.Add(assetInfo.LpAmount, assetDelta.LpAmount)
		if assetDelta.OfferCanceledOrFinalized == nil {
			assetDelta.OfferCanceledOrFinalized = types.ZeroBigInt
		}
		if assetDelta.OfferCanceledOrFinalized.Cmp(types.NilOfferCanceledOrFinalized) != 0 {
			assetInfo.OfferCanceledOrFinalized = assetDelta.OfferCanceledOrFinalized
		}
		newBalance = assetInfo.String()
	case types.LiquidityAssetType:
		// balance: LiquidityInfo
		liquidityInfo, err := types.ParseLiquidityInfo(balance)
		if err != nil {
			return "", err
		}
		deltaLiquidity, err := types.ParseLiquidityInfo(balanceDelta)
		if err != nil {
			return "", err
		}
		liquidityInfo.AssetAId = deltaLiquidity.AssetAId
		liquidityInfo.AssetBId = deltaLiquidity.AssetBId
		liquidityInfo.AssetA = ffmath.Add(liquidityInfo.AssetA, deltaLiquidity.AssetA)
		liquidityInfo.AssetB = ffmath.Add(liquidityInfo.AssetB, deltaLiquidity.AssetB)
		liquidityInfo.LpAmount = ffmath.Add(liquidityInfo.LpAmount, deltaLiquidity.LpAmount)
		if deltaLiquidity.KLast.Cmp(types.ZeroBigInt) != 0 {
			liquidityInfo.KLast = deltaLiquidity.KLast
		}
		liquidityInfo.FeeRate = deltaLiquidity.FeeRate
		liquidityInfo.TreasuryAccountIndex = deltaLiquidity.TreasuryAccountIndex
		liquidityInfo.TreasuryRate = deltaLiquidity.TreasuryRate
		newBalance = liquidityInfo.String()
	case types.NftAssetType:
		// just set the old one as the new one
		newBalance = balanceDelta
	default:
		return "", errors.New("[ComputeNewBalance] invalid asset type")
	}
	return newBalance, nil
}
