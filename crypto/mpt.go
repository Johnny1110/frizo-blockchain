package crypto

import (
	"bytes"
	"errors"
	"fmt"
	"frizo-blockchain/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

// Modified Merkle Patricia Tree (MPT)

type MPTNodeType uint8 // 1 byte

func (nt MPTNodeType) String() string {
	switch nt {
	case LEAF:
		return "LEAF"
	case BRANCH:
		return "BRANCH"
	case EXTENSION:
		return "EXT"
	case BLANK:
		return "BLANK"
	}
	return "<unknown>"
}

const (
	// BLANK nil node
	BLANK MPTNodeType = iota

	// LEAF store actual value
	// format: [path-encode, value(mandatory)]
	LEAF

	// EXTENSION path compress
	// format: [shared_path, next-node-hash]
	EXTENSION

	// BRANCH
	// format: [16 sub-node-hash, value(optional)]
	BRANCH
)

// Node type in bytes length
const (
	// using decode bytes length to refer node type.
	branchListLength = 17 // BRANCH node have 17 elements（16 sub node + 1 value
	leafAndExtLength = 2  // LEAF and EXT have 2 elements
)

// MPTNode support 4 MPTNodeType
// include path info (hex encoding)
// BRANCH node have 16 sub-node (not 2)
type MPTNode struct {
	NodeType MPTNodeType
	Path     []byte
	Value    []byte
	Children [16]*MPTNode // only BRANCH node have 16 Children, EXTENSION node have 1 child
	Hash     *common.Hash // cached hash
	Dirty    bool         // remark is modified, optimized hash data
}

// ModifiedMerklePatriciaTree MPT main struct
type ModifiedMerklePatriciaTree struct {
	root     *MPTNode
	db       map[common.Hash][]byte
	hashFunc func([]byte) common.Hash
}

// NewMPT create new ModifiedMerklePatriciaTree
func NewMPT() *ModifiedMerklePatriciaTree {
	return &ModifiedMerklePatriciaTree{
		root:     nil,
		db:       make(map[common.Hash][]byte),
		hashFunc: defaultHashFunc,
	}
}

// ========== Encoding func =============================================================================

// HexToCompact convert hex path to Compact(緊湊編碼) for saving space
// encoding rule:
// - first nibble represent node type: [00=extension even, 01=extension odd, 10=leaf even, 11=leaf odd]
func HexToCompact(hexPath []byte, isLeaf bool) []byte {
	if len(hexPath) == 0 {
		return nil
	}
	// terminator (leaf is 1, extension is 0)
	terminator := byte(0) // default is extension (0000 0000)
	if isLeaf {
		terminator = 1 // leaf (0000 0001)
	}

	// process odd even hex path len (convert terminator)
	// - ex: even leaf -> from 1 -> 10
	// - ex: odd  leaf -> from 1 -> 11
	if len(hexPath)%2 == 1 {
		// odd len: first byte = (terminator*2 + 1) << 4 | hex[0]
		// combined 2 hex to 1 compact
		compact := make([]byte, len(hexPath)/2+1)     // why +1 ? because first one represent node-type
		compact[0] = (terminator*2+1)<<4 | hexPath[0] // odd situation: merge first hex path into compact head. (terminator*2+1) -> 添加奇偶屬性
		for i := 1; i < len(hexPath); i += 2 {        // skip first one, already filled first hexPath in previous line.
			compact[i/2+1] = hexPath[i]<<4 | hexPath[i+1] // combined 2 hexPath to 1 compact [0x(hex_1 hex_2), 0x(hex_3 hex_4), 0x(hex_5 hex_6)]
		}
		return compact
	} else {
		// even len: first byte = terminator * 2 << 4
		compact := make([]byte, len(hexPath)/2+1)
		compact[0] = terminator * 2 << 4
		for i := 0; i < len(hexPath); i += 2 {
			compact[i/2+1] = hexPath[i]<<4 | hexPath[i+1] // combined 2 hexPath to 1 compact [0x(hex_1 hex_2), 0x(hex_3 hex_4), 0x(hex_5 hex_6)]
		}
		return compact
	}
}

// CompactToHex revert compact to hexPath
func CompactToHex(compact []byte) ([]byte, bool) {
	if len(compact) == 0 {
		return nil, false
	}

	// parse first byte
	head := compact[0]
	flag := head >> 4 // node type: [0000 0000=extension even, 0000 0001=extension odd, 0000 0010=leaf even, 0000 0011=leaf odd]

	isLeaf := (flag & 2) != 0 // leaf is '0011' or '0010'
	isOdd := (flag & 1) != 0  // odd is 'xxx1' even is 'xxx0'

	if isOdd { // handle odd
		hexPath := make([]byte, (len(compact)-1)*2+1)
		// split first hexPath from compact[0] -> (xxxx xxxx) & (0000 1111)
		hexPath[0] = compact[0] & 0x0f
		for i := 1; i < len(compact); i++ {
			// split 1 compact to 2 hex
			hexPath[i*2-1] = compact[i] >> 4
			hexPath[i*2] = compact[i] & 0x0f
		}
		return hexPath, isLeaf
	} else { // handle even
		hexPath := make([]byte, (len(compact)-1)*2) // compact len remove head then * 2
		for i := 1; i < len(compact); i++ {
			hexPath[i*2-2] = compact[i] >> 4
			hexPath[i*2-1] = compact[i] & 0x0f
		}
		return hexPath, isLeaf
	}
}

// KeyToHex convert key to hexPath
func KeyToHex(key []byte) []byte {
	// split 1 key byte to 2 nibble, ex: 0x12 -> 0001 0002
	hexPath := make([]byte, len(key)*2)
	for idx, val := range key {
		hexPath[idx*2] = val >> 4
		hexPath[idx*2+1] = val & 0x0f
	}
	return hexPath
}

// ========== MPT Core Access Func ==================================================================

func (t *ModifiedMerklePatriciaTree) GetRoot() common.Hash {
	if t.root == nil {
		return common.Hash{}
	}
	// recursive calculate all dirty node HASH
	t.updateHashes(t.root)
	return *t.root.Hash
}

func (t *ModifiedMerklePatriciaTree) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("key can not be empty")
	}

	// convert key to key
	hexPath := KeyToHex(key)
	return t.get(t.root, hexPath)
}

