/*
 * Copyright © 2021 ZkBAS Protocol
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
	"sort"

	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/types"
)

const TxDetailTableName = `tx_detail`

type (
	TxDetailModel interface {
		CreateTxDetailTable() error
		DropTxDetailTable() error
		GetTxDetailByAccountIndex(accountIndex int64) (txDetails []*TxDetail, err error)
		GetTxIdsByAccountIndex(accountIndex int64) (txIds []int64, err error)
	}

	defaultTxDetailModel struct {
		table string
		DB    *gorm.DB
	}

	TxDetail struct {
		gorm.Model
		TxId            int64 `gorm:"index"`
		AssetId         int64
		AssetType       int64
		AccountIndex    int64 `gorm:"index"`
		AccountName     string
		Balance         string
		BalanceDelta    string
		Order           int64
		AccountOrder    int64
		Nonce           int64
		CollectionNonce int64
	}
)

func NewTxDetailModel(db *gorm.DB) TxDetailModel {
	return &defaultTxDetailModel{
		table: TxDetailTableName,
		DB:    db,
	}
}

func (*TxDetail) TableName() string {
	return TxDetailTableName
}

func (m *defaultTxDetailModel) CreateTxDetailTable() error {
	return m.DB.AutoMigrate(TxDetail{})
}

func (m *defaultTxDetailModel) DropTxDetailTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultTxDetailModel) GetTxDetailByAccountIndex(accountIndex int64) (txDetails []*TxDetail, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Find(&txDetails)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return txDetails, nil
}

func (m *defaultTxDetailModel) GetTxIdsByAccountIndex(accountIndex int64) (txIds []int64, err error) {
	dbTx := m.DB.Table(m.table).Select("tx_id").Where("account_index = ?", accountIndex).Group("tx_id").Find(&txIds)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	sort.Slice(txIds, func(i, j int) bool {
		return txIds[i] > txIds[j]
	})
	return txIds, nil
}
