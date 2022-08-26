package executor

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	common2 "github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/common/chain"
	"github.com/bnb-chain/zkbas/core/statedb"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/nft"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/types"
)

type DepositNftExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.DepositNftTxInfo

	isNewNft bool
}

func NewDepositNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseDepositNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse deposit nft tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &DepositNftExecutor{
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *DepositNftExecutor) Prepare() error {
	bc := e.bc
	txInfo := e.txInfo

	// The account index from txInfo isn't true, find account by account name hash.
	accountNameHash := common.Bytes2Hex(txInfo.AccountNameHash)
	account, err := bc.DB().AccountModel.GetAccountByNameHash(accountNameHash)
	if err != nil {
		for index := range bc.StateDB().PendingNewAccountIndexMap {
			if accountNameHash == bc.StateDB().AccountMap[index].AccountNameHash {
				account, err = chain.FromFormatAccountInfo(bc.StateDB().AccountMap[index])
				break
			}
		}

		if err != nil {
			return errors.New("invalid account name hash")
		}
	}

	// Set the right account index.
	txInfo.AccountIndex = account.AccountIndex

	accounts := []int64{txInfo.AccountIndex, txInfo.CreatorAccountIndex}
	assets := []int64{0} // Just used for generate an empty tx detail.
	err = e.bc.StateDB().PrepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return err
	}

	// Check if it is a new nft, or it is a nft previously withdraw from layer2.
	if txInfo.NftIndex == 0 && txInfo.CollectionId == 0 && txInfo.CreatorAccountIndex == 0 && txInfo.CreatorTreasuryRate == 0 {
		e.isNewNft = true
		// Set new nft index for new nft.
		txInfo.NftIndex = bc.StateDB().GetNextNftIndex()
	} else {
		err = e.bc.StateDB().PrepareNft(txInfo.NftIndex)
		if err != nil {
			logx.Errorf("prepare nft failed")
			return err
		}
	}

	return nil
}

func (e *DepositNftExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	if e.isNewNft {
		if bc.StateDB().NftMap[txInfo.NftIndex] != nil {
			return errors.New("invalid nft index, already exist")
		}
	} else {
		if bc.StateDB().NftMap[txInfo.NftIndex].OwnerAccountIndex != types.NilAccountIndex {
			return errors.New("invalid nft index, already exist")
		}
	}

	return nil
}

func (e *DepositNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	bc.StateDB().NftMap[txInfo.NftIndex] = &nft.L2Nft{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.AccountIndex,
		NftContentHash:      common.Bytes2Hex(txInfo.NftContentHash),
		NftL1Address:        txInfo.NftL1Address,
		NftL1TokenId:        txInfo.NftL1TokenId.String(),
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		CollectionId:        txInfo.CollectionId,
	}

	stateCache := e.bc.StateDB()
	if e.isNewNft {
		stateCache.PendingNewNftIndexMap[txInfo.NftIndex] = statedb.StateCachePending
	} else {
		stateCache.PendingUpdateNftIndexMap[txInfo.NftIndex] = statedb.StateCachePending
	}

	return nil
}

func (e *DepositNftExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeDepositNft))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(common2.AddressStrToBytes(txInfo.NftL1Address))
	chunk1 := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
	chunk2 := common2.PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(common2.Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(common2.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PriorityOperations++
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *DepositNftExecutor) UpdateTrees() error {
	bc := e.bc
	txInfo := e.txInfo

	return bc.StateDB().UpdateNftTree(txInfo.NftIndex)
}

func (e *DepositNftExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.NftIndex = e.txInfo.NftIndex
	e.tx.AccountIndex = e.txInfo.AccountIndex
	return e.BaseExecutor.GetExecutedTx()
}

func (e *DepositNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo
	depositAccount := e.bc.StateDB().AccountMap[txInfo.AccountIndex]
	txDetails := make([]*tx.TxDetail, 0, 2)

	// user info
	accountOrder := int64(0)
	order := int64(0)
	baseBalance := depositAccount.AssetInfo[0]
	deltaBalance := &types.AccountAsset{
		AssetId:                  0,
		Balance:                  big.NewInt(0),
		LpAmount:                 big.NewInt(0),
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         0,
		AssetType:       types.FungibleAssetType,
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     depositAccount.AccountName,
		Balance:         baseBalance.String(),
		BalanceDelta:    deltaBalance.String(),
		AccountOrder:    accountOrder,
		Order:           order,
		Nonce:           depositAccount.Nonce,
		CollectionNonce: depositAccount.CollectionNonce,
	})
	// nft info
	order++
	baseNft := types.EmptyNftInfo(txInfo.NftIndex)
	newNft := types.ConstructNftInfo(
		txInfo.NftIndex,
		txInfo.CreatorAccountIndex,
		txInfo.AccountIndex,
		common.Bytes2Hex(txInfo.NftContentHash),
		txInfo.NftL1TokenId.String(),
		txInfo.NftL1Address,
		txInfo.CreatorTreasuryRate,
		txInfo.CollectionId,
	)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.NftIndex,
		AssetType:       types.NftAssetType,
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     depositAccount.AccountName,
		Balance:         baseNft.String(),
		BalanceDelta:    newNft.String(),
		AccountOrder:    types.NilAccountOrder,
		Order:           order,
		Nonce:           depositAccount.Nonce,
		CollectionNonce: depositAccount.CollectionNonce,
	})

	return txDetails, nil
}

func (e *DepositNftExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	return nil, nil
}
