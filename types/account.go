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

package types

import (
	"encoding/json"
	"github.com/bnb-chain/zkbnb-crypto/legend/circuit/bn254/std"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

const (
	FungibleAssetType        = 1
	LiquidityAssetType       = 2
	NftAssetType             = 3
	CollectionNonceAssetType = 4

	BuyOfferType  = 0
	SellOfferType = 1
)

type AccountAsset std.AccountAsset

func (asset *AccountAsset) DeepCopy() *AccountAsset {
	return &AccountAsset{
		AssetId:                  asset.AssetId,
		Balance:                  big.NewInt(0).Set(asset.Balance),
		LpAmount:                 big.NewInt(0).Set(asset.LpAmount),
		OfferCanceledOrFinalized: big.NewInt(0).Set(asset.OfferCanceledOrFinalized),
	}
}

func ParseAccountAsset(balance string) (asset *AccountAsset, err error) {
	err = json.Unmarshal([]byte(balance), &asset)
	if err != nil {
		return nil, JsonErrUnmarshal
	}
	return asset, nil
}

func (asset *AccountAsset) String() (info string) {
	infoBytes, _ := json.Marshal(asset)
	return string(infoBytes)
}

type AccountInfo struct {
	AccountId       uint
	AccountIndex    int64
	AccountName     string
	PublicKey       string
	AccountNameHash string
	L1Address       string
	Nonce           int64
	CollectionNonce int64
	// map[int64]*AccountAsset
	AssetInfo map[int64]*AccountAsset // key: index, value: balance
	AssetRoot string
	Status    int
}

func (ai *AccountInfo) DeepCopy() (*AccountInfo, error) {
	assetInfo := make(map[int64]*AccountAsset)
	for assetId, asset := range ai.AssetInfo {
		assetInfo[assetId] = asset.DeepCopy()
	}

	newAccountInfo := &AccountInfo{
		AccountId:       ai.AccountId,
		AccountIndex:    ai.AccountIndex,
		AccountName:     ai.AccountName,
		PublicKey:       ai.PublicKey,
		AccountNameHash: ai.AccountNameHash,
		L1Address:       ai.L1Address,
		Nonce:           ai.Nonce,
		CollectionNonce: ai.CollectionNonce,
		AssetInfo:       assetInfo,
		AssetRoot:       ai.AssetRoot,
		Status:          ai.Status,
	}
	return newAccountInfo, nil
}

func (ai *AccountInfo) ToStdAccount() (*std.Account, error) {
	pk, err := ParsePubKey(ai.PublicKey)
	if err != nil {
		return nil, err
	}
	return &std.Account{
		AccountIndex:    ai.AccountIndex,
		AccountNameHash: common.FromHex(ai.AccountNameHash),
		AccountPk:       pk,
		Nonce:           ai.Nonce,
		CollectionNonce: ai.CollectionNonce,
		AssetRoot:       common.FromHex(ai.AssetRoot),
	}, nil
}

func ParsePubKey(pkStr string) (pk *eddsa.PublicKey, err error) {
	pkBytes := common.FromHex(pkStr)
	pk = new(eddsa.PublicKey)
	_, err = pk.A.SetBytes(pkBytes)
	if err != nil {
		return nil, err
	}
	return pk, nil
}
