package preconflicts

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/utils"
	"github.com/cosmos/evm/x/vm/keeper"
	"github.com/cosmos/evm/x/vm/statedb"
	"github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

type ConflictAnalyzer struct {
	*AccessCounter
	evmKeeper *keeper.Keeper
}

func NewConflictAnalyzer(evmKeeper *keeper.Keeper) *ConflictAnalyzer {
	return &ConflictAnalyzer{
		AccessCounter: NewAccessCounter(),
		evmKeeper:     evmKeeper,
	}
}

func (ca *ConflictAnalyzer) DetectWithTxRun(ctx sdk.Context, ethTx *ethtypes.Transaction, txIdx int) (*AccessResult, error) {
	signer := ethtypes.MakeSigner(types.GetEthChainConfig(), big.NewInt(ctx.BlockHeight()), uint64(ctx.BlockTime().Unix()))
	msg, err := core.TransactionToMessage(ethTx, signer, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	tracker := NewAccessTracker(txIdx, ca.AccessChannel())

	cfg := &statedb.EVMConfig{
		Params:   ca.evmKeeper.GetParams(ctx),
		CoinBase: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		BaseFee:  ca.evmKeeper.GetBaseFee(ctx),
	}

	txConfig := statedb.NewEmptyTxConfig(ethTx.Hash())

	trackingDB := NewTrackingDB(ctx, ca.evmKeeper, txConfig, tracker)

	evm := ca.evmKeeper.NewEVM(ctx, *msg, cfg, nil, trackingDB)

	value, err := utils.Uint256FromBigInt(msg.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to convert value to uint256: %w", err)
	}

	gasLimit := uint64(0xffffffffffffffff)

	sender := vm.AccountRef(msg.From)

	if msg.To == nil {
		_, _, _, err = evm.Create(sender.Address(), msg.Data, gasLimit, value)
	} else {
		_, _, err = evm.Call(sender.Address(), *msg.To, msg.Data, gasLimit, value)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute transaction: %w", err)
	}

	return tracker.ToAccessResult(), nil
}
