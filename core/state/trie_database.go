package state

import (
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/db"
	"frizo-blockchain/trie"
	"sync"
)

// trieDatabase implements the Database interface
type trieDatabase struct {
	db            db.Database // Underlying database
	codeSizeCache *lruCache   // Cache for contract code size
	codeCache     *lruCache   // Cache for contract code

	mu sync.RWMutex
}

func (t *trieDatabase) TrieDB() db.Database {
	return t.db
}

func (t *trieDatabase) OpenTrie(root common.Hash) (Trie, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	mpt, err := trie.NewMPTWithDB(t.db, root)
	if err != nil {
		return nil, err
	}
	return &trieMPT{mpt: mpt, db: t.db}, nil
}

func (t *trieDatabase) OpenStorageTrie(addrHash, root common.Hash) (Trie, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	mpt, err := trie.NewMPTWithDB(t.db, root)
	if err != nil {
		return nil, err
	}
	return &trieMPT{mpt: mpt, db: t.db, addrHash: addrHash}, nil
}

// CopyTrie creates a copy of the trie
func (t *trieDatabase) CopyTrie(trie Trie) Trie {
	switch t := trie.(type) {
	case *trieMPT:
		return &trieMPT{
			mpt:      t.mpt.Copy(),
			db:       t.db,
			addrHash: t.addrHash,
		}
	default:
		panic(fmt.Errorf("unknown trie type %T", t))
	}
}

// ContractCode retrieves the code of a contract
func (t *trieDatabase) ContractCode(addrHash, codeHash common.Hash) ([]byte, error) {
	t.mu.RLock()
	// check cache first
	if code, ok := t.codeCache.Get(codeHash); ok {
		t.mu.RUnlock()
		return code.([]byte), nil
	}

	t.mu.RUnlock()

	// Load from database
	t.mu.Lock()
	defer t.mu.Unlock()

	code, err := t.db.Get(codeKey(codeHash))
	if err != nil {
		return nil, err
	}

	// Update code cache
	t.codeCache.Add(codeHash, code)
	return code, nil
}

func (t *trieDatabase) ContractCodeSize(addrHash, codeHash common.Hash) (int, error) {
	t.mu.RLock()
	// check cache first
	if size, ok := t.codeSizeCache.Get(codeHash); ok {
		t.mu.RUnlock()
		return size.(int), nil
	}
	t.mu.RUnlock()

	// load from db
	code, err := t.ContractCode(addrHash, codeHash)
	if err != nil {
		return 0, err
	}

	size := len(code)
	// store into cache
	t.mu.Lock()
	t.codeSizeCache.Add(codeHash, size)
	t.mu.Unlock()

	return size, nil
}

// ContractCodeWithPrefix retrieves the code with a prefix
func (t *trieDatabase) ContractCodeWithPrefix(addrHash, codeHash common.Hash) ([]byte, error) {
	return t.ContractCode(addrHash, codeHash)
}

func (t *trieDatabase) Debug() {
	t.db.Debug()
}

// -------------------------------------------------------------------------------------

// trieMPT wraps the MPT implementation to implement the Trie interface
type trieMPT struct {
	mpt      *trie.ModifiedMerklePatriciaTree
	db       db.Database
	addrHash common.Hash // this is for storage tries
}

func (t *trieMPT) TryGet(key []byte) ([]byte, error) {
	if t.mpt == nil {
		return nil, fmt.Errorf("trie not open")
	}
	return t.mpt.Get(key)
}

func (t *trieMPT) TryUpdate(key, value []byte) error {
	if t.mpt == nil {
		return fmt.Errorf("trie not open")
	}
	return t.mpt.Put(key, value)
}

func (t *trieMPT) TryDelete(key []byte) error {
	if t.mpt == nil {
		return fmt.Errorf("trie not open")
	}
	return t.mpt.Delete(key)
}

func (t *trieMPT) Commit() (common.Hash, error) {
	batch := t.db.NewBatch()
	// iterate all dirty node and put
	root, nodes, err := t.mpt.Commit()

	if err != nil {
		return common.Hash{}, err
	}

	for hash, node := range nodes {
		err := batch.Put(hash.Bytes(), node)
		if err != nil {
			return common.Hash{}, err
		}
	}

	return root, batch.Write()
}

func (t *trieMPT) Hash() common.Hash {
	return t.mpt.GetRoot()
}

func (t *trieMPT) NodeIterator(startKey []byte) NodeIterator {
	//TODO implement me
	panic("implement me")
}

// Prove generates a merkle proof for a key
func (t *trieMPT) Prove(key []byte, fromLevel uint, proofDb db.Database) error {
	proof, err := t.mpt.GenerateProof(key)
	if err != nil {
		return err
	}

	// Store proof nodes in the proofDb
	for _, node := range proof.Proof {
		hash := common.Keccak256Hash(node)
		if err := proofDb.Put(hash.Bytes(), node); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------------------------------------------------------------------

// NewTrieDatabase creates a new trie database
func NewTrieDatabase(db db.Database) Database {
	return &trieDatabase{
		db:            db,
		codeSizeCache: newLRUCache(100),
		codeCache:     newLRUCache(100),
	}
}

// Simple LRU cache implementation
type lruCache struct {
	cache map[common.Hash]interface{}
	size  int
	mu    sync.RWMutex
}

func newLRUCache(size int) *lruCache {
	return &lruCache{
		cache: make(map[common.Hash]interface{}),
		size:  size,
	}
}

// Get get data from cache
func (c *lruCache) Get(key common.Hash) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.cache[key]
	return val, ok
}

// Add add data into cache
func (c *lruCache) Add(key common.Hash, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cache) >= c.size {
		// remove random entry
		for hash := range c.cache {
			delete(c.cache, hash)
			break
		}
	}
	c.cache[key] = value
}

// Helper function to generate code storage key
func codeKey(hash common.Hash) []byte {
	return append([]byte("code-"), hash.Bytes()...)
}
