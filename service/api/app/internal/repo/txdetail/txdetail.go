package txdetail

import (
	"context"
	"sort"
	"time"

	table "github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/errcode"
	"gorm.io/gorm"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

func (m *model) GetTxsTotalCountByAccountIndex(ctx context.Context, accountIndex int64) (int64, error) {
	f := func() (interface{}, error) {
		var count int64
		dbTx := m.db.Table(m.table).Select("tx_id").
			Where("account_index = ? and deleted_at is NULL", accountIndex).Group("tx_id").Count(&count)
		if dbTx.Error != nil {
			return nil, dbTx.Error
		} else if dbTx.RowsAffected == 0 {
			return nil, nil
		}
		return &count, nil
	}
	var countType int64
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyTxsCount(), &countType, 5, f)
	if err != nil {
		return 0, err
	}
	count, _ := value.(*int64)
	return *count, nil
}

func (m *model) GetTxDetailByAccountIndex(ctx context.Context, accountIndex int64) ([]*table.TxDetail, error) {
	result := make([]*table.TxDetail, 0)
	dbTx := m.db.Table(m.table).Where("account_index = ?", accountIndex).Find(&result)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, errcode.ErrDataNotExist
	}
	return result, nil
}

func (m *model) GetTxIdsByAccountIndex(ctx context.Context, accountIndex int64) ([]int64, error) {
	txIds := make([]int64, 0)
	dbTx := m.db.Table(m.table).Select("tx_id").Where("account_index = ?", accountIndex).Group("tx_id").Find(&txIds)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, errcode.ErrDataNotExist
	}
	sort.Slice(txIds, func(i, j int) bool {
		return txIds[i] > txIds[j]
	})
	return txIds, nil
}

func (m *model) GetDauInTxDetail(ctx context.Context, data string) (count int64, err error) {
	var (
		from time.Time
		to   time.Time
	)
	now := time.Now()
	today := now.Round(24 * time.Hour).Add(-8 * time.Hour)
	switch data {
	case "yesterday":
		from = today.Add(-24 * time.Hour)
		to = today
	case "today":
		from = today
		to = now
	}
	dbTx := m.db.Raw("SELECT account_index FROM tx_detail WHERE created_at BETWEEN ? AND ? AND account_index != -1 GROUP BY account_index", from, to).Count(&count)
	if dbTx.Error != nil {
		return 0, dbTx.Error
	}
	return count, nil
}