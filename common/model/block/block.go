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

package block

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/errorcode"
)

var (
	cacheBlockIdPrefix = "cache::block:id:"

	CacheBlockStatusPrefix         = "cache::block:blockStatus:"
	cacheBlockListLimitPrefix      = "cache::block:blockList:"
	cacheBlockCommittedCountPrefix = "cache::block:committed_count"
	cacheBlockVerifiedCountPrefix  = "cache::block:verified_count"
)

type (
	BlockModel interface {
		CreateBlockTable() error
		DropBlockTable() error
		GetBlocksList(limit int64, offset int64) (blocks []*Block, err error)
		GetBlocksBetween(start int64, end int64) (blocks []*Block, err error)
		GetBlocksForSender(status int, limit int) (blocks []*Block, err error)
		GetBlocksForSenderBetween(start int64, end int64, status int, maxBlocksCount int) (blocks []*Block, err error)
		GetBlocksForSenderHigherThanBlockHeight(blockHeight int64, status int, limit int) (blocks []*Block, err error)
		GetBlocksLowerThanHeight(end int64, status int) (rowsAffected int64, blocks []*Block, err error)
		GetBlocksHigherThanBlockHeight(blockHeight int64) (blocks []*Block, err error)
		GetBlockByCommitment(blockCommitment string) (block *Block, err error)
		GetBlockByBlockHeight(blockHeight int64) (block *Block, err error)
		GetBlockByBlockHeightWithoutTx(blockHeight int64) (block *Block, err error)
		GetNotVerifiedOrExecutedBlocks() (blocks []*Block, err error)
		GetCommittedBlocksCount() (count int64, err error)
		GetVerifiedBlocksCount() (count int64, err error)
		GetLatestVerifiedBlockHeight() (height int64, err error)
		GetBlocksForProverBetween(start, end int64) (blocks []*Block, err error)
		CreateBlock(block *Block) error
		CreateGenesisBlock(block *Block) error
		UpdateBlock(block *Block) error
		GetCurrentBlockHeight() (blockHeight int64, err error)
		GetBlocksTotalCount() (count int64, err error)
		UpdateBlockStatusCacheByBlockHeight(blockHeight int64, blockStatusInfo *BlockStatusInfo) error
		GetBlockStatusCacheByBlockHeight(blockHeight int64) (blockStatusInfo *BlockStatusInfo, err error)
		CreateBlockForCommitter(
			oBlock *Block,
			oBlockForCommit *blockForCommit.BlockForCommit,
			pendingMempoolTxs []*mempool.MempoolTx,
			pendingDeleteMempoolTxs []*mempool.MempoolTx,
			pendingUpdateAccounts []*account.Account,
			pendingNewAccountHistories []*account.AccountHistory,
			pendingUpdateLiquidity []*liquidity.Liquidity,
			pendingNewLiquidityHistories []*liquidity.LiquidityHistory,
			pendingUpdateNft []*nft.L2Nft,
			pendingNewNftHistories []*nft.L2NftHistory,
			pendingNewNftWithdrawHistories []*nft.L2NftWithdrawHistory,
		) (err error)
	}

	defaultBlockModel struct {
		sqlc.CachedConn
		table     string
		DB        *gorm.DB
		RedisConn *redis.Redis
	}

	Block struct {
		gorm.Model
		BlockSize uint16
		// pubdata
		BlockCommitment                 string
		BlockHeight                     int64 `gorm:"uniqueIndex"`
		StateRoot                       string
		PriorityOperations              int64
		PendingOnChainOperationsHash    string
		PendingOnChainOperationsPubData string
		CommittedTxHash                 string
		CommittedAt                     int64
		VerifiedTxHash                  string
		VerifiedAt                      int64
		Txs                             []*tx.Tx `gorm:"foreignKey:BlockId"`
		BlockStatus                     int64
	}
)

func NewBlockModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB, redisConn *redis.Redis) BlockModel {
	return &defaultBlockModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      BlockTableName,
		DB:         db,
		RedisConn:  redisConn,
	}
}

func (*Block) TableName() string {
	return BlockTableName
}

