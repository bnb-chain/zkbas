package rollback

import (
	"fmt"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/l1rolluptx"
	"github.com/bnb-chain/zkbnb/service/sender/config"
	"github.com/bnb-chain/zkbnb/tools/revertblock"
	"github.com/bnb-chain/zkbnb/tools/rollback/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"time"
)

//If the smt tree data is incorrect, automatic rollback cannot be used

func RollbackAll(configFile string, height int64) error {
	c := config.Config{}
	if err := config.InitSystemConfiguration(&c, configFile); err != nil {
		logx.Severef("failed to initiate system configuration, %v", err)
		panic("failed to initiate system configuration, err:" + err.Error())
	}
	ctx := svc.NewServiceContext(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})

	if !c.EnableRollback {
		return fmt.Errorf("rollback switch not turned on")
	}

	start := time.Now()

	logx.Infof("revert CommittedBlocks,start height=%d", height)
	err := revertblock.RevertCommittedBlocks(configFile, height, true)
	if err != nil {
		return err
	}

	logx.Infof("delete L1RollupTx,start height=%d", height)
	err = ctx.L1RollupTxModel.DeleteGreaterOrEqualToHeight(height, l1rolluptx.TxTypeCommit)
	if err != nil {
		return err
	}

	logx.Infof("update block status to StatusPending,start height=%d", height)
	err = ctx.BlockModel.UpdateGreaterOrEqualHeight(height, block.StatusPending)
	if err != nil {
		return err
	}

	logx.Infof("delete proof,start height=%d", height)
	err = ctx.ProofModel.DeleteGreaterOrEqualToHeight(height)
	if err != nil {
		logx.Severe(err)
		return err
	}

	logx.Infof("delete block witness,start height=%d", height)
	err = ctx.BlockWitnessModel.DeleteGreaterOrEqualToHeight(height)
	if err != nil {
		return err
	}

	logx.Infof("update block status to StatusProposing,start height=%d", height)
	err = ctx.BlockModel.UpdateGreaterOrEqualHeight(height, block.StatusProposing)
	if err != nil {
		return err
	}
	logx.Infof("rollback success,start height=%d,cost time %v", height, time.Since(start))
	return nil
}
