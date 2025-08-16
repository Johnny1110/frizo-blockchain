package state

import (
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/storage"
	"github.com/stretchr/testify/assert"
	"math/big"
	"sort"
	"testing"
)

type testStruct struct {
	id   int
	name string
}

func Test_search(t *testing.T) {
	revisionId := 3

	testArr := make([]*testStruct, 0)
	testArr = append(testArr, &testStruct{1, "a"})
	testArr = append(testArr, &testStruct{2, "a"})
	testArr = append(testArr, &testStruct{3, "a"})
	idx := sort.Search(len(testArr), func(i int) bool {
		return testArr[i].id >= revisionId
	})

	fmt.Println(idx)
}

func Test_journal_simple_revert(t *testing.T) {
	sdb := mockStateDB(t)
	assert.NotNil(t, sdb)
	myAddr := common.HexToAddress("0x521147b68d948f24A341E5C0Da50Fa5AA39A06D7")
	version := sdb.Snapshot()
	fmt.Println("snapshot version:", version)
	sdb.CreateAccount(myAddr)
	fmt.Println("init balance:", sdb.GetBalance(myAddr))
	fmt.Println("init nonce:", sdb.GetNonce(myAddr))
	err := sdb.AddBalance(myAddr, big.NewInt(100))
	assert.Nil(t, err)
	sdb.SetNonce(myAddr, 1)
	fmt.Println("after add balance:", sdb.GetBalance(myAddr))
	fmt.Println("after add nonce:", sdb.GetNonce(myAddr))
	// do revert:
	sdb.Revert(version)
	fmt.Println("revert balance:", sdb.GetBalance(myAddr))
	fmt.Println("revert nonce:", sdb.GetNonce(myAddr))
	assert.Equal(t, sdb.GetBalance(myAddr), big.NewInt(0))
	assert.Equal(t, sdb.GetNonce(myAddr), uint64(0))
}

func Test_journal_commit(t *testing.T) {
	sdb := mockStateDB(t)
	assert.NotNil(t, sdb)
	myAddr := common.HexToAddress("0x521147b68d948f24A341E5C0Da50Fa5AA39A06D7")
	version := sdb.Snapshot()
	fmt.Println("snapshot version:", version)
	sdb.CreateAccount(myAddr)
	fmt.Println("init balance:", sdb.GetBalance(myAddr))
	fmt.Println("init nonce:", sdb.GetNonce(myAddr))
	err := sdb.AddBalance(myAddr, big.NewInt(100))
	assert.Nil(t, err)
	sdb.SetNonce(myAddr, 1)
	fmt.Println("after add balance:", sdb.GetBalance(myAddr))
	fmt.Println("after add nonce:", sdb.GetNonce(myAddr))
	// do revert:
	commit, err := sdb.Commit(true)
	assert.Nil(t, err)
	fmt.Println("commit hash:", commit)

	fmt.Println("revert balance:", sdb.GetBalance(myAddr))
	fmt.Println("revert nonce:", sdb.GetNonce(myAddr))
	assert.Equal(t, big.NewInt(100), sdb.GetBalance(myAddr))
	assert.Equal(t, uint64(1), sdb.GetNonce(myAddr))

	sdb.Database().Debug()
}

func Test_journal_do_commit_then_revert_then_commit(t *testing.T) {
	sdb := mockStateDB(t)
	assert.NotNil(t, sdb)
	myAddr := common.HexToAddress("0x521147b68d948f24A341E5C0Da50Fa5AA39A06D7")
	// DO commit 1 txn
	version_1 := sdb.Snapshot()
	fmt.Println("snapshot version:", version_1)
	sdb.CreateAccount(myAddr)
	fmt.Println("init balance:", sdb.GetBalance(myAddr))
	fmt.Println("init nonce:", sdb.GetNonce(myAddr))
	err := sdb.AddBalance(myAddr, big.NewInt(100))
	assert.Nil(t, err)
	sdb.SetNonce(myAddr, uint64(1))
	fmt.Println("after add balance:", sdb.GetBalance(myAddr))
	fmt.Println("after add nonce:", sdb.GetNonce(myAddr))
	root, err := sdb.Commit(true)
	assert.Nil(t, err)
	fmt.Println("commit root:", root)

	// DO another txn then snapshot again
	_ = sdb.Snapshot()
	err = sdb.SubBalance(myAddr, big.NewInt(50))
	assert.Nil(t, err)

	// DO another txn
	version_3 := sdb.Snapshot()
	sdb.SetNonce(myAddr, uint64(2))
	sdb.Revert(version_3) // revert to version 2

	// final commit
	root, err = sdb.Commit(true)
	fmt.Println("final root:", root)
	fmt.Println("final balance:", sdb.GetBalance(myAddr))
	fmt.Println("final nonce:", sdb.GetNonce(myAddr))

	assert.Equal(t, big.NewInt(50), sdb.GetBalance(myAddr))
	assert.Equal(t, uint64(1), sdb.GetNonce(myAddr))
}