/*
	Func: CreateBlockTable
	Params:
	Return: err error
	Description: create Block table
*/

func (m *defaultBlockModel) CreateBlockTable() error {
	return m.DB.AutoMigrate(Block{})
}

/*
	Func: DropBlockTable
	Params:
	Return: err error
	Description: drop block table
*/

func (m *defaultBlockModel) DropBlockTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: GetBlocksList
	Params: limit int64, offset int64
	Return: err error
	Description:  For API /api/v1/block/getBlocksList

*/
func (m *defaultBlockModel) GetBlocksList(limit int64, offset int64) (blocks []*Block, err error) {
	var (
		//blockForeignKeyColumn = `BlockDetails`
		txForeignKeyColumn = `Txs`
	)
	key := fmt.Sprintf("%s%v:%v", cacheBlockListLimitPrefix, limit, offset)
	cacheBlockListLimitVal, err := m.RedisConn.Get(key)

	if err != nil {
		errInfo := fmt.Sprintf("[block.GetBlocksList] Get Redis Error: %s, key:%s", err.Error(), key)
		logx.Errorf(errInfo)
		return nil, err
	} else if cacheBlockListLimitVal == "" {
		dbTx := m.DB.Table(m.table).Limit(int(limit)).Offset(int(offset)).Order("block_height desc").Find(&blocks)
		if dbTx.Error != nil {
			logx.Errorf("[block.GetBlocksList] %s", dbTx.Error.Error())
			return nil, errorcode.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			logx.Error("[block.GetBlocksList] Get Blocks Error")
			return nil, errorcode.DbErrNotFound
		}

		for _, block := range blocks {
			cacheBlockIdKey := fmt.Sprintf("%s%v", cacheBlockIdPrefix, block.ID)
			cacheBlockIdVal, err := m.RedisConn.Get(cacheBlockIdKey)
			if err != nil {
				errInfo := fmt.Sprintf("[block.GetBlocksList] Get Redis Error: %s, key:%s", err.Error(), key)
				logx.Errorf(errInfo)
				return nil, err
			} else if cacheBlockIdVal == "" {
				/*
					err = m.DB.Model(&block).Association(blockForeignKeyColumn).Find(&block.BlockDetails)
					if err != nil {
						logx.Error("[block.GetBlocksList] Get Associate BlockDetails Error")
						return nil, err
					}
				*/
				txLength := m.DB.Model(&block).Association(txForeignKeyColumn).Count()
				block.Txs = make([]*tx.Tx, txLength)

				// json string
				jsonString, err := json.Marshal(block)
				if err != nil {
					logx.Errorf("[block.GetBlocksList] json.Marshal Error: %s, value: %v", err.Error(), block)
					return nil, err
				}
				// todo
				err = m.RedisConn.Setex(key, string(jsonString), 60)
				if err != nil {
					logx.Errorf("[block.GetBlocksList] redis set error: %s", err.Error())
					return nil, err
				}
			} else {
				// json string unmarshal
				var (
					nBlock *Block
				)
				err = json.Unmarshal([]byte(cacheBlockIdVal), &nBlock)
				if err != nil {
					logx.Errorf("[tblock.GetBlocksList] json.Unmarshal error: %s, value : %s", err.Error(), cacheBlockIdVal)
					return nil, err
				}
				block = nBlock
			}
		}
		// json string
		jsonString, err := json.Marshal(blocks)
		if err != nil {
			logx.Errorf("[block.GetBlocksList] json.Marshal Error: %s, value: %v", err.Error(), blocks)
			return nil, err
		}
		// todo
		err = m.RedisConn.Setex(key, string(jsonString), 30)
		if err != nil {
			logx.Errorf("[block.GetBlocksList] redis set error: %s", err.Error())
			return nil, err
		}

	} else {
		// json string unmarshal
		var (
			nBlocks []*Block
		)
		err = json.Unmarshal([]byte(cacheBlockListLimitVal), &nBlocks)
		if err != nil {
			logx.Errorf("[block.GetBlocksList] json.Unmarshal error: %s, value : %s", err.Error(), cacheBlockListLimitVal)
			return nil, err
		}
		blocks = nBlocks
	}

	return blocks, nil
}

