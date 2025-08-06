package storage

import (
	"frizo-blockchain/common"
	"frizo-blockchain/trie"
)

type StateStore struct {
	db *ChainDatabase
}

func NewStateStore(db *ChainDatabase) *StateStore {
	return &StateStore{db: db}
}

func (s *StateStore) Get(key []byte) ([]byte, error) {
	return s.db.stateDB.Get(key)
}
func (s *StateStore) Put(key []byte, value []byte) error {
	return s.db.stateDB.Put(key, value)
}

func (s *StateStore) Delete(key []byte) error {
	return s.db.stateDB.Delete(key)
}

func (s *StateStore) Has(key []byte) (bool, error) {
	return s.db.stateDB.Has(key)
}

func (s *StateStore) WriteCode(codeHash common.Hash, code []byte) error {
	return s.db.stateDB.Put(codeKey(codeHash), code)
}

func (s *StateStore) ReadCode(codeHash common.Hash) ([]byte, error) {
	return s.db.stateDB.Get(codeKey(codeHash))
}

func (s *StateStore) CommitTrie(mpt *trie.ModifiedMerklePatriciaTree) (common.Hash, error) {
	batch := s.db.stateDB.NewBatch()
	// iterate all dirty node and put
	root, nodes := trie.Commit()
	for hash, node := range nodes {
		batch.Put(hash.Bytes(), node)
	}

	return root, batch.Write()
}
