package state

import (
	"frizo-blockchain/common"
	"math/big"
	"sync"
)

type Account struct {
	Balance *big.Int
	Nonce   uint64
}

// SimpleStateDB TODO: replace this to MPT implement
type SimpleStateDB struct {
	mu       sync.RWMutex
	accounts map[common.Address]*Account

	// for rollback
	snapshots []map[common.Address]*Account
}

func NewSimpleStateDB() *SimpleStateDB {
	return &SimpleStateDB{
		accounts: make(map[common.Address]*Account),
	}
}

func (s *SimpleStateDB) GetBalance(addr common.Address) *big.Int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if acc, exists := s.accounts[addr]; exists {
		return new(big.Int).Set(acc.Balance)
	}
	return big.NewInt(0)
}

func (s *SimpleStateDB) GetNonce(addr common.Address) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if acc, exists := s.accounts[addr]; exists {
		return acc.Nonce
	}
	return 0
}

func (s *SimpleStateDB) SetBalance(addr common.Address, amount *big.Int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	acc := s.getOrCreateAccount(addr)
	acc.Balance = new(big.Int).Set(amount)
}

func (s *SimpleStateDB) AddBalance(addr common.Address, amount *big.Int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	acc := s.getOrCreateAccount(addr)
	acc.Balance.Add(acc.Balance, amount)
}

func (s *SimpleStateDB) SubBalance(addr common.Address, amount *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	acc := s.getOrCreateAccount(addr)
	if acc.Balance.Cmp(amount) < 0 {
		return common.ErrInsufficientFunds
	}
	acc.Balance.Sub(acc.Balance, amount)
	return nil
}

func (s *SimpleStateDB) SetNonce(addr common.Address, nonce uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	acc := s.getOrCreateAccount(addr)
	acc.Nonce = nonce
}

func (s *SimpleStateDB) Transfer(from, to common.Address, amount *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fromAcc := s.getOrCreateAccount(from)
	if fromAcc.Balance.Cmp(amount) < 0 {
		return common.ErrInsufficientFunds
	}

	toAcc := s.getOrCreateAccount(to)
	fromAcc.Balance.Sub(fromAcc.Balance, amount)
	toAcc.Balance.Add(toAcc.Balance, amount)

	return nil
}

func (s *SimpleStateDB) Snapshot() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := make(map[common.Address]*Account)
	for addr, acc := range s.accounts {
		snapshot[addr] = &Account{
			Balance: new(big.Int).Set(acc.Balance),
			Nonce:   acc.Nonce,
		}
	}

	s.snapshots = append(s.snapshots, snapshot)
	return len(s.snapshots) - 1
}

func (s *SimpleStateDB) RevertToSnapshot(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id < 0 || id >= len(s.snapshots) {
		return
	}

	s.accounts = s.snapshots[id]
	s.snapshots = s.snapshots[:id]
}

// ComputeRoot
func (s *SimpleStateDB) ComputeRoot() common.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := []byte{}
	for addr, acc := range s.accounts {

		rawData := []interface{}{
			addr.Bytes(),
			acc.Balance,
			acc.Nonce,
		}

		bytes, err := common.RlpEncodeToBytes(rawData)
		if err != nil {
			panic(err)
		}

		data = append(data, bytes...)
	}

	return common.Keccak256Hash(data)
}

// 內部方法
func (s *SimpleStateDB) getOrCreateAccount(addr common.Address) *Account {
	if acc, exists := s.accounts[addr]; exists {
		return acc
	}

	acc := &Account{
		Balance: big.NewInt(0),
		Nonce:   0,
	}
	s.accounts[addr] = acc
	return acc
}

func (s *SimpleStateDB) Copy() *SimpleStateDB {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newState := NewSimpleStateDB()
	for addr, acc := range s.accounts {
		newState.accounts[addr] = &Account{
			Balance: new(big.Int).Set(acc.Balance),
			Nonce:   acc.Nonce,
		}
	}
	return newState
}
