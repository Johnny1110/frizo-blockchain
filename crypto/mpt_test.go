package crypto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_HexToCompact_oddLeaf(t *testing.T) {
	mockEvenPath := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	// head: odd leaf: 0011 0000 | 0x01 = 0011 0001 = 0x31
	// even leaf should be [0x31, 0x23, 0x45, 0x67]
	res := HexToCompact(mockEvenPath, true)
	assert.Equal(t, []byte{0x31, 0x23, 0x45, 0x67}, res)
}

func Test_HexToCompact_evenLeaf(t *testing.T) {
	mockEvenPath := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	// head: even leaf: 0010 0000 = 0x20
	// even leaf should be [0x20, 0x12, 0x34, 0x56]
	res := HexToCompact(mockEvenPath, true)
	assert.Equal(t, []byte{0x20, 0x12, 0x34, 0x56}, res)
}

func Test_HexToCompact_oddExt(t *testing.T) {
	mockEvenPath := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	// head: odd ext: 0001 0000 | 0x01 = 0001 0001 = 0x11
	// odd ext should be [0x11, 0x23, 0x45, 0x67]
	res := HexToCompact(mockEvenPath, false)
	assert.Equal(t, []byte{0x11, 0x23, 0x45, 0x67}, res)
}

func Test_HexToCompact_evenExt(t *testing.T) {
	mockEvenPath := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	// head: even dxt: 0000 0000 = 0x00
	// even ext should be [0x00, 0x12, 0x34, 0x56]
	res := HexToCompact(mockEvenPath, false)
	assert.Equal(t, []byte{0x00, 0x12, 0x34, 0x56}, res)
}

func Test_CompactToHex_oddLeaf(t *testing.T) {
	mockEvenPath := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	// head: odd leaf: 0011 0000 | 0x01 = 0011 0001 = 0x31
	// even leaf should be [0x31, 0x23, 0x45, 0x67]
	comp := HexToCompact(mockEvenPath, true)
	hex, isLeaf := CompactToHex(comp)
	assert.True(t, isLeaf)
	assert.Equal(t, mockEvenPath, hex)
}

func Test_CompactToHex_evenLeaf(t *testing.T) {
	mockEvenPath := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	comp := HexToCompact(mockEvenPath, true)
	hex, isLeaf := CompactToHex(comp)
	assert.True(t, isLeaf)
	assert.Equal(t, mockEvenPath, hex)
}

func Test_CompactToHex_oddExt(t *testing.T) {
	mockEvenPath := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	comp := HexToCompact(mockEvenPath, false)
	hex, isLeaf := CompactToHex(comp)
	assert.False(t, isLeaf)
	assert.Equal(t, mockEvenPath, hex)
}

func Test_CompactToHex_evenExt(t *testing.T) {
	mockEvenPath := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	comp := HexToCompact(mockEvenPath, false)
	hex, isLeaf := CompactToHex(comp)
	assert.False(t, isLeaf)
	assert.Equal(t, mockEvenPath, hex)
}

func Test_KeyToHex(t *testing.T) {
	key := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	hex := KeyToHex(key)
	fmt.Println(hex)
	assert.Equal(t, []byte{0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04, 0x00, 0x05, 0x00, 0x06, 0x00, 0x07}, hex)
}

func Test_Slice(t *testing.T) {
	pathA := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	pathB := []byte{0x01, 0x02, 0x03, 0x04, 0x08, 0x09, 0x0A}
	commonLen := commonPrefixLen(pathA, pathB)
	fmt.Println(commonLen)

	remainingPath := pathA[commonLen+1:] // ex: a-b-c-d-e => d-e
	fmt.Println(remainingPath)
}

func Test_MPT_Put_2DiffPath(t *testing.T) {
	mpt := NewMPT()

	key_1 := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	key_2 := []byte{0x11, 0x12, 0x13, 0x14, 0x15}
	err := mpt.Put(key_1, key_1)
	assert.Nil(t, err)
	fmt.Println(mpt.root)

	err = mpt.Put(key_2, key_2)
	assert.Nil(t, err)

	fmt.Println("==============================================")
	fmt.Println("root: ", mpt.root)

	fmt.Println("first leaf: ", mpt.root.Children[0])
	fmt.Println("sec leaf: ", mpt.root.Children[1])
}

