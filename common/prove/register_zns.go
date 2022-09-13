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
	"strings"

	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbnb/types"
)

func fillRegisterZnsTxWitness(cryptoTx *TxWitness, oTx *Tx) error {
	txInfo, err := types.ParseRegisterZnsTxInfo(oTx.TxInfo)
	if err != nil {
		return err
	}
	cryptoTxInfo, err := toCryptoRegisterZnsTx(txInfo)
	if err != nil {
		return err
	}
	cryptoTx.Signature = std.EmptySignature()
	cryptoTx.RegisterZnsTxInfo = cryptoTxInfo
	return nil
}

func toCryptoRegisterZnsTx(txInfo *legendTxTypes.RegisterZnsTxInfo) (info *CryptoRegisterZnsTx, err error) {
	accountName := make([]byte, 32)
	realName := strings.Split(txInfo.AccountName, types.AccountNameSuffix)[0]
	copy(accountName[:], realName)
	pk, err := types.ParsePubKey(txInfo.PubKey)
	if err != nil {
		return nil, err
	}
	info = &CryptoRegisterZnsTx{
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     accountName,
		AccountNameHash: txInfo.AccountNameHash,
		PubKey:          pk,
	}
	return info, nil
}
