package sender

import (
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"time"

	"github.com/bnb-chain/zkbnb/service/sender/config"
	"github.com/bnb-chain/zkbnb/service/sender/sender"
)

const GracefulShutdownTimeout = 10 * time.Second

func Run(configFile string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	s := sender.NewSender(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()

	// new cron
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	_, err := cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start commit task =========================")
		err := s.CommitBlocks()
		if err != nil {
			logx.Errorf("failed to rollup block, %v", err)
		}
	})
	if err != nil {
		panic(err)
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start verify task =========================")
		err = s.VerifyAndExecuteBlocks()
		if err != nil {
			logx.Errorf("failed to send verify transaction, %v", err)
		}
	})
	if err != nil {
		panic(err)
	}

	_, err = cronJob.AddFunc("@every 10s", func() {
		logx.Info("========================= start update txs task =========================")
		err = s.UpdateSentTxs()
		if err != nil {
			logx.Errorf("failed to update update tx status, %v", err)
		}
	})
	if err != nil {
		panic(err)
	}

	cronJob.Start()

	exit := make(chan struct{})
	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown sender")
		_ = logx.Close()
		<-cronJob.Stop().Done()
		exit <- struct{}{}
	})

	logx.Info("sender cronjob is starting......")
	select {
	case <-exit:
		break
	}
	return nil
}
