package tests

import (
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/storage"
	"frizo-blockchain/trie"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_state_store_commit_trie(t *testing.T) {
	config := &storage.DatabaseConfig{
		DataDir:   "/Users/johnny.wang/projects",
		Cache:     2,
		Handles:   3,
		Namespace: "test",
	}
	chainDB, err := storage.NewChainDatabaseInMemory(config)
	assert.Nil(t, err)

	stateStore := storage.NewStateStore(chainDB)
	assert.NotNil(t, stateStore)

	// create MPT
	mpt, err := trie.NewMPTWithDB(chainDB.StateDB(), common.Hash{})
	assert.Nil(t, err)
	assert.NotNil(t, mpt)

	putTestData(mpt)

	rootHash, err := stateStore.CommitTrie(mpt)
	assert.Nil(t, err)
	assert.NotNil(t, rootHash)
	fmt.Println("commit rootHash:", rootHash)

	chainDB.StateDB().Debug()

	// restore MPT from database
	mpt, err = trie.NewMPTWithDB(chainDB.StateDB(), rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, mpt)

	mpt.PrintTree()
}

func putTestData(mpt *trie.ModifiedMerklePatriciaTree) {
	testData := map[string]string{
		"cat":   "animal",
		"car":   "vehicle",
		"card":  "payment",
		"care":  "emotion",
		"dog":   "animal",
		"dodge": "action",
		"door":  "entrance",
	}

	for key, value := range testData {
		_ = mpt.Put([]byte(key), []byte(value))
	}

	mpt.PrintTree()
	mpt.PrintStats()
}

func Test_state_store_commit_trie_levelDB(t *testing.T) {
	config := &storage.DatabaseConfig{
		DataDir:   "/Users/johnny.wang/projects",
		Cache:     2,
		Handles:   3,
		Namespace: "test",
	}
	chainDB, err := storage.NewChainDatabaseLevelDB(config)
	assert.Nil(t, err)

	stateStore := storage.NewStateStore(chainDB)
	assert.NotNil(t, stateStore)

	// create MPT
	mpt, err := trie.NewMPTWithDB(chainDB.StateDB(), common.Hash{})
	assert.Nil(t, err)
	assert.NotNil(t, mpt)

	putTestData(mpt)

	rootHash, err := stateStore.CommitTrie(mpt)
	assert.Nil(t, err)
	assert.NotNil(t, rootHash)
	fmt.Println("commit rootHash:", rootHash)

	chainDB.StateDB().Debug()

	// restore MPT from database
	mpt, err = trie.NewMPTWithDB(chainDB.StateDB(), rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, mpt)

	mpt.PrintTree()
}