/*
	Func: GetBlocksForSender
	Params: limit int64
	Return: err error
	Description:  For API /api/v1/block/getBlocksList

*/
func (m *defaultBlockModel) GetBlocksForSender(status int, limit int) (blocks []*Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status = ?", status).Limit(limit).Order("block_height").Find(&blocks)
	if dbTx.Error != nil {
		logx.Errorf("[block.GetBlocksList] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[block.GetBlocksList] Get Blocks Error")
		return nil, errorcode.DbErrNotFound
	}
	return blocks, nil
}

func (m *defaultBlockModel) GetBlocksForSenderBetween(start int64, end int64, status int, maxBlocksCount int) (blocks []*Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status = ? AND block_height > ? AND block_height <= ?", status, start, end).
		Order("block_height").
		Limit(maxBlocksCount).
		Find(&blocks)
	if dbTx.Error != nil {
		logx.Errorf("[block.GetBlocksList] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[block.GetBlocksList] Get Blocks Error")
		return nil, errorcode.DbErrNotFound
	}
	return blocks, nil
}

func (m *defaultBlockModel) GetBlocksBetween(start int64, end int64) (blocks []*Block, err error) {
	var (
		txForeignKeyColumn        = `Txs`
		txDetailsForeignKeyColumn = `TxDetails`
	)
	dbTx := m.DB.Table(m.table).Where("block_height >= ? AND block_height <= ?", start, end).
		Order("block_height").
		Find(&blocks)
	if dbTx.Error != nil {
		logx.Errorf("[block.GetBlocksList] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[block.GetBlocksList] Blocks not found")
		return nil, errorcode.DbErrNotFound
	}

	for _, block := range blocks {
		err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
		if err != nil {
			logx.Error("[block.GetBlocksList] Get Associate Txs Error")
			return nil, err
		}
		sort.Slice(block.Txs, func(i, j int) bool {
			return block.Txs[i].TxIndex < block.Txs[j].TxIndex
		})

		for _, txInfo := range block.Txs {
			err = m.DB.Model(&txInfo).Association(txDetailsForeignKeyColumn).Find(&txInfo.TxDetails)
			if err != nil {
				logx.Error("[block.GetBlocksList] Get Associate Tx details Error")
				return nil, err
			}
			sort.Slice(txInfo.TxDetails, func(i, j int) bool {
				return txInfo.TxDetails[i].Order < txInfo.TxDetails[j].Order
			})
		}
	}
	return blocks, nil
}

func (m *defaultBlockModel) GetBlocksLowerThanHeight(end int64, status int) (rowsAffected int64, blocks []*Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status = ? AND block_height <= ?", status, end).Order("block_height").Find(&blocks)
	if dbTx.Error != nil {
		logx.Errorf("[block.GetBlocksLowerThanHeight] %s", dbTx.Error.Error())
		return 0, nil, dbTx.Error
	}
	return dbTx.RowsAffected, blocks, nil
}

func (m *defaultBlockModel) GetBlocksForSenderHigherThanBlockHeight(blockHeight int64, status int, limit int) (blocks []*Block, err error) {
	var (
		txForeignKeyColumn = `Txs`
	)
	dbTx := m.DB.Table(m.table).Limit(limit).Where("block_height > ? AND block_status = ?", blockHeight, status).Order("block_height").Find(&blocks)
	if dbTx.Error != nil {
		logx.Errorf("[block.GetBlocksList] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[block.GetBlocksList] Get Blocks Error")
		return nil, errorcode.DbErrNotFound
	}
	for _, block := range blocks {
		err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
		sort.Slice(block.Txs, func(i, j int) bool {
			return block.Txs[i].TxIndex < block.Txs[j].TxIndex
		})
		if err != nil {
			logx.Error("[block.GetBlocksList] Get Associate Txs Error")
			return nil, err
		}
	}
	return blocks, nil
}

