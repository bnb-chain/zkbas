package mempool

import (
	"context"
	"sort"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	table "github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

/*
	Func: GetMempoolTxs
	Params: offset uint64, limit uint64
	Return: mempoolTx []*mempoolModel.MempoolTx, err error
	Description: query txs from db that sit in the range
*/
func (m *model) GetMempoolTxs(offset int, limit int) (mempoolTxs []*table.MempoolTx, err error) {
	var mempoolForeignKeyColumn = `MempoolDetails`
	dbTx := m.db.Table(m.table).Where("status = ? and deleted_at is NULL", PendingTxStatus).Order("created_at, id").Offset(offset).Limit(limit).Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsList] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	}
	for _, mempoolTx := range mempoolTxs {
		err := m.db.Model(&mempoolTx).Association(mempoolForeignKeyColumn).Find(&mempoolTx.MempoolDetails)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxsList] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTxs, nil
}

func (m *model) GetMempoolTxsTotalCount() (count int64, err error) {
	dbTx := m.db.Table(m.table).Where("status = ? and deleted_at is NULL", PendingTxStatus).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[txVerification.GetTxsTotalCount] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *model) GetMempoolTxByTxHash(hash string) (mempoolTx *table.MempoolTx, err error) {
	var mempoolForeignKeyColumn = `MempoolDetails`
	dbTx := m.db.Table(m.table).Where("status = ? and tx_hash = ?", PendingTxStatus, hash).Find(&mempoolTx)
	if dbTx.Error != nil {
		logx.Errorf("[GetMempoolTxByTxHash] %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	if err = m.db.Model(&mempoolTx).Association(mempoolForeignKeyColumn).Find(&mempoolTx.MempoolDetails); err != nil {
		logx.Errorf("[mempool.GetMempoolTxByTxHash] Get Associate MempoolDetails Error")
		return nil, err
	}
	return mempoolTx, nil
}

func (m *model) GetMempoolTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error) {
	var (
		mempoolDetailTable = `mempool_tx_detail`
		mempoolIds         []int64
	)
	var mempoolTxDetails []*table.MempoolTxDetail
	dbTx := m.db.Table(mempoolDetailTable).Select("tx_id").Where("account_index = ?", accountIndex).Find(&mempoolTxDetails).Group("tx_id").Find(&mempoolIds)
	if dbTx.Error != nil {
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	dbTx = m.db.Table(m.table).Where("status = ? and id in (?) and deleted_at is NULL", PendingTxStatus, mempoolIds).Count(&count)
	if dbTx.Error != nil {
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *model) GetMempoolTxByTxId(ctx context.Context, txID int64) (*table.MempoolTx, error) {
	f := func() (interface{}, error) {
		tx := &table.MempoolTx{}
		dbTx := m.db.Table(m.table).Where("id = ? and deleted_at is NULL", txID).Find(&tx)
		if dbTx.Error != nil {
			logx.Errorf("fail to get mempool tx by id: %d, error: %s", txID, dbTx.Error.Error())
			return nil, errorcode.DbErrSqlOperation
		} else if dbTx.RowsAffected == 0 {
			return nil, errorcode.DbErrNotFound
		}
		err := m.db.Model(&tx).Association(`MempoolDetails`).Find(&tx.MempoolDetails)
		if err != nil {
			return nil, err
		}
		sort.SliceStable(tx.MempoolDetails, func(i, j int) bool {
			return tx.MempoolDetails[i].Order < tx.MempoolDetails[j].Order
		})
		return tx, nil
	}
	tx := &table.MempoolTx{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyTxByTxId(txID), tx, multcache.MempoolTxTtl, f)
	if err != nil {
		return nil, err
	}
	tx, _ = value.(*table.MempoolTx)
	return tx, nil
}
