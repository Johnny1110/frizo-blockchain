package trie

import (
	"bytes"
	"errors"
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/crypto"
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

// VerifyMPTProof verifies a Merkle proof
func VerifyMPTProof(rootHash common.Hash, key []byte, proof *MPTProof) (bool, error) {
	if proof == nil || len(proof.Proof) == 0 {
		return false, errors.New("invalid proof")
	}

	// Verify Root Hash
	if rootHash != crypto.Keccak256Hash(proof.Proof[0]) {
		return false, errors.New("invalid proof, root hash mismatch")
	}

	nibblesPath := KeyToHex(key)

	// recursive call verify func
	computedVal, err := verifyMPTProofPath(nibblesPath, proof.Proof, 0)
	if err != nil {
		return false, fmt.Errorf("failed to compute proof for key %x: %w", key, err)
	}

	// Check if the computed value matches the claimed value
	if !bytes.Equal(computedVal, proof.Value) {
		return false, errors.New("invalid proof, value mismatch")
	}

	return true, nil
}

// verifyMPTProofPath recursive call verify MPT ProofPath, return computedVal and errors
func verifyMPTProofPath(path []byte, proof [][]byte, proofIdx int) ([]byte, error) {

	if proofIdx >= len(proof) {
		return nil, fmt.Errorf("input proof index out of range: %d", proofIdx)
	}

	// decode proof bytes to proofNode
	proofNode, err := decodeProofNode(proof[proofIdx])
	if err != nil {
		return nil, fmt.Errorf("failed to decode proof node at index %d: %w", proofIdx, err)
	}

	switch proofNode.NodeType {
	case LEAF:
		if bytes.Equal(path, proofNode.Path) {
			return proofNode.Value, nil
		}
		return nil, errors.New("input proof path mismatch with leaf")

	case EXTENSION:
		extPathLen := len(proofNode.Path)

		if len(path) < extPathLen {
			return nil, errors.New("input proof path too short for extension")
		}

		if !bytes.Equal(path[:extPathLen], proofNode.Path) {
			return nil, errors.New("input proof path mismatch with extension")
		}
		return verifyMPTProofPath(path[extPathLen:], proof, proofIdx+1)

	case BRANCH:
		if len(path) == 0 {
			// current branch is the target node
			return proofNode.Value, nil
		}

		targetChildIdx := path[0]
		if targetChildIdx < 0 || targetChildIdx >= 16 {
			log.Error("invalid child index for branch", "index", targetChildIdx)
			return nil, errors.New("verify proof failed, invalid target child index")
		}
		targetChild := proofNode.Children[targetChildIdx]
		if !targetChild.IsZero() {
			remainPath := path[1:]
			return verifyMPTProofPath(remainPath, proof, proofIdx+1)
		} else {
			log.Error("invalid child index for branch", "index", targetChildIdx)
			return nil, errors.New("input proof path mismatch with branch")
		}
	default:
		return nil, errors.New("invalid node type")
	}
}

// decodeProofNode decode proof data to proof node
// NodeType:
// 1. LEAF: rlp(compactPath, nodeValue) -> size = 2
// 2. EXTENSION rlp(compactPath, nodeValue) -> size = 2
// 3. BRANCH rlp(16 * subNodes ref, nodeValue) -> size = 17
// RlpDecode proofData and recover: identify node by size 2: LEAF | EXTENSION, 17: BRANCH
// LEAF | EXTENSION can be identified by compactPath rule (defined HexToCompact() in mpt.go)
func decodeProofNode(proofData []byte) (*MPTProofNode, error) {
	if len(proofData) == 0 {
		return nil, errors.New("decode proof node failed, invalid proof data")
	}

	var decoded []interface{}
	err := common.RlpDecodeBytes(proofData, &decoded)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("rlp decode failed, err %v", err)
	}

	switch len(decoded) {
	case leafAndExtLength:
		compactPath, ok := decoded[0].([]byte)
		if !ok {
			return nil, errors.New("decode proof node failed, decode leaf or ext failed, compactPath error")
		}
		hexPath, isLeaf := CompactToHex(compactPath)

		if isLeaf {
			value, ok := decoded[1].([]byte)
			if !ok {
				return nil, errors.New("decode proof node failed, decode leaf or ext failed, value error")
			}

			return &MPTProofNode{
				NodeType: LEAF,
				Path:     hexPath,
				Value:    value,
			}, nil
		} else {
			// is EXTENSION
			return &MPTProofNode{
				NodeType: EXTENSION,
				Path:     hexPath,
				// Child ref not required.
			}, nil
		}
	case branchListLength:
		// is BRANCH
		node := &MPTProofNode{
			NodeType: BRANCH,
			Children: make([]common.Hash, 16),
		}

		// First 16 elements are child references
		for i := 0; i < 16; i++ {
			if childRef, ok := decoded[i].([]byte); ok && childRef != nil && len(childRef) > 0 {
				node.Children[i] = common.Keccak256Hash(childRef)
			}
		}

		// No.17 element(index-16) is node value
		if val, ok := decoded[16].([]byte); ok && val != nil && len(val) > 0 {
			node.Value = val
		}

		return node, nil
	default:
		return nil, errors.New("decode proof node failed, rlp decode data length invalid (must be 2 or 17)")
	}
}