/*
	Func: GetBlocksList
	Params: limit int64, offset int64
	Return: err error
	Description:  For API /api/v1/block/getBlocksList

*/
func (m *defaultBlockModel) GetBlocksHigherThanBlockHeight(blockHeight int64) (blocks []*Block, err error) {
	var (
		txForeignKeyColumn = `Txs`
	)
	dbTx := m.DB.Table(m.table).Where("block_height > ?", blockHeight).Order("block_height desc").Find(&blocks)
	if dbTx.Error != nil {
		logx.Errorf("[block.GetBlocksHigherThanBlockHeight] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[block.GetBlocksHigherThanBlockHeight] Get Blocks Error")
		return nil, errorcode.DbErrNotFound
	}
	for _, block := range blocks {
		err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
		sort.Slice(block.Txs, func(i, j int) bool {
			return block.Txs[i].TxIndex < block.Txs[j].TxIndex
		})
		if err != nil {
			logx.Error("[block.GetBlocksHigherThanBlockHeight] Get Associate Txs Error")
			return nil, err
		}
	}
	return blocks, nil
}

/*
	Func: GetBlockByCommitment
	Params: blockCommitment string
	Return: err error
	Description:  For API /api/v1/block/getBlockByCommitment
*/
func (m *defaultBlockModel) GetBlockByCommitment(blockCommitment string) (block *Block, err error) {
	var (
		txForeignKeyColumn = `Txs`
	)
	dbTx := m.DB.Table(m.table).Where("block_commitment = ?", blockCommitment).Find(&block)
	if dbTx.Error != nil {
		logx.Errorf("[block.GetBlockByCommitment] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[block.GetBlockByCommitment] Get Block Error")
		return nil, errorcode.DbErrNotFound
	}
	err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
	sort.Slice(block.Txs, func(i, j int) bool {
		return block.Txs[i].TxIndex < block.Txs[j].TxIndex
	})
	if err != nil {
		logx.Error("[block.GetBlockByCommitment] Get Associate Txs Error")
		return nil, err
	}
	return block, nil
}

/*
	Func: GetBlockByBlockStatus
	Params: blockStatus int64
	Return: err error
*/
func (m *defaultBlockModel) GetNotVerifiedOrExecutedBlocks() (blocks []*Block, err error) {
	var (
		txForeignKeyColumn = `Txs`
	)
	dbTx := m.DB.Table(m.table).Where("block_status < ?", StatusVerifiedAndExecuted).Find(&blocks)
	if dbTx.Error != nil {
		logx.Errorf("[block.GetBlockByBlockHeight] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[block.GetBlockByBlockHeight] Get Block Error")
		return nil, errorcode.DbErrNotFound
	}
	for _, block := range blocks {
		err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
		sort.Slice(block.Txs, func(i, j int) bool {
			return block.Txs[i].TxIndex < block.Txs[j].TxIndex
		})
		if err != nil {
			logx.Error("[block.GetBlockByBlockHeight] Get Associate Txs Error")
			return nil, err
		}
	}
	return blocks, nil
}

/*
	Func: GetBlockByBlockHeight
	Params: blockHeight int64
	Return: err error
	Description:  For API /api/v1/block/getBlockByBlockHeight
*/
func (m *defaultBlockModel) GetBlockByBlockHeight(blockHeight int64) (block *Block, err error) {
	var (
		txForeignKeyColumn = `Txs`
	)
	dbTx := m.DB.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		logx.Errorf("[block.GetBlockByBlockHeight] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[block.GetBlockByBlockHeight] Get Block Error")
		return nil, errorcode.DbErrNotFound
	}
	err = m.DB.Model(&block).Association(txForeignKeyColumn).Find(&block.Txs)
	sort.Slice(block.Txs, func(i, j int) bool {
		return block.Txs[i].TxIndex < block.Txs[j].TxIndex
	})
	if err != nil {
		logx.Error("[block.GetBlockByBlockHeight] Get Associate Txs Error")
		return nil, err
	}

	return block, nil
}

/*
	Func: GetBlockByBlockHeightWithoutTx
	Params: blockHeight int64
	Return: err error
	Description:  For API /api/v1/block/getBlockByBlockHeight
*/
func (m *defaultBlockModel) GetBlockByBlockHeightWithoutTx(blockHeight int64) (block *Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_height = ?", blockHeight).Find(&block)
	if dbTx.Error != nil {
		logx.Errorf("[block.GetBlockByBlockHeight] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[block.GetBlockByBlockHeight] Get Block Error")
		return nil, errorcode.DbErrNotFound
	}
	return block, nil
}

