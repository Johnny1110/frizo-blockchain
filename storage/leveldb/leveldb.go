package leveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type LevelDB struct {
	db *leveldb.DB
}

func (l *LevelDB) Close() error {
	return l.db.Close()
}

func NewLevelDB(path string, cache int, handles int) (*LevelDB, error) {
	options := &opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB,
		Filter:                 filter.NewBloomFilter(10),
	}

	db, err := leveldb.OpenFile(path, options)
	if err != nil {
		return nil, err
	}

	return &LevelDB{db: db}, nil
}
