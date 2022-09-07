package statedb

import (
	"github.com/bnb-chain/zkbas/dao/compressedblock"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/dao/account"
	"github.com/bnb-chain/zkbas/dao/asset"
	"github.com/bnb-chain/zkbas/dao/block"
	"github.com/bnb-chain/zkbas/dao/liquidity"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/nft"
	"github.com/bnb-chain/zkbas/dao/tx"
)

type ChainDB struct {
	DB *gorm.DB
	// Block Chain data
	BlockModel           block.BlockModel
	CompressedBlockModel compressedblock.CompressedBlockModel
	TxModel              tx.TxModel

	// State DB
	AccountModel          account.AccountModel
	AccountHistoryModel   account.AccountHistoryModel
	L2AssetInfoModel      asset.AssetModel
	LiquidityModel        liquidity.LiquidityModel
	LiquidityHistoryModel liquidity.LiquidityHistoryModel
	L2NftModel            nft.L2NftModel
	L2NftHistoryModel     nft.L2NftHistoryModel
	MempoolModel          mempool.MempoolModel
}

func NewChainDB(db *gorm.DB) *ChainDB {
	return &ChainDB{
		DB:                   db,
		BlockModel:           block.NewBlockModel(db),
		CompressedBlockModel: compressedblock.NewCompressedBlockModel(db),
		TxModel:              tx.NewTxModel(db),

		AccountModel:          account.NewAccountModel(db),
		AccountHistoryModel:   account.NewAccountHistoryModel(db),
		L2AssetInfoModel:      asset.NewAssetModel(db),
		LiquidityModel:        liquidity.NewLiquidityModel(db),
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(db),
		L2NftModel:            nft.NewL2NftModel(db),
		L2NftHistoryModel:     nft.NewL2NftHistoryModel(db),
		MempoolModel:          mempool.NewMempoolModel(db),
	}
}
