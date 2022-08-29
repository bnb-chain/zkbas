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

package prove

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"

	"github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/types"
)

func (w *WitnessHelper) constructAddLiquidityTxWitness(witness *TxWitness, oTx *Tx) (*TxWitness, error) {
	txInfo, err := types.ParseAddLiquidityTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoAddLiquidityTx(txInfo)
	if err != nil {
		return nil, err
	}
	witness.AddLiquidityTxInfo = cryptoTxInfo
	witness.ExpiredAt = txInfo.ExpiredAt
	witness.Signature = new(eddsa.Signature)
	_, err = witness.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		return nil, err
	}
	return witness, nil
}

func toCryptoAddLiquidityTx(txInfo *types.AddLiquidityTxInfo) (info *CryptoAddLiquidityTx, err error) {
	packedAAmount, err := common.ToPackedAmount(txInfo.AssetAAmount)
	if err != nil {
		return nil, err
	}
	packedBAmount, err := common.ToPackedAmount(txInfo.AssetBAmount)
	if err != nil {
		return nil, err
	}
	packedLpAmount, err := common.ToPackedAmount(txInfo.LpAmount)
	if err != nil {
		return nil, err
	}
	packedTreasuryAmount, err := common.ToPackedAmount(txInfo.TreasuryAmount)
	if err != nil {
		return nil, err
	}
	packedKLast, err := common.ToPackedAmount(txInfo.KLast)
	if err != nil {
		return nil, err
	}
	packedFee, err := common.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		return nil, err
	}
	info = &CryptoAddLiquidityTx{
		FromAccountIndex:  txInfo.FromAccountIndex,
		PairIndex:         txInfo.PairIndex,
		AssetAId:          txInfo.AssetAId,
		AssetAAmount:      packedAAmount,
		AssetBId:          txInfo.AssetBId,
		AssetBAmount:      packedBAmount,
		LpAmount:          packedLpAmount,
		KLast:             packedKLast,
		TreasuryAmount:    packedTreasuryAmount,
		GasAccountIndex:   txInfo.GasAccountIndex,
		GasFeeAssetId:     txInfo.GasFeeAssetId,
		GasFeeAssetAmount: packedFee,
	}
	return info, nil
}
