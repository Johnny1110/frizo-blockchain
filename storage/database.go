package storage

import (
	"fmt"
	"frizo-blockchain/db"
	"path/filepath"
)

type DatabaseConfig struct {
	DataDir   string // category
	Cache     int    // cache size
	Handles   int    // file handles (已打開文件數)
	Namespace string
}

type ChainDatabase struct {
	blockDB db.Database // block, txn, receipt
	stateDB db.Database // state,mpt-node, contract-code
	indexDB db.Database // index

	config *DatabaseConfig
}

func (cdb *ChainDatabase) BlockDB() db.Database {
	return cdb.blockDB
}
func (cdb *ChainDatabase) StateDB() db.Database {
	return cdb.stateDB
}
func (cdb *ChainDatabase) IndexDB() db.Database {
	return cdb.indexDB
}

// NewChainDatabaseLevelDB create new chain database
func NewChainDatabaseLevelDB(config *DatabaseConfig) (*ChainDatabase, error) {
	// create category
	blockDir := filepath.Join(config.DataDir, "block")
	stateDir := filepath.Join(config.DataDir, "state")
	indexDir := filepath.Join(config.DataDir, "index")

	// cache alloc
	blockCacheSize := config.Cache * 40 / 100 // 40%
	stateCacheSize := config.Cache * 50 / 100 // 50%
	indexCacheSize := config.Cache * 10 / 100 // 10%

	// create block db
	blockDB, err := NewLevelDB(blockDir, blockCacheSize, config.Handles/3)
	if err != nil {
		return nil, fmt.Errorf("failed to create block database: %v", err)
	}

	stateDB, err := NewLevelDB(stateDir, stateCacheSize, config.Handles/3)
	if err != nil {
		_ = blockDB.Close()
		return nil, fmt.Errorf("failed to create state database: %v", err)
	}

	indexDB, err := NewLevelDB(indexDir, indexCacheSize, config.Handles/3)
	if err != nil {
		_ = blockDB.Close()
		_ = stateDB.Close()
		return nil, fmt.Errorf("failed to create index database: %v", err)
	}

	return &ChainDatabase{
		blockDB: blockDB,
		stateDB: stateDB,
		indexDB: indexDB,
		config:  config,
	}, nil
}

// NewChainDatabaseInMemory create new chain database
func NewChainDatabaseInMemory(config *DatabaseConfig) (*ChainDatabase, error) {
	return &ChainDatabase{
		blockDB: NewMemoryDatabase(),
		stateDB: NewMemoryDatabase(),
		indexDB: NewMemoryDatabase(),
		config:  config,
	}, nil
}

func (db *ChainDatabase) Close() error {
	var errs []error

	if err := db.blockDB.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := db.stateDB.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := db.indexDB.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close databases: %v", errs)
	}
	return nil
}