/*
	Func: GetCommitedBlocksCount
	Params:
	Return: count int64, err error
	Description:  For API /api/v1/info/getLayer2BasicInfo
*/
func (m *defaultBlockModel) GetCommittedBlocksCount() (count int64, err error) {
	key := fmt.Sprintf("%s", cacheBlockCommittedCountPrefix)
	val, err := m.RedisConn.Get(key)
	if err != nil {
		errInfo := fmt.Sprintf("[block.GetCommittedBlocksCount] Get Redis Error: %s, key:%s", err.Error(), key)
		logx.Errorf(errInfo)
		return 0, err

	} else if val == "" {
		dbTx := m.DB.Table(m.table).Where("block_status >= ? and deleted_at is NULL", StatusCommitted).Count(&count)

		if dbTx.Error != nil {
			if dbTx.Error == errorcode.DbErrNotFound {
				return 0, nil
			}
			logx.Error("[block.GetCommittedBlocksCount] Get block Count Error")
			return 0, err
		}

		err = m.RedisConn.Setex(key, strconv.FormatInt(count, 10), 120)
		if err != nil {
			logx.Errorf("[block.GetCommittedBlocksCount] redis set error: %s", err.Error())
			return 0, err
		}
	} else {
		count, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			logx.Errorf("[block.GetCommittedBlocksCount] strconv.ParseInt error: %s, value : %s", err.Error(), val)
			return 0, err
		}
	}

	return count, nil

}

/*
	Func: GetVerifiedBlocksCount
	Params:
	Return: count int64, err error
	Description:  For API /api/v1/info/getLayer2BasicInfo
*/
func (m *defaultBlockModel) GetVerifiedBlocksCount() (count int64, err error) {
	key := fmt.Sprintf("%s", cacheBlockVerifiedCountPrefix)
	val, err := m.RedisConn.Get(key)
	if err != nil {
		errInfo := fmt.Sprintf("[block.GetVerifiedBlocksCount] Get Redis Error: %s, key:%s", err.Error(), key)
		logx.Errorf(errInfo)
		return 0, err

	} else if val == "" {
		dbTx := m.DB.Table(m.table).Where("block_status = ? and deleted_at is NULL", StatusVerifiedAndExecuted).Count(&count)

		if dbTx.Error != nil {
			if dbTx.Error == errorcode.DbErrNotFound {
				return 0, nil
			}
			logx.Error("[block.GetVerifiedBlocksCount] Get block Count Error")
			return 0, err
		}

		err = m.RedisConn.Setex(key, strconv.FormatInt(count, 10), 120)
		if err != nil {
			logx.Errorf("[block.GetVerifiedBlocksCount] redis set error: %s", err.Error())
			return 0, err
		}
	} else {
		count, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			logx.Errorf("[block.GetVerifiedBlocksCount] strconv.ParseInt error: %s, value : %s", err.Error(), val)
			return 0, err
		}
	}

	return count, nil
}

/*
	Func: CreateBlock
	Params: *Block
	Return: error
	Description: Insert Block when committerProto completing packing new L2Block.
*/
func (m *defaultBlockModel) CreateBlock(block *Block) error {
	dbTx := m.DB.Table(m.table).Create(block)

	if dbTx.Error != nil {
		logx.Errorf("[block.CreateBlock] %s", dbTx.Error.Error())
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Error("[block.CreateBlock] Create Invalid Block")
		return errorcode.DbErrFailToCreateBlock
	}
	return nil
}

func (m *defaultBlockModel) CreateGenesisBlock(block *Block) error {
	dbTx := m.DB.Table(m.table).Omit("BlockDetails").Omit("Txs").Create(block)

	if dbTx.Error != nil {
		logx.Errorf("[block.CreateBlock] %s", dbTx.Error.Error())
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Error("[block.CreateBlock] Create Invalid Block")
		return errorcode.DbErrFailToCreateBlock
	}
	return nil
}

