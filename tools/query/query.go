package query

import (
	"encoding/json"
	bsmt "github.com/bnb-chain/zkbnb-smt"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/log"
	committerConfig "github.com/bnb-chain/zkbnb/service/committer/config"
	witnessConfig "github.com/bnb-chain/zkbnb/service/witness/config"
	"github.com/bnb-chain/zkbnb/tools/query/config"
	"github.com/bnb-chain/zkbnb/tools/query/svc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"strconv"

	"github.com/bnb-chain/zkbnb/tree"
)

func QueryTreeDB(
	configFile string,
	blockHeight int64,
	serviceName string,
	fromHistory bool, AccountIndexesStr string,

) {
	configInfo := BuildConfig(configFile, serviceName)
	ctx := svc.NewServiceContext(configInfo)
	logx.MustSetup(configInfo.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})
	ctxLog := log.NewCtxWithKV(log.BlockHeightContext, blockHeight)

	var AccountIndexes []int64
	if AccountIndexesStr != "" {
		err := json.Unmarshal([]byte(AccountIndexesStr), &AccountIndexes)
		if err != nil {
			logx.Errorf("json.Unmarshal failed: %s", err)
			return
		}
	}
	treeCtx, err := tree.NewContext(serviceName, configInfo.TreeDB.Driver, false, true, configInfo.TreeDB.RoutinePoolSize, &configInfo.TreeDB.LevelDBOption, &configInfo.TreeDB.RedisDBOption, configInfo.TreeDB.AssetTreeCacheSize,
		fromHistory, configInfo.DbRoutineSize)
	if err != nil {
		logx.Errorf("Init tree database failed: %s", err)
		return
	}

	treeCtx.SetOptions(bsmt.InitializeVersion(0))
	treeCtx.SetBatchReloadSize(configInfo.DbBatchSize)
	err = tree.SetupTreeDB(treeCtx)
	if err != nil {
		logx.Errorf("Init tree database failed: %s", err)
		return
	}

	// dbinitializer accountTree and accountStateTrees
	accountTree, accountAssetTrees, err := tree.InitAccountTree(
		ctx.AccountModel,
		ctx.AccountHistoryModel,
		make([]int64, 0),
		blockHeight,
		treeCtx,
	)
	if err != nil {
		logx.Error("InitMerkleTree error:", err)
		return
	}
	if len(AccountIndexes) > 0 {
		for _, accountIndex := range AccountIndexes {
			ctxLog := log.UpdateCtxWithKV(ctxLog, log.AccountIndexCtx, accountIndex)
			assetRoot := common.Bytes2Hex(accountAssetTrees.Get(accountIndex).Root())
			logx.WithContext(ctxLog).Infof("asset tree root accountIndex=%s,assetRoot=%s,versions=%s,latestVersion=%s", strconv.FormatInt(accountIndex, 10), assetRoot,
				common2.FormatVersion(accountAssetTrees.Get(accountIndex).Versions()), strconv.FormatUint(uint64(accountAssetTrees.Get(accountIndex).LatestVersion()), 10))
			for i := 0; i < 20; i++ {
				assetOne, err := accountAssetTrees.Get(accountIndex).Get(uint64(i), nil)
				if err != nil {
					continue
				}
				logx.WithContext(ctxLog).Infof("asset tree accountIndex=%s,assetId=%s,assetRoot=%s", strconv.FormatInt(accountIndex, 10), strconv.FormatInt(int64(i), 10), common.Bytes2Hex(assetOne))
			}
			//accountAssetTrees.Get(accountIndex).PrintLeaves()
		}
	}
	ctxLog = log.NewCtxWithKV(log.BlockHeightContext, blockHeight)

	stateRoot := common.Bytes2Hex(accountTree.Root())
	logx.WithContext(ctxLog).Infof("account tree accountRoot=%s,versions=%s,,latestVersion=%s", stateRoot, common2.FormatVersion(accountTree.Versions()), strconv.FormatUint(uint64(accountTree.LatestVersion()), 10))
	// dbinitializer nftTree
	nftTree, err := tree.InitNftTree(
		ctx.NftModel,
		ctx.NftHistoryModel,
		blockHeight,
		treeCtx)
	if err != nil {
		logx.WithContext(ctxLog).Errorf("InitNftTree error: %s", err.Error())
		return
	}
	nftRoot := common.Bytes2Hex(nftTree.Root())
	logx.WithContext(ctxLog).Infof("nft tree nftRoot=%s,versions=%s,,latestVersion=%s", nftRoot, common2.FormatVersion(nftTree.Versions()), strconv.FormatUint(uint64(nftTree.LatestVersion()), 10))
}

func BuildConfig(configFile string, serviceName string) config.Config {
	configInfo := config.Config{}
	if serviceName == "committer" {
		c := committerConfig.Config{}
		if err := committerConfig.InitSystemConfiguration(&c, configFile); err != nil {
			logx.Severef("failed to initiate system configuration, %v", err)
			panic("failed to initiate system configuration, err:" + err.Error())
		}
		configInfo.TreeDB = c.TreeDB
		configInfo.Postgres = c.Postgres
		configInfo.CacheRedis = c.CacheRedis
		configInfo.LogConf = c.LogConf
		configInfo.EnableRollback = c.EnableRollback
		configInfo.DbRoutineSize = c.DbRoutineSize
		configInfo.DbBatchSize = c.DbBatchSize
	} else if serviceName == "witness" {
		c := witnessConfig.Config{}
		if err := witnessConfig.InitSystemConfiguration(&c, configFile); err != nil {
			logx.Severef("failed to initiate system configuration, %v", err)
			panic("failed to initiate system configuration, err:" + err.Error())
		}
		configInfo.TreeDB = c.TreeDB
		configInfo.Postgres = c.Postgres
		configInfo.LogConf = c.LogConf
		configInfo.EnableRollback = c.EnableRollback
		configInfo.DbRoutineSize = c.DbRoutineSize
		configInfo.DbBatchSize = c.DbBatchSize
	} else {
		logx.Error("there is no serviceName,%s", serviceName)
	}
	return configInfo
}