func (t *ModifiedMerklePatriciaTree) Contains(key []byte) bool {
	value, err := t.Get(key)
	if err != nil {
		log.Error("[MPT] Contains failed", err)
		return false
	}
	return value != nil
}

func (t *ModifiedMerklePatriciaTree) Put(key, value []byte) error {
	if len(key) == 0 {
		return errors.New("key can not be empty")
	}

	// split 1 byte to 2 nibble.
	hexPath := KeyToHex(key)

	newRoot, err := t.insert(t.root, hexPath, value)
	if err != nil {
		return err
	}
	t.root = newRoot
	return nil
}

// Delete removes a key-value pair from the MPT
func (t *ModifiedMerklePatriciaTree) Delete(key []byte) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}

	// Convert key to hex path
	hexPath := KeyToHex(key)

	// Delete from root
	newRoot, err := t.delete(t.root, hexPath)
	if err != nil {
		return err
	}

	t.root = newRoot
	return nil
}

// Hash calculate Node HASH（with cache）
func (t *ModifiedMerklePatriciaTree) Hash(node *MPTNode) common.Hash {
	if node == nil {
		return common.Hash{}
	}

	// not dirty and have hash, just return cached HASH
	if !node.Dirty && node.Hash != nil {
		return *node.Hash
	} else { // calculate HASH
		// RLP encoding
		encoded := t.encodeNode(node)
		hash := t.hashFunc(encoded)

		node.Hash = &hash
		node.Dirty = false

		// encoded grater than 32 need to be store in levelDB (Ethereum std rule)
		if len(encoded) > 32 {
			t.db[hash] = encoded
		}

		return hash
	}
}

