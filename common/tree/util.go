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

package tree

import (
	"math/big"

	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/pkg/errors"
)

func EmptyAccountNodeHash() []byte {
	hFunc := mimc.NewMiMC()
	zero := big.NewInt(0).FillBytes(make([]byte, 32))
	/*
		AccountNameHash
		PubKey
		Nonce
		CollectionNonce
		AssetRoot
	*/
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	// asset root
	hFunc.Write(NilAccountAssetRoot)
	return hFunc.Sum(nil)
}

func EmptyAccountAssetNodeHash() []byte {
	hFunc := mimc.NewMiMC()
	zero := big.NewInt(0).FillBytes(make([]byte, 32))
	/*
		balance
		lpAmount
		offerCanceledOrFinalized
	*/
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	return hFunc.Sum(nil)
}

func EmptyLiquidityNodeHash() []byte {
	hFunc := mimc.NewMiMC()
	zero := big.NewInt(0).FillBytes(make([]byte, 32))
	/*
		assetAId
		assetA
		assetBId
		assetB
		lpAmount
		kLast
		feeRate
		treasuryAccountIndex
		treasuryRate
	*/
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	return hFunc.Sum(nil)
}

func EmptyNftNodeHash() []byte {
	hFunc := mimc.NewMiMC()
	zero := big.NewInt(0).FillBytes(make([]byte, 32))
	/*
		creatorAccountIndex
		ownerAccountIndex
		nftContentHash
		nftL1Address
		nftL1TokenId
		creatorTreasuryRate
		collectionId
	*/
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	hFunc.Write(zero)
	return hFunc.Sum(nil)
}

func CommitTrees(version uint64,
	accountTree bsmt.SparseMerkleTree,
	assetTrees *[]bsmt.SparseMerkleTree,
	liquidityTree bsmt.SparseMerkleTree,
	nftTree bsmt.SparseMerkleTree) error {

	accPrunedVersion := bsmt.Version(version)
	if accountTree.LatestVersion() < accPrunedVersion {
		accPrunedVersion = accountTree.LatestVersion()
	}
	ver, err := accountTree.Commit(&accPrunedVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to commit account tree, tree ver: %d, prune ver: %d", ver, accPrunedVersion)
	}
	for idx, assetTree := range *assetTrees {
		assetPrunedVersion := bsmt.Version(version)
		if assetTree.LatestVersion() < assetPrunedVersion {
			assetPrunedVersion = assetTree.LatestVersion()
		}
		ver, err := assetTree.Commit(&assetPrunedVersion)
		if err != nil {
			return errors.Wrapf(err, "unable to commit asset tree [%d], tree ver: %d, prune ver: %d", idx, ver, assetPrunedVersion)
		}
	}
	liquidityPrunedVersion := bsmt.Version(version)
	if liquidityTree.LatestVersion() < liquidityPrunedVersion {
		liquidityPrunedVersion = liquidityTree.LatestVersion()
	}
	ver, err = liquidityTree.Commit(&liquidityPrunedVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to commit liquidity tree, tree ver: %d, prune ver: %d", ver, liquidityPrunedVersion)
	}
	nftPrunedVersion := bsmt.Version(version)
	if nftTree.LatestVersion() < nftPrunedVersion {
		nftPrunedVersion = nftTree.LatestVersion()
	}
	ver, err = nftTree.Commit(&nftPrunedVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to commit nft tree, tree ver: %d, prune ver: %d", ver, nftPrunedVersion)
	}
	return nil
}

func RollBackTrees(version uint64,
	accountTree bsmt.SparseMerkleTree,
	assetTrees *[]bsmt.SparseMerkleTree,
	liquidityTree bsmt.SparseMerkleTree,
	nftTree bsmt.SparseMerkleTree) error {

	ver := bsmt.Version(version)
	if accountTree.LatestVersion() > ver && !accountTree.IsEmpty() {
		err := accountTree.Rollback(ver)
		if err != nil {
			return errors.Wrapf(err, "unable to rollback account tree, ver: %d", ver)
		}
	}

	for idx, assetTree := range *assetTrees {
		if assetTree.LatestVersion() > ver && !assetTree.IsEmpty() {
			err := assetTree.Rollback(ver)
			if err != nil {
				return errors.Wrapf(err, "unable to rollback asset tree [%d], ver: %d", idx, ver)
			}
		}
	}

	if liquidityTree.LatestVersion() > ver && !liquidityTree.IsEmpty() {
		err := liquidityTree.Rollback(ver)
		if err != nil {
			return errors.Wrapf(err, "unable to rollback liquidity tree, ver: %d", ver)
		}
	}

	if nftTree.LatestVersion() > ver && !nftTree.IsEmpty() {
		err := nftTree.Rollback(ver)
		if err != nil {
			return errors.Wrapf(err, "unable to rollback nft tree, tree ver: %d", ver)
		}
	}
	return nil
}
