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

package proverUtil

import (
	"errors"

	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/pkg/treedb"
)

func ConstructDepositNftCryptoTx(
	oTx *Tx,
	treeCtx *treedb.Context,
	finalityBlockNr uint64,
	accountTree bsmt.SparseMerkleTree,
	accountAssetsTree *[]bsmt.SparseMerkleTree,
	liquidityTree bsmt.SparseMerkleTree,
	nftTree bsmt.SparseMerkleTree,
	accountModel AccountModel,
) (cryptoTx *CryptoTx, err error) {
	if oTx.TxType != commonTx.TxTypeDepositNft {
		logx.Errorf("[ConstructCreatePairCryptoTx] invalid tx type")
		return nil, errors.New("[ConstructCreatePairCryptoTx] invalid tx type")
	}
	if oTx == nil || accountTree == nil || accountAssetsTree == nil || liquidityTree == nil || nftTree == nil {
		logx.Errorf("[ConstructDepositNftCryptoTx] invalid params")
		return nil, errors.New("[ConstructDepositNftCryptoTx] invalid params")
	}
	txInfo, err := commonTx.ParseDepositNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructDepositNftCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoDepositNftTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructDepositNftCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, err := ConstructProverInfo(oTx, accountModel)
	if err != nil {
		logx.Errorf("[ConstructDepositNftCryptoTx] unable to construct prover info: %s", err.Error())
		return nil, err
	}
	cryptoTx, err = ConstructWitnessInfo(
		oTx,
		accountModel,
		treeCtx,
		finalityBlockNr,
		accountTree,
		accountAssetsTree,
		liquidityTree,
		nftTree,
		accountKeys,
		proverAccounts,
		proverLiquidityInfo,
		proverNftInfo,
	)
	if err != nil {
		logx.Errorf("[ConstructDepositNftCryptoTx] unable to construct witness info: %s", err.Error())
		return nil, err
	}
	cryptoTx.TxType = uint8(oTx.TxType)
	cryptoTx.DepositNftTxInfo = cryptoTxInfo
	cryptoTx.Nonce = oTx.Nonce
	cryptoTx.Signature = std.EmptySignature()
	return cryptoTx, nil
}

func ToCryptoDepositNftTx(txInfo *commonTx.DepositNftTxInfo) (info *CryptoDepositNftTx, err error) {
	info = &CryptoDepositNftTx{
		AccountIndex:        txInfo.AccountIndex,
		NftIndex:            txInfo.NftIndex,
		NftL1Address:        txInfo.NftL1Address,
		AccountNameHash:     txInfo.AccountNameHash,
		NftContentHash:      txInfo.NftContentHash,
		NftL1TokenId:        txInfo.NftL1TokenId,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		CollectionId:        txInfo.CollectionId,
	}
	return info, nil
}
