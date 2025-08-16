package storage

import (
	"errors"
	"fmt"
	"frizo-blockchain/common"
	"sync"
)

// MemoryDatabase in-memory KVStore
type MemoryDatabase struct {
	db     map[string][]byte // mock level db
	mu     sync.RWMutex
	closed bool
}

// NewInMemoryKVStore create in-memory KVStore
func NewInMemoryKVStore() IKVStore {
	return &MemoryDatabase{
		db:     make(map[string][]byte),
		closed: false,
	}
}

func (db *MemoryDatabase) Debug() {
	fmt.Println("===========================<In-Memory DB>===========================")
	for k, v := range db.db {
		fmt.Println("key: ", common.Bytes2Hex([]byte(k)), "value: ", common.Bytes2Hex(v))
	}
}

// Get 獲取值
func (db *MemoryDatabase) Get(key []byte) ([]byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return nil, errors.New("db is closed")
	}

	if key == nil {
		return nil, fmt.Errorf("key cannot be nil")
	}

	if value, ok := db.db[string(key)]; ok {
		result := make([]byte, len(value))
		copy(result, value)
		return result, nil
	}

	return nil, errors.New("key not found")
}

// Put
func (db *MemoryDatabase) Put(key []byte, value []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return errors.New("db is closed")
	}

	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	storedValue := make([]byte, len(value))
	copy(storedValue, value)
	db.db[string(key)] = storedValue

	return nil
}

// Delete
func (db *MemoryDatabase) Delete(key []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return errors.New("db is closed")
	}

	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	delete(db.db, string(key))
	return nil
}

// Has
func (db *MemoryDatabase) Has(key []byte) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return false, errors.New("db is closed")
	}

	if key == nil {
		return false, fmt.Errorf("key cannot be nil")
	}

	_, exists := db.db[string(key)]
	return exists, nil
}

// NewBatch
func (db *MemoryDatabase) NewBatch() IKVStoreBatch {
	return &memBatch{
		db:     db,
		writes: make([]batchOp, 0),
	}
}

// Close
func (db *MemoryDatabase) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return errors.New("db is closed")
	}

	db.closed = true
	db.db = nil

	return nil
}

// Compact
func (db *MemoryDatabase) Compact(start []byte, limit []byte) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return errors.New("db is closed")
	}

	// 內存數據庫不需要壓縮，直接返回
	return nil
}

// === Batch ===

// batchOp
type batchOp struct {
	isDelete bool
	key      []byte
	value    []byte
}

// memBatch
type memBatch struct {
	db     *MemoryDatabase
	writes []batchOp
	size   int
	mu     sync.Mutex
}

// Put
func (b *memBatch) Put(key []byte, value []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	b.writes = append(b.writes, batchOp{
		isDelete: false,
		key:      keyCopy,
		value:    valueCopy,
	})

	b.size += len(key) + len(value)

	return nil
}

// Delete
func (b *memBatch) Delete(key []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)

	b.writes = append(b.writes, batchOp{
		isDelete: true,
		key:      keyCopy,
	})

	b.size += len(key)

	return nil
}

func (b *memBatch) Write() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.db.mu.Lock()
	defer b.db.mu.Unlock()

	if b.db.closed {
		return errors.New("db is closed")
	}

	for _, op := range b.writes {
		if op.isDelete {
			delete(b.db.db, string(op.key))
		} else {
			b.db.db[string(op.key)] = op.value
		}
	}

	return nil
}

// Reset
func (b *memBatch) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.writes = b.writes[:0]
	b.size = 0
}

// ValueSize
func (b *memBatch) ValueSize() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.size
}

// Replay
func (b *memBatch) Replay(target IKVStoreBatch) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, op := range b.writes {
		if op.isDelete {
			if err := target.Delete(op.key); err != nil {
				return err
			}
		} else {
			if err := target.Put(op.key, op.value); err != nil {
				return err
			}
		}
	}

	return nil
}
