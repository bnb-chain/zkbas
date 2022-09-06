package committer

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/core"
	"github.com/bnb-chain/zkbas/dao/block"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/tx"
)

const (
	MaxCommitterInterval = 60 * 1
)

type Config struct {
	core.ChainConfig

	BlockConfig struct {
		OptionalBlockSizes []int
	}
}

type Committer struct {
	config             *Config
	maxTxsPerBlock     int
	optionalBlockSizes []int

	bc *core.BlockChain

	executedMemPoolTxs []*mempool.MempoolTx
}

func NewCommitter(config *Config) (*Committer, error) {
	if len(config.BlockConfig.OptionalBlockSizes) == 0 {
		return nil, errors.New("nil optional block sizes")
	}

	bc, err := core.NewBlockChain(&config.ChainConfig, "committer")
	if err != nil {
		return nil, fmt.Errorf("new blockchain error: %v", err)
	}

	committer := &Committer{
		config:             config,
		maxTxsPerBlock:     config.BlockConfig.OptionalBlockSizes[len(config.BlockConfig.OptionalBlockSizes)-1],
		optionalBlockSizes: config.BlockConfig.OptionalBlockSizes,

		bc: bc,

		executedMemPoolTxs: make([]*mempool.MempoolTx, 0),
	}
	return committer, nil
}

func (c *Committer) Run() {
	curBlock, err := c.restoreExecutedTxs()
	if err != nil {
		panic("restore executed tx failed: " + err.Error())
	}

	for {
		if curBlock.BlockStatus > block.StatusProposing {
			curBlock, err = c.bc.ProposeNewBlock()
			if err != nil {
				panic("propose new block failed: " + err.Error())
			}
		}

		// Read pending transactions from mempool_tx table.
		pendingTxs, err := c.bc.MempoolModel.GetMempoolTxsByStatus(mempool.PendingTxStatus)
		if err != nil {
			logx.Error("get pending transactions from mempool failed:", err)
			return
		}
		for len(pendingTxs) == 0 {
			if c.shouldCommit(curBlock) {
				break
			}

			time.Sleep(100 * time.Millisecond)
			pendingTxs, err = c.bc.MempoolModel.GetMempoolTxsByStatus(mempool.PendingTxStatus)
			if err != nil {
				logx.Error("get pending transactions from mempool failed:", err)
				return
			}
		}

		pendingUpdateMempoolTxs := make([]*mempool.MempoolTx, 0, len(pendingTxs))
		pendingDeleteMempoolTxs := make([]*mempool.MempoolTx, 0, len(pendingTxs))
		for _, mempoolTx := range pendingTxs {
			if c.shouldCommit(curBlock) {
				break
			}

			tx := convertMempoolTxToTx(mempoolTx)
			err = c.bc.ApplyTransaction(tx)
			if err != nil {
				logx.Errorf("apply mempool tx ID: %d failed, err %v ", mempoolTx.ID, err)
				mempoolTx.Status = mempool.FailTxStatus
				pendingDeleteMempoolTxs = append(pendingDeleteMempoolTxs, mempoolTx)
				continue
			}
			mempoolTx.Status = mempool.ExecutedTxStatus
			pendingUpdateMempoolTxs = append(pendingUpdateMempoolTxs, mempoolTx)

			// Write the proposed block into database when the first transaction executed.
			if len(c.bc.Statedb.Txs) == 1 {
				err = c.createNewBlock(curBlock)
				if err != nil {
					panic("create new block failed" + err.Error())
				}
			}
		}

		err = c.bc.StateDB().SyncStateCacheToRedis()
		if err != nil {
			panic("sync redis cache failed: " + err.Error())
		}

		err = c.bc.MempoolModel.UpdateMempoolTxs(pendingUpdateMempoolTxs, pendingDeleteMempoolTxs)
		if err != nil {
			panic("update mempool failed: " + err.Error())
		}
		c.executedMemPoolTxs = append(c.executedMemPoolTxs, pendingUpdateMempoolTxs...)

		if c.shouldCommit(curBlock) {
			curBlock, err = c.commitNewBlock(curBlock)
			if err != nil {
				panic("commit new block failed: " + err.Error())
			}
		}
	}
}

func (c *Committer) restoreExecutedTxs() (*block.Block, error) {
	bc := c.bc
	curHeight, err := bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		return nil, err
	}
	curBlock, err := bc.BlockModel.GetBlockByHeight(curHeight)
	if err != nil {
		return nil, err
	}

	executedTxs, err := c.bc.MempoolModel.GetMempoolTxsByStatus(mempool.ExecutedTxStatus)
	if err != nil {
		return nil, err
	}

	if curBlock.BlockStatus > block.StatusProposing {
		if len(executedTxs) != 0 {
			return nil, errors.New("no proposing block but exist executed txs")
		}
		return curBlock, nil
	}

	for _, mempoolTx := range executedTxs {
		tx := convertMempoolTxToTx(mempoolTx)
		err = c.bc.ApplyTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	c.executedMemPoolTxs = append(c.executedMemPoolTxs, executedTxs...)
	return curBlock, nil
}

func (c *Committer) createNewBlock(curBlock *block.Block) error {
	return c.bc.BlockModel.CreateNewBlock(curBlock)
}

