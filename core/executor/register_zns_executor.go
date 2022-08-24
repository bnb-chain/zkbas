package executor

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	common2 "github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/common/chain"
	"github.com/bnb-chain/zkbas/core/statedb"
	"github.com/bnb-chain/zkbas/dao/account"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/tree"
	"github.com/bnb-chain/zkbas/types"
)

type RegisterZnsExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.RegisterZnsTxInfo
}

func NewRegisterZnsExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseRegisterZnsTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse register tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &RegisterZnsExecutor{
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *RegisterZnsExecutor) Prepare() error {
	return nil
}

func (e *RegisterZnsExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	_, err := bc.DB().AccountModel.GetAccountByName(txInfo.AccountName)
	if err != sqlx.ErrNotFound {
		return errors.New("invalid account name, already registered")
	}

	for index := range bc.StateDB().PendingNewAccountIndexMap {
		if txInfo.AccountName == bc.StateDB().AccountMap[index].AccountName {
			return errors.New("invalid account name, already registered")
		}
	}

	if txInfo.AccountIndex != bc.StateDB().GetNextAccountIndex() {
		return errors.New("invalid account index")
	}

	return nil
}

func (e *RegisterZnsExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo
	var err error

	newAccount := &account.Account{
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     txInfo.AccountName,
		PublicKey:       txInfo.PubKey,
		AccountNameHash: common.Bytes2Hex(txInfo.AccountNameHash),
		L1Address:       e.tx.NativeAddress,
		Nonce:           types.NilNonce,
		CollectionNonce: types.NilNonce,
		AssetInfo:       types.NilAssetInfo,
		AssetRoot:       common.Bytes2Hex(tree.NilAccountAssetRoot),
		Status:          account.AccountStatusConfirmed,
	}
	bc.StateDB().AccountMap[txInfo.AccountIndex], err = chain.ToFormatAccountInfo(newAccount)
	if err != nil {
		return err
	}

	stateCache := e.bc.StateDB()
	stateCache.PendingNewAccountIndexMap[txInfo.AccountIndex] = statedb.StateCachePending
	return nil
}

func (e *RegisterZnsExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeRegisterZns))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	chunk := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(common2.PrefixPaddingBufToChunkSize(common2.AccountNameToBytes32(txInfo.AccountName)))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	pk, err := common2.ParsePubKey(txInfo.PubKey)
	if err != nil {
		logx.Errorf("unable to parse pub key: %s", err.Error())
		return err
	}
	// because we can get Y from X, so we only need to store X is enough
	buf.Write(common2.PrefixPaddingBufToChunkSize(pk.A.X.Marshal()))
	buf.Write(common2.PrefixPaddingBufToChunkSize(pk.A.Y.Marshal()))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PriorityOperations++
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *RegisterZnsExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo
	accounts := []int64{txInfo.AccountIndex}

	emptyAssetTree, err := tree.NewEmptyAccountAssetTree(bc.StateDB().TreeCtx, txInfo.AccountIndex, uint64(bc.CurrentBlock().BlockHeight))
	if err != nil {
		logx.Errorf("new empty account asset tree failed: %s", err.Error())
		return err
	}
	bc.StateDB().AccountAssetTrees = append(bc.StateDB().AccountAssetTrees, emptyAssetTree)

	return bc.StateDB().UpdateAccountTree(accounts, nil)
}

func (e *RegisterZnsExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.AccountIndex = e.txInfo.AccountIndex
	return e.BaseExecutor.GetExecutedTx()
}

func (e *RegisterZnsExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	return nil, nil
}

func (e *RegisterZnsExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	return nil, nil
}
