package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type FullExitNftExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.FullExitNftTxInfo

	exitNft *nft.L2Nft
}

func NewFullExitNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseFullExitNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse full exit nft tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &FullExitNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo),
		txInfo:       txInfo,
	}, nil
}

func (e *FullExitNftExecutor) Prepare() error {
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

	// Default withdraw an empty nft.
	// Case1: the nft index isn't exist.
	// Case2: the account isn't the owner of the nft.
	emptyNftInfo := types.EmptyNftInfo(txInfo.NftIndex)
	exitNft := &nft.L2Nft{
		NftIndex:            emptyNftInfo.NftIndex,
		CreatorAccountIndex: emptyNftInfo.CreatorAccountIndex,
		OwnerAccountIndex:   emptyNftInfo.OwnerAccountIndex,
		NftContentHash:      emptyNftInfo.NftContentHash,
		NftL1Address:        emptyNftInfo.NftL1Address,
		NftL1TokenId:        emptyNftInfo.NftL1TokenId,
		CreatorTreasuryRate: emptyNftInfo.CreatorTreasuryRate,
		CollectionId:        emptyNftInfo.CollectionId,
	}

	var isExitEmptyNft = true
	err = e.bc.StateDB().PrepareNft(txInfo.NftIndex)
	if err == nil &&
		bc.StateDB().NftMap[txInfo.NftIndex].OwnerAccountIndex == account.AccountIndex {
		// Set the right nft if the owner is correct.
		exitNft = bc.StateDB().NftMap[txInfo.NftIndex]
		isExitEmptyNft = false
	}
	err = e.BaseExecutor.Prepare(context.Background())
	if err != nil {
		return err
	}

	// Set the right tx info.
	txInfo.CreatorAccountIndex = exitNft.CreatorAccountIndex
	txInfo.CreatorTreasuryRate = exitNft.CreatorTreasuryRate
	txInfo.CreatorAccountNameHash = common.FromHex(types.EmptyAccountNameHash)
	if isExitEmptyNft {
		txInfo.CreatorAccountNameHash = common.FromHex(bc.StateDB().AccountMap[exitNft.CreatorAccountIndex].AccountNameHash)
	}
	txInfo.NftL1Address = exitNft.NftL1Address
	txInfo.NftL1TokenId, _ = new(big.Int).SetString(exitNft.NftL1TokenId, 10)
	txInfo.NftContentHash = common.FromHex(exitNft.NftContentHash)
	txInfo.CollectionId = exitNft.CollectionId

	e.exitNft = exitNft
	return nil
}

func (e *FullExitNftExecutor) VerifyInputs() error {
	bc := e.bc
	txInfo := e.txInfo

	if bc.StateDB().NftMap[txInfo.NftIndex] == nil || txInfo.AccountIndex != bc.StateDB().NftMap[txInfo.NftIndex].OwnerAccountIndex {
		// The check is not fully enough, just avoid explicit error.
		if !bytes.Equal(txInfo.NftContentHash, common.FromHex(types.EmptyNftContentHash)) {
			return errors.New("invalid nft content hash")
		}
	} else {
		// The check is not fully enough, just avoid explicit error.
		if !bytes.Equal(txInfo.NftContentHash, common.FromHex(bc.StateDB().NftMap[txInfo.NftIndex].NftContentHash)) {
			return errors.New("invalid nft content hash")
		}
	}

	return nil
}

func (e *FullExitNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	if bc.StateDB().NftMap[txInfo.NftIndex] == nil || txInfo.AccountIndex != bc.StateDB().NftMap[txInfo.NftIndex].OwnerAccountIndex {
		// Do nothing.
		return nil
	}

	// Set nft to empty nft.
	emptyNftInfo := types.EmptyNftInfo(txInfo.NftIndex)
	emptyNft := &nft.L2Nft{
		NftIndex:            emptyNftInfo.NftIndex,
		CreatorAccountIndex: emptyNftInfo.CreatorAccountIndex,
		OwnerAccountIndex:   emptyNftInfo.OwnerAccountIndex,
		NftContentHash:      emptyNftInfo.NftContentHash,
		NftL1Address:        emptyNftInfo.NftL1Address,
		NftL1TokenId:        emptyNftInfo.NftL1TokenId,
		CreatorTreasuryRate: emptyNftInfo.CreatorTreasuryRate,
		CollectionId:        emptyNftInfo.CollectionId,
	}
	bc.StateDB().NftMap[txInfo.NftIndex] = emptyNft

	stateCache := e.bc.StateDB()
	stateCache.PendingUpdateNftIndexMap[txInfo.NftIndex] = statedb.StateCachePending
	return e.BaseExecutor.ApplyTransaction()
}

func (e *FullExitNftExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeFullExitNft))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CreatorTreasuryRate)))
	buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.CollectionId)))
	chunk1 := common2.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(common2.AddressStrToBytes(txInfo.NftL1Address))
	chunk2 := common2.PrefixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk1)
	buf.Write(chunk2)
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.AccountNameHash))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.CreatorAccountNameHash))
	buf.Write(common2.PrefixPaddingBufToChunkSize(txInfo.NftContentHash))
	buf.Write(common2.Uint256ToBytes(txInfo.NftL1TokenId))
	pubData := buf.Bytes()

	stateCache := e.bc.StateDB()
	stateCache.PriorityOperations++
	stateCache.PubDataOffset = append(stateCache.PubDataOffset, uint32(len(stateCache.PubData)))
	stateCache.PendingOnChainOperationsPubData = append(stateCache.PendingOnChainOperationsPubData, pubData)
	stateCache.PendingOnChainOperationsHash = common2.ConcatKeccakHash(stateCache.PendingOnChainOperationsHash, pubData)
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *FullExitNftExecutor) GetExecutedTx() (*tx.Tx, error) {
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
