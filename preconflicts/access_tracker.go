package preconflicts

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type AccessTracker struct {
	txIdx int

	accessedAddresses map[common.Address]bool
	accessedHashes    map[common.Hash]bool

	accessChannel chan<- AccessMessage

	mu sync.RWMutex
}

func NewAccessTracker(txIdx int, accessChannel chan<- AccessMessage) *AccessTracker {
	return &AccessTracker{
		txIdx:             txIdx,
		accessedAddresses: make(map[common.Address]bool),
		accessedHashes:    make(map[common.Hash]bool),

		accessChannel: accessChannel,
	}
}

func (at *AccessTracker) TrackBalanceRead(addr common.Address) {
	at.accessChannel <- AccessMessage{
		txIdx: at.txIdx,
		item:  addr,
	}
	at.mu.Lock()
	at.accessedAddresses[addr] = true
	at.mu.Unlock()
}

func (at *AccessTracker) TrackBalanceWrite(addr common.Address) {
	at.accessChannel <- AccessMessage{
		txIdx: at.txIdx,
		item:  addr,
	}
	at.mu.Lock()
	at.accessedAddresses[addr] = true
	at.mu.Unlock()
}

func (at *AccessTracker) TrackNonceRead(addr common.Address) {
	at.accessChannel <- AccessMessage{
		txIdx: at.txIdx,
		item:  addr,
	}
	at.mu.Lock()
	at.accessedAddresses[addr] = true
	at.mu.Unlock()
}

func (at *AccessTracker) TrackNonceWrite(addr common.Address) {
	at.accessChannel <- AccessMessage{
		txIdx: at.txIdx,
		item:  addr,
	}
	at.mu.Lock()
	at.accessedAddresses[addr] = true
	at.mu.Unlock()
}

func (at *AccessTracker) TrackCodeRead(addr common.Address) {
	at.accessChannel <- AccessMessage{
		txIdx: at.txIdx,
		item:  addr,
	}
	at.mu.Lock()
	at.accessedAddresses[addr] = true
	at.mu.Unlock()
}

func (at *AccessTracker) TrackCodeWrite(addr common.Address) {
	at.accessChannel <- AccessMessage{
		txIdx: at.txIdx,
		item:  addr,
	}
	at.mu.Lock()
	at.accessedAddresses[addr] = true
	at.mu.Unlock()
}

func (at *AccessTracker) TrackStorageRead(addr common.Address, key common.Hash) {
	at.accessChannel <- AccessMessage{
		txIdx: at.txIdx,
		item:  key,
	}
	at.mu.Lock()
	at.accessedAddresses[addr] = true
	at.accessedHashes[key] = true
	at.mu.Unlock()
}

func (at *AccessTracker) TrackStorageWrite(addr common.Address, key common.Hash) {
	at.accessChannel <- AccessMessage{
		txIdx: at.txIdx,
		item:  key,
	}
	at.mu.Lock()
	at.accessedAddresses[addr] = true
	at.accessedHashes[key] = true
	at.mu.Unlock()
}

func (at *AccessTracker) ToAccessResult() *AccessResult {
	at.mu.RLock()
	defer at.mu.RUnlock()

	addressList := make([]common.Address, 0, len(at.accessedAddresses))
	hashList := make([]common.Hash, 0, len(at.accessedHashes))

	for addr := range at.accessedAddresses {
		addressList = append(addressList, addr)
	}

	for hash := range at.accessedHashes {
		hashList = append(hashList, hash)
	}

	return &AccessResult{
		TxIdx:       at.txIdx,
		AddressList: addressList,
		HashList:    hashList,
	}
}
