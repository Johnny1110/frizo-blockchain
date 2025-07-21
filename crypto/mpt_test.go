package crypto

import (
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
