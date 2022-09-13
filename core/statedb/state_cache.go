package statedb

import (
	"github.com/bnb-chain/zkbnb/common/prove"
	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/zkbnb-crypto/legend/circuit/bn254/std"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

const (
	_ = iota
	StateCachePending
	StateCacheCached
)

type StateCache struct {
	StateRoot string
	// Updated in executor's GeneratePubData method.
	PubData                         []byte
	PriorityOperations              int64
	PubDataOffset                   []uint32
	PendingOnChainOperationsPubData [][]byte
	PendingOnChainOperationsHash    []byte
	Txs                             []*tx.Tx
	Witnesses                       []*prove.TxWitness

	// Record the flat states that should be updated.
	PendingNewAccountIndexMap      map[int64]int
	PendingNewLiquidityIndexMap    map[int64]int
	PendingNewNftIndexMap          map[int64]int
	PendingUpdateAccountIndexMap   map[int64]int
	PendingUpdateLiquidityIndexMap map[int64]int
	PendingUpdateNftIndexMap       map[int64]int

	// Record the tree states that should be updated.
	dirtyAccountsAndAssetsMap map[int64]map[int64]bool
	dirtyLiquidityMap         map[int64]bool
	dirtyNftMap               map[int64]bool
}

func NewStateCache(stateRoot string) *StateCache {
	return &StateCache{
		StateRoot: stateRoot,
		Txs:       make([]*tx.Tx, 0),
		Witnesses: make([]*prove.TxWitness, 0),

		PendingNewAccountIndexMap:      make(map[int64]int, 0),
		PendingNewLiquidityIndexMap:    make(map[int64]int, 0),
		PendingNewNftIndexMap:          make(map[int64]int, 0),
		PendingUpdateAccountIndexMap:   make(map[int64]int, 0),
		PendingUpdateLiquidityIndexMap: make(map[int64]int, 0),
		PendingUpdateNftIndexMap:       make(map[int64]int, 0),

		PubData:                         make([]byte, 0),
		PriorityOperations:              0,
		PubDataOffset:                   make([]uint32, 0),
		PendingOnChainOperationsPubData: make([][]byte, 0),
		PendingOnChainOperationsHash:    common.FromHex(types.EmptyStringKeccak),

		dirtyAccountsAndAssetsMap: make(map[int64]map[int64]bool, 0),
		dirtyLiquidityMap:         make(map[int64]bool, 0),
		dirtyNftMap:               make(map[int64]bool, 0),
	}
}

func (c *StateCache) AlignPubData(blockSize int) {
	emptyPubdata := make([]byte, (blockSize-len(c.Txs))*32*std.PubDataSizePerTx)
	c.PubData = append(c.PubData, emptyPubdata...)
}

func (c *StateCache) MarkAccountAssetsDirty(accountIndex int64, assets []int64) {
	if accountIndex < 0 {
		return
	}

	if _, ok := c.dirtyAccountsAndAssetsMap[accountIndex]; !ok {
		c.dirtyAccountsAndAssetsMap[accountIndex] = make(map[int64]bool, 0)
	}

	for _, assetIndex := range assets {
		// Should never happen, but protect here.
		if assetIndex < 0 {
			continue
		}
		c.dirtyAccountsAndAssetsMap[accountIndex][assetIndex] = true
	}
}

func (c *StateCache) MarkLiquidityDirty(pairIndex int64) {
	c.dirtyLiquidityMap[pairIndex] = true
}

func (c *StateCache) MarkNftDirty(nftIndex int64) {
	c.dirtyNftMap[nftIndex] = true
}
