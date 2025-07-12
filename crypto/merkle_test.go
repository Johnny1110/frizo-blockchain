package crypto

import (
	"fmt"
	"frizo-blockchain/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalculateTreeHeight(t *testing.T) {
	r := calculateMerkleTreeHeight(12)
	assert.Equal(t, 4, r)
}

func TestBuildMerkleTree(t *testing.T) {
	data := [][]byte{}

	for i := 0; i < 4; i++ {
		data = append(data, common.RandomHash().Bytes())
	}

	tree := NewMerkleTree(data, nil)

	fmt.Println("root--------------------------")
	fmt.Println(tree.Root.Hash)
	fmt.Println("sec layer --------------------------")

	tree.PrintTree()
}

func TestGenerateProof(t *testing.T) {
	data := [][]byte{}
	for i := 0; i < 1000; i++ {
		data = append(data, common.RandomHash().Bytes())
	}
	tree := NewMerkleTree(data, nil)
	tree.PrintTree()

	// generate proof
	proof, err := tree.GenerateProof(0)
	assert.Nil(t, err)
	fmt.Println("--------------------- Proof ---------------------")
	fmt.Println("root hash:", proof.RootHash)
	fmt.Println("Direction:", proof.Directions)
	fmt.Println("Path Hash:", proof.Proof)

	assert.True(t, VerifyProof(proof, nil))
	// make a fake txn
	proof.LeafHash = common.RandomHash()
	assert.False(t, VerifyProof(proof, nil))
}
