package commglobalmap

import (
	"context"
	"fmt"
	"github.com/bnb-chain/zkbas/errorcode"
	"strconv"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

func (m *model) GetLatestAccountInfoWithCache(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error) {
	f := func() (interface{}, error) {
		accountInfo, err := m.GetLatestAccountInfo(ctx, accountIndex)
		if err != nil {
			return nil, err
		}
		account, err := commonAsset.FromFormatAccountInfo(accountInfo)
		if err != nil {
			return nil, errorcode.RepoErrFromFormatAccountInfo.RefineError(fmt.Sprint(accountInfo))
		}
		return account, nil
	}
	accountInfo := &account.Account{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex), accountInfo, 10, f)
	if err != nil {
		return nil, err
	}
	account, _ := value.(*account.Account)
	res, err := commonAsset.ToFormatAccountInfo(account)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *model) SetLatestAccountInfoInToCache(ctx context.Context, accountIndex int64) error {
	accountInfo, err := m.GetLatestAccountInfo(ctx, accountIndex)
	if err != nil {
		return err
	}
	account, err := commonAsset.FromFormatAccountInfo(accountInfo)
	if err != nil {
		return err
	}
	if err := m.cache.Set(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex), account, 10); err != nil {
		return err
	}
	return nil
}

func (m *model) DeleteLatestAccountInfoInCache(ctx context.Context, accountIndex int64) error {
	return m.cache.Delete(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex))
}

func (m *model) GetLatestAccountInfo(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error) {
	oAccountInfo, err := m.accountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		if err == account.ErrNotFound {

		}
		return nil, errorcode.RepoErrSqlOperation.RefineError(fmt.Sprintf("GetAccountByAccountIndex:%v", err))
	}
	accountInfo, err := commonAsset.ToFormatAccountInfo(oAccountInfo)
	if err != nil {
		return nil, errorcode.RepoErrToFormatAccountInfo.RefineError(err.Error())
	}
	mempoolTxs, err := m.mempoolModel.GetPendingMempoolTxsByAccountIndex(accountIndex)
	if err != nil && err != mempool.ErrNotFound {
		return nil, errorcode.RepoErrSqlOperation.RefineError(fmt.Sprintf("GetPendingMempoolTxsByAccountIndex:%v", err))
	}
	for _, mempoolTx := range mempoolTxs {
		if mempoolTx.Nonce != commonConstant.NilNonce {
			accountInfo.Nonce = mempoolTx.Nonce
		}
		for _, mempoolTxDetail := range mempoolTx.MempoolDetails {
			if mempoolTxDetail.AccountIndex != accountIndex {
				continue
			}
			switch mempoolTxDetail.AssetType {
			case commonAsset.GeneralAssetType:
				if accountInfo.AssetInfo[mempoolTxDetail.AssetId] == nil {
					accountInfo.AssetInfo[mempoolTxDetail.AssetId] = &commonAsset.AccountAsset{
						AssetId:                  mempoolTxDetail.AssetId,
						Balance:                  util.ZeroBigInt,
						LpAmount:                 util.ZeroBigInt,
						OfferCanceledOrFinalized: util.ZeroBigInt,
					}
				}
				nBalance, err := commonAsset.ComputeNewBalance(commonAsset.GeneralAssetType,
					accountInfo.AssetInfo[mempoolTxDetail.AssetId].String(), mempoolTxDetail.BalanceDelta)
				if err != nil {
					return nil, errorcode.RepoErrComputeNewBalance.RefineError(err.Error())
				}
				accountInfo.AssetInfo[mempoolTxDetail.AssetId], err = commonAsset.ParseAccountAsset(nBalance)
				if err != nil {
					return nil, errorcode.RepoErrParseAccountAsset.RefineError(err.Error())
				}
			case commonAsset.CollectionNonceAssetType:
				accountInfo.CollectionNonce, err = strconv.ParseInt(mempoolTxDetail.BalanceDelta, 10, 64)
				if err != nil {
					return nil, errorcode.RepoErrParseInt.RefineError(mempoolTxDetail.BalanceDelta)
				}
			case commonAsset.LiquidityAssetType:
			case commonAsset.NftAssetType:
			default:
				return nil, errorcode.RepoErrInvalidAssetType
			}
		}
	}
	accountInfo.Nonce = accountInfo.Nonce + 1
	accountInfo.CollectionNonce = accountInfo.CollectionNonce + 1
	return accountInfo, nil
}

func (m *model) GetBasicAccountInfoWithCache(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error) {
	f := func() (interface{}, error) {
		accountInfo, err := m.GetBasicAccountInfo(ctx, accountIndex)
		if err != nil {
			return nil, err
		}
		account, err := commonAsset.FromFormatAccountInfo(accountInfo)
		if err != nil {
			return nil, errorcode.RepoErrFromFormatAccountInfo.RefineError(fmt.Sprint(accountInfo))
		}
		return account, nil
	}
	accountInfo := &account.Account{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyBasicAccountByAccountIndex(accountIndex), accountInfo, 10, f)
	if err != nil {
		return nil, err
	}
	account, _ := value.(*account.Account)
	res, err := commonAsset.ToFormatAccountInfo(account)
	if err != nil {
		return nil, errorcode.RepoErrToFormatAccountInfo.RefineError(fmt.Sprint(accountInfo))
	}
	return res, nil
}

func (m *model) GetBasicAccountInfo(ctx context.Context, accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error) {
	oAccountInfo, err := m.accountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		return nil, errorcode.RepoErrSqlOperation.RefineError(fmt.Sprint("GetAccountByAccountIndex:", err.Error()))
	}
	accountInfo, err = commonAsset.ToFormatAccountInfo(oAccountInfo)
	if err != nil {
		return nil, errorcode.RepoErrToFormatAccountInfo.RefineError(fmt.Sprint(accountInfo))
	}
	return accountInfo, nil
}
