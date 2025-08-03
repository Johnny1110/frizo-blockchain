package trie

import (
	"fmt"
	"frizo-blockchain/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func test_Generate_MPTProof(t *testing.T) (*ModifiedMerklePatriciaTree, *MPTProof) {
	mpt := NewMPT()

	keys := [][]byte{
		[]byte("aaa"),
		[]byte("aab"),
		[]byte("aac"),
		[]byte("aaaa"),
		[]byte("aaab"),
		[]byte("aaac"),

		[]byte("aba"),
		[]byte("abb"),
		[]byte("abc"),
	}

	// Put all keys
	for i, key := range keys {
		_ = mpt.Put(key, []byte(fmt.Sprintf("value%d", i)))
	}

	mpt.PrintTree()

	println("========================================================================")
	println("<generate proof>")
	println("========================================================================")

	proof, err := mpt.GenerateProof([]byte("aaaa"))
	assert.Nil(t, err)

	for i, p := range proof.Proof {
		fmt.Printf("%d: %x\n", i, p)
	}

	fmt.Println("first-1(root) Hash: ", crypto.Keccak256Hash(proof.Proof[0]))
	assert.Equal(t, *mpt.root.Hash, crypto.Keccak256Hash(proof.Proof[0]))

	fmt.Println("last-6(leaf) Hash: ", crypto.Keccak256Hash(proof.Proof[6]))
	lastNode := mpt.root.getExtensionChild().Children[1].getExtensionChild().Children[1].Children[6].Children[1]
	assert.Equal(t, *lastNode.Hash, crypto.Keccak256Hash(proof.Proof[6]))

	return mpt, proof
}

func Test_Verify_MPTProof(t *testing.T) {
	mpt, proof := test_Generate_MPTProof(t)
	ok, err := VerifyMPTProof(mpt.GetRoot(), []byte("aaa"), proof)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VerifyMPTProof(mpt.GetRoot(), []byte("aab"), proof)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VerifyMPTProof(mpt.GetRoot(), []byte("aac"), proof)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VerifyMPTProof(mpt.GetRoot(), []byte("aba"), proof)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VerifyMPTProof(mpt.GetRoot(), []byte("abb"), proof)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VerifyMPTProof(mpt.GetRoot(), []byte("abc"), proof)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VerifyMPTProof(mpt.GetRoot(), []byte("aad"), proof)
	assert.NotNil(t, err)
	assert.False(t, ok)

	ok, err = VerifyMPTProof(mpt.GetRoot(), []byte("aae"), proof)
	assert.NotNil(t, err)
	assert.False(t, ok)

	ok, err = VerifyMPTProof(mpt.GetRoot(), []byte("aaad"), proof)
	assert.NotNil(t, err)
	assert.False(t, ok)

	ok, err = VerifyMPTProof(mpt.GetRoot(), []byte("aaaaa"), proof)
	assert.NotNil(t, err)
	assert.False(t, ok)

	ok, err = VerifyMPTProof(mpt.GetRoot(), []byte("aaaa"), proof)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VerifyMPTProof(mpt.GetRoot(), []byte("aaab"), proof)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = VerifyMPTProof(mpt.GetRoot(), []byte("aaac"), proof)
	assert.Nil(t, err)
	assert.True(t, ok)
}
