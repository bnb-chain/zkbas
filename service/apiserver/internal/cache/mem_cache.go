package cache

import (
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"

	accdao "github.com/bnb-chain/zkbas/dao/account"
	assetdao "github.com/bnb-chain/zkbas/dao/asset"
	blockdao "github.com/bnb-chain/zkbas/dao/block"
	"github.com/bnb-chain/zkbas/dao/sysconfig"
	"github.com/bnb-chain/zkbas/dao/tx"
)

const (
	cacheDefaultExpiration    = time.Hour * 1   //gocache default expiration
	cacheDefaultPurgeInterval = time.Minute * 5 // gocache purge interval

	AccountIndexNameKeyPrefix  = "in:" //key for cache: accountIndex -> accountName
	AccountIndexPkKeyPrefix    = "ip:" //key for cache: accountIndex -> accountPk
	AccountNameKeyPrefix       = "n:"  //key for cache: accountName -> accountIndex
	AccountPkKeyPrefix         = "k:"  //key for cache: accountPk -> accountIndex
	AccountByIndexKeyPrefix    = "a:"  //key for cache: accountIndex -> account
	AccountCountKeyPrefix      = "ac"  //key for cache: total account count
	BlockByHeightKeyPrefix     = "h:"  //key for cache: blockHeight -> block
	BlockByCommitmentKeyPrefix = "c:"  //key for cache: blockCommitment -> block
	BlockCountKeyPrefix        = "bc"  //key for cache: total block count
	TxByHashKeyPrefix          = "h:"  //key for cache: txHash -> tx
	TxCountKeyPrefix           = "tc"  //key for cache: total tx count
	AssetCountKeyKeyPrefix     = "AC"  //key for cache: total asset count
	AssetIdNameKeyPrefix       = "IN:" //key for cache: assetId -> assetName
	AssetByIdKeyPrefix         = "I:"  //key for cache: assetId -> asset
	AssetBySymbolKeyPrefix     = "S:"  //key for cache: assetSymbol -> asset
	PriceKeyPrefix             = "p:"  //key for cache: symbol -> price
	SysConfigKeyPrefix         = "s:"  //key for cache: configName -> sysconfig
)

type fallback func() (interface{}, error)

type MemCache struct {
	goCache           *gocache.Cache
	accountModel      accdao.AccountModel
	assetModel        assetdao.AssetModel
	accountExpiration time.Duration
	blockExpiration   time.Duration
	txExpiration      time.Duration
	assetExpiration   time.Duration
	priceExpiration   time.Duration
}

func NewMemCache(accountModel accdao.AccountModel, assetModel assetdao.AssetModel,
	accountExpiration, blockExpiration, txExpiration,
	assetExpiration, priceExpiration int) *MemCache {
	memCache := &MemCache{
		goCache:           gocache.New(cacheDefaultExpiration, cacheDefaultPurgeInterval),
		accountModel:      accountModel,
		assetModel:        assetModel,
		accountExpiration: time.Duration(accountExpiration) * time.Millisecond,
		blockExpiration:   time.Duration(blockExpiration) * time.Millisecond,
		txExpiration:      time.Duration(txExpiration) * time.Millisecond,
		assetExpiration:   time.Duration(assetExpiration) * time.Millisecond,
		priceExpiration:   time.Duration(priceExpiration) * time.Millisecond,
	}
	return memCache
}

func (m *MemCache) getWithSet(key string, duration time.Duration, f fallback) (interface{}, error) {
	result, found := m.goCache.Get(key)
	if found {
		return result, nil
	}
	result, err := f()
	if err != nil {
		return nil, err
	}
	m.goCache.Set(key, result, duration)
	return result, nil
}