func Test_journal_complex_revert(t *testing.T) {
	sdb := mockStateDB(t)
	assert.NotNil(t, sdb)
	myAddr := common.HexToAddress("0x521147b68d948f24A341E5C0Da50Fa5AA39A06D7")
	// DO commit 1 txn
	version_1 := sdb.Snapshot()
	fmt.Println("snapshot version:", version_1)
	sdb.CreateAccount(myAddr)
	fmt.Println("init balance:", sdb.GetBalance(myAddr))
	fmt.Println("init nonce:", sdb.GetNonce(myAddr))
	err := sdb.AddBalance(myAddr, big.NewInt(100))
	assert.Nil(t, err)
	sdb.SetNonce(myAddr, uint64(1))
	fmt.Println("after add balance:", sdb.GetBalance(myAddr))
	fmt.Println("after add nonce:", sdb.GetNonce(myAddr))
	root, err := sdb.Commit(true)
	assert.Nil(t, err)
	fmt.Println("commit root:", root)

	// DO another txn then snapshot again
	_ = sdb.Snapshot()
	err = sdb.SubBalance(myAddr, big.NewInt(50))
	assert.Nil(t, err)

	// DO another txn
	version_3 := sdb.Snapshot()
	sdb.SetNonce(myAddr, uint64(2))

	// DO another txn
	_ = sdb.Snapshot()
	sdb.SetNonce(myAddr, uint64(10))
	err = sdb.AddBalance(myAddr, big.NewInt(10000))
	assert.Nil(t, err)

	sdb.Revert(version_3) // revert to version 2

	// final commit
	root, err = sdb.Commit(true)
	fmt.Println("final root:", root)
	fmt.Println("final balance:", sdb.GetBalance(myAddr))
	fmt.Println("final nonce:", sdb.GetNonce(myAddr))

	assert.Equal(t, big.NewInt(50), sdb.GetBalance(myAddr))
	assert.Equal(t, uint64(1), sdb.GetNonce(myAddr))
}

func Test_journal_complex_commit(t *testing.T) {
	sdb := mockStateDB(t)
	assert.NotNil(t, sdb)
	myAddr := common.HexToAddress("0x521147b68d948f24A341E5C0Da50Fa5AA39A06D7")
	// DO commit 1 txn
	version_1 := sdb.Snapshot()
	fmt.Println("snapshot version:", version_1)
	sdb.CreateAccount(myAddr)
	fmt.Println("init balance:", sdb.GetBalance(myAddr))
	fmt.Println("init nonce:", sdb.GetNonce(myAddr))
	err := sdb.AddBalance(myAddr, big.NewInt(100))
	assert.Nil(t, err)
	sdb.SetNonce(myAddr, uint64(1))
	fmt.Println("after add balance:", sdb.GetBalance(myAddr))
	fmt.Println("after add nonce:", sdb.GetNonce(myAddr))
	root, err := sdb.Commit(true)
	assert.Nil(t, err)
	fmt.Println("commit root:", root)

	// DO another txn then snapshot again
	_ = sdb.Snapshot()
	err = sdb.SubBalance(myAddr, big.NewInt(50))
	assert.Nil(t, err)

	// DO another txn
	_ = sdb.Snapshot()
	sdb.SetNonce(myAddr, uint64(2))

	// DO another txn
	_ = sdb.Snapshot()
	sdb.SetNonce(myAddr, uint64(3))
	err = sdb.AddBalance(myAddr, big.NewInt(300))
	assert.Nil(t, err)
	// final commit
	root, err = sdb.Commit(true)
	fmt.Println("final root:", root)
	fmt.Println("final balance:", sdb.GetBalance(myAddr))
	fmt.Println("final nonce:", sdb.GetNonce(myAddr))

	assert.Equal(t, big.NewInt(350), sdb.GetBalance(myAddr))
	assert.Equal(t, uint64(3), sdb.GetNonce(myAddr))
}

func mockStateDB(t *testing.T) StateDB {
	db := NewTrieDatabase(storage.NewInMemoryKVStore())
	rootHash := common.Hash{}
	sdb, err := NewStateDB(db, rootHash)
	assert.Nil(t, err)
	return sdb
}
