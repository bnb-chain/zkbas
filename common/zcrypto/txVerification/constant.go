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

package txVerification

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
)

type (
	TransferTxInfo         = commonTx.TransferTxInfo
	SwapTxInfo             = commonTx.SwapTxInfo
	AddLiquidityTxInfo     = commonTx.AddLiquidityTxInfo
	RemoveLiquidityTxInfo  = commonTx.RemoveLiquidityTxInfo
	WithdrawTxInfo         = commonTx.WithdrawTxInfo
	CreateCollectionTxInfo = commonTx.CreateCollectionTxInfo
	MintNftTxInfo          = commonTx.MintNftTxInfo
	TransferNftTxInfo      = commonTx.TransferNftTxInfo
	OfferTxInfo            = commonTx.OfferTxInfo
	AtomicMatchTxInfo      = commonTx.AtomicMatchTxInfo
	CancelOfferTxInfo      = commonTx.CancelOfferTxInfo
	WithdrawNftTxInfo      = commonTx.WithdrawNftTxInfo

	PublicKey = eddsa.PublicKey

	MempoolTxDetail = mempool.MempoolTxDetail

	AccountInfo   = commonAsset.AccountInfo
	LiquidityInfo = commonAsset.LiquidityInfo
	NftInfo       = commonAsset.NftInfo
)

const (
	OfferPerAsset = 128

	TenThousand = 10000

	GeneralAssetType         = commonAsset.GeneralAssetType
	LiquidityAssetType       = commonAsset.LiquidityAssetType
	NftAssetType             = commonAsset.NftAssetType
	CollectionNonceAssetType = commonAsset.CollectionNonceAssetType
)

var (
	ZeroBigInt = big.NewInt(0)
)
