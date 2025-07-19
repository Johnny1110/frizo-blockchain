package crypto

import (
	"errors"
	"frizo-blockchain/common"
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

// MPTNode support 4 MPTNodeType
// include path info (hex encoding)
// BRANCH node have 16 sub-node (not 2)
type MPTNode struct {
	NodeType MPTNodeType
	Path     []byte
	Value    []byte
	Children [16]*MPTNode // only BRANCH node have Children
	Child    *MPTNode     // only for EXTENSION node have 1 child
	Hash     *common.Hash
	Dirty    bool // remark is modified, optimized hash data
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

// ========== Encoding func ==========

// HexToCompact convert hex path to Compact(緊湊編碼) for saving space
// encoding rule:
// - first nibble represent node type: [00=extension even, 01=extension odd, 10=leaf even, 11=leaf odd]
func HexToCompact(hexPath []byte, isLeaf bool) []byte {
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

// ========== MPT Core Access Func ==========

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
		originIdx := leaf.Path[0]
		leaf.Path = leaf.Path[1:]
		newBranch.Children[originIdx] = leaf
		// leaf-2:
		newIndex := path[0]
		newBranch.Children[newIndex] = &MPTNode{
			NodeType: LEAF,
			Path:     path[1:],
			Value:    value,
			Dirty:    true,
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
			Child:    newBranch, // EXTENSION only have 1 child BRANCH.
		}

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
			Child:    newBranch, // EXTENSION only have 1 child BRANCH.
		}

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
			Child:    newBranch, // EXTENSION only have 1 child BRANCH.
		}

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
		ext.Child.Value = value
		ext.Child.Dirty = true
		return ext, nil
	}

	// ---------------------------------------------------------------------------
	// S-1: EXTENSION path length equals to commonLen
	// ---------------------------------------------------------------------------
	if len(ext.Path) == commonLen {
		remainingPath := path[commonLen:]
		// insert remainPath and value into child Branch.
		newChild, err := t.insert(ext.Child, remainingPath, value)
		if err != nil {
			return nil, err
		}
		ext.Child = newChild
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
				newBranch.Children[extPathIndex] = ext.Child
			} else {
				newExt := &MPTNode{
					NodeType: EXTENSION,
					Path:     extPathRemain[1:], // remove first element
					Child:    ext.Child,
					Dirty:    true,
				}
				newBranch.Children[extPathIndex] = newExt
			}

			if commonLen > 0 {
				ext.Path = ext.Path[:commonLen]
				ext.Child = newBranch
				return ext, nil
			} else {
				return newBranch, nil
			}
		}
	}

	// 2-1: input path have some part diff with common >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	if commonLen < len(path) {
		// TODO: implements.
	}

	return nil, errors.New("invalid insert MPT EXTENSION node conditions")
}

// insertIntoBranch Handle insert into BRANCH node
func (t *ModifiedMerklePatriciaTree) insertIntoBranch(branch *MPTNode, path []byte, value []byte) (*MPTNode, error) {
	// TODO
	return nil, nil
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