func Test_MPT_Put_bothPathHaveRemaining(t *testing.T) {
	mpt := NewMPT()

	key_1 := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	key_2 := []byte{0x01, 0x02, 0x13, 0x14, 0x15}
	err := mpt.Put(key_1, key_1)
	assert.Nil(t, err)
	fmt.Println(mpt.root)

	err = mpt.Put(key_2, key_2)
	assert.Nil(t, err)

	fmt.Println("==============================================")
	fmt.Println("root: ", mpt.root)
	fmt.Println("child-branch: ", mpt.root.Children[0])

	fmt.Println("first leaf: ", mpt.root.Children[0].Children[0])
	fmt.Println("sec leaf: ", mpt.root.Children[0].Children[1])
}

func Test_MPT_Put_OriginLeafHasRemaining(t *testing.T) {
	mpt := NewMPT()

	key_1 := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	key_2 := []byte{0x01, 0x02, 0x03}
	err := mpt.Put(key_1, key_1)
	assert.Nil(t, err)
	fmt.Println(mpt.root)

	err = mpt.Put(key_2, key_2)
	assert.Nil(t, err)

	fmt.Println("==============================================")
	fmt.Println("root: ", mpt.root)
	fmt.Println("child-branch: ", mpt.root.Children[0])

	fmt.Println("first leaf: ", mpt.root.Children[0].Children[0])
}

func Test_MPT_Put_NewPathHasRemaining(t *testing.T) {
	mpt := NewMPT()

	key_1 := []byte{0x01, 0x02, 0x03}
	key_2 := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	err := mpt.Put(key_1, key_1)
	assert.Nil(t, err)
	fmt.Println(mpt.root)

	err = mpt.Put(key_2, key_2)
	assert.Nil(t, err)

	fmt.Println("==============================================")
	fmt.Println("root: ", mpt.root)
	fmt.Println("child-branch: ", mpt.root.Children[0])
	fmt.Println("first: ", mpt.root.Children[0].Children[0])
}

func mockAExtNode(t *testing.T) *ModifiedMerklePatriciaTree {
	mpt := NewMPT()

	key_1 := []byte{0x01, 0x02, 0x03}
	key_2 := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	err := mpt.Put(key_1, key_1)
	assert.Nil(t, err)
	err = mpt.Put(key_2, key_2)
	assert.Nil(t, err)
	return mpt
}

func Test_MPT_Put_insertIntoExt_FullMatch(t *testing.T) {
	mpt := mockAExtNode(t)
	fmt.Println("==============================================")
	fmt.Println("root: ", mpt.root)
	fmt.Println("child-branch: ", mpt.root.Children[0])
	fmt.Println("first: ", mpt.root.Children[0].Children[0])
	fmt.Println("==============================================")

	key := []byte{0x01, 0x02, 0x03}
	err := mpt.Put(key, []byte{0xAA, 0xBB, 0xCC})
	assert.Nil(t, err)

	fmt.Println("root:", mpt.root)
	fmt.Println("root.child:", mpt.root.Children[0])
}

func Test_MPT_Put_insertIntoExt_S_2_1(t *testing.T) {
	mpt := mockAExtNode(t)
	fmt.Println("==============================================")
	fmt.Println("root: ", mpt.root)
	fmt.Println("child-branch: ", mpt.root.Children[0])
	fmt.Println("first: ", mpt.root.Children[0].Children[0])
	fmt.Println("==============================================")

	key := []byte{0x01, 0x02}
	err := mpt.Put(key, []byte{0xAA, 0xBB, 0xCC})
	assert.Nil(t, err)

	fmt.Println("root:", mpt.root)
	fmt.Println("root.child:", mpt.root.Children[0])
	fmt.Println("origin.ext:", mpt.root.Children[0].Children[0])
	fmt.Println("origin.ext.ChildBranch:", mpt.root.Children[0].Children[0].Children[0])
	fmt.Println("?:", mpt.root.Children[0].Children[0].Children[0].Children[0])
}

