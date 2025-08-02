# MPT Enhancement

<br>

---

<br>

## 需要實現的目標

1. 實現哈希計算和節點編碼
2. 完成 Merkle Proof 功能
3. 實現 RLP 編碼
4. 添加並發支持
5. 實現 LevelDB 存儲
6. 添加快照功能

<br>
<br>

## MPT 哈希計算實現指南

### Dirty 標記的作用

Dirty 標記是一個優化機制，用來追蹤哪些節點被修改過，需要重新計算哈希：

```go
// 節點被修改時的流程
修改節點 → 設置 Dirty = true → 計算哈希時檢查 Dirty → 只重算 Dirty 節點
```

<br>

MPT Proof:

```go
package mpt

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ChenYuTingJerry/frizo/common"
	"github.com/ChenYuTingJerry/frizo/crypto"
)

// MPTProof represents a Merkle proof in MPT
type MPTProof struct {
	Key      []byte          // Original key
	Value    []byte          // Value (if exists)
	Proof    [][]byte        // Encoded nodes in the proof path
	RootHash common.Hash     // Root hash of the trie
}

// ProofNode represents a decoded node in the proof path
type ProofNode struct {
	NodeType NodeType
	Path     []byte
	Value    []byte
	Children []common.Hash // For branch nodes, stores 16 children hashes
}

// GenerateProof generates a Merkle proof for a given key
func (t *MPT) GenerateProof(key []byte) (*MPTProof, error) {
	if len(key) == 0 {
		return nil, errors.New("key cannot be empty")
	}

	// Convert key to nibbles (hex encoding)
	nibbles := keyToNibbles(key)

	proof := &MPTProof{
		Key:      key,
		RootHash: t.Root.Hash(),
		Proof:    [][]byte{},
	}

	// Collect proof nodes along the path
	value, err := t.collectProof(t.Root, nibbles, &proof.Proof)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %w", err)
	}

	proof.Value = value
	return proof, nil
}

// collectProof traverses the trie and collects proof nodes
func (t *MPT) collectProof(node *MPTNode, path []byte, proof *[][]byte) ([]byte, error) {
	if node == nil {
		return nil, nil
	}

	// Encode and add current node to proof
	encoded := encodeNodeForProof(node)
	*proof = append(*proof, encoded)

	switch node.NodeType {
	case LeafNode:
		// Check if this is the target leaf
		if bytes.Equal(node.Key, path) {
			return node.Value, nil
		}
		return nil, nil

	case ExtensionNode:
		// Check if path matches the extension
		if len(path) >= len(node.Key) && bytes.Equal(node.Key, path[:len(node.Key)]) {
			// Continue with the remaining path
			return t.collectProof(node.Children[0], path[len(node.Key):], proof)
		}
		return nil, nil

	case BranchNode:
		if len(path) == 0 {
			// We've reached the end of the path
			return node.Value, nil
		}
		// Follow the path to the next child
		nextIndex := path[0]
		if nextIndex >= 16 {
			return nil, errors.New("invalid nibble in path")
		}
		return t.collectProof(node.Children[nextIndex], path[1:], proof)

	default:
		return nil, fmt.Errorf("unknown node type: %v", node.NodeType)
	}
}

// VerifyProof verifies a Merkle proof
func VerifyProof(rootHash common.Hash, key []byte, proof *MPTProof) (bool, error) {
	if proof == nil {
		return false, errors.New("proof cannot be nil")
	}

	if len(proof.Proof) == 0 {
		return false, errors.New("proof cannot be empty")
	}

	// Convert key to nibbles
	nibbles := keyToNibbles(key)

	// Verify the proof by traversing the path
	computedValue, err := verifyProofPath(nibbles, proof.Proof, 0, 0)
	if err != nil {
		return false, fmt.Errorf("proof verification failed: %w", err)
	}

	// Check if the computed value matches the claimed value
	if !bytes.Equal(computedValue, proof.Value) {
		return false, errors.New("value mismatch")
	}

	// Verify the root hash
	rootNode, err := decodeProofNode(proof.Proof[0])
	if err != nil {
		return false, fmt.Errorf("failed to decode root node: %w", err)
	}

	computedRootHash := hashNode(rootNode)
	return computedRootHash == rootHash, nil
}

// verifyProofPath recursively verifies the proof path
func verifyProofPath(path []byte, proofNodes [][]byte, nodeIndex int, pathIndex int) ([]byte, error) {
	if nodeIndex >= len(proofNodes) {
		return nil, errors.New("proof path incomplete")
	}

	node, err := decodeProofNode(proofNodes[nodeIndex])
	if err != nil {
		return nil, fmt.Errorf("failed to decode proof node: %w", err)
	}

	remainingPath := path[pathIndex:]

	switch node.NodeType {
	case LeafNode:
		// Check if this is the target leaf
		if bytes.Equal(node.Path, remainingPath) {
			return node.Value, nil
		}
		return nil, errors.New("leaf path mismatch")

	case ExtensionNode:
		// Check if path matches the extension
		if len(remainingPath) < len(node.Path) {
			return nil, errors.New("path too short for extension")
		}
		if !bytes.Equal(node.Path, remainingPath[:len(node.Path)]) {
			return nil, errors.New("extension path mismatch")
		}
		// Continue with the next node
		return verifyProofPath(path, proofNodes, nodeIndex+1, pathIndex+len(node.Path))

	case BranchNode:
		if len(remainingPath) == 0 {
			// We've reached the end of the path
			return node.Value, nil
		}
		// Follow the path to the next child
		nextIndex := remainingPath[0]
		if nextIndex >= 16 {
			return nil, errors.New("invalid nibble in path")
		}
		// Verify the child exists
		if node.Children[nextIndex] == (common.Hash{}) {
			return nil, errors.New("child not found in branch")
		}
		// Continue with the next node
		return verifyProofPath(path, proofNodes, nodeIndex+1, pathIndex+1)

	default:
		return nil, fmt.Errorf("unknown node type: %v", node.NodeType)
	}
}

// encodeNodeForProof encodes a node for inclusion in a proof
// This must match the encoding used in MPT for hash calculation
func encodeNodeForProof(node *MPTNode) []byte {
	switch node.NodeType {
	case LeafNode:
		// Use RLP encoding like in MPT
		compactPath := nibblesToCompact(node.Key, true)
		encoded, err := common.RlpEncodeToBytes([]interface{}{
			compactPath,
			node.Value,
		})
		if err != nil {
			panic(fmt.Errorf("failed to RLP encode leaf node: %v", err))
		}
		return encoded

	case ExtensionNode:
		// Use RLP encoding like in MPT
		compactPath := nibblesToCompact(node.Key, false)
		childRef := nodeRef(node.Children[0])
		encoded, err := common.RlpEncodeToBytes([]interface{}{
			compactPath,
			childRef,
		})
		if err != nil {
			panic(fmt.Errorf("failed to RLP encode extension node: %v", err))
		}
		return encoded

	case BranchNode:
		// Use RLP encoding like in MPT
		var refs []interface{}

		// Add 16 child references
		for i := 0; i < 16; i++ {
			if node.Children[i] != nil {
				refs = append(refs, nodeRef(node.Children[i]))
			} else {
				refs = append(refs, []byte{})
			}
		}

		// Add value
		refs = append(refs, node.Value)

		encoded, err := common.RlpEncodeToBytes(refs)
		if err != nil {
			panic(fmt.Errorf("failed to RLP encode branch node: %v", err))
		}
		return encoded

	default:
		return nil
	}
}

// nodeRef returns the reference bytes for a node (matching MPT's nodeRef)
func nodeRef(node *MPTNode) []byte {
	if node == nil {
		return []byte{}
	}

	// Encode the node
	encoded := encodeNodeForProof(node)

	// Ethereum standard: if encoded length < 32 bytes, return the encoded data directly
	if len(encoded) < 32 {
		return encoded
	}

	// Otherwise, return the hash
	hash := crypto.Keccak256Hash(encoded)
	return hash.Bytes()
}

// decodeProofNode decodes a node from proof data using RLP
func decodeProofNode(data []byte) (*ProofNode, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var decoded []interface{}
	if err := common.RlpDecodeBytes(data, &decoded); err != nil {
		return nil, fmt.Errorf("failed to RLP decode node: %v", err)
	}

	// Determine node type based on decoded length
	switch len(decoded) {
	case 2:
		// Leaf or Extension node
		pathData, ok := decoded[0].([]byte)
		if !ok {
			return nil, errors.New("invalid path data in leaf/extension node")
		}

		hexPath, isLeaf := compactToHex(pathData)
		
		if isLeaf {
			// Leaf node
			value, ok := decoded[1].([]byte)
			if !ok {
				return nil, errors.New("invalid value in leaf node")
			}
			
			return &ProofNode{
				NodeType: LeafNode,
				Path:     hexPath,
				Value:    value,
			}, nil
		} else {
			// Extension node
			// The second element is a node reference (hash or embedded node)
			return &ProofNode{
				NodeType: ExtensionNode,
				Path:     hexPath,
				// Children will be populated during verification
			}, nil
		}

	case 17:
		// Branch node
		node := &ProofNode{
			NodeType: BranchNode,
			Children: make([]common.Hash, 16),
		}

		// First 16 elements are child references
		for i := 0; i < 16; i++ {
			if childRef, ok := decoded[i].([]byte); ok && len(childRef) > 0 {
				if len(childRef) == 32 {
					copy(node.Children[i][:], childRef)
				}
			}
		}

		// 17th element is the value (if any)
		if value, ok := decoded[16].([]byte); ok && len(value) > 0 {
			node.Value = value
		}

		return node, nil

	default:
		return nil, fmt.Errorf("invalid node format: wrong element count %d", len(decoded))
	}
}

// compactToHex converts compact encoding to hex (nibbles)
// Returns the hex path and whether it's a leaf
func compactToHex(compact []byte) ([]byte, bool) {
	if len(compact) == 0 {
		return []byte{}, false
	}

	flag := compact[0]
	isLeaf := (flag & 0x20) != 0
	
	var nibbles []byte
	
	// Check if odd length
	if (flag & 0x10) != 0 {
		// Odd length, first nibble is in the flag
		nibbles = append(nibbles, flag&0x0f)
	}
	
	// Decode remaining bytes
	for i := 1; i < len(compact); i++ {
		nibbles = append(nibbles, compact[i]>>4, compact[i]&0x0f)
	}
	
	return nibbles, isLeaf
}

// hashNode computes the hash of a proof node
func hashNode(node *ProofNode) common.Hash {
	// Convert ProofNode back to MPTNode for encoding
	mptNode := proofNodeToMPTNode(node)
	
	// Encode using the same logic as MPT
	encoded := encodeMPTNode(mptNode)
	
	// If encoded data is less than 32 bytes, hash it directly
	// Otherwise, it should already be a hash reference
	if len(encoded) < 32 {
		return crypto.Keccak256Hash(encoded)
	}
	
	// For larger nodes, the encoding should be hashed
	return crypto.Keccak256Hash(encoded)
}

// proofNodeToMPTNode converts ProofNode to MPTNode for encoding
func proofNodeToMPTNode(node *ProofNode) *MPTNode {
	mptNode := &MPTNode{
		NodeType: node.NodeType,
		Value:    node.Value,
	}

	switch node.NodeType {
	case LeafNode, ExtensionNode:
		mptNode.Key = node.Path
		if node.NodeType == ExtensionNode && len(node.Children) > 0 {
			// For extension nodes, we need to handle the child reference
			mptNode.Children[0] = &MPTNode{} // Placeholder
		}
	case BranchNode:
		// Convert children hashes to node references
		for i := 0; i < 16; i++ {
			if node.Children[i] != (common.Hash{}) {
				mptNode.Children[i] = &MPTNode{} // Placeholder
			}
		}
	}

	return mptNode
}

// encodeMPTNode encodes an MPT node using RLP (matching the MPT implementation)
func encodeMPTNode(node *MPTNode) []byte {
	switch node.NodeType {
	case LeafNode:
		compactPath := nibblesToCompact(node.Key, true)
		encoded, err := common.RlpEncodeToBytes([]interface{}{
			compactPath,
			node.Value,
		})
		if err != nil {
			panic(err)
		}
		return encoded

	case ExtensionNode:
		compactPath := nibblesToCompact(node.Key, false)
		// For proof verification, we don't have the actual child node
		// but we have its hash stored in the proof
		childRef := []byte{} // This would be filled from the proof data
		encoded, err := common.RlpEncodeToBytes([]interface{}{
			compactPath,
			childRef,
		})
		if err != nil {
			panic(err)
		}
		return encoded

	case BranchNode:
		var refs []interface{}
		for i := 0; i < 16; i++ {
			if node.Children[i] != nil {
				refs = append(refs, []byte{}) // Placeholder for child references
			} else {
				refs = append(refs, []byte{})
			}
		}
		refs = append(refs, node.Value)
		
		encoded, err := common.RlpEncodeToBytes(refs)
		if err != nil {
			panic(err)
		}
		return encoded

	default:
		return nil
	}
}

// Helper functions - removed duplicates and simplified

// nibblesToCompact converts nibbles to compact encoding
func nibblesToCompact(nibbles []byte, isLeaf bool) []byte {
	l := len(nibbles)
	if l == 0 {
		if isLeaf {
			return []byte{0x20}
		}
		return []byte{0x00}
	}
	
	odd := l&1 != 0
	compact := make([]byte, l/2+1)
	
	// Set header byte
	if odd {
		compact[0] = 0x10 | nibbles[0]
		if isLeaf {
			compact[0] |= 0x20
		}
		for i := 1; i < l; i += 2 {
			compact[i/2+1] = nibbles[i]<<4 | nibbles[i+1]
		}
	} else {
		compact[0] = 0x00
		if isLeaf {
			compact[0] = 0x20
		}
		for i := 0; i < l; i += 2 {
			compact[i/2+1] = nibbles[i]<<4 | nibbles[i+1]
		}
	}
	
	return compact
}

// keyToNibbles converts a key to nibbles (hex encoding)
func keyToNibbles(key []byte) []byte {
	nibbles := make([]byte, len(key)*2)
	for i, b := range key {
		nibbles[i*2] = b >> 4
		nibbles[i*2+1] = b & 0x0f
	}
	return nibbles
}

// PrintProof prints proof details for debugging
func PrintProof(proof *MPTProof) {
	fmt.Println("\n========== MPT Proof Details ==========")
	fmt.Printf("Key: %x (%s)\n", proof.Key, string(proof.Key))
	fmt.Printf("Value: %x (%s)\n", proof.Value, string(proof.Value))
	fmt.Printf("Root Hash: %s\n", proof.RootHash.Hex())
	fmt.Printf("Proof Path Length: %d nodes\n", len(proof.Proof))
	
	fmt.Println("\nProof Path:")
	for i, nodeData := range proof.Proof {
		node, err := decodeProofNode(nodeData)
		if err != nil {
			fmt.Printf("  [%d] Error decoding node: %v\n", i, err)
			continue
		}
		
		fmt.Printf("  [%d] %s Node\n", i, nodeTypeString(node.NodeType))
		if len(node.Path) > 0 {
			fmt.Printf("      Path: %x\n", node.Path)
		}
		if node.Value != nil {
			fmt.Printf("      Value: %x (%s)\n", node.Value, string(node.Value))
		}
		if node.NodeType == BranchNode {
			fmt.Printf("      Children: ")
			for j, child := range node.Children {
				if child != (common.Hash{}) {
					fmt.Printf("%x ", j)
				}
			}
			fmt.Println()
		}
	}
	fmt.Println("=====================================")
}

func nodeTypeString(nt NodeType) string {
	switch nt {
	case LeafNode:
		return "Leaf"
	case ExtensionNode:
		return "Extension"
	case BranchNode:
		return "Branch"
	default:
		return "Unknown"
	}
}
```