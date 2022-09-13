/*
 * Copyright © 2021 ZkBNB Protocol
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
	"github.com/bnb-chain/zkbnb-crypto/legend/circuit/bn254/std"
	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbnb/types"
)

func fillFullExitNftTxWitness(cryptoTx *TxWitness, oTx *Tx) error {
	txInfo, err := types.ParseFullExitNftTxInfo(oTx.TxInfo)
	if err != nil {
		return err
	}
	cryptoTxInfo, err := toCryptoFullExitNftTx(txInfo)
	if err != nil {
		return err
	}
	cryptoTx.FullExitNftTxInfo = cryptoTxInfo
	cryptoTx.Signature = std.EmptySignature()
	return nil
}

func toCryptoFullExitNftTx(txInfo *legendTxTypes.FullExitNftTxInfo) (info *CryptoFullExitNftTx, err error) {
	info = &CryptoFullExitNftTx{
		AccountIndex:           txInfo.AccountIndex,
		AccountNameHash:        txInfo.AccountNameHash,
		CreatorAccountIndex:    txInfo.CreatorAccountIndex,
		CreatorAccountNameHash: txInfo.CreatorAccountNameHash,
		CreatorTreasuryRate:    txInfo.CreatorTreasuryRate,
		NftIndex:               txInfo.NftIndex,
		CollectionId:           txInfo.CollectionId,
		NftContentHash:         txInfo.NftContentHash,
		NftL1Address:           txInfo.NftL1Address,
		NftL1TokenId:           txInfo.NftL1TokenId,
	}
	return info, nil
}