func Test_MPT_Put_insertIntoExt_S_2_2(t *testing.T) {
	mpt := mockAExtNode(t)
	fmt.Println("==============================================")
	fmt.Println("root: ", mpt.root)
	fmt.Println("child-branch: ", mpt.root.Children[0])
	fmt.Println("first: ", mpt.root.Children[0].Children[0])
	fmt.Println("==============================================")

	key := []byte{0x01, 0x02, 0xAB, 0xCD}
	err := mpt.Put(key, []byte{0xAA, 0xBB, 0xCC})
	assert.Nil(t, err)

	fmt.Println("root:", mpt.root)
	fmt.Println("root.childBranch:", mpt.root.Children[0])
	fmt.Println("index-0:", mpt.root.Children[0].Children[0])
	// idx 12 should be a leaf
	fmt.Println("index-10:", mpt.root.Children[0].Children[10])

	fmt.Println("check index-0: ----------------------")
	fmt.Println("index-0(ext)-Child Branch:", mpt.root.Children[0].Children[0].Children[0])
	fmt.Println("index-0(ext)-Child Branch.idx-0:", mpt.root.Children[0].Children[0].Children[0].Children[0])
}

func Test_MPT_Debug_Mode(t *testing.T) {
	mpt := NewMPT()

	mpt.Put([]byte("test"), []byte("value1"))
	fmt.Println("PUT: ", []byte("test"))

	mpt.Put([]byte("team"), []byte("value2"))
	fmt.Println("PUT: ", []byte("team"))

	mpt.Put([]byte("testing"), []byte("value3"))
	fmt.Println("PUT: ", []byte("testing"))

	mpt.Put([]byte("apple"), []byte("fruit"))
	fmt.Println("PUT: ", []byte("apple"))

	mpt.Put([]byte("app"), []byte("application"))
	fmt.Println("PUT: ", []byte("app"))

	// 打印樹結構
	mpt.PrintTree()

	// 打印統計信息
	mpt.PrintStats()

	// 打印所有鍵值對
	mpt.PrintAllKeys()
}

func Test_MPT_Debug_Mode_2(t *testing.T) {
	mpt := NewMPT()

	mpt.Put([]byte{0x01, 0x02, 0x11}, []byte("value1"))
	mpt.Put([]byte{0x01, 0x02, 0x12}, []byte("value2"))
	mpt.Put([]byte{0x01, 0x02, 0x13}, []byte("value3"))
	mpt.Put([]byte{0x01, 0x02, 0x14}, []byte("value4"))

	mpt.Put([]byte{0x01, 0x02}, []byte("value5"))

	mpt.Put([]byte{0x01, 0x02, 0x11, 0xDD}, []byte("valueK"))

	mpt.Put([]byte{0x01, 0x02, 0x15}, []byte("valueB"))
	mpt.Put([]byte{0x01, 0x02, 0x11, 0xAA}, []byte("valueAA"))
	mpt.Put([]byte{0x01, 0x02, 0x11, 0xAB}, []byte("valueAB"))
	mpt.Put([]byte{0x01, 0x02, 0x11, 0xAC}, []byte("valueAC"))

	mpt.Put([]byte{0x01}, []byte("Value 01"))

	// 打印樹結構
	mpt.PrintTree()
	mpt.PrintAllKeys()

	fmt.Println(mpt.root.Children[0].Children[0].Children[0].Children[1])
}

// 工具函數：hex string -> []byte
func mustHex(s string) []byte {
	bz, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bz
}

// 工具函數：打平 MPT 結構方便比較（只看 type/path/value，不看 hash/db）
func printMPT(node *MPTNode, indent string) {
	if node == nil {
		fmt.Println(indent + "nil")
		return
	}
	fmt.Printf("%s[%s] Path: %x, Value: %x\n", indent, node.NodeType.String(), node.Path, node.Value)
	switch node.NodeType {
	case EXTENSION:
		printMPT(node.Children[0], indent+"  ")
	case BRANCH:
		for i, child := range node.Children {
			if child != nil {
				fmt.Printf("%s  Child[%x]:\n", indent, i)
				printMPT(child, indent+"    ")
			}
		}
	}
}

