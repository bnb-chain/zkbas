package main

import (
	"fmt"
	"os"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/api/app/internal/config"
	"github.com/bnb-chain/zkbas/service/api/app/internal/handler"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

var (
	CodeVersion   = ""
	GitCommitHash = ""
)

func main() {
	args := os.Args
	if len(args) == 2 && (args[1] == "--version" || args[1] == "-v") {
		fmt.Printf("Git Commit Hash: %s\n", GitCommitHash)
		fmt.Printf("Git Code Version : %s\n", CodeVersion)
		return
	}

	configFile := util.ReadConfigFileFlag()
	var c config.Config
	conf.MustLoad(configFile, &c)

	logx.DisableStat()
	ctx := svc.NewServiceContext(c)
	ctx.CodeVersion = CodeVersion
	ctx.GitCommitHash = GitCommitHash
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()
	handler.RegisterHandlers(server, ctx)
	logx.Infof("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
