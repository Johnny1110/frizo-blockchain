package storage

type Database interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Delete(key []byte) error
	Has(key []byte) (bool, error)

	NewBatch() Batch
	NewIterator(prefix []byte, start []byte) Iterator

	Close() error
	Compact(start []byte, limit []byte) error
	Debug()
}

// Batch
type Batch interface {
	Put(key []byte, value []byte) error
	Delete(key []byte) error
	Write() error
	Reset()
	ValueSize() int
	Replay(target Batch) error
}

// Iterator
type Iterator interface {
	Next() bool
	Prev() bool
	Seek(key []byte) bool
	First() bool
	Last() bool

	Key() []byte
	Value() []byte

	Valid() bool
	Error() error
	Release()
}
