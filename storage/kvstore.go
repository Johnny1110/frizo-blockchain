package storage

// IKVStoreReader
type IKVStoreReader interface {
	Has(key []byte) (bool, error)
	Get(key []byte) ([]byte, error)
}

// IKVStoreWriter
type IKVStoreWriter interface {
	Put(key []byte, value []byte) error
	Delete(key []byte) error
}

// IKVStoreDeleter
type IKVStoreDeleter interface {
	Delete(key []byte) error
}

// Database
type IKVStore interface {
	IKVStoreReader
	IKVStoreWriter
	IKVStoreDeleter

	NewBatch() IKVStoreBatch
	Close() error
}

// IKVStoreBatch batch access
type IKVStoreBatch interface {
	IKVStoreWriter
	ValueSize() int
	Write() error
	Reset()
	Replay(w IKVStoreBatch) error
}
