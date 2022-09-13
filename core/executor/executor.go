package executor

import (
	"errors"
	"github.com/bnb-chain/zkbnb/common/prove"

	sdb "github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type IBlockchain interface {
	VerifyExpiredAt(expiredAt int64) error
	VerifyNonce(accountIndex int64, nonce int64) error
	VerifyGas(gasAccountIndex, gasFeeAssetId int64) error
	StateDB() *sdb.StateDB
	DB() *sdb.ChainDB
	CurrentBlock() *block.Block
}

type TxExecutor interface {
	Prepare() error
	VerifyInputs() error
	ApplyTransaction() error
	GeneratePubData() error
	GetExecutedTx() (*tx.Tx, error)
	GenerateWitness() (*prove.TxWitness, error)
}

func NewTxExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	switch tx.TxType {
	case types.TxTypeRegisterZns:
		return NewRegisterZnsExecutor(bc, tx)
	case types.TxTypeCreatePair:
		return NewCreatePairExecutor(bc, tx)
	case types.TxTypeUpdatePairRate:
		return NewUpdatePairRateExecutor(bc, tx)
	case types.TxTypeDeposit:
		return NewDepositExecutor(bc, tx)
	case types.TxTypeDepositNft:
		return NewDepositNftExecutor(bc, tx)
	case types.TxTypeTransfer:
		return NewTransferExecutor(bc, tx)
	case types.TxTypeSwap:
		return NewSwapExecutor(bc, tx)
	case types.TxTypeAddLiquidity:
		return NewAddLiquidityExecutor(bc, tx)
	case types.TxTypeRemoveLiquidity:
		return NewRemoveLiquidityExecutor(bc, tx)
	case types.TxTypeWithdraw:
		return NewWithdrawExecutor(bc, tx)
	case types.TxTypeCreateCollection:
		return NewCreateCollectionExecutor(bc, tx)
	case types.TxTypeMintNft:
		return NewMintNftExecutor(bc, tx)
	case types.TxTypeTransferNft:
		return NewTransferNftExecutor(bc, tx)
	case types.TxTypeAtomicMatch:
		return NewAtomicMatchExecutor(bc, tx)
	case types.TxTypeCancelOffer:
		return NewCancelOfferExecutor(bc, tx)
	case types.TxTypeWithdrawNft:
		return NewWithdrawNftExecutor(bc, tx)
	case types.TxTypeFullExit:
		return NewFullExitExecutor(bc, tx)
	case types.TxTypeFullExitNft:
		return NewFullExitNftExecutor(bc, tx)
	}

	return nil, errors.New("unsupported tx type")
}
