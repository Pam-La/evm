package preconflicts

import "github.com/ethereum/go-ethereum/common"

type accessInfo struct {
	accessTxs map[int]bool
	count     uint64
}

func newAccessInfo() *accessInfo {
	return &accessInfo{
		accessTxs: make(map[int]bool),
		count:     0,
	}
}

type AccessCounter struct {
	addressCount map[common.Address]*accessInfo
	hashCount    map[common.Hash]*accessInfo

	txMax map[int]int // txIdx -> max count

	accessChannel chan AccessMessage
}

func NewAccessCounter() *AccessCounter {
	return &AccessCounter{
		addressCount:  make(map[common.Address]*accessInfo),
		hashCount:     make(map[common.Hash]*accessInfo),
		txMax:         make(map[int]int),
		accessChannel: make(chan AccessMessage, 1024),
	}
}

func (ac *AccessCounter) StartCounter() {
	go func() {
		for msg := range ac.accessChannel {
			txIdx := msg.txIdx
			item := msg.item

			switch item := item.(type) {
			case common.Address:
				if _, ok := ac.addressCount[item]; !ok {
					ac.addressCount[item] = newAccessInfo()
				}
				if _, ok := ac.addressCount[item].accessTxs[txIdx]; !ok {
					ac.addressCount[item].accessTxs[txIdx] = true
					ac.addressCount[item].count++
				}

				// 최대 접근 횟수 업데이트
				if _, ok := ac.txMax[txIdx]; !ok {
					ac.txMax[txIdx] = 0
				} else {
					ac.txMax[txIdx] = max(ac.txMax[txIdx], len(ac.addressCount[item].accessTxs))
				}
			case common.Hash:
				if _, ok := ac.hashCount[item]; !ok {
					ac.hashCount[item] = newAccessInfo()
				}
				if _, ok := ac.hashCount[item].accessTxs[txIdx]; !ok {
					ac.hashCount[item].accessTxs[txIdx] = true
					ac.hashCount[item].count++
				}

				// 최대 접근 횟수 업데이트
				if _, ok := ac.txMax[txIdx]; !ok {
					ac.txMax[txIdx] = 0
				} else {
					ac.txMax[txIdx] = max(ac.txMax[txIdx], len(ac.hashCount[item].accessTxs))
				}
			}
		}
	}()
}

func (ac *AccessCounter) AccessChannel() chan<- AccessMessage {
	return ac.accessChannel
}

func (ac *AccessCounter) IsConflict(txIdx int) bool {
	return ac.txMax[txIdx] > 1
}

func (ac *AccessCounter) GetAddressCount(addr common.Address) uint64 {
	return ac.addressCount[addr].count
}

func (ac *AccessCounter) GetAddressAccessTxs(addr common.Address) []int {
	accessTxs := ac.addressCount[addr].accessTxs
	accessTxsList := make([]int, 0, len(accessTxs))
	for txIdx := range accessTxs {
		accessTxsList = append(accessTxsList, txIdx)
	}

	return accessTxsList
}

func (ac *AccessCounter) GetHashCount(hash common.Hash) uint64 {
	return ac.hashCount[hash].count
}

func (ac *AccessCounter) GetHashAccessTxs(hash common.Hash) []int {
	accessTxs := ac.hashCount[hash].accessTxs
	accessTxsList := make([]int, 0, len(accessTxs))
	for txIdx := range accessTxs {
		accessTxsList = append(accessTxsList, txIdx)
	}

	return accessTxsList
}

func (ac *AccessCounter) StopCounter() {
	close(ac.accessChannel)
}
