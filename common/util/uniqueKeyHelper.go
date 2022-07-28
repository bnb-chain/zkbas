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

package util

import (
	"strconv"
)

const (
	AccountPrefix        = "AccountIndex::"
	LiquidityReadPrefix  = "LiquidityRead::"
	NftReadPrefix        = "NftRead::"
	NftIndexWritePrefix  = "NftIndexWrite::"
	OfferIdWritePrefix   = "OfferIdWrite::"
	LiquidityWritePrefix = "LiquidityWrite::"
	BasicAccountPrefix   = "BasicAccount::"
	LockKeySuffix        = "ByLock"
)

func GetLiquidityKeyForRead(pairIndex int64) string {
	return LiquidityReadPrefix + strconv.FormatInt(pairIndex, 10)
}

func GetNftKeyForRead(nftIndex int64) string {
	return NftReadPrefix + strconv.FormatInt(nftIndex, 10)
}

func GetNftIndexKeyForWrite() string {
	return NftIndexWritePrefix
}

func GetOfferIdKeyForWrite(accountIndex int64) string {
	return OfferIdWritePrefix + strconv.FormatInt(accountIndex, 10)
}

func GetLiquidityKeyForWrite(pairIndex int64) string {
	return LiquidityWritePrefix + strconv.FormatInt(pairIndex, 10)
}

func GetAccountKey(accountIndex int64) string {
	return AccountPrefix + strconv.FormatInt(accountIndex, 10)
}

func GetBasicAccountKey(accountIndex int64) string {
	return BasicAccountPrefix + strconv.FormatInt(accountIndex, 10)
}

func GetLockKey(key string) string {
	return key + LockKeySuffix
}
