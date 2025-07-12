package crypto

import (
	"fmt"
	"frizo-blockchain/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math"
	"strings"
)

type MerkleDirection bool

func (m MerkleDirection) String() string {
	switch m {
	case true:
		return "LEFT"
	case false:
		return "RIGHT"
	}
	return ""
}

const (
	LEFT  MerkleDirection = true
	RIGHT MerkleDirection = false
)

func defaultHashFunc(data []byte) common.Hash {
	//  Using Keccak-256 as Default to get a 32-bytes raw data
	rawHashBytes := crypto.Keccak256(data)
	return common.BytesToHash(rawHashBytes)
}

// MerkleProof Merkle Proof
type MerkleProof struct {
	LeafIndex  int               // leaf index
	LeafHash   common.Hash       // leaf hash
	Proof      []common.Hash     // proof path
	Directions []MerkleDirection // direction（true=left，false=right）
	RootHash   common.Hash       // root hash
}

// MerkleNode Merkle Tree Node
type MerkleNode struct {
	IsLeaf    bool
	Index     int
	Parent    *MerkleNode // for generate proof
	Left      *MerkleNode // only parent node has left
	Right     *MerkleNode // only parent node has right
	Hash      common.Hash
	Direction MerkleDirection // node direction（true=left，false=right）
	Data      []byte          // only leaf node has data !!!
}

// MerkleTree whole MerkleTree structure
type MerkleTree struct {
	Root       *MerkleNode
	Leaves     []*MerkleNode
	TreeHeight int // tree height
	LeafCount  int // all leaf count
	HashFunc   func([]byte) common.Hash
}

// buildTree arrange tree structure by rules. (recursive)
func (t *MerkleTree) buildTree(nodes []*MerkleNode) *MerkleNode {
	if len(nodes) == 0 {
		return nil
	}

	if len(nodes) == 1 {
		// return until recursive to last root.
		return nodes[0]
	}

	// if the nodes count is Odd, we need copy last to make it as Even.
	if len(nodes)%2 == 1 {
		lastNode := nodes[len(nodes)-1]
		copyNode := &MerkleNode{
			IsLeaf: lastNode.IsLeaf,
			Index:  lastNode.Index,
			Hash:   lastNode.Hash,
			Data:   lastNode.Data,
		}
		nodes = append(nodes, copyNode)
	}

	var newParentLevel []*MerkleNode

	for i := 0; i < len(nodes); i += 2 {
		left := nodes[i]
		right := nodes[i+1]

		// combined left & right hash
		leftRightHashCombined := append(left.Hash.Bytes(), right.Hash.Bytes()...)

		parentNode := &MerkleNode{
			IsLeaf: false,
			Hash:   t.HashFunc(leftRightHashCombined),
			Left:   left,
			Right:  right,
		}

		// pre-process data for generate proof
		left.Parent = parentNode
		right.Parent = parentNode
		left.Direction = true
		right.Direction = false

		newParentLevel = append(newParentLevel, parentNode)
	}

	// recursive build.
	return t.buildTree(newParentLevel)
}

func (t *MerkleTree) GetRootHash() common.Hash {
	if t.Root == nil {
		return common.Hash{}
	}
	return t.Root.Hash
}

// NewMerkleTree Create a new MerkleTree
func NewMerkleTree(data [][]byte, hashFunc func([]byte) common.Hash) *MerkleTree {
	if len(data) == 0 {
		return &MerkleTree{}
	}

	if hashFunc == nil {
		hashFunc = defaultHashFunc
	}

	mt := &MerkleTree{
		HashFunc:   hashFunc,
		LeafCount:  len(data),
		TreeHeight: calculateMerkleTreeHeight(len(data)),
	}

	// create leaf node
	leaves := make([]*MerkleNode, len(data))
	for i, datum := range data {
		leaves[i] = &MerkleNode{
			Hash:   hashFunc(datum),
			Data:   datum,
			IsLeaf: true,
			Index:  i,
		}
	}

	mt.Leaves = leaves
	mt.Root = mt.buildTree(mt.Leaves)
	return mt
}

func (t *MerkleTree) GetLeafData(index int) ([]byte, error) {
	if index < 0 || index >= t.LeafCount {
		return nil, common.ErrInvalidData("index out of range")
	}

	return t.Leaves[index].Data, nil
}

// calculateMerkleTreeHeight calculate tree height based on leaf count
// The height of a Merkle tree is determined by the number of leaf nodes (data items).
// If a Merkle tree has N leaf nodes, and N is a power of 2 (N = 2H), then the height of the tree is H.
// The height is the number of levels in the tree, !!!including the root and the leaves!!!
func calculateMerkleTreeHeight(leafCount int) int {
	// height =log2(n)
	if leafCount == 0 {
		return 0
	}
	return int(math.Ceil(math.Log2(float64(leafCount))))
}

