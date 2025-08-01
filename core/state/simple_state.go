package state

import (
	"bytes"
	"fmt"
	"frizo-blockchain/common"
	"math/big"
	"sort"
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

// ComputeRoot 計算狀態根 - 使用有序的方式
func (s *SimpleStateDB) ComputeRoot() common.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 1. 收集所有地址並排序
	addresses := make([]common.Address, 0, len(s.accounts))
	for addr := range s.accounts {
		addresses = append(addresses, addr)
	}

	// 2. 按地址的字節順序排序
	sort.Slice(addresses, func(i, j int) bool {
		return bytes.Compare(addresses[i][:], addresses[j][:]) < 0
	})

	// 3. 按順序序列化數據
	data := []byte{}
	for _, addr := range addresses {
		acc := s.accounts[addr]

		// 只有非空賬戶才包含在狀態根計算中
		if acc.Balance.Sign() == 0 && acc.Nonce == 0 {
			continue
		}

		// RLP 編碼賬戶數據
		accountData := []interface{}{
			acc.Nonce,
			acc.Balance.Bytes(),
		}

		accountRLP, err := common.RlpEncodeToBytes(accountData)
		if err != nil {
			panic(fmt.Sprintf("failed to RLP encode account: %v", err))
		}

		// 組合地址和賬戶數據
		entry := []interface{}{
			addr.Bytes(),
			accountRLP,
		}

		entryRLP, err := common.RlpEncodeToBytes(entry)
		if err != nil {
			panic(fmt.Sprintf("failed to RLP encode entry: %v", err))
		}

		data = append(data, entryRLP...)
	}

	return common.Keccak256Hash(data)
}

// GetAllAccounts 獲取所有賬戶（有序）
func (s *SimpleStateDB) GetAllAccounts() map[common.Address]*Account {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[common.Address]*Account)
	for addr, acc := range s.accounts {
		result[addr] = &Account{
			Balance: new(big.Int).Set(acc.Balance),
			Nonce:   acc.Nonce,
		}
	}
	return result
}

// GetSortedAddresses 獲取排序後的地址列表
func (s *SimpleStateDB) GetSortedAddresses() []common.Address {
	s.mu.RLock()
	defer s.mu.RUnlock()

	addresses := make([]common.Address, 0, len(s.accounts))
	for addr := range s.accounts {
		addresses = append(addresses, addr)
	}

	sort.Slice(addresses, func(i, j int) bool {
		return bytes.Compare(addresses[i][:], addresses[j][:]) < 0
	})

	return addresses
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

func (s *SimpleStateDB) copy() *SimpleStateDB {
	newState := NewSimpleStateDB()
	for addr, acc := range s.accounts {
		newState.accounts[addr] = &Account{
			Balance: new(big.Int).Set(acc.Balance),
			Nonce:   acc.Nonce,
		}
	}
	return newState
}

func (s *SimpleStateDB) Copy() *SimpleStateDB {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.copy()
}

func (s *SimpleStateDB) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 使用有序的地址列表來打印
	addresses := s.GetSortedAddresses()

	content := "State DB: \n"
	for _, addr := range addresses {
		acc := s.accounts[addr]
		content += fmt.Sprintf("Address: %s Balance: %s Nonce: %d\n",
			addr.Hex(), acc.Balance.String(), acc.Nonce)
	}
	return content
}

// IsEmpty 檢查賬戶是否為空
func (s *SimpleStateDB) IsEmpty(addr common.Address) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	acc, exists := s.accounts[addr]
	if !exists {
		return true
	}

	return acc.Balance.Sign() == 0 && acc.Nonce == 0
}

// DeleteEmptyAccounts 刪除空賬戶（可選的優化）
func (s *SimpleStateDB) DeleteEmptyAccounts() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, acc := range s.accounts {
		if acc.Balance.Sign() == 0 && acc.Nonce == 0 {
			delete(s.accounts, addr)
		}
	}
}