func (t *ModifiedMerklePatriciaTree) Commit() (common.Hash, error) {
	if t.root == nil {
		return common.Hash{}, nil
	}

	// calculate HASH
	rootHash := t.GetRoot()

	// store into levelDB
	// TODO: impl

	return rootHash, nil
}

// ==================================================================================================================================

// markDirty after update node, propagate dirty mark from node to root.
func (t *ModifiedMerklePatriciaTree) markDirty(node *MPTNode) {
	if node == nil {
		return
	}
	node.Dirty = true
}

// insert recursive insert node
// MPT core algorithm, handle all kind of NodeType.
func (t *ModifiedMerklePatriciaTree) insert(node *MPTNode, path []byte, value []byte) (*MPTNode, error) {
	// 1. if nil node, create leaf node
	if node == nil {
		return &MPTNode{
			NodeType: LEAF,
			Path:     path,
			Value:    value,
			Dirty:    true,
		}, nil
	}

	// mark as dirty node (require recalculate hash)
	node.Dirty = true

	switch node.NodeType {
	case LEAF:
		return t.insertIntoLeaf(node, path, value)
	case EXTENSION:
		return t.insertIntoExtension(node, path, value)
	case BRANCH:
		return t.insertIntoBranch(node, path, value)
	default:
		return nil, errors.New("invalid MPT node type")
	}
}

// insertIntoLeaf Handle insert into LEAF node
func (t *ModifiedMerklePatriciaTree) insertIntoLeaf(leaf *MPTNode, path []byte, value []byte) (*MPTNode, error) {
	// calculate common prefix length
	commonPathLen := commonPrefixLen(leaf.Path, path)

	// ---------------------------------------------------------------------------
	// S-1: All matched up: update node value
	// ---------------------------------------------------------------------------

	if commonPathLen == len(leaf.Path) && commonPathLen == len(path) {
		leaf.Value = value
		return leaf, nil
	}

	// ---------------------------------------------------------------------------
	// S-2: Need split a new Node
	// ---------------------------------------------------------------------------
	newBranch := &MPTNode{ // create new branch
		NodeType: BRANCH,
		Dirty:    true,
	}

	// 2-1: no common path prefix >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	if commonPathLen == 0 {

		// make 2 path to leaf and add into new branch.
		// leaf-1:
		if len(leaf.Path) == 0 {
			newBranch.Value = value
		} else {
			originIdx := leaf.Path[0]
			leaf.Path = leaf.Path[1:]
			newBranch.Children[originIdx] = leaf
		}
		// leaf-2:
		if len(path) == 0 {
			newBranch.Value = value
		} else {
			newIndex := path[0]
			newBranch.Children[newIndex] = &MPTNode{
				NodeType: LEAF,
				Path:     path[1:],
				Value:    value,
				Dirty:    true,
			}
		}
		return newBranch, nil
	}

	// 2-2: both Path have remaining >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	if commonPathLen < len(leaf.Path) && commonPathLen < len(path) {
		// we need create a EXTENSION node to compress the path
		extension := &MPTNode{
			NodeType: EXTENSION,
			Path:     leaf.Path[:commonPathLen],
			Dirty:    true,
		}
		// EXTENSION only have 1 child BRANCH.
		extension.Children[0] = newBranch

		// leaf-1:
		originIdx := leaf.Path[commonPathLen]
		leaf.Path = leaf.Path[commonPathLen+1:]
		newBranch.Children[originIdx] = leaf
		// leaf-2:
		newIndex := path[commonPathLen]
		newBranch.Children[newIndex] = &MPTNode{
			NodeType: LEAF,
			Path:     path[commonPathLen+1:],
			Value:    value,
			Dirty:    true,
		}

		return extension, nil
	}

	// 2-3: only origin leaf Path has remaining >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	if commonPathLen < len(leaf.Path) && commonPathLen == len(path) {
		// we need create a EXTENSION node to compress the path
		extension := &MPTNode{
			NodeType: EXTENSION,
			Path:     leaf.Path[:commonPathLen],
			Dirty:    true,
		}
		extension.Children[0] = newBranch

		// leaf-1:
		originIdx := leaf.Path[commonPathLen]
		leaf.Path = leaf.Path[commonPathLen+1:]
		newBranch.Children[originIdx] = leaf
		// new path process:
		newBranch.Value = value

		return extension, nil
	}

	// 2-4: only new Path has remaining >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	if commonPathLen == len(leaf.Path) && commonPathLen < len(path) {
		// we need create a EXTENSION node to compress the path
		extension := &MPTNode{
			NodeType: EXTENSION,
			Path:     leaf.Path[:commonPathLen],
			Dirty:    true,
		}
		extension.Children[0] = newBranch

		// leaf-1:
		newBranch.Value = leaf.Value
		// leaf-2:
		newIndex := path[commonPathLen]
		newBranch.Children[newIndex] = &MPTNode{
			NodeType: LEAF,
			Path:     path[commonPathLen+1:],
			Value:    value,
			Dirty:    true,
		}

		return extension, nil
	}

	return nil, errors.New("invalid insert MPT leaf node conditions")
}

