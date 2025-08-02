package trie

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_MPT_Proof(t *testing.T) {
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

	mpt.PrintTree()
	mpt.PrintAllKeys()

	proof, err := mpt.GetProof(keys[0])
	assert.Nil(t, err)

	fmt.Println("proof key:", string(proof.Key))
	fmt.Println("proof value:", string(proof.Value))
	for idx, val := range proof.Proof {
		fmt.Println("proof route idx:", idx, "value:", val)
	}

	fmt.Println("===========================================================================")

	fmt.Println("first ext hash: ", mpt.root.getExtensionChild().Hash)
	fmt.Println("second ext hash: ", mpt.root.getExtensionChild().Children[1].Hash)
}
