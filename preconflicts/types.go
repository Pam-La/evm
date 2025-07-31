package preconflicts

import "github.com/ethereum/go-ethereum/common"

// 트랜잭션이 주소나 해시에 접근할 때, 전송하는 메시지
type AccessMessage struct {
	txIdx int
	item  any
}

// 트랜잭션이 주소나 해시에 접근한 결과
type AccessResult struct {
	TxIdx       int
	AddressList []common.Address
	HashList    []common.Hash
}
