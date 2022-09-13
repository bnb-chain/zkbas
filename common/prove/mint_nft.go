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
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/common"

	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/types"
)

func fillMintNftTxWitness(cryptoTx *TxWitness, oTx *Tx) error {
	txInfo, err := types.ParseMintNftTxInfo(oTx.TxInfo)
	if err != nil {
		return err
	}
	cryptoTxInfo, err := toCryptoMintNftTx(txInfo)
	if err != nil {
		return err
	}
	cryptoTx.MintNftTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		return err
	}
	return nil
}

func toCryptoMintNftTx(txInfo *types.MintNftTxInfo) (info *CryptoMintNftTx, err error) {
	packedFee, err := common2.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		return nil, err
	}
	info = &CryptoMintNftTx{
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		ToAccountIndex:      txInfo.ToAccountIndex,
		ToAccountNameHash:   common.FromHex(txInfo.ToAccountNameHash),
		NftIndex:            txInfo.NftIndex,
		NftContentHash:      common.FromHex(txInfo.NftContentHash),
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		GasAccountIndex:     txInfo.GasAccountIndex,
		GasFeeAssetId:       txInfo.GasFeeAssetId,
		GasFeeAssetAmount:   packedFee,
		CollectionId:        txInfo.NftCollectionId,
		ExpiredAt:           txInfo.ExpiredAt,
	}
	return info, nil
}
