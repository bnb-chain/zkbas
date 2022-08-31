package prover

import (
	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbas/service/prover/config"
	"github.com/bnb-chain/zkbas/service/prover/prover"
)

func Run(configFile string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	p := prover.NewProver(c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})

	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err := cronJob.AddFunc("@every 10s", func() {
		logx.Info("start prover job......")
		// cron job for receiving cryptoBlock and handling
		err := p.ProveBlock()
		if err != nil {
			logx.Errorf("failed to generate proof, %v", err)
		}
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()
	select {}
}
