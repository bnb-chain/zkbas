/*
 * Copyright © 2021 Zkbas Protocol
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

package account

import (
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	AccountModel interface {
		CreateAccountTable() error
		DropAccountTable() error
		IfAccountNameExist(name string) (bool, error)
		IfAccountExistsByAccountIndex(accountIndex int64) (bool, error)
		GetAccountByAccountIndex(accountIndex int64) (account *Account, err error)
		GetVerifiedAccountByAccountIndex(accountIndex int64) (account *Account, err error)
		GetConfirmedAccountByAccountIndex(accountIndex int64) (account *Account, err error)
		GetAccountByPk(pk string) (account *Account, err error)
		GetAccountByAccountName(accountName string) (account *Account, err error)
		GetAccountByAccountNameHash(accountNameHash string) (account *Account, err error)
		GetAccountsList(limit int, offset int64) (accounts []*Account, err error)
		GetAccountsTotalCount() (count int64, err error)
		GetAllAccounts() (accounts []*Account, err error)
		GetLatestAccountIndex() (accountIndex int64, err error)
		GetConfirmedAccounts() (accounts []*Account, err error)
	}

	defaultAccountModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	/*
		always keep the latest data of committer
	*/
	Account struct {
		gorm.Model
		AccountIndex    int64  `gorm:"uniqueIndex"`
		AccountName     string `gorm:"uniqueIndex"`
		PublicKey       string `gorm:"uniqueIndex"`
		AccountNameHash string `gorm:"uniqueIndex"`
		L1Address       string
		Nonce           int64
		CollectionNonce int64
		// map[int64]*AccountAsset
		AssetInfo string
		AssetRoot string
		// 0 - registered, not committer 1 - committer
		Status int
	}
)

func NewAccountModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) AccountModel {
	return &defaultAccountModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      AccountTableName,
		DB:         db,
	}
}

func (*Account) TableName() string {
	return AccountTableName
}

/*
	Func: CreateAccountTable
	Params:
	Return: err error
	Description: create account table
*/
func (m *defaultAccountModel) CreateAccountTable() error {
	return m.DB.AutoMigrate(Account{})
}

/*
	Func: DropAccountTable
	Params:
	Return: err error
	Description: drop account table
*/
func (m *defaultAccountModel) DropAccountTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: IfAccountNameExist
	Params: name string
	Return: bool, error
	Description: check account name existence
*/
func (m *defaultAccountModel) IfAccountNameExist(name string) (bool, error) {
	var res int64
	dbTx := m.DB.Table(m.table).Where("account_name = ? and deleted_at is NULL", strings.ToLower(name)).Count(&res)

	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.IfAccountNameExist] %s", dbTx.Error)
		logx.Error(err)
		return true, dbTx.Error
	} else if res == 0 {
		return false, nil
	} else if res != 1 {
		logx.Errorf("[account.IfAccountNameExist] %s", errorcode.DbErrDuplicatedAccountName)
		return true, errorcode.DbErrDuplicatedAccountName
	} else {
		return true, nil
	}
}

/*
	Func: IfAccountExistsByAccountIndex
	Params: accountIndex int64
	Return: bool, error
	Description: check account index existence
*/
func (m *defaultAccountModel) IfAccountExistsByAccountIndex(accountIndex int64) (bool, error) {
	var res int64
	dbTx := m.DB.Table(m.table).Where("account_index = ? and deleted_at is NULL", accountIndex).Count(&res)

	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.IfAccountExistsByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		// TODO : to be modified
		return true, dbTx.Error
	} else if res == 0 {
		return false, nil
	} else if res != 1 {
		logx.Errorf("[account.IfAccountExistsByAccountIndex] %s", errorcode.DbErrDuplicatedAccountIndex)
		return true, errorcode.DbErrDuplicatedAccountIndex
	} else {
		return true, nil
	}
}

/*
	Func: GetAccountByAccountIndex
	Params: accountIndex int64
	Return: account Account, err error
	Description: get account info by index
*/

func (m *defaultAccountModel) GetAccountByAccountIndex(accountIndex int64) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetAccountByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetAccountByAccountIndex] %s", errorcode.DbErrNotFound)
		logx.Error(err)
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

