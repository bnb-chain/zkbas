/*
 * Copyright © 2021 ZkBNB Protocol
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

package tx

import (
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/types"
)

const (
	PoolTxTableName = `pool_tx`
)

type (
	TxPoolModel interface {
		CreatePoolTxTable() error
		DropPoolTxTable() error
		GetTxs(limit int64, offset int64) (txs []*Tx, err error)
		GetTxsTotalCount() (count int64, err error)
		GetTxByTxHash(hash string) (txs *Tx, err error)
		GetTxsByStatus(status int) (txs []*Tx, err error)
		CreateTxs(txs []*Tx) error
		GetPendingTxsByAccountIndex(accountIndex int64) (txs []*Tx, err error)
		GetMaxNonceByAccountIndex(accountIndex int64) (nonce int64, err error)
		CreateTxsInTransact(tx *gorm.DB, txs []*Tx) error
		UpdateTxsInTransact(tx *gorm.DB, txs []*Tx) error
		DeleteTxsInTransact(tx *gorm.DB, txs []*Tx) error
	}

	defaultTxPoolModel struct {
		table string
		DB    *gorm.DB
	}

	PoolTx struct {
		Tx
	}
)

func NewTxPoolModel(db *gorm.DB) TxPoolModel {
	return &defaultTxPoolModel{
		table: PoolTxTableName,
		DB:    db,
	}
}

func (*PoolTx) TableName() string {
	return PoolTxTableName
}

func (m *defaultTxPoolModel) CreatePoolTxTable() error {
	return m.DB.AutoMigrate(PoolTx{})
}

func (m *defaultTxPoolModel) DropPoolTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultTxPoolModel) GetTxs(limit int64, offset int64) (txs []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_status = ?", StatusPending).Limit(int(limit)).Offset(int(offset)).Order("created_at desc, id desc").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	return txs, nil
}

func (m *defaultTxPoolModel) GetTxsByStatus(status int) (txs []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_status = ?", status).Order("created_at, id").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	}
	return txs, nil
}

func (m *defaultTxPoolModel) GetTxsTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_status = ? and deleted_at is NULL", StatusPending).Count(&count)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

func (m *defaultTxPoolModel) GetTxByTxHash(hash string) (tx *Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_hash = ?", hash).Find(&tx)
	if dbTx.Error != nil {
		if dbTx.Error == types.DbErrNotFound {
			return tx, dbTx.Error
		} else {
			return nil, types.DbErrSqlOperation
		}
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return tx, nil
}

func (m *defaultTxPoolModel) CreateTxs(txs []*Tx) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Create(txs)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToCreatePoolTx
		}
		return nil
	})
}

func (m *defaultTxPoolModel) GetPendingTxsByAccountIndex(accountIndex int64) (txs []*Tx, err error) {
	dbTx := m.DB.Table(m.table).Where("tx_status = ? AND account_index = ?", StatusPending, accountIndex).
		Order("created_at, id").Find(&txs)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txs, nil
}

func (m *defaultTxPoolModel) GetMaxNonceByAccountIndex(accountIndex int64) (nonce int64, err error) {
	dbTx := m.DB.Table(m.table).Select("nonce").Where("deleted_at is null and account_index = ?", accountIndex).Order("nonce desc").Limit(1).Find(&nonce)
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return 0, types.DbErrNotFound
	}
	return nonce, nil
}

func (m *defaultTxPoolModel) CreateTxsInTransact(tx *gorm.DB, txs []*Tx) error {
	dbTx := tx.Table(m.table).CreateInBatches(txs, len(txs))
	if dbTx.Error != nil {
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return types.DbErrFailToCreatePoolTx
	}
	return nil
}

func (m *defaultTxPoolModel) UpdateTxsInTransact(tx *gorm.DB, txs []*Tx) error {
	for _, poolTx := range txs {
		// Don't write tx details when update tx pool.
		dbTx := tx.Table(m.table).Where("id = ?", poolTx.ID).
			Select("*").
			Updates(&poolTx)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToUpdatePoolTx
		}
	}
	return nil
}

func (m *defaultTxPoolModel) DeleteTxsInTransact(tx *gorm.DB, txs []*Tx) error {
	for _, poolTx := range txs {
		dbTx := tx.Table(m.table).Where("id = ?", poolTx.ID).Delete(&poolTx)
		if dbTx.Error != nil {
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return types.DbErrFailToDeletePoolTx
		}
	}
	return nil
}
