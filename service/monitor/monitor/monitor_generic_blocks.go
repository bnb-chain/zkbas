/*
 * Copyright © 2021 ZkBAS Protocol
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

package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	common2 "github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/dao/block"
	"github.com/bnb-chain/zkbas/dao/l1syncedblock"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/priorityrequest"
	types2 "github.com/bnb-chain/zkbas/types"
)

func (m *Monitor) MonitorGenericBlocks() (err error) {
	latestHandledBlock, err := m.L1SyncedBlockModel.GetLatestL1BlockByType(l1syncedblock.TypeGeneric)
	var handledHeight int64
	if err != nil {
		if err == types2.DbErrNotFound {
			handledHeight = m.Config.ChainConfig.StartL1BlockHeight
		} else {
			return fmt.Errorf("failed to get latest l1 monitor block, err: %v", err)
		}
	} else {
		handledHeight = latestHandledBlock.L1BlockHeight
	}

	// get latest l1 block height(latest height - pendingBlocksCount)
	latestHeight, err := m.cli.GetHeight()
	if err != nil {
		return fmt.Errorf("failed to get l1 height, err: %v", err)
	}

	safeHeight := latestHeight - m.Config.ChainConfig.ConfirmBlocksCount
	safeHeight = uint64(common2.MinInt64(int64(safeHeight), handledHeight+m.Config.ChainConfig.MaxHandledBlocksCount))
	if safeHeight <= uint64(handledHeight) {
		return nil
	}

	logx.Infof("syncing l1 blocks from %d to %d", big.NewInt(handledHeight+1), big.NewInt(int64(safeHeight)))

	priorityRequestCount, err := getPriorityRequestCount(m.cli, m.zkbasContractAddress, uint64(handledHeight+1), safeHeight)
	if err != nil {
		return fmt.Errorf("failed to get priority request count, err: %v", err)
	}

	logs, err := getZkbasContractLogs(m.cli, m.zkbasContractAddress, uint64(handledHeight+1), safeHeight)
	if err != nil {
		return fmt.Errorf("failed to get contract logs, err: %v", err)
	}
	var (
		l1EventInfos     []*L1EventInfo
		priorityRequests []*priorityrequest.PriorityRequest

		priorityRequestCountCheck = 0

		relatedBlocks = make(map[int64]*block.Block)
	)
	for _, vlog := range logs {
		l1EventInfo := &L1EventInfo{
			TxHash: vlog.TxHash.Hex(),
		}

		logBlock, err := m.cli.GetBlockHeaderByNumber(big.NewInt(int64(vlog.BlockNumber)))
		if err != nil {
			return fmt.Errorf("failed to get block header, err: %v", err)
		}

		switch vlog.Topics[0].Hex() {
		case zkbasLogNewPriorityRequestSigHash.Hex():
			priorityRequestCountCheck++
			l1EventInfo.EventType = EventTypeNewPriorityRequest

			l2TxEventMonitorInfo, err := convertLogToNewPriorityRequestEvent(vlog)
			if err != nil {
				return fmt.Errorf("failed to convert NewPriorityRequest log, err: %v", err)
			}
			priorityRequests = append(priorityRequests, l2TxEventMonitorInfo)
		case zkbasLogWithdrawalSigHash.Hex():
		case zkbasLogWithdrawalPendingSigHash.Hex():
		case zkbasLogBlockCommitSigHash.Hex():
			l1EventInfo.EventType = EventTypeCommittedBlock

			var event zkbas.ZkbasBlockCommit
			if err := ZkbasContractAbi.UnpackIntoInterface(&event, EventNameBlockCommit, vlog.Data); err != nil {
				return fmt.Errorf("failed to unpack ZkbasBlockCommit event, err: %v", err)
			}

			// update block status
			blockHeight := int64(event.BlockNumber)
			if relatedBlocks[blockHeight] == nil {
				relatedBlocks[blockHeight], err = m.BlockModel.GetBlockByHeightWithoutTx(blockHeight)
				if err != nil {
					return fmt.Errorf("GetBlockByHeightWithoutTx err: %v", err)
				}
			}
			relatedBlocks[blockHeight].CommittedTxHash = vlog.TxHash.Hex()
			relatedBlocks[blockHeight].CommittedAt = int64(logBlock.Time)
			relatedBlocks[blockHeight].BlockStatus = block.StatusCommitted
		case zkbasLogBlockVerificationSigHash.Hex():
			l1EventInfo.EventType = EventTypeVerifiedBlock

			var event zkbas.ZkbasBlockVerification
			if err := ZkbasContractAbi.UnpackIntoInterface(&event, EventNameBlockVerification, vlog.Data); err != nil {
				return fmt.Errorf("failed to unpack ZkbasBlockVerification err: %v", err)
			}

			// update block status
			blockHeight := int64(event.BlockNumber)
			if relatedBlocks[blockHeight] == nil {
				relatedBlocks[blockHeight], err = m.BlockModel.GetBlockByHeightWithoutTx(blockHeight)
				if err != nil {
					return fmt.Errorf("failed to GetBlockByHeightWithoutTx: %v", err)
				}
			}
			relatedBlocks[blockHeight].VerifiedTxHash = vlog.TxHash.Hex()
			relatedBlocks[blockHeight].VerifiedAt = int64(logBlock.Time)
			relatedBlocks[blockHeight].BlockStatus = block.StatusVerifiedAndExecuted
		case zkbasLogBlocksRevertSigHash.Hex():
			l1EventInfo.EventType = EventTypeRevertedBlock
		default:
		}

		l1EventInfos = append(l1EventInfos, l1EventInfo)
	}
	if priorityRequestCount != priorityRequestCountCheck {
		return fmt.Errorf("new priority requests events not match, try it again")
	}

	eventInfosBytes, err := json.Marshal(l1EventInfos)
	if err != nil {
		return err
	}
	l1BlockMonitorInfo := &l1syncedblock.L1SyncedBlock{
		L1BlockHeight: int64(safeHeight),
		BlockInfo:     string(eventInfosBytes),
		Type:          l1syncedblock.TypeGeneric,
	}

	// get pending update blocks
	pendingUpdateBlocks := make([]*block.Block, 0, len(relatedBlocks))
	for _, pendingUpdateBlock := range relatedBlocks {
		pendingUpdateBlocks = append(pendingUpdateBlocks, pendingUpdateBlock)
	}

	// get mempool txs to delete
	pendingDeleteMempoolTxs, err := getMempoolTxsToDelete(pendingUpdateBlocks, m.MempoolModel)
	if err != nil {
		return fmt.Errorf("failed to get mempool txs to delete, err: %v", err)
	}

	if err = m.L1SyncedBlockModel.CreateGenericBlock(l1BlockMonitorInfo, priorityRequests,
		pendingUpdateBlocks, pendingDeleteMempoolTxs); err != nil {
		return fmt.Errorf("failed to store monitor info, err: %v", err)
	}
	logx.Info("create txs count:", len(priorityRequests))
	return nil
}

func getMempoolTxsToDelete(blocks []*block.Block, mempoolModel mempool.MempoolModel) ([]*mempool.MempoolTx, error) {
	var toDeleteMempoolTxs []*mempool.MempoolTx
	for _, pendingUpdateBlock := range blocks {
		if pendingUpdateBlock.BlockStatus == BlockVerifiedStatus {
			_, blockToDleteMempoolTxs, err := mempoolModel.GetMempoolTxsByBlockHeight(pendingUpdateBlock.BlockHeight)
			if err != nil {
				logx.Errorf("GetMempoolTxsByBlockHeight err: %s", err.Error())
				return nil, err
			}
			if len(blockToDleteMempoolTxs) == 0 {
				continue
			}
			toDeleteMempoolTxs = append(toDeleteMempoolTxs, blockToDleteMempoolTxs...)
		}
	}
	return toDeleteMempoolTxs, nil
}

func getZkbasContractLogs(cli *_rpc.ProviderClient, zkbasContract string, startHeight, endHeight uint64) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startHeight)),
		ToBlock:   big.NewInt(int64(endHeight)),
		Addresses: []common.Address{common.HexToAddress(zkbasContract)},
	}
	logs, err := cli.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func getPriorityRequestCount(cli *_rpc.ProviderClient, zkbasContract string, startHeight, endHeight uint64) (int, error) {
	zkbasInstance, err := zkbas.LoadZkbasInstance(cli, zkbasContract)
	if err != nil {
		return 0, err
	}
	priorityRequests, err := zkbasInstance.ZkbasFilterer.
		FilterNewPriorityRequest(&bind.FilterOpts{Start: startHeight, End: &endHeight})
	if err != nil {
		return 0, err
	}
	priorityRequestCount := 0
	for priorityRequests.Next() {
		priorityRequestCount++
	}
	return priorityRequestCount, nil
}

func convertLogToNewPriorityRequestEvent(log types.Log) (*priorityrequest.PriorityRequest, error) {
	var event zkbas.ZkbasNewPriorityRequest
	if err := ZkbasContractAbi.UnpackIntoInterface(&event, EventNameNewPriorityRequest, log.Data); err != nil {
		return nil, err
	}
	request := &priorityrequest.PriorityRequest{
		L1TxHash:        log.TxHash.Hex(),
		L1BlockHeight:   int64(log.BlockNumber),
		SenderAddress:   event.Sender.Hex(),
		RequestId:       int64(event.SerialId),
		TxType:          int64(event.TxType),
		Pubdata:         common.Bytes2Hex(event.PubData),
		ExpirationBlock: event.ExpirationBlock.Int64(),
		Status:          priorityrequest.PendingStatus,
	}
	return request, nil
}