// 測試：單一 key 插入（建立 Leaf 節點）
func TestPut_SingleKey(t *testing.T) {
	tree := NewMPT()

	err := tree.Put([]byte("dog"), []byte("woof"))
	if err != nil {
		t.Fatal(err)
	}

	root := tree.root
	if root.NodeType != LEAF {
		t.Errorf("expected LEAF node, got %s", root.NodeType)
	}
	if !bytes.Equal(root.Value, []byte("woof")) {
		t.Errorf("expected value 'woof', got %s", root.Value)
	}
}

// 測試：插入兩筆相同 prefix 的 key，觸發 Leaf → Extension + Branch 拆解
func TestPut_TwoKeysWithSharedPrefix(t *testing.T) {
	tree := NewMPT()

	_ = tree.Put([]byte("dog"), []byte("woof"))
	_ = tree.Put([]byte("door"), []byte("open"))

	root := tree.root
	printMPT(root, "")

	if root.NodeType != EXTENSION {
		t.Fatalf("expected EXTENSION node at root, got %s", root.NodeType)
	}

	branch := root.getExtensionChild()
	if branch == nil || branch.NodeType != BRANCH {
		t.Fatalf("expected BRANCH as child of EXTENSION")
	}

	// check leaf under different branch slots
	foundDog, foundDoor := false, false
	for i, child := range branch.Children {
		if child == nil {
			continue
		}
		if child.NodeType == LEAF {
			if bytes.Equal(child.Value, []byte("woof")) {
				foundDog = true
			}
			if bytes.Equal(child.Value, []byte("open")) {
				foundDoor = true
			}
		}
		if i > 15 {
			t.Errorf("invalid child index: %d", i)
		}
	}

	if !foundDog || !foundDoor {
		t.Errorf("expected both 'dog' and 'door' to be in trie")
	}
}

// 測試：key 為 prefix（如 "do" -> "dog"），覆蓋 Branch.Value
func TestPut_KeyIsPrefixOfExisting(t *testing.T) {
	tree := NewMPT()

	_ = tree.Put([]byte("dog"), []byte("woof"))
	_ = tree.Put([]byte("do"), []byte("helper"))

	root := tree.root
	printMPT(root, "")

	if root.NodeType != EXTENSION {
		t.Errorf("expected EXTENSION node at root, got %s", root.NodeType)
	}
	branch := root.getExtensionChild()
	if branch.Value == nil || !bytes.Equal(branch.Value, []byte("helper")) {
		t.Errorf("branch node should have value 'helper', got %x", branch.Value)
	}
}

// 測試：完全覆寫相同 key
func TestPut_OverwriteSameKey(t *testing.T) {
	tree := NewMPT()

	_ = tree.Put([]byte("dog"), []byte("woof"))
	_ = tree.Put([]byte("dog"), []byte("bark"))

	root := tree.root
	if root.NodeType != LEAF {
		t.Fatalf("expected LEAF node, got %s", root.NodeType)
	}
	if !bytes.Equal(root.Value, []byte("bark")) {
		t.Errorf("value not updated correctly, got %s", root.Value)
	}
}

