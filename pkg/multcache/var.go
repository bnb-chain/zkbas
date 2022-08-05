package multcache

import (
	"errors"
	"fmt"
)

// error got from other package
var (
	errRedisCacheKeyNotExist = errors.New("redis: nil")
	errGoCacheKeyNotExist    = errors.New("Value not found in GoCache store")
)

const (
	KeyGetBlockByBlockHeight   = "cache:block:blockHeight"
	KeyGetBlockBlockCommitment = "cache::block:blockCommitment:"
	KeyGetBlockWithTxHeight    = "cache::block:blockHeightWithTx:"
	KeyGetBlockList            = "cache::block:blockList:"
	KeyGetCommittedBlocksCount = "cache::block:CommittedBlocksCount:"
	KeyGetVerifiedBlocksCount  = "cache::block:VerifiedBlocksCount:"
	KeyGetBlocksTotalCount     = "cache::block:BlocksTotalCount:"
	KeyGetCurrentBlockHeight   = "cache::block:GetCurrentBlockHeight:"

	KeyGetL2AssetsList               = "cache::l2asset:L2AssetsList:"
	KeyGetL2AssetInfoBySymbol        = "cache::l2asset:L2AssetInfoBySymbol:"
	KeyGetSimpleL2AssetInfoByAssetId = "cache::l2asset:SimpleL2AssetInfoByAssetId:"

	KeyGetSysconfigByName = "cache::sysconf:GetSysconfigByName:"
)

// cache key prefix: account
func SpliceCacheKeyAccountByAccountName(accountName string) string {
	return "cache:account_accountName_" + accountName
}

func SpliceCacheKeyAccountByAccountPk(accountPk string) string {
	return "cache:account_accountPk_" + accountPk
}

func SpliceCacheKeyBasicAccountByAccountIndex(accountIndex int64) string {
	return fmt.Sprintf("cache:basicAccount_accountIndex_%d", accountIndex)
}

func SpliceCacheKeyAccountByAccountIndex(accountIndex int64) string {
	return fmt.Sprintf("cache:account_accountIndex_%d", accountIndex)
}

// cache key prefix: tx
func SpliceCacheKeyTxsCount() string {
	return "cache:txsCount"
}

func SpliceCacheKeyTxByTxHash(txHash string) string {
	return "cache:tx_txHash" + txHash
}

func SpliceCacheKeyTxByTxId(txID int64) string {
	return fmt.Sprintf("cache:tx_txId_%d", txID)
}

func SpliceCacheKeyTxCountByTimeRange(data string) string {
	return "cache:txCount_" + data
}

// cache key prefix: liquidity
func SpliceCacheKeyLiquidityForReadByPairIndex(pairIndex int64) string {
	return fmt.Sprintf("cache:liquidity_pairIndex_%d", pairIndex)
}

func SpliceCacheKeyLiquidityInfoForWriteByPairIndex(pairIndex int64) string {
	return fmt.Sprintf("cache:liquidity_pairIndex_%d", pairIndex)
}

// cache key prefix: nft

func SpliceCacheKeyNftInfoByNftIndex(nftIndex int64) string {
	return fmt.Sprintf("cache:nftInfo_nftIndex_%d", nftIndex)
}

func SpliceCacheKeyAccountTotalNftCount(accountIndex int64) string {
	return fmt.Sprintf("cache:account_nftTotalCount_%d", accountIndex)
}

func SpliceCacheKeyAccountNftList(accountIndex int64, offset, limit int64) string {
	return fmt.Sprintf("cache:account_nftList_%d_%d_%d", accountIndex, offset, limit)
}

// cache key prefix: price
func SpliceCacheKeyCurrencyPrice() string {
	return "cache:currencyPrice:"

}
