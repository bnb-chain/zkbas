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

package logic

import (
	"errors"
	"time"

	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/errorcode"
)

func SendCommittedBlocks(param *SenderParam, l1TxSenderModel L1TxSenderModel,
	blockModel BlockModel, blockForCommitModel BlockForCommitModel) (err error) {
	var (
		cli            = param.Cli
		authCli        = param.AuthCli
		zkbasInstance  = param.ZkbasInstance
		gasPrice       = param.GasPrice
		gasLimit       = param.GasLimit
		maxBlockCount  = param.MaxBlocksCount
		maxWaitingTime = param.MaxWaitingTime
	)
	// scan l1 tx sender table for handled committed height
	lastHandledBlock, getHandleErr := l1TxSenderModel.GetLatestHandledBlock(CommitTxType)
	if getHandleErr != nil && getHandleErr != errorcode.DbErrNotFound {
		logx.Errorf("[SendVerifiedAndExecutedBlocks] GetLatestHandledBlock err: %s", getHandleErr.Error())
		return getHandleErr
	}
	// scan l1 tx sender table for pending committed height that higher than the latest handled height
	pendingSender, getPendingerr := l1TxSenderModel.GetLatestPendingBlock(CommitTxType)
	if getPendingerr != nil {
		if getPendingerr != errorcode.DbErrNotFound {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] GetLatestPendingBlock err: %s", getPendingerr.Error())
			return getPendingerr
		}
	}

	// case 1:
	if getHandleErr == errorcode.DbErrNotFound && getPendingerr == nil {
		_, isPending, err := cli.GetTransactionByHash(pendingSender.L1TxHash)
		// if err != nil, means we cannot get this tx by hash
		if err != nil {
			// if we cannot get it from rpc and the time over 1 min
			lastUpdatedAt := pendingSender.UpdatedAt.UnixMilli()
			now := time.Now().UnixMilli()
			if now-lastUpdatedAt > maxWaitingTime {
				err := l1TxSenderModel.DeleteL1TxSender(pendingSender)
				if err != nil {
					logx.Errorf("[SendCommittedBlocks] unable to delete l1 tx sender: %s", err.Error())
					return err
				}
				return nil
			} else {
				return nil
			}
		}
		// if it is pending, still waiting
		if isPending {
			logx.Infof("[SendCommittedBlocks] tx is still pending, no need to work for anything tx hash: %s", pendingSender.L1TxHash)
			return nil
		} else {
			receipt, err := cli.GetTransactionReceipt(pendingSender.L1TxHash)
			if err != nil {
				logx.Errorf("[SendCommittedBlocks] unable to get transaction receipt: %s", err.Error())
				return err
			}
			if receipt.Status == 0 {
				logx.Infof("[SendCommittedBlocks] the transaction is failure, please check: %s", pendingSender.L1TxHash)
				return nil
			}
		}
	}
	// case 2:
	if getHandleErr == nil && getPendingerr == nil {
		isSuccess, err := cli.WaitingTransactionStatus(pendingSender.L1TxHash)
		// if err != nil, means we cannot get this tx by hash
		if err != nil {
			// if we cannot get it from rpc and the time over 1 min
			lastUpdatedAt := pendingSender.UpdatedAt.UnixMilli()
			now := time.Now().UnixMilli()
			if now-lastUpdatedAt > maxWaitingTime {
				// drop the record
				err := l1TxSenderModel.DeleteL1TxSender(pendingSender)
				if err != nil {
					logx.Errorf("[SendCommittedBlocks] unable to delete l1 tx sender: %s", err.Error())
					return err
				}
				return nil
			} else {
				logx.Infof("[SendCommittedBlocks] tx cannot be found, but not exceed time limit: %s", pendingSender.L1TxHash)
				return nil
			}
		}
		// if it is pending, still waiting
		if !isSuccess {
			logx.Infof("[SendCommittedBlocks] tx is still pending, no need to work for anything tx hash: %s", pendingSender.L1TxHash)
			return nil
		}
	}

	// case 3:
	var lastStoredBlockInfo StorageStoredBlockInfo
	var pendingCommitBlocks []ZkbasCommitBlockInfo
	// if lastHandledBlock == nil, means we haven't committed any blocks, just start from 0
	// if errorcode.DbErrNotFound, means we haven't committed new blocks, just start to commit
	if getHandleErr == errorcode.DbErrNotFound && getPendingerr == errorcode.DbErrNotFound {
		var blocks []*BlockForCommit
		blocks, err = blockForCommitModel.GetBlockForCommitBetween(1, int64(maxBlockCount))
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] GetBlockForCommitBetween err: %d, maxBlockCount: %d", err.Error(), maxBlockCount)
			return err
		}
		pendingCommitBlocks, err = ConvertBlocksForCommitToCommitBlockInfos(blocks)
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] unable to convert blocks to commit block infos: %s", err.Error())
			return err
		}
		// set stored block header to default 0
		lastStoredBlockInfo = DefaultBlockHeader()
	}
	if getHandleErr == nil && getPendingerr == errorcode.DbErrNotFound {
		// if errorcode.DbErrNotFound, means we haven't committed new blocks, just start to commit
		// get blocks higher than last handled blocks
		var blocks []*BlockForCommit
		// commit new blocks
		blocks, err = blockForCommitModel.GetBlockForCommitBetween(lastHandledBlock.L2BlockHeight+1, lastHandledBlock.L2BlockHeight+int64(maxBlockCount))
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] unable to get sender new blocks: %s", err.Error())
			return err
		}
		pendingCommitBlocks, err = ConvertBlocksForCommitToCommitBlockInfos(blocks)
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] unable to convert blocks to commit block infos: %s", err.Error())
			return err
		}
		// get last block info
		lastHandledBlockInfo, err := blockModel.GetBlockByBlockHeight(lastHandledBlock.L2BlockHeight)
		if err != nil && err != errorcode.DbErrNotFound {
			logx.Errorf("[SendCommittedBlocks] unable to get last handled block info: %s", err.Error())
			return err
		}
		// construct last stored block header
		lastStoredBlockInfo = util.ConstructStoredBlockInfo(lastHandledBlockInfo)
	}
	// commit blocks on-chain
	if len(pendingCommitBlocks) != 0 {
		txHash, err := zkbas.CommitBlocks(
			cli, authCli,
			zkbasInstance,
			lastStoredBlockInfo,
			pendingCommitBlocks,
			gasPrice,
			gasLimit)
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] unable to commit blocks: %s", err.Error())
			return err
		}
		for _, pendingCommittedBlock := range pendingCommitBlocks {
			logx.Infof("[SendCommittedBlocks] commit blocks: %v", pendingCommittedBlock.BlockNumber)
		}
		// update l1 tx sender table records
		newSender := &L1TxSender{
			L1TxHash:      txHash,
			TxStatus:      PendingStatus,
			TxType:        CommitTxType,
			L2BlockHeight: int64(pendingCommitBlocks[len(pendingCommitBlocks)-1].BlockNumber),
		}
		isValid, err := l1TxSenderModel.CreateL1TxSender(newSender)
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] unable to create l1 tx sender")
			return err
		}
		if !isValid {
			logx.Errorf("[SendCommittedBlocks] cannot create new senders")
			return errors.New("[SendCommittedBlocks] cannot create new senders")
		}
		logx.Infof("[SendCommittedBlocks] new blocks have been committed(height): %v", newSender.L2BlockHeight)
		return nil
	}
	return nil
}
