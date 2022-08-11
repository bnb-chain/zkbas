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
 */

package svc

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/l1TxSender"
	"github.com/bnb-chain/zkbas/common/model/proofSender"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/common/util"
)

type (
	Block               = block.Block
	BlockForCommit      = blockForCommit.BlockForCommit
	L1TxSenderModel     = l1TxSender.L1TxSenderModel
	L1TxSender          = l1TxSender.L1TxSender
	BlockModel          = block.BlockModel
	BlockForCommitModel = blockForCommit.BlockForCommitModel

	ProviderClient = _rpc.ProviderClient
	AuthClient     = _rpc.AuthClient
	Zkbas          = zkbas.Zkbas

	ZkbasCommitBlockInfo   = zkbas.OldZkbasCommitBlockInfo
	ZkbasVerifyBlockInfo   = zkbas.OldZkbasVerifyAndExecuteBlockInfo
	StorageStoredBlockInfo = zkbas.StorageStoredBlockInfo

	ProofSenderModel = proofSender.ProofSenderModel
)

type SenderParam struct {
	Cli            *ProviderClient
	AuthCli        *AuthClient
	ZkbasInstance  *Zkbas
	MaxWaitingTime int64
	MaxBlocksCount int
	GasPrice       *big.Int
	GasLimit       uint64
}

func DefaultBlockHeader() StorageStoredBlockInfo {
	var (
		pendingOnChainOperationsHash [32]byte
		stateRoot                    [32]byte
		commitment                   [32]byte
	)
	copy(pendingOnChainOperationsHash[:], common.FromHex(util.EmptyStringKeccak)[:])
	copy(stateRoot[:], tree.NilStateRoot[:])
	copy(commitment[:], common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000")[:])
	return StorageStoredBlockInfo{
		BlockSize:                    0,
		BlockNumber:                  0,
		PriorityOperations:           0,
		PendingOnchainOperationsHash: pendingOnChainOperationsHash,
		Timestamp:                    big.NewInt(0),
		StateRoot:                    stateRoot,
		Commitment:                   commitment,
	}
}

func ConvertBlocksForCommitToCommitBlockInfos(oBlocks []*BlockForCommit) (commitBlocks []ZkbasCommitBlockInfo, err error) {
	for _, oBlock := range oBlocks {
		var newStateRoot [32]byte
		var pubDataOffsets []uint32
		copy(newStateRoot[:], common.FromHex(oBlock.StateRoot)[:])
		err = json.Unmarshal([]byte(oBlock.PublicDataOffsets), &pubDataOffsets)
		if err != nil {
			logx.Errorf("[ConvertBlocksForCommitToCommitBlockInfos] unable to unmarshal: %s", err.Error())
			return nil, err
		}
		commitBlock := ZkbasCommitBlockInfo{
			NewStateRoot:      newStateRoot,
			PublicData:        common.FromHex(oBlock.PublicData),
			Timestamp:         big.NewInt(oBlock.Timestamp),
			PublicDataOffsets: pubDataOffsets,
			BlockNumber:       uint32(oBlock.BlockHeight),
			BlockSize:         oBlock.BlockSize,
		}
		commitBlocks = append(commitBlocks, commitBlock)
	}
	return commitBlocks, nil
}

func ConvertBlocksToVerifyAndExecuteBlockInfos(oBlocks []*Block) (verifyAndExecuteBlocks []ZkbasVerifyBlockInfo, err error) {
	for _, oBlock := range oBlocks {
		var pendingOnChainOpsPubData [][]byte
		if oBlock.PendingOnChainOperationsPubData != "" {
			err = json.Unmarshal([]byte(oBlock.PendingOnChainOperationsPubData), &pendingOnChainOpsPubData)
			if err != nil {
				logx.Errorf("[ConvertBlocksToVerifyAndExecuteBlockInfos] unable to unmarshal pending pub data: %s", err.Error())
				return nil, err
			}
		}
		verifyAndExecuteBlock := ZkbasVerifyBlockInfo{
			BlockHeader:              util.ConstructStoredBlockInfo(oBlock),
			PendingOnchainOpsPubData: pendingOnChainOpsPubData,
		}
		verifyAndExecuteBlocks = append(verifyAndExecuteBlocks, verifyAndExecuteBlock)
	}
	return verifyAndExecuteBlocks, nil
}
