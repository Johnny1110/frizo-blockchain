package storage

import (
	"errors"
	"frizo-blockchain/common"
	"sync"
)

// im-memory db --------------------------------------------------------------
type MemoryDatabase struct {
	db map[string][]byte
	mu sync.RWMutex
}

func NewMemoryDataBase() *MemoryDatabase {
	return &MemoryDatabase{
		db: make(map[string][]byte),
	}
}

// Node retrieves a trie node by hash
func (m *MemoryDatabase) Node(hash common.Hash) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if data, ok := m.db[hash.Hex()]; ok {
		return common.CopyBytes(data), nil
	}

	return nil, errors.New("node not found in memory db")
}

// Put store a trie node by hash
func (m *MemoryDatabase) Put(hash common.Hash, blob []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.db[hash.Hex()] = common.CopyBytes(blob)
	return nil
}

// Delete delete a trie node by hash
func (m *MemoryDatabase) Delete(hash common.Hash) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.db, hash.Hex())
	return nil
}

// Has checks if a trie node exists
func (m *MemoryDatabase) Has(hash common.Hash) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.db[hash.Hex()]
	return ok
}

// NewBatch create a batch process for atomic update
func (m *MemoryDatabase) NewBatch() Batch {
	return &memBatch{db: m}
}

// Close close db
func (m *MemoryDatabase) Close() {
	// not support compact op for in-memory db
	return
}

// Compact trigger db compact
func (m *MemoryDatabase) Compact(start []byte, limit []byte) error {
	// not support compact op for in-memory db
	return nil
}

// Batch ------------------------------------------------------------
type memBatch struct {
	db     *MemoryDatabase
	writes []batchOp
	size   int // total data size
}

type batchOp struct {
	hash   common.Hash
	data   []byte
	delete bool
}

func (m memBatch) Put(key []byte, value []byte) error {
	m.writes = append(m.writes, batchOp{
		hash: common.BytesToHash(key),
		data: common.CopyBytes(value),
	})
	m.size += len(value)
	return nil
}

func (m memBatch) Delete(key []byte) error {
	m.writes = append(m.writes, batchOp{
		hash:   common.BytesToHash(key),
		delete: true,
	})
	return nil
}

func (m memBatch) ValueSize() int {
	return m.size
}

func (m memBatch) Write() error {
	m.db.mu.Lock()
	defer m.db.mu.Unlock()

	for _, op := range m.writes {
		if op.delete {
			delete(m.db.db, op.hash.Hex())
		} else {
			m.db.db[op.hash.Hex()] = op.data
		}
	}
	return nil
}

func (m memBatch) Reset() {
	m.writes = m.writes[:0]
	m.size = 0
}

func (m memBatch) Replay(w KeyValueWriter) error {
	for _, op := range m.writes {
		if op.delete {
			if err := w.Delete(op.hash.Bytes()); err != nil {
				return err
			}
		} else {
			if err := w.Put(op.hash.Bytes(), op.data); err != nil {
				return err
			}
		}
	}
	return nil
}