// insertIntoExtension Handle insert into EXTENSION node
// EXTENSION only have 1 child branch and cannot have any value.
func (t *ModifiedMerklePatriciaTree) insertIntoExtension(ext *MPTNode, path []byte, value []byte) (*MPTNode, error) {
	commonLen := commonPrefixLen(ext.Path, path)

	if commonLen == len(ext.Path) && commonLen == len(path) {
		// overwrite value to EXTENSION's child branch -> branch.value
		newChild, err := t.insert(ext.Children[0], []byte{}, value)
		if err != nil {
			return nil, err
		}
		ext.Children[0] = newChild
		return ext, nil
	}

	// ---------------------------------------------------------------------------
	// S-1: EXTENSION path length equals to commonLen
	// ---------------------------------------------------------------------------
	if len(ext.Path) == commonLen {
		remainingPath := path[commonLen:]
		// insert remainPath and value into child Branch.
		newChild, err := t.insert(ext.Children[0], remainingPath, value)
		if err != nil {
			return nil, err
		}
		ext.Children[0] = newChild
		return ext, nil
	}

	// ---------------------------------------------------------------------------
	// S-2: need split a new EXTENSION
	// ---------------------------------------------------------------------------
	if commonLen < len(ext.Path) {

		// 2-1: input path equals to common path >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
		if len(path) == commonLen {
			newBranch := &MPTNode{
				NodeType: BRANCH,
				Value:    value, // store value in new Branch
				Dirty:    true,
			}

			extPathRemain := ext.Path[commonLen:]
			extPathIndex := extPathRemain[0]
			if len(extPathRemain)-1 == 0 {
				// without first byte, no more ext path left.
				newBranch.Children[extPathIndex] = ext.Children[0]
			} else {
				newExt := &MPTNode{
					NodeType: EXTENSION,
					Path:     extPathRemain[1:], // remove first element
					Dirty:    true,
				}
				newExt.Children[0] = ext.Children[0]
				newBranch.Children[extPathIndex] = newExt
			}

			ext.Path = ext.Path[:commonLen]
			ext.Children[0] = newBranch
			return ext, nil
		}

		// 2-2: input path have some part diff with common >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
		if commonLen < len(path) {
			newExt := &MPTNode{
				NodeType: EXTENSION,
				Dirty:    true,
				Path:     ext.Path[:commonLen],
			}

			newBranch := &MPTNode{
				NodeType: BRANCH,
				Dirty:    true,
			}
			newExt.Children[0] = newBranch

			// let origin ext concat to newBranch
			originExtRemainPath := ext.Path[commonLen:]
			originExtIndex := originExtRemainPath[0]
			ext.Path = originExtRemainPath[1:]
			newBranch.Children[originExtIndex] = ext

			// create new Leaf for input
			inputRemainPath := path[commonLen:]
			inputIndex := inputRemainPath[0]
			newLeaf := &MPTNode{
				NodeType: LEAF,
				Path:     inputRemainPath[1:],
				Value:    value,
				Dirty:    true,
			}
			newBranch.Children[inputIndex] = newLeaf
			return newExt, nil
		}
	}

	return nil, errors.New("invalid insert MPT EXTENSION node conditions")
}