func Test_MPT_Get_Basic(t *testing.T) {
	mpt := NewMPT()

	// Test 1: Get from empty tree
	value, err := mpt.Get([]byte("key"))
	assert.Nil(t, err)
	assert.Nil(t, value)

	// Test 2: Put and Get single key
	err = mpt.Put([]byte("hello"), []byte("world"))
	assert.Nil(t, err)

	value, err = mpt.Get([]byte("hello"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("world"), value)

	// Test 3: Get non-existent key
	value, err = mpt.Get([]byte("goodbye"))
	assert.Nil(t, err)
	assert.Nil(t, value)
}

func Test_MPT_Get_Multiple_Keys(t *testing.T) {
	mpt := NewMPT()

	testCases := []struct {
		key   []byte
		value []byte
	}{
		{[]byte("dog"), []byte("woof")},
		{[]byte("cat"), []byte("meow")},
		{[]byte("bird"), []byte("tweet")},
		{[]byte("fish"), []byte("blub")},
	}

	// Put all test cases
	for _, tc := range testCases {
		err := mpt.Put(tc.key, tc.value)
		assert.Nil(t, err)
	}

	// Get all test cases
	for _, tc := range testCases {
		value, err := mpt.Get(tc.key)
		assert.Nil(t, err)
		assert.Equal(t, tc.value, value)
	}
}

func Test_MPT_Get_With_Shared_Prefix(t *testing.T) {
	mpt := NewMPT()

	// Put keys with shared prefixes
	_ = mpt.Put([]byte("test"), []byte("value1"))
	_ = mpt.Put([]byte("testing"), []byte("value2"))
	_ = mpt.Put([]byte("tester"), []byte("value3"))
	_ = mpt.Put([]byte("team"), []byte("value4"))

	// Get each key
	value, _ := mpt.Get([]byte("test"))
	assert.Equal(t, []byte("value1"), value)

	value, _ = mpt.Get([]byte("testing"))
	assert.Equal(t, []byte("value2"), value)

	value, _ = mpt.Get([]byte("tester"))
	assert.Equal(t, []byte("value3"), value)

	value, _ = mpt.Get([]byte("team"))
	assert.Equal(t, []byte("value4"), value)

	// Test non-existent keys with similar prefixes
	value, _ = mpt.Get([]byte("te"))
	assert.Nil(t, value)

	value, _ = mpt.Get([]byte("tests"))
	assert.Nil(t, value)
}

func Test_MPT_Get_After_Update(t *testing.T) {
	mpt := NewMPT()

	// Initial put
	_ = mpt.Put([]byte("key1"), []byte("initial"))

	value, _ := mpt.Get([]byte("key1"))
	assert.Equal(t, []byte("initial"), value)

	// Update the value
	_ = mpt.Put([]byte("key1"), []byte("updated"))

	value, _ = mpt.Get([]byte("key1"))
	assert.Equal(t, []byte("updated"), value)
}

func Test_MPT_Get_Complex_Structure(t *testing.T) {
	mpt := NewMPT()

	// Create a complex tree structure
	keys := [][]byte{
		{0x01, 0x02, 0x03},
		{0x01, 0x02, 0x03, 0x04},
		{0x01, 0x02, 0x03, 0x04, 0x05},
		{0x01, 0x02, 0x13},
		{0x01, 0x02, 0x13, 0x14},
		{0x11, 0x12, 0x13},
	}

	for i, key := range keys {
		value := []byte(fmt.Sprintf("value%d", i))
		_ = mpt.Put(key, value)
	}

	// Verify all keys can be retrieved
	for i, key := range keys {
		expectedValue := []byte(fmt.Sprintf("value%d", i))
		value, err := mpt.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, expectedValue, value)
	}

	// Test non-existent keys
	nonExistentKeys := [][]byte{
		{0x01},
		{0x01, 0x02},
		{0x01, 0x02, 0x03, 0x05},
		{0x01, 0x02, 0x14},
		{0x12},
	}

	for _, key := range nonExistentKeys {
		value, err := mpt.Get(key)
		assert.Nil(t, err)
		assert.Nil(t, value)
	}
}

func Test_MPT_Contains(t *testing.T) {
	mpt := NewMPT()

	_ = mpt.Put([]byte("exists"), []byte("yes"))

	assert.True(t, mpt.Contains([]byte("exists")))
	assert.False(t, mpt.Contains([]byte("not-exists")))
}

// Benchmark tests
func BenchmarkMPT_Get(b *testing.B) {
	mpt := NewMPT()

	// Prepare test data
	numKeys := 10000
	keys := make([][]byte, numKeys)
	for i := 0; i < numKeys; i++ {
		key := []byte(fmt.Sprintf("key%06d", i))
		value := []byte(fmt.Sprintf("value%06d", i))
		keys[i] = key
		_ = mpt.Put(key, value)
	}

	b.ResetTimer()

	// Benchmark Get operations
	for i := 0; i < b.N; i++ {
		keyIndex := i % numKeys
		_, _ = mpt.Get(keys[keyIndex])
	}
}

