package storage

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
)

// LevelDB  goleveldb impl
type LevelDB struct {
	db   *leveldb.DB
	path string
	mu   sync.RWMutex
}

func NewLevelDB(path string, cache int, handles int) (IKVStore, error) {
	// LevelDB setup
	options := &opt.Options{
		OpenFilesCacheCapacity: handles,                   // 文件句柄緩存
		BlockCacheCapacity:     cache / 2 * opt.MiB,       // 塊緩存（讀緩存）
		WriteBuffer:            cache / 4 * opt.MiB,       // 寫緩衝區
		Filter:                 filter.NewBloomFilter(10), // Bloom 過濾器
		Compression:            opt.SnappyCompression,     // 壓縮算法
		CompactionTableSize:    2 * opt.MiB,               // 壓縮表大小
		CompactionTotalSize:    10 * opt.MiB,              // 總壓縮大小
	}

	// open database
	db, err := leveldb.OpenFile(path, options)
	if err != nil {
		// if corrupt, try recover.
		if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
			db, err = leveldb.RecoverFile(path, options)
			if err != nil {
				return nil, fmt.Errorf("failed to recover database: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
	}

	return &LevelDB{
		db:   db,
		path: path,
	}, nil
}

func (l *LevelDB) Debug() {

}

// Get
func (l *LevelDB) Get(key []byte) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("key cannot be nil")
	}

	value, err := l.db.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, errors.New("key not found")
		}
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	result := make([]byte, len(value))
	copy(result, value)

	return result, nil
}

// Put
func (l *LevelDB) Put(key []byte, value []byte) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	err := l.db.Put(key, value, nil)
	if err != nil {
		return fmt.Errorf("failed to put key: %w", err)
	}

	return nil
}

// Delete
func (l *LevelDB) Delete(key []byte) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	err := l.db.Delete(key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	return nil
}

// Has check exists
func (l *LevelDB) Has(key []byte) (bool, error) {
	if key == nil {
		return false, fmt.Errorf("key cannot be nil")
	}

	_, err := l.db.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return true, nil
}

// NewBatch create batch
func (l *LevelDB) NewBatch() IKVStoreBatch {
	return &leveldbBatch{
		db:    l.db,
		batch: new(leveldb.Batch),
		size:  0,
	}
}

// Close 關閉數據庫
func (l *LevelDB) Close() error {
	if l.db == nil {
		return nil
	}

	err := l.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	l.db = nil
	return nil
}

// Compact compact db
func (l *LevelDB) Compact(start []byte, limit []byte) error {
	return l.db.CompactRange(util.Range{
		Start: start,
		Limit: limit,
	})
}

// === Batch implements ===

// leveldbBatch batch
type leveldbBatch struct {
	db    *leveldb.DB
	batch *leveldb.Batch
	size  int
	mu    sync.Mutex
}

// Put 添加寫操作到批處理
func (b *leveldbBatch) Put(key []byte, value []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	b.batch.Put(key, value)
	b.size += len(key) + len(value)

	return nil
}

// Delete
func (b *leveldbBatch) Delete(key []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	b.batch.Delete(key)
	b.size += len(key)

	return nil
}

// Write execute
func (b *leveldbBatch) Write() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	err := b.db.Write(b.batch, nil)
	if err != nil {
		return fmt.Errorf("failed to write batch: %w", err)
	}

	return nil
}

// Reset reset batch
func (b *leveldbBatch) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.batch.Reset()
	b.size = 0
}

// ValueSize return batch size
func (b *leveldbBatch) ValueSize() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.size
}

// Replay replay batch execute to another batch
func (b *leveldbBatch) Replay(target IKVStoreBatch) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	// level db bach replay
	return b.batch.Replay(&replayHandler{target: target})
}

// replayHandler
type replayHandler struct {
	target IKVStoreBatch
}

func (h *replayHandler) Put(key, value []byte) {
	h.target.Put(key, value)
}

func (h *replayHandler) Delete(key []byte) {
	h.target.Delete(key)
}