// insertIntoBranch Handle insert into BRANCH node
func (t *ModifiedMerklePatriciaTree) insertIntoBranch(branch *MPTNode, path []byte, value []byte) (*MPTNode, error) {
	// ---------------------------------------------------------------------------
	// S-1: path is empty -> update directly
	// ---------------------------------------------------------------------------
	if len(path) == 0 {
		branch.Value = value
		return branch, nil
	}

	// ---------------------------------------------------------------------------
	// S-2: path is not empty
	// ---------------------------------------------------------------------------
	branchIdx := path[0]
	remainingPath := path[1:]

	// S-2-1: branch.Children[idx] is nil: create a new Leaf >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	if branch.Children[branchIdx] == nil {
		branch.Children[branchIdx] = &MPTNode{
			NodeType: LEAF,
			Path:     remainingPath,
			Value:    value,
			Dirty:    true,
		}
		return branch, nil
	}

	// S-2-2: branch.Children[idx] is not nil: recursive insert >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	if branch.Children[branchIdx] != nil {
		newChild, err := t.insert(branch.Children[branchIdx], remainingPath, value)
		if err != nil {
			return nil, err
		}
		branch.Children[branchIdx] = newChild
		return branch, nil
	}

	return nil, errors.New("invalid insert MPT BRANCH node conditions")
}

func (t *ModifiedMerklePatriciaTree) get(node *MPTNode, path []byte) ([]byte, error) {
	if node == nil {
		return nil, nil // key not found
	}

	switch node.NodeType {
	case LEAF:
		return t.getFromLeaf(node, path)
	case EXTENSION:
		return t.getFromExtension(node, path)
	case BRANCH:
		return t.getFromBranch(node, path)
	default:
		return nil, errors.New("invalid MPT node type")
	}
}

func (t *ModifiedMerklePatriciaTree) getFromLeaf(leaf *MPTNode, path []byte) ([]byte, error) {
	if bytes.Equal(leaf.Path, path) {
		return leaf.Value, nil
	}
	return nil, nil
}

func (t *ModifiedMerklePatriciaTree) getFromExtension(ext *MPTNode, path []byte) ([]byte, error) {
	if len(path) < len(ext.Path) {
		return nil, nil
	}

	// ext path not match input path prefix
	if !bytes.Equal(path[:len(ext.Path)], ext.Path) {
		return nil, nil // prefix doesn't match, key not found
	}

	remainPath := path[len(ext.Path):]
	return t.get(ext.Children[0], remainPath)
}

func (t *ModifiedMerklePatriciaTree) getFromBranch(branch *MPTNode, path []byte) ([]byte, error) {
	// If path is empty, return the branch's value (if any)
	if len(path) == 0 {
		return branch.Value, nil
	}

	// Get the branch index from the first byte of path
	branchIndex := path[0]
	remainingPath := path[1:]

	// Check if the child at this index exists
	if branch.Children[branchIndex] == nil {
		return nil, nil // key not found
	}

	// Recursively search in the selected child
	return t.get(branch.Children[branchIndex], remainingPath)
}

func (t *ModifiedMerklePatriciaTree) delete(node *MPTNode, path []byte) (*MPTNode, error) {
	if node == nil {
		return nil, nil // key not found, nothing to delete
	}

	node.Dirty = true

	switch node.NodeType {
	case LEAF:
		return t.deleteFromLeaf(node, path)
	case EXTENSION:
		return t.deleteFromExtension(node, path)
	case BRANCH:
		return t.deleteFromBranch(node, path)
	default:
		return nil, errors.New("invalid MPT node type")
	}
}

func (t *ModifiedMerklePatriciaTree) deleteFromLeaf(leaf *MPTNode, path []byte) (*MPTNode, error) {
	if bytes.Equal(leaf.Path, path) {
		return nil, nil // Delete the leaf by returning nil
	}

	// Path doesn't match, key not found
	return leaf, nil
}

