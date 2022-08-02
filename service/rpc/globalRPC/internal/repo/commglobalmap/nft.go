package commglobalmap

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type model struct {
	mempoolModel         mempool.MempoolModel
	mempoolTxDetailModel mempool.MempoolTxDetailModel
	accountModel         account.AccountModel
	liquidityModel       liquidity.LiquidityModel
	redisConnection      *redis.Redis
	offerModel           nft.OfferModel
	nftModel             nft.L2NftModel
	cache                multcache.MultCache
}

func (m *model) GetLatestOfferIdForWrite(ctx context.Context, accountIndex int64) (int64, error) {
	lastOfferId, err := m.offerModel.GetLatestOfferId(accountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return 0, nil
		}
		return -1, err
	}
	return lastOfferId, nil
}

func (m *model) GetLatestNftInfoForRead(ctx context.Context, nftIndex int64) (*commonAsset.NftInfo, error) {
	dbNftInfo, err := m.nftModel.GetNftAsset(nftIndex)
	if err != nil {
		return nil, err
	}
	mempoolTxs, err := m.mempoolModel.GetPendingNftTxs()
	if err != nil && err != errorcode.DbErrNotFound {
		return nil, err
	}
	nftInfo := commonAsset.ConstructNftInfo(nftIndex, dbNftInfo.CreatorAccountIndex, dbNftInfo.OwnerAccountIndex, dbNftInfo.NftContentHash,
		dbNftInfo.NftL1TokenId, dbNftInfo.NftL1Address, dbNftInfo.CreatorTreasuryRate, dbNftInfo.CollectionId)
	for _, mempoolTx := range mempoolTxs {
		for _, txDetail := range mempoolTx.MempoolDetails {
			if txDetail.AssetType != commonAsset.NftAssetType || txDetail.AssetId != nftInfo.NftIndex {
				continue
			}
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.NftAssetType, nftInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				return nil, err
			}
			nftInfo, err = commonAsset.ParseNftInfo(nBalance)
			if err != nil {
				return nil, err
			}
		}
	}
	return nftInfo, nil
}

func (m *model) GetLatestNftInfoForReadWithCache(ctx context.Context, nftIndex int64) (*commonAsset.NftInfo, error) {
	f := func() (interface{}, error) {
		tmpNftInfo, err := m.GetLatestNftInfoForRead(ctx, nftIndex)
		if err != nil {
			return nil, err
		}
		return tmpNftInfo, nil
	}
	nftInfoType := &commonAsset.NftInfo{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyNftInfoByNftIndex(nftIndex), nftInfoType, multcache.NftTtl, f)
	if err != nil {
		return nil, err
	}
	nftInfo, _ := value.(*commonAsset.NftInfo)
	return nftInfo, nil
}

func (m *model) SetLatestNftInfoForReadInCache(ctx context.Context, nftIndex int64) error {
	nftInfo, err := m.GetLatestNftInfoForRead(ctx, nftIndex)
	if err != nil {
		return err
	}
	if err := m.cache.Set(ctx, multcache.SpliceCacheKeyNftInfoByNftIndex(nftIndex), nftInfo, multcache.NftTtl); err != nil {
		return err
	}
	return nil
}

func (m *model) DeleteLatestNftInfoForReadInCache(ctx context.Context, nftIndex int64) error {
	return m.cache.Delete(ctx, multcache.SpliceCacheKeyNftInfoByNftIndex(nftIndex))
}
