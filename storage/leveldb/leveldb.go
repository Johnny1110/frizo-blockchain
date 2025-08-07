package leveldb

import (
	"fmt"
	"frizo-blockchain/storage/interfaces"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
)

// LevelDB 封裝了 goleveldb 的實現
type LevelDB struct {
	db   *leveldb.DB
	path string

	// 性能指標（可選）
	getCounter    uint64
	putCounter    uint64
	deleteCounter uint64

	mu sync.RWMutex
}

// NewLevelDB 創建新的 LevelDB 實例
func NewLevelDB(path string, cache int, handles int) (*LevelDB, error) {
	// 設置 LevelDB 選項
	options := &opt.Options{
		OpenFilesCacheCapacity: handles,                   // 文件句柄緩存
		BlockCacheCapacity:     cache / 2 * opt.MiB,       // 塊緩存（讀緩存）
		WriteBuffer:            cache / 4 * opt.MiB,       // 寫緩衝區
		Filter:                 filter.NewBloomFilter(10), // Bloom 過濾器
		Compression:            opt.SnappyCompression,     // 壓縮算法
		CompactionTableSize:    2 * opt.MiB,               // 壓縮表大小
		CompactionTotalSize:    10 * opt.MiB,              // 總壓縮大小
	}

	// 打開數據庫
	db, err := leveldb.OpenFile(path, options)
	if err != nil {
		// 如果數據庫損壞，嘗試恢復
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

func (leveldb *LevelDB) Debug() {

}

// Get 獲取值
func (l *LevelDB) Get(key []byte) ([]byte, error) {
	if key == nil {
		return nil, fmt.Errorf("key cannot be nil")
	}

	l.mu.RLock()
	l.getCounter++
	l.mu.RUnlock()

	value, err := l.db.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, errors.New("key not found")
		}
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// 返回值的副本，避免數據被修改
	result := make([]byte, len(value))
	copy(result, value)

	return result, nil
}

// Put 存儲鍵值對
func (l *LevelDB) Put(key []byte, value []byte) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	l.mu.RLock()
	l.putCounter++
	l.mu.RUnlock()

	err := l.db.Put(key, value, nil)
	if err != nil {
		return fmt.Errorf("failed to put key: %w", err)
	}

	return nil
}

// Delete 刪除鍵
func (l *LevelDB) Delete(key []byte) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	l.mu.RLock()
	l.deleteCounter++
	l.mu.RUnlock()

	err := l.db.Delete(key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	return nil
}

// Has 檢查鍵是否存在
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

// NewBatch 創建批量操作
func (l *LevelDB) NewBatch() interfaces.Batch {
	return &leveldbBatch{
		db:    l.db,
		batch: new(leveldb.Batch),
		size:  0,
	}
}

// NewIterator 創建迭代器
func (l *LevelDB) NewIterator(prefix []byte, start []byte) interfaces.Iterator {
	var slice *util.Range

	if prefix != nil {
		slice = util.BytesPrefix(prefix)
	} else if start != nil {
		slice = &util.Range{Start: start}
	}

	return &leveldbIterator{
		iter: l.db.NewIterator(slice, nil),
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

// Compact 手動壓縮數據庫
func (l *LevelDB) Compact(start []byte, limit []byte) error {
	return l.db.CompactRange(util.Range{
		Start: start,
		Limit: limit,
	})
}

// Stats 獲取數據庫統計信息
func (l *LevelDB) Stats() (string, error) {
	stats, err := l.db.GetProperty("leveldb.stats")
	if err != nil {
		return "", fmt.Errorf("failed to get stats: %w", err)
	}
	return stats, nil
}

// Path 返回數據庫路徑
func (l *LevelDB) Path() string {
	return l.path
}

// GetMetrics 獲取性能指標
func (l *LevelDB) GetMetrics() map[string]uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return map[string]uint64{
		"get_count":    l.getCounter,
		"put_count":    l.putCounter,
		"delete_count": l.deleteCounter,
	}
}

// === Batch 實現 ===

// leveldbBatch 實現批量操作
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

// Delete 添加刪除操作到批處理
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

// Write 執行批量操作
func (b *leveldbBatch) Write() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	err := b.db.Write(b.batch, nil)
	if err != nil {
		return fmt.Errorf("failed to write batch: %w", err)
	}

	return nil
}

// Reset 重置批處理
func (b *leveldbBatch) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.batch.Reset()
	b.size = 0
}

// ValueSize 返回批處理的大小
func (b *leveldbBatch) ValueSize() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.size
}

// Replay 重放批處理操作到另一個批處理
func (b *leveldbBatch) Replay(target interfaces.Batch) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.batch.Replay(&replayHandler{target: target})
}

// replayHandler 用於重放批處理
type replayHandler struct {
	target interfaces.Batch
}

func (h *replayHandler) Put(key, value []byte) {
	h.target.Put(key, value)
}

func (h *replayHandler) Delete(key []byte) {
	h.target.Delete(key)
}

// === Iterator 實現 ===

// leveldbIterator 封裝 LevelDB 迭代器
type leveldbIterator struct {
	iter iterator.Iterator
	mu   sync.Mutex
}

// Next 移動到下一個元素
func (it *leveldbIterator) Next() bool {
	it.mu.Lock()
	defer it.mu.Unlock()

	return it.iter.Next()
}

// Prev 移動到上一個元素
func (it *leveldbIterator) Prev() bool {
	it.mu.Lock()
	defer it.mu.Unlock()

	return it.iter.Prev()
}

// Seek 定位到指定鍵
func (it *leveldbIterator) Seek(key []byte) bool {
	it.mu.Lock()
	defer it.mu.Unlock()

	return it.iter.Seek(key)
}

// First 移動到第一個元素
func (it *leveldbIterator) First() bool {
	it.mu.Lock()
	defer it.mu.Unlock()

	return it.iter.First()
}

// Last 移動到最後一個元素
func (it *leveldbIterator) Last() bool {
	it.mu.Lock()
	defer it.mu.Unlock()

	return it.iter.Last()
}

// Key 返回當前鍵
func (it *leveldbIterator) Key() []byte {
	it.mu.Lock()
	defer it.mu.Unlock()

	key := it.iter.Key()
	if key == nil {
		return nil
	}

	// 返回副本
	result := make([]byte, len(key))
	copy(result, key)
	return result
}

// Value 返回當前值
func (it *leveldbIterator) Value() []byte {
	it.mu.Lock()
	defer it.mu.Unlock()

	value := it.iter.Value()
	if value == nil {
		return nil
	}

	// 返回副本
	result := make([]byte, len(value))
	copy(result, value)
	return result
}

// Valid 檢查迭代器是否有效
func (it *leveldbIterator) Valid() bool {
	it.mu.Lock()
	defer it.mu.Unlock()

	return it.iter.Valid()
}

// Error 返回迭代器錯誤
func (it *leveldbIterator) Error() error {
	it.mu.Lock()
	defer it.mu.Unlock()

	return it.iter.Error()
}

// Release 釋放迭代器資源
func (it *leveldbIterator) Release() {
	it.mu.Lock()
	defer it.mu.Unlock()

	it.iter.Release()
}
