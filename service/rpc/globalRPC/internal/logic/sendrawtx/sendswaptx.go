package sendrawtx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

func SendSwapTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {
	// parse swap tx info
	txInfo, err := commonTx.ParseSwapTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendSwapTx.ParseSwapTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}
	/*
		Check Params
	*/
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return "", err
	}
	if err := util.CheckPackedAmount(txInfo.AssetAAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetAAmount, err)
		return "", err
	}
	if err := util.CheckPackedAmount(txInfo.AssetBMinAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetBMinAmount, err)
		return "", err
	}
	err = util.CheckRequestParam(util.TypeAssetId, reflect.ValueOf(txInfo.AssetAId))
	if err != nil {
		errInfo := fmt.Sprintf("[sendSwapTx] err: invalid assetAId %v", txInfo.AssetAId)
		logx.Error(errInfo)
		return "", handleCreateFailSwapTx(svcCtx.FailTxModel, txInfo, errors.New(errInfo))
	}
	err = commglobalmap.DeleteLatestAccountInfoInCache(ctx, txInfo.FromAccountIndex)
	if err != nil {
		logx.Errorf("[DeleteLatestAccountInfoInCache] err:%v", err)
	}
	// check gas account index
	gasAccountIndexConfig, err := svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendSwapTx] unable to get sysconfig by name: %s", err.Error())
		return "", handleCreateFailSwapTx(svcCtx.FailTxModel, txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		logx.Error(err)
		return "", handleCreateFailSwapTx(svcCtx.FailTxModel, txInfo, errors.New("[sendSwapTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendSwapTx] invalid gas account index")
		return "", handleCreateFailSwapTx(svcCtx.FailTxModel, txInfo, errors.New("[sendSwapTx] invalid gas account index"))
	}

	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendSwapTx] invalid time stamp")
		return "", handleCreateFailSwapTx(svcCtx.FailTxModel, txInfo, errors.New("[sendSwapTx] invalid time stamp"))
	}

	var (
		redisLock      *redis.RedisLock
		liquidityInfo  *commonAsset.LiquidityInfo
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)

	redisLock, liquidityInfo, err = globalmapHandler.GetLatestLiquidityInfoForWrite(
		svcCtx.LiquidityModel,
		svcCtx.MempoolModel,
		svcCtx.RedisConnection,
		txInfo.PairIndex,
	)
	if err != nil {
		logx.Errorf("[sendSwapTx] unable to get latest liquidity info for write: %s", err.Error())
		return "", handleCreateFailSwapTx(svcCtx.FailTxModel, txInfo, err)
	}
	defer redisLock.Release()

	// check params
	if liquidityInfo.AssetA == nil ||
		liquidityInfo.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidityInfo.AssetB == nil ||
		liquidityInfo.AssetB.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("[sendSwapTx] invalid params")
		return "", errors.New("[sendSwapTx] invalid params")
	}

	// compute delta
	var (
		toDelta *big.Int
	)
	if liquidityInfo.AssetAId == txInfo.AssetAId &&
		liquidityInfo.AssetBId == txInfo.AssetBId {
		toDelta, _, err = util.ComputeDelta(
			liquidityInfo.AssetA,
			liquidityInfo.AssetB,
			liquidityInfo.AssetAId,
			liquidityInfo.AssetBId,
			txInfo.AssetAId,
			true,
			txInfo.AssetAAmount,
			liquidityInfo.FeeRate,
		)
	} else if liquidityInfo.AssetAId == txInfo.AssetBId &&
		liquidityInfo.AssetBId == txInfo.AssetAId {
		toDelta, _, err = util.ComputeDelta(
			liquidityInfo.AssetA,
			liquidityInfo.AssetB,
			liquidityInfo.AssetAId,
			liquidityInfo.AssetBId,
			txInfo.AssetBId,
			true,
			txInfo.AssetAAmount,
			liquidityInfo.FeeRate,
		)
	} else {
		err = errors.New("invalid pair assetIds")
	}
	if err != nil {
		errInfo := fmt.Sprintf("[sendSwapTx] => [util.ComputeDelta]: %s. invalid AssetId: %v/%v/%v",
			err.Error(), txInfo.AssetAId,
			uint32(liquidityInfo.AssetAId),
			uint32(liquidityInfo.AssetBId))
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}
	// check if toDelta is less than minToAmount
	if toDelta.Cmp(txInfo.AssetBMinAmount) < 0 {
		errInfo := fmt.Sprintf("[sendSwapTx] => minToAmount is bigger than toDelta: %s/%s",
			txInfo.AssetBMinAmount.String(), toDelta.String())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}
	// complete tx info
	txInfo.AssetBAmountDelta = toDelta
	// get latest account info for from account index
	if accountInfoMap[txInfo.FromAccountIndex] == nil {
		accountInfoMap[txInfo.FromAccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.FromAccountIndex)
		if err != nil {
			logx.Errorf("[sendSwapTx] unable to get latest account info: %s", err.Error())
			return "", err
		}
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			svcCtx.AccountModel,
			svcCtx.RedisConnection,
			txInfo.GasAccountIndex,
		)
		if err != nil {
			logx.Errorf("[sendSwapTx] unable to get latest account info: %s", err.Error())
			return "", err
		}
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	/*
		Get txDetails
	*/

	// verify swap tx
	txDetails, err = txVerification.VerifySwapTxInfo(
		accountInfoMap,
		liquidityInfo,
		txInfo)

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", handleCreateFailSwapTx(svcCtx.FailTxModel, txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeSwap,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		txInfo.PairIndex,
		commonConstant.NilAssetId,
		txInfo.AssetAAmount.String(),
		"",
		string(txInfoBytes),
		"",
		txInfo.FromAccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	// delete key
	key := util.GetLiquidityKeyForWrite(txInfo.PairIndex)
	key2 := util.GetLiquidityKeyForRead(txInfo.PairIndex)
	_, err = svcCtx.RedisConnection.Del(key)
	if err != nil {
		logx.Errorf("[sendSwapTx] unable to delete key from redis: %s", err.Error())
		return "", handleCreateFailSwapTx(svcCtx.FailTxModel, txInfo, err)
	}
	_, err = svcCtx.RedisConnection.Del(key2)
	if err != nil {
		logx.Errorf("[sendSwapTx] unable to delete key from redis: %s", err.Error())
		return "", handleCreateFailSwapTx(svcCtx.FailTxModel, txInfo, err)
	}
	// insert into mempool
	err = CreateMempoolTx(mempoolTx, svcCtx.RedisConnection, svcCtx.MempoolModel)
	if err != nil {
		return "", handleCreateFailSwapTx(svcCtx.FailTxModel, txInfo, err)
	}
	// update redis
	// get latest liquidity info
	for _, txDetail := range txDetails {
		if txDetail.AssetType == commonAsset.LiquidityAssetType {
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.LiquidityAssetType, liquidityInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[sendAddLiquidityTx] unable to compute new balance: %s", err.Error())
				return txId, nil
			}
			liquidityInfo, err = commonAsset.ParseLiquidityInfo(nBalance)
			if err != nil {
				logx.Errorf("[sendAddLiquidityTx] unable to parse liquidity info: %s", err.Error())
				return txId, nil
			}
		}
	}
	liquidityInfoBytes, err := json.Marshal(liquidityInfo)
	if err != nil {
		logx.Errorf("[sendSwapTx] unable to marshal: %s", err.Error())
		return txId, nil
	}
	_ = svcCtx.RedisConnection.Setex(key, string(liquidityInfoBytes), globalmapHandler.LiquidityExpiryTime)
	return txId, nil
}

func handleCreateFailSwapTx(failTxModel tx.FailTxModel, txInfo *commonTx.SwapTxInfo, err error) error {
	errCreate := createFailSwapTx(failTxModel, txInfo, err.Error())
	if errCreate != nil {
		logx.Errorf("[sendswaptxlogic.HandleCreateFailSwapTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendswaptxlogic.HandleCreateFailSwapTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func createFailSwapTx(failTxModel tx.FailTxModel, info *commonTx.SwapTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	txFeeAssetId := info.GasFeeAssetId
	assetAId := info.AssetAId
	assetBId := info.AssetBId
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailSwapTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeSwap,
		// tx fee
		GasFee: info.GasFeeAssetAmount.String(),
		// tx fee l1asset id
		GasFeeAssetId: int64(txFeeAssetId),
		// tx status, 1 - success(default), 2 - failure
		TxStatus: TxFail,
		// AssetAId
		AssetAId: int64(assetAId),
		// l1asset id
		AssetBId: int64(assetBId),
		// tx amount
		TxAmount: util.ZeroBigInt.String(),
		// layer1 address
		NativeAddress: nativeAddress,
		// tx proof
		TxInfo: string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: extraInfo,
	}
	err = failTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailSwapTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