func Test_MPT_Delete_Basic(t *testing.T) {
	mpt := NewMPT()

	// Test 1: Delete from empty tree
	err := mpt.Delete([]byte("nonexistent"))
	assert.Nil(t, err) // Should not error on non-existent key

	// Test 2: Put and delete single key
	_ = mpt.Put([]byte("hello"), []byte("world"))

	value, _ := mpt.Get([]byte("hello"))
	assert.Equal(t, []byte("world"), value)

	err = mpt.Delete([]byte("hello"))
	assert.Nil(t, err)

	value, _ = mpt.Get([]byte("hello"))
	assert.Nil(t, value) // Should be deleted

	// Tree should be empty
	assert.Nil(t, mpt.root)
}

func Test_MPT_Delete_Branch_To_Extension(t *testing.T) {
	mpt := NewMPT()

	// Create a branch with 2 children
	_ = mpt.Put([]byte("aa"), []byte("value1"))
	_ = mpt.Put([]byte("ab"), []byte("value2"))

	mpt.PrintTree()

	// Delete one child, branch should become extension
	_ = mpt.Delete([]byte("aa"))

	mpt.PrintTree()

	// Check remaining key still exists
	value, _ := mpt.Get([]byte("ab"))
	assert.Equal(t, []byte("value2"), value)
	//
	//// Root should now be a leaf (after optimization)
	//assert.Equal(t, LEAF, mpt.root.NodeType)
}

func Test_MPT_Delete_Branch_With_Value(t *testing.T) {
	mpt := NewMPT()

	// Create structure where branch has a value
	_ = mpt.Put([]byte("a"), []byte("value1"))
	_ = mpt.Put([]byte("aa"), []byte("value2"))
	_ = mpt.Put([]byte("ab"), []byte("value3"))

	// Delete one child
	_ = mpt.Delete([]byte("aa"))

	// Branch should still exist because it has a value
	value, _ := mpt.Get([]byte("a"))
	assert.Equal(t, []byte("value1"), value)

	value, _ = mpt.Get([]byte("ab"))
	assert.Equal(t, []byte("value3"), value)
}

func Test_MPT_Delete_Complex_Scenario(t *testing.T) {
	mpt := NewMPT()

	// Build a complex tree
	keys := []string{
		"do", "dog", "doge", "horse", "house",
	}

	for i, key := range keys {
		_ = mpt.Put([]byte(key), []byte(fmt.Sprintf("value%d", i)))
	}

	mpt.PrintTree()

	// Delete "doge"
	_ = mpt.Delete([]byte("doge"))

	// Verify "doge" is deleted but others remain
	value, _ := mpt.Get([]byte("doge"))
	assert.Nil(t, value)

	for i, key := range keys {
		if key != "doge" {
			value, _ := mpt.Get([]byte(key))
			assert.Equal(t, []byte(fmt.Sprintf("value%d", i)), value)
		}
	}

	fmt.Println("After deleting 'doge':")
	mpt.PrintTree()
}

func Test_MPT_Delete_All_Keys(t *testing.T) {
	mpt := NewMPT()

	keys := [][]byte{
		[]byte("apple"),
		[]byte("app"),
		[]byte("application"),
		[]byte("apply"),
		[]byte("banana"),
		[]byte("band"),
		[]byte("bandana"),
	}

	// Put all keys
	for i, key := range keys {
		_ = mpt.Put(key, []byte(fmt.Sprintf("value%d", i)))
	}

	originalSize := mpt.Size()
	assert.Equal(t, len(keys), originalSize)

	// Delete all keys one by one
	for _, key := range keys {
		err := mpt.Delete(key)
		assert.Nil(t, err)

		// Verify key is deleted
		value, _ := mpt.Get(key)
		assert.Nil(t, value)
	}

	// Tree should be empty
	assert.Nil(t, mpt.root)
	assert.Equal(t, 0, mpt.Size())
}

