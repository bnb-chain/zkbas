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

package nft

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/types"
)

const (
	OfferTableName = `offer`

	OfferFinishedStatus = 1
)

type (
	OfferModel interface {
		CreateOfferTable() error
		DropOfferTable() error
		GetOfferByAccountIndexAndOfferId(accountIndex int64, offerId int64) (offer *Offer, err error)
		GetLatestOfferId(accountIndex int64) (offerId int64, err error)
	}
	defaultOfferModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	Offer struct {
		gorm.Model
		OfferType    int64
		OfferId      int64
		AccountIndex int64
		NftIndex     int64
		AssetId      int64
		AssetAmount  string
		ListedAt     int64
		ExpiredAt    int64
		TreasuryRate int64
		Sig          string
		Status       int
	}
)

func NewOfferModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) OfferModel {
	return &defaultOfferModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      OfferTableName,
		DB:         db,
	}
}

func (*Offer) TableName() string {
	return OfferTableName
}

func (m *defaultOfferModel) CreateOfferTable() error {
	return m.DB.AutoMigrate(Offer{})
}

func (m *defaultOfferModel) DropOfferTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultOfferModel) GetLatestOfferId(accountIndex int64) (offerId int64, err error) {
	var offer *Offer
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Order("offer_id desc").Find(&offer)
	if dbTx.Error != nil {
		return -1, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return -1, types.DbErrNotFound
	}
	return offer.OfferId, nil
}

func (m *defaultOfferModel) GetOfferByAccountIndexAndOfferId(accountIndex int64, offerId int64) (offer *Offer, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? AND offer_id = ?", accountIndex, offerId).Find(&offer)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return offer, nil
}