// GenerateProof input leaf's index to get a MerkleProof
// MerkleProof include all hash of sibling from target leaf to root
// Validator can recalculate from target leaf with all sibling to get a root hash
// validateEquals(calculatedHash, rootHash)
func (t *MerkleTree) GenerateProof(leafIndex int) (*MerkleProof, error) {
	if leafIndex < 0 || leafIndex >= t.LeafCount {
		return nil, common.ErrFailedToGenerateMerkleProof("leaf index out of range")
	}

	targetNode := t.Leaves[leafIndex]
	proof := &MerkleProof{
		RootHash:   t.GetRootHash(),
		LeafIndex:  leafIndex,
		LeafHash:   targetNode.Hash,
		Proof:      []common.Hash{},
		Directions: []MerkleDirection{},
	}

	// iterate from target to root, collect sibling data
	for {
		parentNode := targetNode.Parent
		if parentNode == nil {
			// reach root node
			break
		}

		switch targetNode.Direction {
		case LEFT:
			// find right sibling
			proof.Proof = append(proof.Proof, parentNode.Right.Hash)
			proof.Directions = append(proof.Directions, RIGHT)
			break
		case RIGHT:
			proof.Proof = append(proof.Proof, parentNode.Left.Hash)
			proof.Directions = append(proof.Directions, LEFT)
		}

		targetNode = parentNode
	}

	return proof, nil
}

func (t *MerkleTree) ContainsData(data []byte) (bool, int) {
	hash := t.HashFunc(data)
	for i, leaf := range t.Leaves {
		if leaf.Hash == hash {
			return true, i
		}
	}
	return false, -1
}

// VerifyProof Merkle proof func
func VerifyProof(proof *MerkleProof, hashFunc func([]byte) common.Hash) bool {
	// validate direction len and proof len
	if len(proof.Proof) != len(proof.Directions) {
		return false
	}

	if hashFunc == nil {
		hashFunc = defaultHashFunc
	}

	targetHash := proof.LeafHash

	for i, siblingDirection := range proof.Directions {
		siblingHash := proof.Proof[i]
		switch siblingDirection {
		case LEFT:
			concatHashBytes := append(siblingHash.Bytes(), targetHash.Bytes()...)
			targetHash = hashFunc(concatHashBytes)
			break
		case RIGHT:
			concatHashBytes := append(targetHash.Bytes(), siblingHash.Bytes()...)
			targetHash = hashFunc(concatHashBytes)
			break
		}
	}

	return targetHash == proof.RootHash
}

// for debug ------------------------------------------------------------------------------------------

// PrintTree prints the entire tree structure
func (t *MerkleTree) PrintTree() {
	fmt.Println("\n========== Merkle Tree Structure ==========")
	fmt.Printf("Tree Height: %d\n", t.TreeHeight)
	fmt.Printf("Leaf Count: %d\n", t.LeafCount)
	fmt.Printf("Root Hash: %s\n", t.Root.Hash.Hex())
	fmt.Println("===========================================")

	if t.Root == nil {
		fmt.Println("Empty tree")
		return
	}

	fmt.Println("\nTree Visualization:")
	t.printNode(t.Root, "", true, 0)

	fmt.Println("\n========== Leaf Nodes Details ==========")
	for i, leaf := range t.Leaves {
		fmt.Printf("Leaf[%d]: Hash=%s, Data=%x\n",
			i,
			leaf.Hash.Hex()[:16]+"...",
			leaf.Data[:min(len(leaf.Data), 32)])
	}

	fmt.Println("\n========== Level-by-Level View ==========")
	t.printLevelByLevel()
}

// printNode recursively prints each node
func (t *MerkleTree) printNode(node *MerkleNode, prefix string, isLast bool, level int) {
	if node == nil {
		return
	}

	fmt.Print(prefix)
	if isLast {
		fmt.Print("└── ")
		prefix += "    "
	} else {
		fmt.Print("├── ")
		prefix += "│   "
	}

	nodeType := "Branch"
	if node.IsLeaf {
		nodeType = fmt.Sprintf("Leaf[%d]", node.Index)
	}

	fmt.Printf("[%s] %s (Level %d)\n", nodeType, node.Hash.Hex()[:16]+"...", level)

	if !node.IsLeaf {
		if node.Left != nil {
			t.printNode(node.Left, prefix, false, level+1)
		}
		if node.Right != nil {
			t.printNode(node.Right, prefix, true, level+1)
		}
	}
}

// printLevelByLevel prints the tree level by level
func (t *MerkleTree) printLevelByLevel() {
	if t.Root == nil {
		return
	}

	type nodeLevel struct {
		node  *MerkleNode
		level int
	}

	queue := []nodeLevel{{t.Root, 0}}
	currentLevel := -1
	levelNodes := []string{}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if item.level != currentLevel {
			if currentLevel >= 0 {
				fmt.Printf("Level %d: %s\n", currentLevel, strings.Join(levelNodes, " | "))
			}
			currentLevel = item.level
			levelNodes = []string{}
		}

		nodeInfo := item.node.Hash.Hex()[:8]
		if item.node.IsLeaf {
			nodeInfo = fmt.Sprintf("L%d:%s", item.node.Index, nodeInfo)
		}
		levelNodes = append(levelNodes, nodeInfo)

		if item.node.Left != nil {
			queue = append(queue, nodeLevel{item.node.Left, item.level + 1})
		}
		if item.node.Right != nil {
			queue = append(queue, nodeLevel{item.node.Right, item.level + 1})
		}
	}

	if len(levelNodes) > 0 {
		fmt.Printf("Level %d: %s\n", currentLevel, strings.Join(levelNodes, " | "))
	}
}