func (m *MemCache) setAccount(accountIndex int64, accountName, accountPk string) {
	m.goCache.Set(fmt.Sprintf("%s%d", AccountIndexNameKeyPrefix, accountIndex), accountName, gocache.DefaultExpiration)
	m.goCache.Set(fmt.Sprintf("%s%d", AccountIndexPkKeyPrefix, accountIndex), accountPk, gocache.DefaultExpiration)
	m.goCache.Set(fmt.Sprintf("%s%s", AccountNameKeyPrefix, accountName), accountIndex, gocache.DefaultExpiration)
	m.goCache.Set(fmt.Sprintf("%s%s", AccountPkKeyPrefix, accountPk), accountIndex, gocache.DefaultExpiration)
}

func (m *MemCache) GetAccountIndexByName(accountName string) (int64, error) {
	index, found := m.goCache.Get(fmt.Sprintf("%s%s", AccountNameKeyPrefix, accountName))
	if found {
		return index.(int64), nil
	}
	account, err := m.accountModel.GetAccountByName(accountName)
	if err != nil {
		return 0, err
	}
	m.setAccount(account.AccountIndex, account.AccountName, account.PublicKey)
	return account.AccountIndex, nil
}

func (m *MemCache) GetAccountIndexByPk(accountPk string) (int64, error) {
	index, found := m.goCache.Get(fmt.Sprintf("%s%s", AccountPkKeyPrefix, accountPk))
	if found {
		return index.(int64), nil
	}
	account, err := m.accountModel.GetAccountByPk(accountPk)
	if err != nil {
		return 0, err
	}
	m.setAccount(account.AccountIndex, account.AccountName, account.PublicKey)
	return account.AccountIndex, nil
}

func (m *MemCache) GetAccountNameByIndex(accountIndex int64) (string, error) {
	name, found := m.goCache.Get(fmt.Sprintf("%s%d", AccountIndexNameKeyPrefix, accountIndex))
	if found {
		return name.(string), nil
	}
	account, err := m.accountModel.GetAccountByIndex(accountIndex)
	if err != nil {
		return "", err
	}
	m.setAccount(account.AccountIndex, account.AccountName, account.PublicKey)
	return account.AccountName, nil
}

func (m *MemCache) GetAccountPkByIndex(accountIndex int64) (string, error) {
	pk, found := m.goCache.Get(fmt.Sprintf("%s%d", AccountIndexPkKeyPrefix, accountIndex))
	if found {
		return pk.(string), nil
	}
	account, err := m.accountModel.GetAccountByIndex(accountIndex)
	if err != nil {
		return "", err
	}
	m.setAccount(account.AccountIndex, account.AccountName, account.PublicKey)
	return account.PublicKey, nil
}

func (m *MemCache) GetAccountWithFallback(accountIndex int64, f fallback) (*accdao.Account, error) {
	key := fmt.Sprintf("%s%d", AccountByIndexKeyPrefix, accountIndex)
	a, err := m.getWithSet(key, m.accountExpiration, f)
	if err != nil {
		return nil, err
	}

	account := a.(*accdao.Account)
	m.setAccount(account.AccountIndex, account.AccountName, account.PublicKey)
	return account, nil
}

