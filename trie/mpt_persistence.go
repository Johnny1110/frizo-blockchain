package trie

import (
	"errors"
	"frizo-blockchain/common"
	"sync"
)

type MPTDatabase interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Delete(key []byte) error
	Has(key []byte) (bool, error)
}

type InMemoryMPTDatabase struct {
	mu sync.RWMutex
	db map[string][]byte
}

func NewInMemoryMPTDatabase() MPTDatabase {
	return &InMemoryMPTDatabase{db: make(map[string][]byte)}
}

func (i *InMemoryMPTDatabase) Get(key []byte) ([]byte, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if v, ok := i.db[common.Bytes2Hex(key)]; ok {
		return v, nil
	} else {
		return nil, errors.New("key not found")
	}
}

func (i *InMemoryMPTDatabase) Put(key []byte, value []byte) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.db[common.Bytes2Hex(key)] = value
	return nil
}

func (i *InMemoryMPTDatabase) Delete(key []byte) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.db, common.Bytes2Hex(key))
	return nil
}

func (i *InMemoryMPTDatabase) Has(key []byte) (bool, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	_, ok := i.db[common.Bytes2Hex(key)]
	return ok, nil
}
