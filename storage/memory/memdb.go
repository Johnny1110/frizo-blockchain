package memory

import (
	"bytes"
	"errors"
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/storage/interfaces"
	"sort"
	"sync"
)

// MemoryDatabase 內存數據庫實現
type MemoryDatabase struct {
	db map[string][]byte // 主存儲
	mu sync.RWMutex      // 讀寫鎖

	// 性能指標
	getCounter    uint64
	putCounter    uint64
	deleteCounter uint64

	// 模擬數據庫關閉狀態
	closed bool
}

// NewMemoryDatabase 創建新的內存數據庫
func NewMemoryDatabase() *MemoryDatabase {
	return &MemoryDatabase{
		db:     make(map[string][]byte),
		closed: false,
	}
}

// NewMemoryDatabaseWithCap 創建指定容量的內存數據庫
func NewMemoryDatabaseWithCap(capacity int) *MemoryDatabase {
	return &MemoryDatabase{
		db:     make(map[string][]byte, capacity),
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

	db.getCounter++

	if value, ok := db.db[string(key)]; ok {
		// 返回值的副本，避免外部修改影響內部數據
		result := make([]byte, len(value))
		copy(result, value)
		return result, nil
	}

	return nil, errors.New("key not found")
}

// Put 存儲鍵值對
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

	db.putCounter++

	// 存儲值的副本
	storedValue := make([]byte, len(value))
	copy(storedValue, value)
	db.db[string(key)] = storedValue

	return nil
}

// Delete 刪除鍵
func (db *MemoryDatabase) Delete(key []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return errors.New("db is closed")
	}

	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	db.deleteCounter++

	delete(db.db, string(key))
	return nil
}

// Has 檢查鍵是否存在
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

// NewBatch 創建批量操作
func (db *MemoryDatabase) NewBatch() interfaces.Batch {
	return &memBatch{
		db:     db,
		writes: make([]batchOp, 0),
	}
}

// NewIterator 創建迭代器
func (db *MemoryDatabase) NewIterator(prefix []byte, start []byte) interfaces.Iterator {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// 收集所有符合條件的鍵
	var keys []string
	for key := range db.db {
		keyBytes := []byte(key)

		// 檢查前綴
		if prefix != nil && !bytes.HasPrefix(keyBytes, prefix) {
			continue
		}

		// 檢查起始位置
		if start != nil && bytes.Compare(keyBytes, start) < 0 {
			continue
		}

		keys = append(keys, key)
	}

	// 排序鍵
	sort.Strings(keys)

	// 創建數據快照
	snapshot := make(map[string][]byte, len(keys))
	for _, key := range keys {
		value := db.db[key]
		copiedValue := make([]byte, len(value))
		copy(copiedValue, value)
		snapshot[key] = copiedValue
	}

	return &memIterator{
		keys:     keys,
		snapshot: snapshot,
		index:    -1,
	}
}

// Close 關閉數據庫
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

// Compact 壓縮數據庫（內存數據庫無需實際壓縮）
func (db *MemoryDatabase) Compact(start []byte, limit []byte) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return errors.New("db is closed")
	}

	// 內存數據庫不需要壓縮，直接返回
	return nil
}

// Len 返回數據庫中的鍵值對數量
func (db *MemoryDatabase) Len() int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return len(db.db)
}

// Stats 返回數據庫統計信息
func (db *MemoryDatabase) Stats() string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return fmt.Sprintf("MemoryDatabase: entries=%d, gets=%d, puts=%d, deletes=%d",
		len(db.db), db.getCounter, db.putCounter, db.deleteCounter)
}

// Reset 清空數據庫（用於測試）
func (db *MemoryDatabase) Reset() {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.db = make(map[string][]byte)
	db.getCounter = 0
	db.putCounter = 0
	db.deleteCounter = 0
}

// GetMetrics 獲取性能指標
func (db *MemoryDatabase) GetMetrics() map[string]uint64 {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return map[string]uint64{
		"entries":      uint64(len(db.db)),
		"get_count":    db.getCounter,
		"put_count":    db.putCounter,
		"delete_count": db.deleteCounter,
	}
}

// === Batch 實現 ===

// batchOp 批量操作項
type batchOp struct {
	isDelete bool
	key      []byte
	value    []byte
}

// memBatch 內存數據庫批量操作
type memBatch struct {
	db     *MemoryDatabase
	writes []batchOp
	size   int
	mu     sync.Mutex
}

