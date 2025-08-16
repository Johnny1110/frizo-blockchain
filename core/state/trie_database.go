package state

import (
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/storage"
	"frizo-blockchain/trie"
)

// -------------------------------------------------------------------------------------

// trieMPT wraps the MPT implementation to implement the Trie interface
type trieMPT struct {
	mpt      *trie.ModifiedMerklePatriciaTree
	db       storage.IKVStore
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
func (t *trieMPT) Prove(key []byte, fromLevel uint, proofDb storage.IKVStore) error {
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
