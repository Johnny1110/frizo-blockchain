package trie

import (
	"bytes"
	"errors"
	"frizo-blockchain/common"
	"github.com/ethereum/go-ethereum/log"
)

// MPTProof represents a Merkle proof in MPT
type MPTProof struct {
	Key      []byte      // key -> to path
	Value    []byte      // value data
	Proof    [][]byte    // encode node in mpt path
	RootHash common.Hash // mpt root hash
}

// ProofNode represents a decoded node in the proof path
type MPTProofNode struct {
	NodeType MPTNodeType
	Path     []byte
	Value    []byte
	Children []common.Hash // for branch node, to store all child hash
}

// GenerateProof generate MPTProof by key
func (mpt *ModifiedMerklePatriciaTree) GenerateProof(key []byte) (*MPTProof, error) {
	if len(key) == 0 {
		return nil, errors.New("key is empty")
	}

	nibbles := KeyToHex(key)
	proof := &MPTProof{
		Key:      key,
		RootHash: mpt.GetRoot(),
		Proof:    make([][]byte, 0),
	}

	// collect proof nodes ([][]byte) alone the path
	value, err := mpt.collectProof(mpt.root, nibbles, &proof.Proof)
	if err != nil {
		log.Error("Failed to collect proof", "err", err)
		return nil, errors.New("Failed to collect proof")
	}

	proof.Value = value
	return proof, nil
}

// collectProof recursive call collect mpt proof and store into proof([][]byte) -> Return Leaf or Branch value data
func (mpt *ModifiedMerklePatriciaTree) collectProof(node *MPTNode, path []byte, proof *[][]byte) ([]byte, error) {
	if node == nil {
		return nil, nil
	}

	// Encode and add current node to proof
	encodeVal := mpt.encodeNode(node)
	*proof = append(*proof, encodeVal)

	switch node.NodeType {
	case LEAF:
		if bytes.Equal(path, node.Path) {
			return node.Value, nil
		}
		return nil, nil

	case EXTENSION:
		nodePathLen := len(node.Path)
		if len(path) >= nodePathLen && bytes.Equal(path[:nodePathLen], node.Path) {
			return mpt.collectProof(node.getExtensionChild(), path[nodePathLen:], proof)
		}
		return nil, nil

	case BRANCH:
		if len(path) == 0 {
			// this branch value is the target
			return node.Value, nil
		}

		targetChildIdx := path[0]
		if targetChildIdx < 0 || targetChildIdx >= 16 {
			log.Error("invalid child index", "index", targetChildIdx)
			return nil, errors.New("collect proof failed, invalid target child index")
		}
		targetChild := node.Children[targetChildIdx]
		if targetChild != nil {
			remainPath := path[1:]
			return mpt.collectProof(targetChild, remainPath, proof)
		} else {
			return nil, nil
		}
	default:
		return nil, errors.New("invalid node type")
	}
}