// Put 添加寫操作到批處理
func (b *memBatch) Put(key []byte, value []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	// 存儲副本
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

// Delete 添加刪除操作到批處理
func (b *memBatch) Delete(key []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	// 存儲副本
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)

	b.writes = append(b.writes, batchOp{
		isDelete: true,
		key:      keyCopy,
	})

	b.size += len(key)

	return nil
}

// Write 執行批量操作
func (b *memBatch) Write() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.db.mu.Lock()
	defer b.db.mu.Unlock()

	if b.db.closed {
		return errors.New("db is closed")
	}

	// 執行所有操作
	for _, op := range b.writes {
		if op.isDelete {
			delete(b.db.db, string(op.key))
			b.db.deleteCounter++
		} else {
			b.db.db[string(op.key)] = op.value
			b.db.putCounter++
		}
	}

	return nil
}

// Reset 重置批處理
func (b *memBatch) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.writes = b.writes[:0]
	b.size = 0
}

// ValueSize 返回批處理的大小
func (b *memBatch) ValueSize() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.size
}

// Replay 重放批處理操作到另一個批處理
func (b *memBatch) Replay(target interfaces.Batch) error {
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

// === Iterator 實現 ===

// memIterator 內存數據庫迭代器
type memIterator struct {
	keys     []string          // 排序後的鍵列表
	snapshot map[string][]byte // 數據快照
	index    int               // 當前索引
	mu       sync.Mutex
}

// Next 移動到下一個元素
func (it *memIterator) Next() bool {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.index < len(it.keys)-1 {
		it.index++
		return true
	}
	return false
}

// Prev 移動到上一個元素
func (it *memIterator) Prev() bool {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.index > 0 {
		it.index--
		return true
	}
	return false
}

// Seek 定位到指定鍵
func (it *memIterator) Seek(key []byte) bool {
	it.mu.Lock()
	defer it.mu.Unlock()

	// 二分查找
	index := sort.SearchStrings(it.keys, string(key))

	if index < len(it.keys) {
		it.index = index
		return true
	}

	return false
}

// First 移動到第一個元素
func (it *memIterator) First() bool {
	it.mu.Lock()
	defer it.mu.Unlock()

	if len(it.keys) > 0 {
		it.index = 0
		return true
	}
	return false
}

// Last 移動到最後一個元素
func (it *memIterator) Last() bool {
	it.mu.Lock()
	defer it.mu.Unlock()

	if len(it.keys) > 0 {
		it.index = len(it.keys) - 1
		return true
	}
	return false
}

// Key 返回當前鍵
func (it *memIterator) Key() []byte {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.index >= 0 && it.index < len(it.keys) {
		return []byte(it.keys[it.index])
	}
	return nil
}

// Value 返回當前值
func (it *memIterator) Value() []byte {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.index >= 0 && it.index < len(it.keys) {
		key := it.keys[it.index]
		if value, ok := it.snapshot[key]; ok {
			// 返回副本
			result := make([]byte, len(value))
			copy(result, value)
			return result
		}
	}
	return nil
}

// Valid 檢查迭代器是否有效
func (it *memIterator) Valid() bool {
	it.mu.Lock()
	defer it.mu.Unlock()

	return it.index >= 0 && it.index < len(it.keys)
}

// Error 返回迭代器錯誤
func (it *memIterator) Error() error {
	return nil
}

// Release 釋放迭代器資源
func (it *memIterator) Release() {
	it.mu.Lock()
	defer it.mu.Unlock()

	it.keys = nil
	it.snapshot = nil
}

// === 額外的測試輔助功能 ===

// Snapshot 創建數據庫快照
func (db *MemoryDatabase) Snapshot() map[string][]byte {
	db.mu.RLock()
	defer db.mu.RUnlock()

	snapshot := make(map[string][]byte, len(db.db))
	for key, value := range db.db {
		valueCopy := make([]byte, len(value))
		copy(valueCopy, value)
		snapshot[key] = valueCopy
	}

	return snapshot
}

// LoadSnapshot 從快照恢復數據庫
func (db *MemoryDatabase) LoadSnapshot(snapshot map[string][]byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return errors.New("db is closed")
	}

	db.db = make(map[string][]byte, len(snapshot))
	for key, value := range snapshot {
		valueCopy := make([]byte, len(value))
		copy(valueCopy, value)
		db.db[key] = valueCopy
	}

	return nil
}

// Keys 返回所有鍵（排序後）
func (db *MemoryDatabase) Keys() [][]byte {
	db.mu.RLock()
	defer db.mu.RUnlock()

	keys := make([]string, 0, len(db.db))
	for key := range db.db {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	result := make([][]byte, len(keys))
	for i, key := range keys {
		result[i] = []byte(key)
	}

	return result
}