func (t *ModifiedMerklePatriciaTree) deleteFromExtension(ext *MPTNode, path []byte) (*MPTNode, error) {
	if len(path) < len(ext.Path) {
		return ext, nil // input path shorter than ext path, return ext as key not found
	}

	if !bytes.Equal(path[:len(ext.Path)], ext.Path) {
		return ext, nil // Prefix doesn't match, key not found
	}

	remainingPath := path[len(ext.Path):]
	newChild, err := t.delete(ext.Children[0], remainingPath)
	if err != nil {
		return nil, err
	}

	if newChild == nil {
		return nil, nil // Child Branch was deleted, so delete this ext also.
	}

	ext.Children[0] = newChild // replace child with newChild
	return t.mergeExtensionIfNeeded(ext)
}

func (t *ModifiedMerklePatriciaTree) deleteFromBranch(branch *MPTNode, path []byte) (*MPTNode, error) {
	if len(path) == 0 {
		// Delete the value at this branch
		branch.Value = nil
		return t.normalizeBranch(branch)
	}

	// Get the branch index
	branchIndex := path[0]
	remainingPath := path[1:]

	// Check if child exists
	if branch.Children[branchIndex] == nil {
		return branch, nil // Key not found
	}

	// Recursively delete from child
	newChild, err := t.delete(branch.Children[branchIndex], remainingPath)
	if err != nil {
		return nil, err
	}

	// Update the child
	branch.Children[branchIndex] = newChild

	// Normalize the branch (it might need to be converted)
	return t.normalizeBranch(branch)
}

// mergeExtensionIfNeeded checks if an extension needs to be merged with its child
func (t *ModifiedMerklePatriciaTree) mergeExtensionIfNeeded(ext *MPTNode) (*MPTNode, error) {
	// if child node is ext -> merge path as 1 ext
	// if child is leaf, merge a new leaf
	child := ext.getExtensionChild()
	switch child.NodeType {
	case LEAF:
		// Merge extension and leaf into a single leaf
		newPath := append(ext.Path, child.Path...)
		return &MPTNode{
			NodeType: LEAF,
			Path:     newPath,
			Value:    child.Value,
			Dirty:    true,
		}, nil
	case EXTENSION:
		// Merge two extensions into one
		newPath := append(ext.Path, child.Path...)
		return &MPTNode{
			NodeType: EXTENSION,
			Path:     newPath,
			Children: [16]*MPTNode{child.Children[0]},
			Dirty:    true,
		}, nil
	case BRANCH:
		// Branch can not be merged.
		return ext, nil
	default:
		return nil, errors.New("invalid MPT node type")
	}
}

// normalizeBranch handles branch node transformation after deletion
func (t *ModifiedMerklePatriciaTree) normalizeBranch(branch *MPTNode) (*MPTNode, error) {
	// Count non-nil children
	var nonNilChildrenIdx []int
	for i, child := range branch.Children {
		if child != nil {
			nonNilChildrenIdx = append(nonNilChildrenIdx, i)
		}
	}

	switch len(nonNilChildrenIdx) {
	case 0: // No children left
		if branch.Value != nil {
			// Convert to leaf with empty path
			return &MPTNode{
				NodeType: LEAF,
				Path:     []byte{},
				Value:    branch.Value,
				Dirty:    true,
			}, nil
		} else {
			// No children and no value, delete the branch
			return nil, nil
		}
	case 1: // Only one child left
		childIndex := nonNilChildrenIdx[0]
		child := branch.Children[childIndex]
		if branch.Value == nil {
			// No value at branch, can merge with child
			newNode, err := t.mergeBranchWithSingleChild(branch, childIndex, child)
			return newNode, err
		} else {
			// Has value, must keep as branch
			return branch, nil
		}
	default:
		// Multiple children, keep as branch
		return branch, nil
	}

}

