package svc

import (
	"github.com/bnb-chain/zkbas/common/model/l1BlockMonitor"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/service/cronjob/blockMonitor/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config         config.Config
	L1BlockMonitor l1BlockMonitor.L1BlockMonitorModel
	SysConfig      sysconfig.SysconfigModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)

	return &ServiceContext{
		Config:         c,
		L1BlockMonitor: l1BlockMonitor.NewL1BlockMonitorModel(conn, c.CacheRedis, gormPointer),
		SysConfig:      sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
	}
}
