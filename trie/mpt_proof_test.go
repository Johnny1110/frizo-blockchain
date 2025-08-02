package trie

import (
	"fmt"
	"testing"
)

func Test_Generate_MPTProof(t *testing.T) {
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
}