// mergeBranchWithSingleChild merges a branch with its single child
func (t *ModifiedMerklePatriciaTree) mergeBranchWithSingleChild(branch *MPTNode, childIndex int, child *MPTNode) (*MPTNode, error) {
	switch child.NodeType {
	case LEAF:
		// merge into a single leaf
		return &MPTNode{
			NodeType: LEAF,
			Path:     append([]byte{byte(childIndex)}, child.Path...),
			Value:    child.Value,
			Dirty:    true,
		}, nil
	case EXTENSION:
		// Merge into a single extension
		return &MPTNode{
			NodeType: EXTENSION,
			Path:     append([]byte{byte(childIndex)}, child.Path...),
			Children: [16]*MPTNode{child.Children[0]}, // Copy the child's child
			Dirty:    true,
		}, nil
	case BRANCH:
		// Create an extension pointing to the branch
		return &MPTNode{
			NodeType: EXTENSION,
			Path:     []byte{byte(childIndex)},
			Children: [16]*MPTNode{child},
			Dirty:    true,
		}, nil
	default:
		return nil, errors.New("invalid MPT node type")
	}
}

// encodeNode RLP Encoding (go-ethereum impl)
// doc: https://ethbook.abyteahead.com/ch4/rlp.html
func (t *ModifiedMerklePatriciaTree) encodeNode(node *MPTNode) []byte {
	switch node.NodeType {
	case LEAF:
		return t.encodeLeaf(node)
	case EXTENSION:
		return t.encodeExtension(node)
	case BRANCH:
		return t.encodeBranch(node)
	default:
		return nil
	}
}

// encodeLeaf RLF encoding LEAF Node
func (t *ModifiedMerklePatriciaTree) encodeLeaf(node *MPTNode) []byte {
	compactPath := HexToCompact(node.Path, true)
	// RLP encoding
	encoded, err := rlp.EncodeToBytes([]interface{}{
		compactPath,
		node.Value,
	})

	if err != nil {
		panic(err)
	}

	return encoded
}

func (t *ModifiedMerklePatriciaTree) encodeExtension(node *MPTNode) []byte {
	compactPath := HexToCompact(node.Path, false)
	childRef := t.nodeRef(node.Children[0])
	encoded, err := rlp.EncodeToBytes([]interface{}{
		compactPath,
		childRef,
	})

	if err != nil {
		panic(err)
	}

	return encoded
}

func (t *ModifiedMerklePatriciaTree) encodeBranch(node *MPTNode) []byte {
	// Branch Node: [all subNodes ref, node.value]
	var refs []interface{}

	// add 16 sub node ref
	for i := 0; i < 16; i++ {
		if node.Children[i] != nil {
			refs = append(refs, t.nodeRef(node.Children[i]))
		} else {
			refs = append(refs, []byte{})
		}
	}

	// add node.Value
	refs = append(refs, node.Value)

	encoded, err := rlp.EncodeToBytes(refs)

	if err != nil {
		panic(err)
	}

	return encoded
}

func (t *ModifiedMerklePatriciaTree) nodeRef(node *MPTNode) []byte {
	if node == nil {
		return []byte{}
	}

	// encode node.
	encoded := t.encodeNode(node)

	// ethereum std rule: < 32 bytes just return node encode
	if len(encoded) < 32 {
		return encoded
	}

	// ethereum std rule: > 32 bytes return HASH
	hash := t.Hash(node)
	return hash.Bytes()
}

func (t *ModifiedMerklePatriciaTree) updateHashes(node *MPTNode) {
	if node == nil {
		return
	}

	// update all sub nodes (BRANCH and EXT)
	switch node.NodeType {
	case BRANCH:
		for _, child := range node.Children {
			if child != nil {
				t.updateHashes(child)
			}
		}
	case EXTENSION:
		if node.Children[0] != nil {
			t.updateHashes(node.Children[0])
		}
	}

	// calculate current node
	t.Hash(node)
}

