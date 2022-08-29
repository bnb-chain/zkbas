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
	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbas/types"
)

func (w *WitnessHelper) constructFullExitTxWitness(cryptoTx *TxWitness, oTx *Tx) (*TxWitness, error) {
	txInfo, err := types.ParseFullExitTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoFullExitTx(txInfo)
	if err != nil {
		return nil, err
	}
	cryptoTx.FullExitTxInfo = cryptoTxInfo
	cryptoTx.Signature = std.EmptySignature()
	return cryptoTx, nil
}

func toCryptoFullExitTx(txInfo *legendTxTypes.FullExitTxInfo) (info *CryptoFullExitTx, err error) {
	info = &CryptoFullExitTx{
		AccountIndex:    txInfo.AccountIndex,
		AssetId:         txInfo.AssetId,
		AssetAmount:     txInfo.AssetAmount,
		AccountNameHash: txInfo.AccountNameHash,
	}
	return info, nil
}
