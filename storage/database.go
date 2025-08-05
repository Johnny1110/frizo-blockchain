package storage

import (
	"frizo-blockchain/common"
	"frizo-blockchain/core/state"
	"frizo-blockchain/core/types"
	"math/big"
	"sync"
)

type Database interface {
	LoadState(root common.Hash) *state.SimpleStateDB
	LoadBlockHash(number *big.Int) common.Hash
	LoadBlock(hash common.Hash) (*types.Block, error)
	StoreBlock(block *types.Block) error
	StoreState(state *state.SimpleStateDB)

	// ----------------------------------------------------

	// Node retrieves a trie node by hash
	Node(hash common.Hash) ([]byte, error)
	// Put store a trie node by hash
	Put(hash common.Hash, blob []byte) error
	// Delete delete a trie node by hash
	Delete(hash common.Hash) error
	// Has checks if a trie node exists
	Has(hash common.Hash) bool
	// NewBatch create a batch process for atomic update
	NewBatch() Batch
	// Close close db
	Close()
	// Compact trigger db compact
	Compact(start []byte, limit []byte) error
}

// Batch is a write-only batch that commits changes atomically
type Batch interface {
	Put(key []byte, value []byte) error
	Delete(key []byte) error
	// ValueSize return size of pending writes
	ValueSize() int
	// Write commits all pending operations atomically
	Write() error
	// Reset clears the batch
	Reset()
	// Replay replays the batch contents to another batch
	Replay(w KeyValueWriter) error
}

// KeyValueWriter is a minimal interface for batch replay
type KeyValueWriter interface {
	Put(key []byte, value []byte) error
	Delete(key []byte) error
}

type MockDatabase struct {
	mu sync.RWMutex

	stateMap     map[common.Hash]*state.SimpleStateDB
	blockHashMap map[*big.Int]common.Hash
	blockMap     map[common.Hash]*types.Block
}

func NewMockDatabase() Database {
	return &MockDatabase{
		stateMap:     make(map[common.Hash]*state.SimpleStateDB),
		blockHashMap: make(map[*big.Int]common.Hash),
		blockMap:     make(map[common.Hash]*types.Block),
	}
}

func (m *MockDatabase) StoreState(state *state.SimpleStateDB) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.stateMap[state.ComputeRoot()] = state
}

func (m *MockDatabase) LoadState(root common.Hash) *state.SimpleStateDB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stateMap[root]
}

func (m *MockDatabase) LoadBlockHash(number *big.Int) common.Hash {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.blockHashMap[number]
}

func (m *MockDatabase) LoadBlock(hash common.Hash) (*types.Block, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.blockMap[hash], nil
}

func (m *MockDatabase) StoreBlock(block *types.Block) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	blockNumber := block.Number()
	m.blockHashMap[blockNumber] = block.Hash()
	m.blockMap[block.Hash()] = block
	return nil
}