/*
	Func: UpdateBlock
	Params: *Block
	Return: error
	Description: Update Block when committer completing packing new L2Block. And insert txVerification
*/
func (m *defaultBlockModel) UpdateBlock(block *Block) error {
	dbTx := m.DB.Save(block)

	if dbTx.Error != nil {
		logx.Errorf("[block.UpdateBlock] %s", dbTx.Error.Error())
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Error("[block.UpdateBlock] Update Invalid Block")
		return errorcode.DbErrFailToCreateBlock
	}
	return nil
}

/*
	Func: GetCurrentBlockHeight
	Params:
	Return: blockHeight int64, err error
	Description: get latest block height
*/
func (m *defaultBlockModel) GetCurrentBlockHeight() (blockHeight int64, err error) {
	dbTx := m.DB.Table(m.table).Select("block_height").Order("block_height desc").Limit(1).Find(&blockHeight)
	if dbTx.Error != nil {
		logx.Errorf("[block.GetCurrentBlockHeight] %s", dbTx.Error.Error())
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[block.GetCurrentBlockHeight] No block yet")
		return 0, errorcode.DbErrNotFound
	}
	return blockHeight, nil
}

/*
	Func: GetBlocksTotalCount
	Params:
	Return: count int64, err error
	Description: used for counting total blocks for explorer dashboard
*/
func (m *defaultBlockModel) GetBlocksTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[block.GetBlocksTotalCount] %s", dbTx.Error.Error())
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[block.GetBlocksTotalCount] No Blocks in Block Table")
		return 0, nil
	}
	return count, nil
}

/*
	Func: UpdateBlockStatusCacheByBlockHeight
	Params: blockHeight int64, blockStatus int64
	Return: err
	Description: update blockStatus cache by blockHeight
*/
func (m *defaultBlockModel) UpdateBlockStatusCacheByBlockHeight(blockHeight int64, blockStatusInfo *BlockStatusInfo) error {
	key := fmt.Sprintf("%s%v", CacheBlockStatusPrefix, blockHeight)

	jsonBytes, err := json.Marshal(blockStatusInfo)
	if err != nil {
		logx.Errorf("[blockModel.UpdateBlockStatusCacheByBlockHeight] json.Marshal Error: %s, value: %v", err.Error(), blockStatusInfo)
		return err
	}
	err = m.RedisConn.Setex(key, string(jsonBytes), 60)
	if err != nil {
		logx.Errorf("[blockModel.UpdateBlockStatusCacheByBlockHeight] error: %s", err.Error())
		return err
	}

	logx.Infof("[blockModel.UpdateBlockStatusCacheByBlockHeight] Set Block Status Cache, BlockHeight: %d, BlockStatus: %s", blockHeight, string(jsonBytes))

	return nil
}

/*
	Func: GetBlockStatusCacheByBlockHeight
	Params: blockHeight int64
	Return: blockStatus int64, err
	Description: get blockStatus cache by blockHeight
*/

type BlockStatusInfo struct {
	BlockStatus int64
	CommittedAt int64
	VerifiedAt  int64
}

func (m *defaultBlockModel) GetBlockStatusCacheByBlockHeight(blockHeight int64) (blockStatusInfo *BlockStatusInfo, err error) {

	key := fmt.Sprintf("%s%v", CacheBlockStatusPrefix, blockHeight)
	blockStatusInfoFromCache, err := m.RedisConn.Get(key)
	if err != nil {
		errInfo := fmt.Sprintf("[blockModel.GetBlockStatusCacheByBlockHeight] %s %s", key, err)
		logx.Error(errInfo)
		return blockStatusInfo, err
	} else if blockStatusInfoFromCache == "" {
		errInfo := fmt.Sprintf("[blockModel.GetBlockStatusCacheByBlockHeight] %s not found", key)
		logx.Info(errInfo)
		return blockStatusInfo, errorcode.DbErrNotFound
	} else {
		err = json.Unmarshal([]byte(blockStatusInfoFromCache), &blockStatusInfo)
		if err != nil {
			logx.Errorf("[txVerification.GetBlockStatusCacheByBlockHeight] json.Unmarshal error: %s, value : %s", err.Error(), blockStatusInfoFromCache)
			return nil, err
		}
	}

	return blockStatusInfo, nil
}

