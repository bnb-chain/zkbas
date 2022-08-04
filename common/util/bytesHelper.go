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
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	AccountNameSuffix = ".legend"
)

func PrefixPaddingBufToChunkSize(buf []byte) []byte {
	return new(big.Int).SetBytes(buf).FillBytes(make([]byte, 32))
}

func SuffixPaddingBufToChunkSize(buf []byte) []byte {
	res := make([]byte, 32)
	copy(res[:], buf[:])
	return res
}

func AccountNameToBytes32(accountName string) []byte {
	realName := strings.Split(accountName, AccountNameSuffix)[0]
	buf := make([]byte, 32)
	copy(buf[:], realName)
	return buf
}

func AddressStrToBytes(addr string) []byte {
	return new(big.Int).SetBytes(common.FromHex(addr)).FillBytes(make([]byte, 20))
}

func Uint16ToBytes(a uint16) []byte {
	return new(big.Int).SetUint64(uint64(a)).FillBytes(make([]byte, 2))
}

func Uint24ToBytes(a int64) []byte {
	return new(big.Int).SetInt64(a).FillBytes(make([]byte, 3))
}

func Uint32ToBytes(a uint32) []byte {
	return new(big.Int).SetUint64(uint64(a)).FillBytes(make([]byte, 4))
}

func Uint40ToBytes(a int64) []byte {
	return new(big.Int).SetInt64(a).FillBytes(make([]byte, 5))
}

func Uint128ToBytes(a *big.Int) []byte {
	return a.FillBytes(make([]byte, 16))
}

func Uint256ToBytes(a *big.Int) []byte {
	return a.FillBytes(make([]byte, 32))
}

func AmountToPackedAmountBytes(a *big.Int) (res []byte, err error) {
	packedAmount, err := ToPackedAmount(a)
	if err != nil {
		logx.Errorf("[AmountToPackedAmountBytes] invalid amount: %s", err.Error())
		return nil, err
	}
	return Uint40ToBytes(packedAmount), nil
}

func FeeToPackedFeeBytes(a *big.Int) (res []byte, err error) {
	packedFee, err := ToPackedFee(a)
	if err != nil {
		logx.Errorf("[FeeToPackedFeeBytes] invalid fee amount: %s", err.Error())
		return nil, err
	}
	return Uint16ToBytes(uint16(packedFee)), nil
}

func FromHex(s string) ([]byte, error) {
	if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
	}

	if len(s)%2 == 1 {
		s = "0" + s
	}
	return hex.DecodeString(s)
}