func (c *Committer) shouldCommit(curBlock *block.Block) bool {
	var now = time.Now()
	if (len(c.bc.Statedb.Txs) > 0 && now.Unix()-curBlock.CreatedAt.Unix() >= MaxCommitterInterval) ||
		len(c.bc.Statedb.Txs) >= c.maxTxsPerBlock {
		return true
	}

	return false
}

func (c *Committer) commitNewBlock(curBlock *block.Block) (*block.Block, error) {
	for _, tx := range c.executedMemPoolTxs {
		tx.Status = mempool.SuccessTxStatus
	}

	blockSize := c.computeCurrentBlockSize()
	blockStates, err := c.bc.CommitNewBlock(blockSize, curBlock.CreatedAt.UnixMilli())
	if err != nil {
		return nil, err
	}

	// update db
	err = c.bc.DB().DB.Transaction(func(tx *gorm.DB) error {
		// update mempool
		err := c.bc.DB().MempoolModel.UpdateMempoolTxsInTransact(tx, c.executedMemPoolTxs)
		if err != nil {
			return err
		}
		// update block
		err = c.bc.DB().BlockModel.UpdateBlocksInTransact(tx, []*block.Block{blockStates.Block})
		if err != nil {
			return err
		}
		// create block for commit
		if blockStates.CompressedBlock != nil {
			err = c.bc.DB().CompressedBlockModel.CreateCompressedBlockInTransact(tx, blockStates.CompressedBlock)
			if err != nil {
				return err
			}
		}
		// create new account
		if len(blockStates.PendingNewAccount) != 0 {
			err = c.bc.DB().AccountModel.CreateAccountsInTransact(tx, blockStates.PendingNewAccount)
			if err != nil {
				return err
			}
		}
		// update account
		if len(blockStates.PendingUpdateAccount) != 0 {
			err = c.bc.DB().AccountModel.UpdateAccountsInTransact(tx, blockStates.PendingUpdateAccount)
			if err != nil {
				return err
			}
		}
		// create new account history
		if len(blockStates.PendingNewAccountHistory) != 0 {
			err = c.bc.DB().AccountHistoryModel.CreateAccountHistoriesInTransact(tx, blockStates.PendingNewAccountHistory)
			if err != nil {
				return err
			}
		}
		// create new liquidity
		if len(blockStates.PendingNewLiquidity) != 0 {
			err = c.bc.DB().LiquidityModel.CreateLiquidityInTransact(tx, blockStates.PendingNewLiquidity)
			if err != nil {
				return err
			}
		}
		// update liquidity
		if len(blockStates.PendingUpdateLiquidity) != 0 {
			err = c.bc.DB().LiquidityModel.UpdateLiquidityInTransact(tx, blockStates.PendingUpdateLiquidity)
			if err != nil {
				return err
			}
		}
		// create new liquidity history
		if len(blockStates.PendingNewLiquidityHistory) != 0 {
			err = c.bc.DB().LiquidityHistoryModel.CreateLiquidityHistoriesInTransact(tx, blockStates.PendingNewLiquidityHistory)
			if err != nil {
				return err
			}
		}
		// create new nft
		if len(blockStates.PendingNewNft) != 0 {
			err = c.bc.DB().L2NftModel.CreateNftsInTransact(tx, blockStates.PendingNewNft)
			if err != nil {
				return err
			}
		}
		// update nft
		if len(blockStates.PendingUpdateNft) != 0 {
			err = c.bc.DB().L2NftModel.UpdateNftsInTransact(tx, blockStates.PendingUpdateNft)
			if err != nil {
				return err
			}
		}
		// new nft history
		if len(blockStates.PendingNewNftHistory) != 0 {
			err = c.bc.DB().L2NftHistoryModel.CreateNftHistoriesInTransact(tx, blockStates.PendingNewNftHistory)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	c.executedMemPoolTxs = make([]*mempool.MempoolTx, 0)
	return blockStates.Block, nil
}

func (c *Committer) computeCurrentBlockSize() int {
	var blockSize int
	for i := 0; i < len(c.optionalBlockSizes); i++ {
		if len(c.bc.Statedb.Txs) <= c.optionalBlockSizes[i] {
			blockSize = c.optionalBlockSizes[i]
			break
		}
	}
	return blockSize
}

func convertMempoolTxToTx(mempoolTx *mempool.MempoolTx) *tx.Tx {
	tx := &tx.Tx{
		TxHash:        mempoolTx.TxHash,
		TxType:        mempoolTx.TxType,
		GasFee:        mempoolTx.GasFee,
		GasFeeAssetId: mempoolTx.GasFeeAssetId,
		TxStatus:      tx.StatusPending,
		NftIndex:      mempoolTx.NftIndex,
		PairIndex:     mempoolTx.PairIndex,
		AssetId:       mempoolTx.AssetId,
		TxAmount:      mempoolTx.TxAmount,
		NativeAddress: mempoolTx.NativeAddress,
		TxInfo:        mempoolTx.TxInfo,
		ExtraInfo:     mempoolTx.ExtraInfo,
		Memo:          mempoolTx.Memo,
		AccountIndex:  mempoolTx.AccountIndex,
		Nonce:         mempoolTx.Nonce,
		ExpiredAt:     mempoolTx.ExpiredAt,
	}
	return tx
}
