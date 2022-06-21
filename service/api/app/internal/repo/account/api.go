/*
 * Copyright © 2021 Zecrey Protocol
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

//go:generate mockgen -source api.go -destination api_mock.go -package account

package account

import (
	"context"

	table "github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type AccountModel interface {
	IfAccountNameExist(name string) (bool, error)
	IfAccountExistsByAccountIndex(accountIndex int64) (bool, error)
	GetAccountByAccountIndex(accountIndex int64) (account *table.Account, err error)
	GetVerifiedAccountByAccountIndex(accountIndex int64) (account *table.Account, err error)
	GetConfirmedAccountByAccountIndex(accountIndex int64) (account *table.Account, err error)
	GetAccountByPk(pk string) (account *table.Account, err error)
	GetAccountByAccountName(ctx context.Context, accountName string) (account *table.Account, err error)
	GetAccountByAccountNameHash(accountNameHash string) (account *table.Account, err error)
	GetAccountsList(limit int, offset int64) (accounts []*table.Account, err error)
	GetAccountsTotalCount() (count int64, err error)
	GetAllAccounts() (accounts []*table.Account, err error)
	GetLatestAccountIndex() (accountIndex int64, err error)
	GetConfirmedAccounts() (accounts []*table.Account, err error)
}

func New(svcCtx *svc.ServiceContext) AccountModel {
	return &accountModel{
		table: `account`,
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