func (m *defaultBlockModel) CreateBlockForCommitter(
	oBlock *Block,
	oBlockForCommit *blockForCommit.BlockForCommit,
	pendingMempoolTxs []*mempool.MempoolTx,
	pendingDeleteMempoolTxs []*mempool.MempoolTx,
	pendingUpdateAccounts []*account.Account,
	pendingNewAccountHistories []*account.AccountHistory,
	pendingUpdateLiquiditys []*liquidity.Liquidity,
	pendingNewLiquidityHistories []*liquidity.LiquidityHistory,
	pendingUpdateNfts []*nft.L2Nft,
	pendingNewNftHistories []*nft.L2NftHistory,
	pendingNewNftWithdrawHistory []*nft.L2NftWithdrawHistory,
) (err error) {
	err = m.DB.Transaction(func(tx *gorm.DB) error { // transact
		// create block
		if oBlock != nil {
			dbTx := tx.Table(m.table).Create(oBlock)
			if dbTx.Error != nil {
				logx.Errorf("[CreateBlockForCommitter] unable to create block: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				blockInfo, err := json.Marshal(oBlock)
				if err != nil {
					logx.Errorf("[CreateBlockForCommitter] unable to marshal block")
					return err
				}
				logx.Errorf("[CreateBlockForCommitter] invalid block info: %s", string(blockInfo))
				return errors.New("[CreateBlockForCommitter] invalid block info")
			}
		}
		if oBlockForCommit != nil {
			// create block for commit
			dbTx := tx.Table(blockForCommit.BlockForCommitTableName).Create(oBlockForCommit)
			if dbTx.Error != nil {
				logx.Errorf("[CreateBlockForCommitter] unable to create block for commit: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				blockInfo, err := json.Marshal(oBlockForCommit)
				if err != nil {
					logx.Errorf("[CreateBlockForCommitter] unable to marshal block for commit")
					return err
				}
				logx.Errorf("[CreateBlockForCommitter] invalid block for commit info: %s", string(blockInfo))
				return errors.New("[CreateBlockForCommitter] invalid block for commit info")
			}
		}
		// update mempool
		for _, mempoolTx := range pendingMempoolTxs {
			dbTx := tx.Table(mempool.MempoolTableName).Where("id = ?", mempoolTx.ID).
				Select("*").
				Updates(&mempoolTx)
			if dbTx.Error != nil {
				logx.Errorf("[CreateBlockForCommitter] unable to update mempool tx: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Errorf("[CreateBlockForCommitter] no new mempoolTx")
				return errors.New("[CreateBlockForCommitter] no new mempoolTx")
			}
		}
		for _, pendingDeleteMempoolTx := range pendingDeleteMempoolTxs {
			for _, detail := range pendingDeleteMempoolTx.MempoolDetails {
				dbTx := tx.Table(mempool.DetailTableName).Where("id = ?", detail.ID).Delete(&detail)
				if dbTx.Error != nil {
					logx.Errorf("[CreateBlockForCommitter] %s", dbTx.Error.Error())
					return dbTx.Error
				}
				if dbTx.RowsAffected == 0 {
					logx.Errorf("[CreateBlockForCommitter] Delete Invalid Mempool Tx")
					return errors.New("[CreateBlockForCommitter] Delete Invalid Mempool Tx")
				}
			}
			dbTx := tx.Table(mempool.MempoolTableName).Where("id = ?", pendingDeleteMempoolTx.ID).Delete(&pendingDeleteMempoolTx)
			if dbTx.Error != nil {
				logx.Errorf("[CreateBlockForCommitter] %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Error("[CreateBlockForCommitter] Delete Invalid Mempool Tx")
				return errors.New("[CreateBlockForCommitter] Delete Invalid Mempool Tx")
			}
		}
		// update account
		for _, pendignUpdateAccount := range pendingUpdateAccounts {
			dbTx := tx.Table(account.AccountTableName).Where("id = ?", pendignUpdateAccount.ID).
				Select("*").
				Updates(&pendignUpdateAccount)
			if dbTx.Error != nil {
				logx.Errorf("[CreateBlockForCommitter] unable to update account: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Errorf("[CreateBlockForCommitter] no new account")
				return errors.New("[CreateBlockForCommitter] no new account")
			}
		}
		// create new account history
		if len(pendingNewAccountHistories) != 0 {
			dbTx := tx.Table(account.AccountHistoryTableName).CreateInBatches(pendingNewAccountHistories, len(pendingNewAccountHistories))
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(pendingNewAccountHistories)) {
				logx.Errorf("[CreateBlockForCommitter] unable to create new account history")
				return errors.New("[CreateBlockForCommitter] unable to create new account history")
			}
		}
		// update liquidity
		for _, entity := range pendingUpdateLiquiditys {
			dbTx := tx.Table(liquidity.LiquidityTable).Where("id = ?", entity.ID).
				Select("*").
				Updates(&entity)
			if dbTx.Error != nil {
				logx.Errorf("[CreateBlockForCommitter] unable to update liquidity: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Errorf("[CreateBlockForCommitter] no new liquidity")
				return errors.New("[CreateBlockForCommitter] no new liquidity")
			}
		}
		// create new liquidity history
		if len(pendingNewLiquidityHistories) != 0 {
			dbTx := tx.Table(liquidity.LiquidityHistoryTable).CreateInBatches(pendingNewLiquidityHistories, len(pendingNewLiquidityHistories))
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(pendingNewLiquidityHistories)) {
				logx.Errorf("[CreateBlockForCommitter] unable to create new liquidity history")
				return errors.New("[CreateBlockForCommitter] unable to create new liquidity history")
			}
		}
		// new nft
		if len(pendingNewNftWithdrawHistory) != 0 {
			dbTx := tx.Table(nft.L2NftWithdrawHistoryTableName).CreateInBatches(pendingNewNftWithdrawHistory, len(pendingNewNftWithdrawHistory))
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(pendingNewNftWithdrawHistory)) {
				logx.Errorf("[CreateBlockForCommitter] unable to create new nft withdraw ")
				return errors.New("[CreateBlockForCommitter] unable to create new nft withdraw")
			}
		}
		// update nft
		for _, entity := range pendingUpdateNfts {
			dbTx := tx.Table(nft.L2NftTableName).Where("id = ?", entity.ID).
				Select("*").
				Updates(&entity)
			if dbTx.Error != nil {
				logx.Errorf("[CreateBlockForCommitter] unable to update nft: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Errorf("[CreateBlockForCommitter] no new nft")
				return errors.New("[CreateBlockForCommitter] no new nft")
			}
		}
		// new nft history
		if len(pendingNewNftHistories) != 0 {
			dbTx := tx.Table(nft.L2NftHistoryTableName).CreateInBatches(pendingNewNftHistories, len(pendingNewNftHistories))
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(pendingNewNftHistories)) {
				logx.Errorf("[CreateBlockForCommitter] unable to create new nft history")
				return errors.New("[CreateBlockForCommitter] unable to create new nft history")
			}
		}
		return nil
	})
	return err
}

func (m *defaultBlockModel) GetBlocksForProverBetween(start, end int64) (blocks []*Block, err error) {
	dbTx := m.DB.Table(m.table).Where("block_status = ? AND block_height >= ? AND block_height <= ?", StatusCommitted, start, end).
		Order("block_height").
		Find(&blocks)
	if dbTx.Error != nil {
		logx.Errorf("[GetBlocksForProverBetween] unable to get block between: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return blocks, nil
}

func (m *defaultBlockModel) GetLatestVerifiedBlockHeight() (height int64, err error) {
	block := &Block{}
	dbTx := m.DB.Table(m.table).Where("block_status = ?", StatusVerifiedAndExecuted).
		Order("block_height DESC").
		Limit(1).
		First(&block)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestVerifiedBlockHeight] unable to get block: %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, errorcode.DbErrNotFound
	}
	return block.BlockHeight, nil
}