func (m *MemCache) GetAccountTotalCountWiltFallback(f fallback) (int64, error) {
	count, err := m.getWithSet(AccountCountKeyPrefix, m.accountExpiration, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetBlockByHeightWithFallback(blockHeight int64, f fallback) (*blockdao.Block, error) {
	key := fmt.Sprintf("%s%d", BlockByHeightKeyPrefix, blockHeight)
	b, err := m.getWithSet(key, m.blockExpiration, f)
	if err != nil {
		return nil, err
	}

	block := b.(*blockdao.Block)
	key = fmt.Sprintf("%s%s", BlockByCommitmentKeyPrefix, block.BlockCommitment)
	m.goCache.Set(key, block, m.blockExpiration)
	return block, nil
}

func (m *MemCache) GetBlockByCommitmentWithFallback(blockCommitment string, f fallback) (*blockdao.Block, error) {
	key := fmt.Sprintf("%s%s", BlockByCommitmentKeyPrefix, blockCommitment)
	b, err := m.getWithSet(key, m.blockExpiration, f)
	if err != nil {
		return nil, err
	}

	block := b.(*blockdao.Block)
	key = fmt.Sprintf("%s%d", BlockByHeightKeyPrefix, block.BlockHeight)
	m.goCache.Set(key, block, m.blockExpiration)
	return block, nil
}

func (m *MemCache) GetBlockTotalCountWithFallback(f fallback) (int64, error) {
	count, err := m.getWithSet(BlockCountKeyPrefix, m.blockExpiration, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetTxByHashWithFallback(txHash string, f fallback) (*tx.Tx, error) {
	key := fmt.Sprintf("%s%s", TxByHashKeyPrefix, txHash)
	t, err := m.getWithSet(key, m.txExpiration, f)
	if err != nil {
		return nil, err
	}
	return t.(*tx.Tx), nil
}

func (m *MemCache) GetTxTotalCountWithFallback(f fallback) (int64, error) {
	count, err := m.getWithSet(TxCountKeyPrefix, m.txExpiration, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetAssetTotalCountWithFallback(f fallback) (int64, error) {
	count, err := m.getWithSet(AssetCountKeyKeyPrefix, m.txExpiration, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetAssetByIdWithFallback(assetId int64, f fallback) (*assetdao.Asset, error) {
	key := fmt.Sprintf("%s%d", AssetByIdKeyPrefix, assetId)
	a, err := m.getWithSet(key, m.assetExpiration, f)
	if err != nil {
		return nil, err
	}

	asset := a.(*assetdao.Asset)
	key = fmt.Sprintf("%s%s", AssetBySymbolKeyPrefix, asset.AssetSymbol)
	m.goCache.Set(key, asset, m.assetExpiration)

	key = fmt.Sprintf("%s%d", AssetIdNameKeyPrefix, assetId)
	m.goCache.Set(key, asset.AssetName, gocache.DefaultExpiration)
	return asset, nil
}

func (m *MemCache) GetAssetBySymbolWithFallback(assetSymbol string, f fallback) (*assetdao.Asset, error) {
	key := fmt.Sprintf("%s%s", AssetBySymbolKeyPrefix, assetSymbol)
	a, err := m.getWithSet(key, m.assetExpiration, f)
	if err != nil {
		return nil, err
	}

	asset := a.(*assetdao.Asset)
	key = fmt.Sprintf("%s%d", AssetByIdKeyPrefix, asset.AssetId)
	m.goCache.Set(key, asset, m.assetExpiration)

	key = fmt.Sprintf("%s%d", AssetIdNameKeyPrefix, asset.AssetId)
	m.goCache.Set(key, asset.AssetName, gocache.DefaultExpiration)
	return asset, nil
}

func (m *MemCache) GetAssetNameById(assetId int64) (string, error) {
	key := fmt.Sprintf("%s%d", AssetIdNameKeyPrefix, assetId)
	name, found := m.goCache.Get(key)
	if found {
		return name.(string), nil
	}
	asset, err := m.assetModel.GetAssetById(assetId)
	if err != nil {
		return "", err
	}

	m.goCache.Set(key, asset.AssetName, gocache.DefaultExpiration)
	return asset.AssetName, nil
}

func (m *MemCache) GetPriceWithFallback(symbol string, f fallback) (float64, error) {
	key := fmt.Sprintf("%s%s", PriceKeyPrefix, symbol)
	price, err := m.getWithSet(key, m.priceExpiration, f)
	if err != nil {
		return 0, err
	}
	return price.(float64), nil
}

func (m *MemCache) GetSysConfigWithFallback(configName string, f fallback) (*sysconfig.SysConfig, error) {
	key := fmt.Sprintf("%s%s", SysConfigKeyPrefix, configName)
	c, err := m.getWithSet(key, gocache.DefaultExpiration, f)
	if err != nil {
		return nil, err
	}
	return c.(*sysconfig.SysConfig), nil
}
