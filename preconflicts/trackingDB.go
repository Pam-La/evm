package preconflicts

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/x/vm/statedb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
)

type TrackingDB struct {
	*statedb.StateDB
	tracker *AccessTracker
}

func NewTrackingDB(ctx sdk.Context, keeper statedb.Keeper, txConfig statedb.TxConfig, tracker *AccessTracker) *TrackingDB {
	originalStateDB := statedb.New(ctx, keeper, txConfig)
	return &TrackingDB{
		StateDB: originalStateDB,
		tracker: tracker,
	}
}

func (ts *TrackingDB) GetBalance(addr common.Address) *uint256.Int {
	ts.tracker.TrackBalanceRead(addr)
	return ts.StateDB.GetBalance(addr)
}

func (ts *TrackingDB) AddBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	ts.tracker.TrackBalanceRead(addr)
	ts.tracker.TrackBalanceWrite(addr)
	return ts.StateDB.AddBalance(addr, amount, reason)
}

func (ts *TrackingDB) SubBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) uint256.Int {
	ts.tracker.TrackBalanceRead(addr)
	ts.tracker.TrackBalanceWrite(addr)
	return ts.StateDB.SubBalance(addr, amount, reason)
}

func (ts *TrackingDB) GetNonce(addr common.Address) uint64 {
	ts.tracker.TrackNonceRead(addr)
	return ts.StateDB.GetNonce(addr)
}

func (ts *TrackingDB) SetNonce(addr common.Address, nonce uint64, reason tracing.NonceChangeReason) {
	ts.tracker.TrackNonceWrite(addr)
	ts.StateDB.SetNonce(addr, nonce, reason)
}

func (ts *TrackingDB) GetState(addr common.Address, key common.Hash) common.Hash {
	ts.tracker.TrackStorageRead(addr, key)
	return ts.StateDB.GetState(addr, key)
}

func (ts *TrackingDB) SetState(addr common.Address, key, value common.Hash) common.Hash {
	ts.tracker.TrackStorageWrite(addr, key)
	return ts.StateDB.SetState(addr, key, value)
}

func (ts *TrackingDB) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	ts.tracker.TrackStorageRead(addr, key)
	return ts.StateDB.GetCommittedState(addr, key)
}

func (ts *TrackingDB) GetStorageRoot(addr common.Address) common.Hash {
	root := ts.StateDB.GetStorageRoot(addr)
	ts.tracker.TrackStorageRead(addr, root)
	return root
}

func (ts *TrackingDB) GetCode(addr common.Address) []byte {
	ts.tracker.TrackCodeRead(addr)
	return ts.StateDB.GetCode(addr)
}

func (ts *TrackingDB) SetCode(addr common.Address, code []byte) []byte {
	ts.tracker.TrackCodeWrite(addr)
	return ts.StateDB.SetCode(addr, code)
}

func (ts *TrackingDB) GetCodeSize(addr common.Address) int {
	ts.tracker.TrackCodeRead(addr)
	return ts.StateDB.GetCodeSize(addr)
}

func (ts *TrackingDB) GetCodeHash(addr common.Address) common.Hash {
	ts.tracker.TrackCodeRead(addr)
	return ts.StateDB.GetCodeHash(addr)
}

func (ts *TrackingDB) Exist(addr common.Address) bool {
	ts.tracker.TrackBalanceRead(addr)
	return ts.StateDB.Exist(addr)
}

func (ts *TrackingDB) Empty(addr common.Address) bool {
	ts.tracker.TrackBalanceRead(addr)
	ts.tracker.TrackNonceRead(addr)
	ts.tracker.TrackCodeRead(addr)
	return ts.StateDB.Empty(addr)
}

func (ts *TrackingDB) CreateAccount(addr common.Address) {
	ts.tracker.TrackBalanceWrite(addr)
	ts.tracker.TrackNonceWrite(addr)
	ts.tracker.TrackCodeWrite(addr)
	ts.StateDB.CreateAccount(addr)
}

func (ts *TrackingDB) CreateContract(addr common.Address) {
	ts.tracker.TrackCodeWrite(addr)
	ts.StateDB.CreateContract(addr)
}

func (ts *TrackingDB) SelfDestruct(addr common.Address) uint256.Int {
	ts.tracker.TrackBalanceWrite(addr)
	ts.tracker.TrackNonceWrite(addr)
	ts.tracker.TrackCodeWrite(addr)
	return ts.StateDB.SelfDestruct(addr)
}

func (ts *TrackingDB) SelfDestruct6780(addr common.Address) (uint256.Int, bool) {
	ts.tracker.TrackBalanceWrite(addr)
	ts.tracker.TrackNonceWrite(addr)
	ts.tracker.TrackCodeWrite(addr)
	return ts.StateDB.SelfDestruct6780(addr)
}

// 계정 상태 확인
func (ts *TrackingDB) HasSelfDestructed(addr common.Address) bool {
	return ts.StateDB.HasSelfDestructed(addr)
}

func (ts *TrackingDB) GetRefund() uint64 {
	return ts.StateDB.GetRefund()
}

func (ts *TrackingDB) AddRefund(amount uint64) {
	ts.StateDB.AddRefund(amount)
}

func (ts *TrackingDB) SubRefund(amount uint64) {
	ts.StateDB.SubRefund(amount)
}

func (ts *TrackingDB) AddLog(log *ethtypes.Log) {
	ts.StateDB.AddLog(log)
}

func (ts *TrackingDB) AddPreimage(hash common.Hash, preimage []byte) {
	ts.StateDB.AddPreimage(hash, preimage)
}

func (ts *TrackingDB) Snapshot() int {
	return ts.StateDB.Snapshot()
}

func (ts *TrackingDB) RevertToSnapshot(snapshot int) {
	ts.StateDB.RevertToSnapshot(snapshot)
}

func (ts *TrackingDB) SetTransientState(addr common.Address, key, value common.Hash) {
	ts.StateDB.SetTransientState(addr, key, value)
}

func (ts *TrackingDB) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	return ts.StateDB.GetTransientState(addr, key)
}

func (ts *TrackingDB) PointCache() *utils.PointCache {
	return ts.StateDB.PointCache()
}

func (ts *TrackingDB) Witness() *stateless.Witness {
	return ts.StateDB.Witness()
}

func (ts *TrackingDB) AccessEvents() *state.AccessEvents {
	return ts.StateDB.AccessEvents()
}

func (ts *TrackingDB) Finalise(deleteEmptyObjects bool) {
	ts.StateDB.Finalise(deleteEmptyObjects)
}

func (ts *TrackingDB) AddAddressToAccessList(addr common.Address) {
	ts.StateDB.AddAddressToAccessList(addr)
}

func (ts *TrackingDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	ts.StateDB.AddSlotToAccessList(addr, slot)
}

func (ts *TrackingDB) AddressInAccessList(addr common.Address) bool {
	return ts.StateDB.AddressInAccessList(addr)
}

func (ts *TrackingDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressPresent bool, slotPresent bool) {
	return ts.StateDB.SlotInAccessList(addr, slot)
}

func (ts *TrackingDB) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses ethtypes.AccessList) {
	ts.StateDB.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}
