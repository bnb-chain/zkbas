package core

import (
	"bytes"
	"encoding/json"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

type MintNftExecutor struct {
	bc     *BlockChain
	tx     *tx.Tx
	txInfo *legendTxTypes.MintNftTxInfo
}

func NewMintNftExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	return &MintNftExecutor{
		bc: bc,
		tx: tx,
	}, nil
}

func (e *MintNftExecutor) Prepare() error {
	txInfo, err := commonTx.ParseMintNftTxInfo(e.tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return errors.New("invalid tx info")
	}

	accounts := []int64{txInfo.CreatorAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.GasFeeAssetId}
	err = e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return err
	}

	e.txInfo = txInfo
	return nil
}

func (e *MintNftExecutor) VerifyInputs() error {
	txInfo := e.txInfo

	err := txInfo.Validate()
	if err != nil {
		return err
	}

	if txInfo.ExpiredAt < e.bc.currentBlock.CreatedAt.UnixMilli() {
		return errors.New("tx expired")
	}

	creatorAccount := e.bc.accountMap[txInfo.CreatorAccountIndex]
	if txInfo.Nonce != creatorAccount.Nonce {
		return errors.New("invalid nonce")
	}

	if creatorAccount.CollectionNonce < txInfo.NftCollectionId {
		return errors.New("nft collection id is less than account collection nonce")
	}

	if creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return errors.New("balance is not enough")
	}

	toAccount := e.bc.accountMap[txInfo.ToAccountIndex]
	if txInfo.ToAccountNameHash != toAccount.AccountNameHash {
		return errors.New("invalid ToAccountNameHash")
	}

	err = txInfo.VerifySignature(creatorAccount.PublicKey)
	if err != nil {
		return err
	}
	return nil
}

func (e *MintNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	// add nft index to tx info
	nextNftIndex, err := e.bc.getNextNftIndex()
	if err != nil {
		return err
	}
	txInfo.NftIndex = nextNftIndex

	// generate tx details
	e.tx.TxDetails = e.GenerateTxDetails()

	// apply changes
	creatorAccount := bc.accountMap[txInfo.CreatorAccountIndex]
	gasAccount := bc.accountMap[txInfo.GasAccountIndex]

	creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	creatorAccount.Nonce++

	bc.nftMap[txInfo.NftIndex] = &nft.L2Nft{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
		NftContentHash:      txInfo.NftContentHash,
		NftL1Address:        commonConstant.NilL1Address,
		NftL1TokenId:        commonConstant.NilL1TokenId,
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		CollectionId:        txInfo.NftCollectionId,
	}

	stateCache := e.bc.stateCache
	stateCache.pendingUpdateAccountIndexMap[txInfo.CreatorAccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = StateCachePending
	stateCache.pendingNewNftIndexMap[txInfo.NftIndex] = StateCachePending
	return nil
}

func (e *MintNftExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeMintNft))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(util.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := util.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	buf.Write(util.Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.NftCollectionId)))
	chunk := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(util.PrefixPaddingBufToChunkSize(common.FromHex(txInfo.NftContentHash)))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))

	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *MintNftExecutor) UpdateTrees() error {
	txInfo := e.txInfo

	accounts := []int64{txInfo.CreatorAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.GasFeeAssetId}

	err := e.bc.updateAccountTree(accounts, assets)
	if err != nil {
		logx.Errorf("update account tree error, err: %s", err.Error())
		return err
	}

	err = e.bc.updateNftTree(txInfo.NftIndex)
	if err != nil {
		logx.Errorf("update nft tree error, err: %s", err.Error())
		return err
	}
	return nil
}

func (e *MintNftExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.BlockHeight = e.bc.currentBlock.BlockHeight
	e.tx.StateRoot = e.bc.getStateRoot()
	e.tx.TxInfo = string(txInfoBytes)
	e.tx.TxStatus = tx.StatusPending

	return e.tx, nil
}

func (e *MintNftExecutor) GenerateTxDetails() []*tx.TxDetail {
	txInfo := e.txInfo
	creatorAccount := e.bc.accountMap[txInfo.CreatorAccountIndex]
	toAccount := e.bc.accountMap[txInfo.ToAccountIndex]
	gasAccount := e.bc.accountMap[txInfo.GasAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 4)

	// from account gas asset
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.CreatorAccountIndex,
		AccountName:  creatorAccount.AccountName,
		Balance:      creatorAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           creatorAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: creatorAccount.CollectionNonce,
	})

	// to account empty delta
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  toAccount.AccountName,
		Balance:      toAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ZeroBigInt,
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           toAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: toAccount.CollectionNonce,
	})

	// to account nft delta
	oldNftInfo := commonAsset.EmptyNftInfo(txInfo.NftIndex)
	newNftInfo := &commonAsset.NftInfo{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
		NftContentHash:      txInfo.NftContentHash,
		NftL1TokenId:        commonConstant.NilL1TokenId,
		NftL1Address:        commonConstant.NilL1Address,
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		CollectionId:        txInfo.NftCollectionId,
	}
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.NftIndex,
		AssetType:       commonAsset.NftAssetType,
		AccountIndex:    txInfo.ToAccountIndex,
		AccountName:     toAccount.AccountName,
		Balance:         oldNftInfo.String(),
		BalanceDelta:    newNftInfo.String(),
		Order:           order,
		Nonce:           toAccount.Nonce,
		AccountOrder:    commonConstant.NilAccountOrder,
		CollectionNonce: toAccount.CollectionNonce,
	})

	// gas account gas asset
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  gasAccount.AccountName,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			txInfo.GasFeeAssetAmount,
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           gasAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: gasAccount.CollectionNonce,
	})
	return txDetails
}