func (m *defaultAccountModel) GetVerifiedAccountByAccountIndex(accountIndex int64) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? and status = ?", accountIndex, AccountStatusVerified).Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetAccountByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetAccountByAccountIndex] %s", errorcode.DbErrNotFound)
		logx.Error(err)
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

/*
	Func: GetAccountByPk
	Params: pk string
	Return: account Account, err error
	Description: get account info by public key
*/

func (m *defaultAccountModel) GetAccountByPk(pk string) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("public_key = ?", pk).Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetAccountByPk] %s", dbTx.Error)
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetAccountByPk] %s", errorcode.DbErrNotFound)
		logx.Error(err)
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

/*
	Func: GetAccountByAccountName
	Params: accountName string
	Return: account Account, err error
	Description: get account info by account name
*/

func (m *defaultAccountModel) GetAccountByAccountName(accountName string) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_name = ?", accountName).Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetAccountByAccountName] %s", dbTx.Error)
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetAccountByAccountName] %s", errorcode.DbErrNotFound)
		logx.Info(err)
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

/*
	Func: GetAccountsList
	Params: limit int, offset int64
	Return: err error
	Description:  For API /api/v1/info/getAccountsList

*/
func (m *defaultAccountModel) GetAccountsList(limit int, offset int64) (accounts []*Account, err error) {
	dbTx := m.DB.Table(m.table).Limit(limit).Offset(int(offset)).Order("account_index desc").Find(&accounts)
	if dbTx.Error != nil {
		logx.Errorf("[account.GetAccountsList] error: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[account.GetAccountsList] Get Accounts Error")
		return nil, errorcode.DbErrNotFound
	}
	return accounts, nil
}

/*
	Func: GetAccountsTotalCount
	Params:
	Return: count int64, err error
	Description: used for counting total accounts for explorer dashboard
*/
func (m *defaultAccountModel) GetAccountsTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[account.GetAccountsTotalCount] error: %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[account.GetAccountsTotalCount] No Accounts in Account Table")
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetAllAccounts
	Params:
	Return: count int64, err error
	Description: used for construct MPT
*/
func (m *defaultAccountModel) GetAllAccounts() (accounts []*Account, err error) {
	dbTx := m.DB.Table(m.table).Order("account_index").Find(&accounts)
	if dbTx.Error != nil {
		logx.Errorf("[account.GetAllAccounts] %s", dbTx.Error.Error())
		return accounts, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Error("[account.GetAllAccounts] No Account in Account Table")
		return accounts, nil
	}
	return accounts, nil
}

/*
	Func: GetLatestAccountIndex
	Params:
	Return: accountIndex int64, err error
	Description: get max accountIndex
*/
func (m *defaultAccountModel) GetLatestAccountIndex() (accountIndex int64, err error) {
	dbTx := m.DB.Table(m.table).Select("account_index").Order("account_index desc").Limit(1).Find(&accountIndex)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetLatestAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[account.GetLatestAccountIndex] No Account in Account Table")
		return 0, errorcode.DbErrNotFound
	}
	logx.Info(accountIndex)
	return accountIndex, nil
}

func (m *defaultAccountModel) GetAccountByAccountNameHash(accountNameHash string) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_name_hash = ?", accountNameHash).Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetAccountByAccountNameHash] %s", dbTx.Error)
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetAccountByAccountNameHash] %s", errorcode.DbErrNotFound)
		logx.Info(err)
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}

func (m *defaultAccountModel) GetConfirmedAccounts() (accounts []*Account, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ?", AccountStatusConfirmed).Order("account_index").Find(&accounts)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetConfirmedAccounts] %s", dbTx.Error)
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetConfirmedAccounts] %s", errorcode.DbErrNotFound)
		logx.Info(err)
		return nil, errorcode.DbErrNotFound
	}
	return accounts, nil
}

func (m *defaultAccountModel) GetConfirmedAccountByAccountIndex(accountIndex int64) (account *Account, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? and status = ?", accountIndex, AccountStatusConfirmed).Find(&account)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[account.GetAccountByAccountIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[account.GetAccountByAccountIndex] %s", errorcode.DbErrNotFound)
		logx.Error(err)
		return nil, errorcode.DbErrNotFound
	}
	return account, nil
}