// decodeNode decoding bytes with RLP, return a MPT node
func (t *ModifiedMerklePatriciaTree) decodeNode(data []byte) (*MPTNode, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var decoded []interface{}
	if err := rlp.DecodeBytes(data, &decoded); err != nil {
		return nil, fmt.Errorf("failed to using RLP to decode node: %v", err)
	}

	switch len(decoded) {
	case leafAndExtLength:
		// Leaf Node: path + value
		// Ext Node: path + sub_ref
		return t.decodeLeafOrExtension(decoded)
	case branchListLength:
		// Branch Node: [all subNodes ref, node.value]
		return t.decodeBranch(decoded)
	default:
		return nil, fmt.Errorf("invalid node format: wrong element count %d", len(decoded))
	}
}

// decodeLeafOrExtension RLP decode leaf or ext node
func (t *ModifiedMerklePatriciaTree) decodeLeafOrExtension(decoded []interface{}) (*MPTNode, error) {
	if len(decoded) != leafAndExtLength {
		return nil, fmt.Errorf("invalid branch node: expected %d elements, got %d",
			leafAndExtLength, len(decoded))
	}

	// first element is compact path
	pathData, ok := decoded[0].([]byte)
	if !ok {
		return nil, errors.New("invalid path data in leaf/extension node")
	}

	hexPath, isLeaf := CompactToHex(pathData)

	if isLeaf {
		// second element is value
		value, ok := decoded[1].([]byte)
		if !ok {
			return nil, errors.New("invalid value in leaf node")
		}

		return &MPTNode{
			NodeType: LEAF,
			Path:     hexPath,
			Value:    value,
			Dirty:    false,
		}, nil
	} else { // EXTENSION node
		return &MPTNode{
			NodeType: EXTENSION,
			Path:     hexPath,
			Dirty:    false,
			// children[0] using loadChild to load data later
		}, nil
	}
}

// decodeBranch RLP decode branch node
func (t *ModifiedMerklePatriciaTree) decodeBranch(decoded []interface{}) (*MPTNode, error) {
	if len(decoded) != branchListLength {
		return nil, fmt.Errorf("invalid branch node: expected %d elements, got %d",
			branchListLength, len(decoded))
	}

	branch := &MPTNode{
		NodeType: BRANCH,
		Dirty:    false,
	}

	// no.17 element is branch.value（nullable）
	if value, ok := decoded[16].([]byte); ok && len(value) > 0 {
		branch.Value = value
	}

	// no.1 ~ no.16 is all children, will be restored by loadChild() later

	return branch, nil
}

// loadNode load node from db
func (t *ModifiedMerklePatriciaTree) loadNode(hash common.Hash) (*MPTNode, error) {
	if hash == (common.Hash{}) {
		return nil, nil
	}

	// load data form db
	data, exists := t.db[hash]
	if !exists {
		return nil, fmt.Errorf("node not found: %s", hash.Hex())
	}

	// decode node
	node, err := t.decodeNode(data)
	if err != nil {
		return nil, err
	}

	// set hash to node (load data is not dirty)
	node.Hash = &hash

	return node, nil
}

// resolveNode resolve node ref
func (t *ModifiedMerklePatriciaTree) resolveNode(ref interface{}) (*MPTNode, error) {
	if ref == nil {
		return nil, nil
	}

	switch r := ref.(type) {
	case []byte:
		if len(r) == 0 {
			return nil, nil
		}

		// HASH
		if len(r) == common.HashLength {
			// Hash ref, load from db
			hash := common.BytesToHash(r)
			return t.loadNode(hash)
		} else {
			// node decode directly
			return t.decodeNode(r)
		}

	default:
		return nil, fmt.Errorf("invalid node reference type: %T", ref)
	}
}

func (n *MPTNode) getExtensionChild() *MPTNode {
	if n.NodeType == EXTENSION {
		return n.Children[0]
	}
	return nil
}

func (n *MPTNode) setExtensionChild(child *MPTNode) {
	if n.NodeType == EXTENSION {
		n.Children[0] = child
	}
}

// ========== Tool func ==========

// commonPrefixLen calculate 2 path common path prefix len
func commonPrefixLen(pathA []byte, pathB []byte) int {
	minLen := min(len(pathA), len(pathB))

	for i := 0; i < minLen; i++ {
		if pathA[i] != pathB[i] {
			return i
		}
	}
	return minLen
}
