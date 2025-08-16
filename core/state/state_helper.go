package state

import (
	"frizo-blockchain/common"
	"frizo-blockchain/storage"
	"frizo-blockchain/trie"
)

func OpenTrie(source storage.IKVStore, root common.Hash) (Trie, error) {
	mpt, err := trie.NewMPTWithDB(source, root)
	if err != nil {
		return nil, err
	}
	return &trieMPT{mpt: mpt, db: source}, nil
}

func OpenStorageTrie(source storage.IKVStore, addrHash, root common.Hash) (Trie, error) {
	mpt, err := trie.NewMPTWithDB(source, root)
	if err != nil {
		return nil, err
	}
	return &trieMPT{mpt: mpt, db: source, addrHash: addrHash}, nil
}