func Test_MPT_Delete_Extension_Merge(t *testing.T) {
	mpt := NewMPT()

	// Create structure that will result in extension merging
	_ = mpt.Put([]byte{0x01, 0x02, 0x03, 0x04}, []byte("A"))
	_ = mpt.Put([]byte{0x01, 0x02, 0x03, 0x05}, []byte("B"))
	_ = mpt.Put([]byte{0x01, 0x02, 0x13}, []byte("C"))

	fmt.Println("Before deletion:")
	mpt.PrintTree()

	// Delete C, should trigger extension merging
	_ = mpt.Delete([]byte{0x01, 0x02, 0x13})

	fmt.Println("After deletion:")
	mpt.PrintTree()

	// Verify remaining keys
	value, _ := mpt.Get([]byte{0x01, 0x02, 0x03, 0x04})
	assert.Equal(t, []byte("A"), value)

	value, _ = mpt.Get([]byte{0x01, 0x02, 0x03, 0x05})
	assert.Equal(t, []byte("B"), value)
}

func Test_MPT_Delete_Stress(t *testing.T) {
	mpt := NewMPT()

	// Add many keys
	numKeys := 1000
	keys := make([][]byte, numKeys)

	for i := 0; i < numKeys; i++ {
		key := []byte(fmt.Sprintf("key%06d", i))
		value := []byte(fmt.Sprintf("value%06d", i))
		keys[i] = key
		_ = mpt.Put(key, value)
	}

	// Delete half of them
	for i := 0; i < numKeys; i += 2 {
		err := mpt.Delete(keys[i])
		assert.Nil(t, err)
	}

	// Verify correct keys remain
	for i := 0; i < numKeys; i++ {
		value, _ := mpt.Get(keys[i])
		if i%2 == 0 {
			assert.Nil(t, value) // Should be deleted
		} else {
			assert.Equal(t, []byte(fmt.Sprintf("value%06d", i)), value)
		}
	}
}

func Test_MPT_Delete_And_Reinsert(t *testing.T) {
	mpt := NewMPT()

	key := []byte("test")
	value1 := []byte("value1")
	value2 := []byte("value2")

	// Insert
	_ = mpt.Put(key, value1)
	v, _ := mpt.Get(key)
	assert.Equal(t, value1, v)

	// Delete
	_ = mpt.Delete(key)
	v, _ = mpt.Get(key)
	assert.Nil(t, v)

	// Reinsert with different value
	_ = mpt.Put(key, value2)
	v, _ = mpt.Get(key)
	assert.Equal(t, value2, v)
}

func Test_MPT_Clear(t *testing.T) {
	mpt := NewMPT()

	// Add some keys
	_ = mpt.Put([]byte("key1"), []byte("value1"))
	_ = mpt.Put([]byte("key2"), []byte("value2"))
	_ = mpt.Put([]byte("key3"), []byte("value3"))

	assert.Equal(t, 3, mpt.Size())

	// Clear all
	mpt.Clear()

	assert.Nil(t, mpt.root)
	assert.Equal(t, 0, mpt.Size())

	// Verify all keys are gone
	value, _ := mpt.Get([]byte("key1"))
	assert.Nil(t, value)
}

func TestHashCalculation(t *testing.T) {
	mpt := NewMPT()

	mpt.Put([]byte("cat"), []byte("animal"))
	mpt.Put([]byte("car"), []byte("vehicle"))

	// 第一次計算根哈希
	root1 := mpt.GetRoot()

	// 第二次計算（應該使用緩存）
	root2 := mpt.GetRoot()

	// 應該相等
	assert.Equal(t, root1, root2)

	fmt.Println("first hash root:", root1)

	// 修改數據
	mpt.Put([]byte("cat"), []byte("feline"))

	// 根哈希應該改變
	root3 := mpt.GetRoot()
	assert.NotEqual(t, root1, root3)

	fmt.Println("last hash root:", root3)
}
