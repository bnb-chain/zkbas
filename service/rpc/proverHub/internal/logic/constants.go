/*
 * Copyright © 2021 Zecrey Protocol
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
 */

package logic

import (
	"sync"
	"sync/atomic"
	"unsafe"

	cryptoBlock "github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/bnb-chain/zkbas/common/model/block"
)

type CryptoBlockInfo struct {
	BlockInfo *CryptoBlock
	Status    int64
}

var (
	UnProvedCryptoBlocks []*CryptoBlockInfo
	M                    Mutex
	VerifyingKeyPath     string
)

const mutexLocked = 1 << iota

type Mutex struct {
	sync.Mutex
}

func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.Mutex)), 0, mutexLocked)
}

const (
	PUBLISHED = iota
	RECEIVED
	VERIFIED
)

type (
	Block                  = block.Block
	StorageStoredBlockInfo = legend.StorageStoredBlockInfo
	CryptoTx               = cryptoBlock.Tx
	CryptoBlock            = cryptoBlock.Block
)
