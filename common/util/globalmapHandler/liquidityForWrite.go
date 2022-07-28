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

package globalmapHandler

import (
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/util"
)

func GetLatestLiquidityInfoForWrite(
	liquidityModel LiquidityModel,
	mempoolTxModel MempoolModel,
	redisConnection *Redis,
	pairIndex int64,
) (
	redisLock *RedisLock,
	liquidityInfo *LiquidityInfo,
	err error,
) {
	key := util.GetLiquidityKeyForWrite(pairIndex)
	lockKey := util.GetLockKey(key)
	redisLock = GetRedisLockByKey(redisConnection, lockKey)
	err = TryAcquireLock(redisLock)
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to get lock: %s", err.Error())
		return nil, nil, err
	}
	liquidityInfoStr, err := redisConnection.Get(key)
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to get data from redis: %s", err.Error())
		return nil, nil, err
	}
	var (
		dbLiquidityInfo *liquidity.Liquidity
	)
	if liquidityInfoStr == "" {
		// get latest info from liquidity table
		dbLiquidityInfo, err = liquidityModel.GetLiquidityByPairIndex(pairIndex)
		if err != nil {
			logx.Errorf("[GetLatestLiquidityInfoForRead] unable to get latest liquidity by pair index: %s", err.Error())
			return nil, nil, err
		}

		mempoolTxs, err := mempoolTxModel.GetPendingLiquidityTxs()
		if err != nil {
			if err != mempool.ErrNotFound {
				logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to get mempool txs by account index: %s", err.Error())
				return nil, nil, err
			}
		}
		liquidityInfo, err = commonAsset.ConstructLiquidityInfo(
			pairIndex,
			dbLiquidityInfo.AssetAId,
			dbLiquidityInfo.AssetA,
			dbLiquidityInfo.AssetBId,
			dbLiquidityInfo.AssetB,
			dbLiquidityInfo.LpAmount,
			dbLiquidityInfo.KLast,
			dbLiquidityInfo.FeeRate,
			dbLiquidityInfo.TreasuryAccountIndex,
			dbLiquidityInfo.TreasuryRate,
		)
		if err != nil {
			logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to construct pool info: %s", err.Error())
			return nil, nil, err
		}
		for _, mempoolTx := range mempoolTxs {
			for _, txDetail := range mempoolTx.MempoolDetails {
				if txDetail.AssetType != commonAsset.LiquidityAssetType || liquidityInfo.PairIndex != txDetail.AssetId {
					continue
				}
				nBalance, err := commonAsset.ComputeNewBalance(commonAsset.LiquidityAssetType, liquidityInfo.String(), txDetail.BalanceDelta)
				if err != nil {
					logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to compute new balance: %s", err.Error())
					return nil, nil, err
				}
				liquidityInfo, err = commonAsset.ParseLiquidityInfo(nBalance)
				if err != nil {
					logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to parse pool info: %s", err.Error())
					return nil, nil, err
				}
			}
		}
	} else {
		err = json.Unmarshal([]byte(liquidityInfoStr), &liquidityInfo)
		if err != nil {
			logx.Errorf("[GetLatestLiquidityInfoForWrite] unable to unmarshal liquidity info: %s", err.Error())
			return nil, nil, err
		}
	}
	return redisLock, liquidityInfo, nil
}
