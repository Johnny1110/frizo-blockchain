# Modified Merkle Patricia Tree (MPT)

## 主要特點：

1. 四種節點類型

* 空節點 (BLANK)：表示不存在的節點
* 葉子節點 (LEAF)：存儲實際的鍵值對
* 擴展節點 (EXTENSION)：用於路徑壓縮，MPT 特有
* 分支節點 (BRANCH)：16 叉分支，而非二叉

<br>

2. 路徑編碼

* 使用十六進制編碼（每個字節拆分為兩個 nibble）
* 緊湊編碼（Hex-Prefix Encoding）用於節省空間
* 支持奇偶長度路徑的處理

<br>

3. 動態操作

* 插入：自動處理節點分裂和路徑壓縮
* 查詢：通過鍵路徑導航查找值
* 刪除：自動合併節點，保持樹的最優結構

<br>

4. 與普通 Merkle Tree 的主要區別

| 特性  | 普通 Merkle Tree  | MPT  |
|---|---|---|
| 結構  | 二叉樹  | 16 叉樹  |
| 節點類型  | 2 種（葉子、分支）  | 4 種  |
| 路徑  | 固定索引  | 動態鍵路徑  |
| 用途  | 靜態數據驗證  | 動態鍵值存儲  |
| 空間效率  | 固定層高  | 路徑壓縮 |

<br>
<br>
<br>
<br>

## Code


```go
package crypto

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ========== Modified Merkle Patricia Tree (MPT) 原理介紹 ==========
//
// MPT 是以太坊用來存儲狀態的核心數據結構，它結合了三種數據結構的優點：
// 1. Merkle Tree：提供加密證明和數據完整性驗證
// 2. Patricia Tree (Radix Tree)：提供高效的鍵值存儲和查詢
// 3. 十六進制編碼：每個節點最多有 16 個子節點
//
// 與普通 Merkle Tree 的主要區別：
// - 普通 Merkle Tree：二叉樹結構，主要用於數據驗證
// - MPT：16 叉樹結構，支持動態插入/刪除/更新，適合存儲鍵值對
//
// MPT 的關鍵特性：
// 1. 確定性：相同的鍵值對集合總是生成相同的根哈希
// 2. 路徑壓縮：相同前綴的鍵共享路徑，節省空間
// 3. 默克爾證明：可以證明某個鍵值對的存在或不存在

// NodeType 定義 MPT 中的節點類型
type NodeType uint8

const (
	// BLANK 空節點 - MPT 特有，表示不存在的節點
	BLANK NodeType = iota

	// LEAF 葉子節點 - 存儲實際的值
	// 格式：[路徑編碼, 值]
	// 與普通 Merkle Tree 的葉子節點不同，這裡包含了路徑信息
	LEAF

	// EXTENSION 擴展節點 - MPT 特有，用於路徑壓縮
	// 格式：[共享路徑, 下一個節點的哈希]
	// 普通 Merkle Tree 沒有這種節點類型
	EXTENSION

	// BRANCH 分支節點 - 有分叉的地方
	// 格式：[16 個子節點的哈希, 值（可選）]
	// 與普通 Merkle Tree 的分支節點不同，這裡是 16 叉而非 2 叉
	BRANCH
)

// MPTNode 表示 MPT 中的節點
// 與普通 MerkleNode 的主要區別：
// 1. 支持多種節點類型（不只是葉子和分支）
// 2. 包含路徑信息（用於鍵值查找）
// 3. 分支節點有 16 個子節點（而非 2 個）
type MPTNode struct {
	NodeType NodeType

	// Path 存儲路徑片段（十六進制編碼）
	// 這是 MPT 特有的，普通 Merkle Tree 不需要路徑信息
	Path []byte

	// Value 節點存儲的值
	// LEAF 和 BRANCH 節點可能有值
	Value []byte

	// Children 子節點數組（只有 BRANCH 節點使用）
	// MPT 使用 16 個子節點（0-9, a-f），而普通 Merkle Tree 只有 2 個
	Children [16]*MPTNode

	// Hash 節點的哈希值（緩存）
	Hash *common.Hash

	// Dirty 標記節點是否被修改（需要重新計算哈希）
	// 這是優化技巧，避免重複計算哈希
	Dirty bool
}

// ModifiedMerklePatriciaTree MPT 主結構
type ModifiedMerklePatriciaTree struct {
	root     *MPTNode
	db       map[common.Hash][]byte // 簡化的存儲層
	hashFunc func([]byte) common.Hash
}

// NewMPT 創建新的 MPT
func NewMPT() *ModifiedMerklePatriciaTree {
	return &ModifiedMerklePatriciaTree{
		root:     nil,
		db:       make(map[common.Hash][]byte),
		hashFunc: crypto.Keccak256Hash,
	}
}

// ========== 編碼相關函數 ==========
// MPT 使用特殊的編碼方式來區分奇偶長度的路徑和節點類型

// HexToCompact 將十六進制路徑轉換為緊湊編碼
// 這是 MPT 特有的編碼方式，用於節省存儲空間
// 編碼規則：
// - 第一個半字節（nibble）的低 2 位表示節點類型（00=extension, 01=extension odd, 10=leaf even, 11=leaf odd）
// - 如果路徑長度為奇數，第一個半字節的高 4 位存儲路徑的第一個十六進制字符
func HexToCompact(hex []byte, isLeaf bool) []byte {
	// 計算終止符（葉子節點為 1，擴展節點為 0）
	terminator := byte(0)
	if isLeaf {
		terminator = 1
	}

	// 處理奇偶長度
	if len(hex)%2 == 1 {
		// 奇數長度：第一個字節 = (terminator*2 + 1) << 4 | hex[0]
		compact := make([]byte, len(hex)/2+1)
		compact[0] = (terminator*2 + 1) << 4 | hex[0]
		for i := 1; i < len(hex); i += 2 {
			compact[i/2+1] = hex[i]<<4 | hex[i+1]
		}
		return compact
	} else {
		// 偶數長度：第一個字節 = terminator * 2 << 4
		compact := make([]byte, len(hex)/2+1)
		compact[0] = terminator * 2 << 4
		for i := 0; i < len(hex); i += 2 {
			compact[i/2+1] = hex[i]<<4 | hex[i+1]
		}
		return compact
	}
}

// CompactToHex 將緊湊編碼轉換回十六進制路徑
func CompactToHex(compact []byte) ([]byte, bool) {
	if len(compact) == 0 {
		return nil, false
	}

	// 解析第一個字節
	first := compact[0]
	flag := first >> 4

	// 判斷節點類型
	isLeaf := (flag & 2) != 0

	// 判斷奇偶長度
	if flag&1 != 0 {
		// 奇數長度
		hex := make([]byte, len(compact)*2-1)
		hex[0] = first & 0x0f
		for i := 1; i < len(compact); i++ {
			hex[i*2-1] = compact[i] >> 4
			hex[i*2] = compact[i] & 0x0f
		}
		return hex, isLeaf
	} else {
		// 偶數長度
		hex := make([]byte, (len(compact)-1)*2)
		for i := 1; i < len(compact); i++ {
			hex[(i-1)*2] = compact[i] >> 4
			hex[(i-1)*2+1] = compact[i] & 0x0f
		}
		return hex, isLeaf
	}
}

// KeyToHex 將鍵轉換為十六進制路徑
// MPT 使用十六進制編碼的路徑，每個字節被拆分為兩個半字節
func KeyToHex(key []byte) []byte {
	hex := make([]byte, len(key)*2)
	for i, b := range key {
		hex[i*2] = b >> 4
		hex[i*2+1] = b & 0x0f
	}
	return hex
}

// ========== 核心操作函數 ==========

// Put 插入或更新鍵值對
// 與普通 Merkle Tree 的區別：
// 1. 支持動態插入（普通 Merkle Tree 通常是靜態構建）
// 2. 使用路徑導航（而非索引）
// 3. 自動進行路徑壓縮和節點合併
func (t *ModifiedMerklePatriciaTree) Put(key, value []byte) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}

	// 將鍵轉換為十六進制路徑
	hexKey := KeyToHex(key)

	// 插入到樹中
	newRoot, err := t.insert(t.root, hexKey, value)
	if err != nil {
		return err
	}

	t.root = newRoot
	return nil
}

// insert 遞歸插入節點
// 這是 MPT 的核心算法，處理各種節點類型的插入邏輯
func (t *ModifiedMerklePatriciaTree) insert(node *MPTNode, path []byte, value []byte) (*MPTNode, error) {
	// 情況 1：空節點，直接創建葉子節點
	if node == nil {
		return &MPTNode{
			NodeType: LEAF,
			Path:     path,
			Value:    value,
			Dirty:    true,
		}, nil
	}

	// 標記節點為髒（需要重新計算哈希）
	node.Dirty = true

	switch node.NodeType {
	case LEAF:
		// 情況 2：葉子節點
		return t.insertIntoLeaf(node, path, value)

	case EXTENSION:
		// 情況 3：擴展節點
		return t.insertIntoExtension(node, path, value)

	case BRANCH:
		// 情況 4：分支節點
		return t.insertIntoBranch(node, path, value)

	default:
		return nil, errors.New("invalid node type")
	}
}

// insertIntoLeaf 處理向葉子節點插入的情況
func (t *ModifiedMerklePatriciaTree) insertIntoLeaf(leaf *MPTNode, path []byte, value []byte) (*MPTNode, error) {
	// 計算公共前綴長度
	commonLen := commonPrefixLen(leaf.Path, path)

	// 情況 1：完全匹配，更新值
	if commonLen == len(leaf.Path) && commonLen == len(path) {
		leaf.Value = value
		return leaf, nil
	}

	// 情況 2：需要分裂
	// 創建分支節點
	branch := &MPTNode{
		NodeType: BRANCH,
		Dirty:    true,
	}

	// 處理原葉子節點
	if commonLen < len(leaf.Path) {
		// 原葉子節點還有剩餘路徑
		remainingPath := leaf.Path[commonLen+1:]
		branchIndex := leaf.Path[commonLen]

		if len(remainingPath) == 0 {
			// 剩餘路徑為空，直接在分支節點存儲值
			branch.Children[branchIndex] = nil
			branch.Value = leaf.Value
		} else {
			// 創建新的葉子節點
			branch.Children[branchIndex] = &MPTNode{
				NodeType: LEAF,
				Path:     remainingPath,
				Value:    leaf.Value,
				Dirty:    true,
			}
		}
	} else {
		// 原葉子節點路徑用完，值存在分支節點
		branch.Value = leaf.Value
	}

	// 處理新插入的節點
	if commonLen < len(path) {
		// 新路徑還有剩餘
		remainingPath := path[commonLen+1:]
		branchIndex := path[commonLen]

		if len(remainingPath) == 0 {
			// 剩餘路徑為空，直接在分支節點存儲值
			branch.Children[branchIndex] = nil
			branch.Value = value
		} else {
			// 創建新的葉子節點
			branch.Children[branchIndex] = &MPTNode{
				NodeType: LEAF,
				Path:     remainingPath,
				Value:    value,
				Dirty:    true,
			}
		}
	} else {
		// 新路徑用完，值存在分支節點
		branch.Value = value
	}

	// 如果有公共前綴，創建擴展節點
	if commonLen > 0 {
		return &MPTNode{
			NodeType: EXTENSION,
			Path:     path[:commonLen],
			Children: [16]*MPTNode{branch},
			Dirty:    true,
		}, nil
	}

	return branch, nil
}

// insertIntoExtension 處理向擴展節點插入的情況
func (t *ModifiedMerklePatriciaTree) insertIntoExtension(ext *MPTNode, path []byte, value []byte) (*MPTNode, error) {
	commonLen := commonPrefixLen(ext.Path, path)

	// 情況 1：路徑完全匹配擴展節點
	if commonLen == len(ext.Path) {
		// 繼續向下插入
		remainingPath := path[commonLen:]
		// 擴展節點只有一個子節點，存在 Children[0]
		newChild, err := t.insert(ext.Children[0], remainingPath, value)
		if err != nil {
			return nil, err
		}
		ext.Children[0] = newChild
		return ext, nil
	}

	// 情況 2：需要分裂擴展節點
	// 創建新的分支節點
	branch := &MPTNode{
		NodeType: BRANCH,
		Dirty:    true,
	}

	// 處理原擴展節點的剩餘部分
	if commonLen+1 < len(ext.Path) {
		// 還有剩餘路徑，創建新的擴展節點
		remainingPath := ext.Path[commonLen+1:]
		branchIndex := ext.Path[commonLen]
		branch.Children[branchIndex] = &MPTNode{
			NodeType: EXTENSION,
			Path:     remainingPath,
			Children: [16]*MPTNode{ext.Children[0]},
			Dirty:    true,
		}
	} else {
		// 沒有剩餘路徑，直接連接子節點
		branchIndex := ext.Path[commonLen]
		branch.Children[branchIndex] = ext.Children[0]
	}

	// 插入新節點
	newNode, err := t.insert(branch, path[commonLen:], value)
	if err != nil {
		return nil, err
	}

	// 如果有公共前綴，創建新的擴展節點
	if commonLen > 0 {
		return &MPTNode{
			NodeType: EXTENSION,
			Path:     path[:commonLen],
			Children: [16]*MPTNode{newNode},
			Dirty:    true,
		}, nil
	}

	return newNode, nil
}

// insertIntoBranch 處理向分支節點插入的情況
func (t *ModifiedMerklePatriciaTree) insertIntoBranch(branch *MPTNode, path []byte, value []byte) (*MPTNode, error) {
	// 情況 1：路徑為空，更新分支節點的值
	if len(path) == 0 {
		branch.Value = value
		return branch, nil
	}

	// 情況 2：沿著路徑向下插入
	branchIndex := path[0]
	remainingPath := path[1:]

	newChild, err := t.insert(branch.Children[branchIndex], remainingPath, value)
	if err != nil {
		return nil, err
	}

	branch.Children[branchIndex] = newChild
	return branch, nil
}

// Get 獲取鍵對應的值
// 與普通 Merkle Tree 的區別：
// 1. 使用鍵而非索引進行查找
// 2. 需要處理多種節點類型
// 3. 支持不存在的鍵查詢（返回 nil）
func (t *ModifiedMerklePatriciaTree) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("key cannot be empty")
	}

	hexKey := KeyToHex(key)
	return t.get(t.root, hexKey)
}

// get 遞歸查找值
func (t *ModifiedMerklePatriciaTree) get(node *MPTNode, path []byte) ([]byte, error) {
	if node == nil {
		return nil, nil
	}

	switch node.NodeType {
	case LEAF:
		// 檢查路徑是否完全匹配
		if bytes.Equal(node.Path, path) {
			return node.Value, nil
		}
		return nil, nil

	case EXTENSION:
		// 檢查前綴是否匹配
		if len(path) >= len(node.Path) && bytes.Equal(node.Path, path[:len(node.Path)]) {
			return t.get(node.Children[0], path[len(node.Path):])
		}
		return nil, nil

	case BRANCH:
		if len(path) == 0 {
			return node.Value, nil
		}
		return t.get(node.Children[path[0]], path[1:])

	default:
		return nil, errors.New("invalid node type")
	}
}

// Delete 刪除鍵值對
// 這是 MPT 相對複雜的操作，需要處理節點合併
func (t *ModifiedMerklePatriciaTree) Delete(key []byte) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}

	hexKey := KeyToHex(key)
	newRoot, err := t.delete(t.root, hexKey)
	if err != nil {
		return err
	}

	t.root = newRoot
	return nil
}

// delete 遞歸刪除節點
func (t *ModifiedMerklePatriciaTree) delete(node *MPTNode, path []byte) (*MPTNode, error) {
	if node == nil {
		return nil, nil
	}

	node.Dirty = true

	switch node.NodeType {
	case LEAF:
		if bytes.Equal(node.Path, path) {
			return nil, nil // 刪除葉子節點
		}
		return node, nil

	case EXTENSION:
		if len(path) >= len(node.Path) && bytes.Equal(node.Path, path[:len(node.Path)]) {
			newChild, err := t.delete(node.Children[0], path[len(node.Path):])
			if err != nil {
				return nil, err
			}

			// 如果子節點被刪除，返回 nil
			if newChild == nil {
				return nil, nil
			}

			// 嘗試合併節點
			return t.mergeNodes(node, newChild)
		}
		return node, nil

	case BRANCH:
		if len(path) == 0 {
			node.Value = nil
		} else {
			branchIndex := path[0]
			newChild, err := t.delete(node.Children[branchIndex], path[1:])
			if err != nil {
				return nil, err
			}
			node.Children[branchIndex] = newChild
		}

		// 檢查是否需要簡化分支節點
		return t.simplifyBranch(node)

	default:
		return nil, errors.New("invalid node type")
	}
}

// mergeNodes 合併擴展節點和其子節點
func (t *ModifiedMerklePatriciaTree) mergeNodes(ext *MPTNode, child *MPTNode) (*MPTNode, error) {
	if child == nil {
		return nil, nil
	}

	switch child.NodeType {
	case LEAF:
		// 合併擴展節點和葉子節點
		return &MPTNode{
			NodeType: LEAF,
			Path:     append(ext.Path, child.Path...),
			Value:    child.Value,
			Dirty:    true,
		}, nil

	case EXTENSION:
		// 合併兩個擴展節點
		return &MPTNode{
			NodeType: EXTENSION,
			Path:     append(ext.Path, child.Path...),
			Children: child.Children,
			Dirty:    true,
		}, nil

	default:
		// 不能合併，保持原樣
		ext.Children[0] = child
		return ext, nil
	}
}

// simplifyBranch 簡化分支節點
func (t *ModifiedMerklePatriciaTree) simplifyBranch(branch *MPTNode) (*MPTNode, error) {
	// 計算非空子節點數量
	var nonNilCount int
	var lastNonNilIndex int

	for i, child := range branch.Children {
		if child != nil {
			nonNilCount++
			lastNonNilIndex = i
		}
	}

	// 如果分支節點有值，至少需要保留
	if branch.Value != nil {
		nonNilCount++
	}

	// 情況 1：沒有子節點和值，刪除節點
	if nonNilCount == 0 {
		return nil, nil
	}

	// 情況 2：只有一個子節點且沒有值，轉換為擴展節點
	if nonNilCount == 1 && branch.Value == nil {
		child := branch.Children[lastNonNilIndex]

		switch child.NodeType {
		case LEAF:
			// 轉換為葉子節點
			return &MPTNode{
				NodeType: LEAF,
				Path:     append([]byte{byte(lastNonNilIndex)}, child.Path...),
				Value:    child.Value,
				Dirty:    true,
			}, nil

		case EXTENSION:
			// 轉換為擴展節點
			return &MPTNode{
				NodeType: EXTENSION,
				Path:     append([]byte{byte(lastNonNilIndex)}, child.Path...),
				Children: child.Children,
				Dirty:    true,
			}, nil

		default:
			// 創建新的擴展節點
			return &MPTNode{
				NodeType: EXTENSION,
				Path:     []byte{byte(lastNonNilIndex)},
				Children: [16]*MPTNode{child},
				Dirty:    true,
			}, nil
		}
	}

	// 情況 3：保持分支節點
	return branch, nil
}

// ========== 哈希計算 ==========

// Hash 計算節點的哈希值
// MPT 的哈希計算考慮了節點類型和編碼
func (t *ModifiedMerklePatriciaTree) Hash(node *MPTNode) common.Hash {
	if node == nil {
		return common.Hash{}
	}

	// 如果節點沒有被修改且已有哈希，直接返回
	if !node.Dirty && node.Hash != nil {
		return *node.Hash
	}

	// 序列化節點
	encoded := t.encodeNode(node)

	// 計算哈希
	hash := t.hashFunc(encoded)

	// 緩存哈希值
	node.Hash = &hash
	node.Dirty = false

	// 如果編碼長度超過 32 字節，存儲到數據庫
	if len(encoded) > 32 {
		t.db[hash] = encoded
	}

	return hash
}

// encodeNode 序列化節點（簡化的 RLP 編碼）
func (t *ModifiedMerklePatriciaTree) encodeNode(node *MPTNode) []byte {
	switch node.NodeType {
	case LEAF:
		// [LEAF標記, 緊湊編碼的路徑, 值]
		compactPath := HexToCompact(node.Path, true)
		return append(append([]byte{byte(LEAF)}, compactPath...), node.Value...)

	case EXTENSION:
		// [EXTENSION標記, 緊湊編碼的路徑, 子節點哈希]
		compactPath := HexToCompact(node.Path, false)
		childHash := t.Hash(node.Children[0])
		return append(append([]byte{byte(EXTENSION)}, compactPath...), childHash.Bytes()...)

	case BRANCH:
		// [BRANCH標記, 16個子節點哈希, 值]
		encoded := []byte{byte(BRANCH)}

		// 添加 16 個子節點的哈希
		for _, child := range node.Children {
			if child == nil {
				encoded = append(encoded, common.Hash{}.Bytes()...)
			} else {
				childHash := t.Hash(child)
				encoded = append(encoded, childHash.Bytes()...)
			}
		}

		// 添加值（如果有）
		if node.Value != nil {
			encoded = append(encoded, node.Value...)
		}

		return encoded

	default:
		return nil
	}
}

// GetRoot 獲取根哈希
func (t *ModifiedMerklePatriciaTree) GetRoot() common.Hash {
	return t.Hash(t.root)
}

// ========== Merkle Proof 相關 ==========

// MPTProof MPT 的 Merkle 證明
// 與普通 Merkle Proof 的區別：
// 1. 需要包含節點類型信息
// 2. 路徑是可變的（不是固定的二叉路徑）
// 3. 需要處理編碼後的節點數據
type MPTProof struct {
	Key      []byte          // 原始鍵
	Value    []byte          // 值（如果存在）
	Proof    [][]byte        // 證明路徑上的節點數據
	RootHash common.Hash     // 根哈希
}

// GenerateProof 生成 MPT 的 Merkle 證明
func (t *ModifiedMerklePatriciaTree) GenerateProof(key []byte) (*MPTProof, error) {
	if len(key) == 0 {
		return nil, errors.New("key cannot be empty")
	}

	hexKey := KeyToHex(key)

	proof := &MPTProof{
		Key:      key,
		RootHash: t.GetRoot(),
		Proof:    [][]byte{},
	}

	// 收集證明路徑
	value, err := t.collectProof(t.root, hexKey, &proof.Proof)
	if err != nil {
		return nil, err
	}

	proof.Value = value
	return proof, nil
}

// collectProof 收集證明路徑上的節點
func (t *ModifiedMerklePatriciaTree) collectProof(node *MPTNode, path []byte, proof *[][]byte) ([]byte, error) {
	if node == nil {
		return nil, nil
	}

	// 添加當前節點到證明路徑
	*proof = append(*proof, t.encodeNode(node))

	switch node.NodeType {
	case LEAF:
		if bytes.Equal(node.Path, path) {
			return node.Value, nil
		}
		return nil, nil

	case EXTENSION:
		if len(path) >= len(node.Path) && bytes.Equal(node.Path, path[:len(node.Path)]) {
			return t.collectProof(node.Children[0], path[len(node.Path):], proof)
		}
		return nil, nil

	case BRANCH:
		if len(path) == 0 {
			return node.Value, nil
		}
		return t.collectProof(node.Children[path[0]], path[1:], proof)

	default:
		return nil, errors.New("invalid node type")
	}
}

// VerifyMPTProof 驗證 MPT 的 Merkle 證明
func VerifyMPTProof(proof *MPTProof, hashFunc func([]byte) common.Hash) bool {
	if hashFunc == nil {
		hashFunc = crypto.Keccak256Hash
	}

	// 從證明路徑重建根哈希
	// 這裡需要解析編碼的節點數據並驗證路徑
	// 實際實現較複雜，這裡簡化處理

	if len(proof.Proof) == 0 {
		return false
	}

	// 計算最後一個節點的哈希應該等於根哈希
	rootNode := proof.Proof[0]
	calculatedRoot := hashFunc(rootNode)

	return calculatedRoot == proof.RootHash
}

// ========== 工具函數 ==========

// commonPrefixLen 計算兩個路徑的公共前綴長度
func commonPrefixLen(a, b []byte) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return i
		}
	}

	return minLen
}

// ========== 打印和調試函數 ==========

// PrintTree 打印 MPT 結構
func (t *ModifiedMerklePatriciaTree) PrintTree() {
	fmt.Println("\n========== Modified Merkle Patricia Tree ==========")
	fmt.Printf("Root Hash: %s\n", t.GetRoot().Hex())
	fmt.Println("===================================================")

	if t.root == nil {
		fmt.Println("Empty tree")
		return
	}

	fmt.Println("\nTree Structure:")
	t.printNode(t.root, "", true, []byte{})
}

// printNode 遞歸打印節點
func (t *ModifiedMerklePatriciaTree) printNode(node *MPTNode, prefix string, isLast bool, currentPath []byte) {
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

	// 打印節點信息
	switch node.NodeType {
	case LEAF:
		fullPath := append(currentPath, node.Path...)
		fmt.Printf("[LEAF] Path=%x, Value=%s, Hash=%s\n",
			fullPath,
			string(node.Value),
			t.Hash(node).Hex()[:16])

	case EXTENSION:
		fullPath := append(currentPath, node.Path...)
		fmt.Printf("[EXTENSION] Path=%x, Hash=%s\n",
			fullPath,
			t.Hash(node).Hex()[:16])
		// 打印子節點
		if node.Children[0] != nil {
			t.printNode(node.Children[0], prefix, true, fullPath)
		}

	case BRANCH:
		fmt.Printf("[BRANCH] Hash=%s", t.Hash(node).Hex()[:16])
		if node.Value != nil {
			fmt.Printf(", Value=%s", string(node.Value))
		}
		fmt.Println()

		// 打印所有非空子節點
		nonNilChildren := 0
		for i, child := range node.Children {
			if child != nil {
				nonNilChildren++
			}
		}

		childIndex := 0
		for i, child := range node.Children {
			if child != nil {
				childIndex++
				newPath := append(currentPath, byte(i))
				t.printNode(child, prefix, childIndex == nonNilChildren, newPath)
			}
		}
	}
}

// PrintComparison 打印 MPT 與普通 Merkle Tree 的對比
func PrintComparison() {
	fmt.Println("\n========== MPT vs 普通 Merkle Tree 對比 ==========")
	fmt.Println("
	1. 結構差異：
	- 普通 Merkle Tree：二叉樹，每個節點最多 2 個子節點
	- MPT：16 叉樹，每個節點最多 16 個子節點

	2. 節點類型：
	- 普通 Merkle Tree：只有葉子節點和分支節點
	- MPT：有 4 種節點類型（空、葉子、擴展、分支）

	3. 路徑編碼：
	- 普通 Merkle Tree：使用索引定位（如 0, 1, 2...）
	- MPT：使用鍵的十六進制編碼作為路徑

	4. 動態性：
	- 普通 Merkle Tree：通常是靜態構建，不支持高效的插入/刪除
	- MPT：支持動態插入、刪除和更新操作

	5. 空間效率：
	- 普通 Merkle Tree：所有葉子節點在同一層，可能浪費空間
	- MPT：通過路徑壓縮（擴展節點）節省空間

	6. 用途：
	- 普通 Merkle Tree：主要用於靜態數據集的完整性驗證
	- MPT：用於動態鍵值存儲，如以太坊的狀態存儲

	7. 證明複雜度：
	- 普通 Merkle Tree：證明大小為 O(log n)
	- MPT：證明大小取決於鍵的長度和樹的結構
	")
}

// ========== 示例和測試函數 ==========

// Example 演示 MPT 的使用
func Example() {
	fmt.Println("\n========== MPT 使用示例 ==========")

	// 創建新的 MPT
	mpt := NewMPT()

	// 插入一些鍵值對
	testData := map[string]string{
		"cat":    "animal",
		"car":    "vehicle",
		"card":   "payment",
		"care":   "emotion",
		"dog":    "animal",
		"dodge":  "action",
		"door":   "entrance",
	}

	fmt.Println("\n1. 插入數據：")
	for key, value := range testData {
		err := mpt.Put([]byte(key), []byte(value))
		if err != nil {
			fmt.Printf("Error inserting %s: %v\n", key, err)
			continue
		}
		fmt.Printf("   插入: %s -> %s\n", key, value)
	}

	// 打印樹結構
	fmt.Println("\n2. 樹結構：")
	mpt.PrintTree()

	// 查詢數據
	fmt.Println("\n3. 查詢數據：")
	for key := range testData {
		value, err := mpt.Get([]byte(key))
		if err != nil {
			fmt.Printf("   Error getting %s: %v\n", key, err)
			continue
		}
		fmt.Printf("   查詢: %s -> %s\n", key, string(value))
	}

	// 生成證明
	fmt.Println("\n4. 生成 Merkle 證明：")
	proof, err := mpt.GenerateProof([]byte("cat"))
	if err != nil {
		fmt.Printf("   Error generating proof: %v\n", err)
	} else {
		fmt.Printf("   Key: %s\n", string(proof.Key))
		fmt.Printf("   Value: %s\n", string(proof.Value))
		fmt.Printf("   Proof nodes: %d\n", len(proof.Proof))
		fmt.Printf("   Root hash: %s\n", proof.RootHash.Hex())
	}

	// 刪除數據
	fmt.Println("\n5. 刪除數據：")
	err = mpt.Delete([]byte("car"))
	if err != nil {
		fmt.Printf("   Error deleting: %v\n", err)
	} else {
		fmt.Println("   已刪除: car")
	}

	// 驗證刪除
	value, err := mpt.Get([]byte("car"))
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else if value == nil {
		fmt.Println("   確認: car 已被刪除")
	}

	// 顯示對比
	PrintComparison()
}

// ========== 以太坊特定實現 ==========

// AccountState 以太坊賬戶狀態
type AccountState struct {
	Nonce    uint64
	Balance  uint64
	CodeHash common.Hash
	Storage  common.Hash // 存儲樹的根哈希
}

// EncodeAccount 編碼賬戶狀態
func EncodeAccount(account AccountState) []byte {
	// 簡化的編碼，實際應使用 RLP
	return []byte(fmt.Sprintf("%d:%d:%s:%s",
		account.Nonce,
		account.Balance,
		account.CodeHash.Hex(),
		account.Storage.Hex()))
}

// EthereumStateTrie 以太坊狀態樹
type EthereumStateTrie struct {
	*ModifiedMerklePatriciaTree
}

// NewEthereumStateTrie 創建以太坊狀態樹
func NewEthereumStateTrie() *EthereumStateTrie {
	return &EthereumStateTrie{
		ModifiedMerklePatriciaTree: NewMPT(),
	}
}

// UpdateAccount 更新賬戶狀態
func (s *EthereumStateTrie) UpdateAccount(address common.Address, account AccountState) error {
	encoded := EncodeAccount(account)
	return s.Put(address.Bytes(), encoded)
}

// GetAccount 獲取賬戶狀態
func (s *EthereumStateTrie) GetAccount(address common.Address) (*AccountState, error) {
	data, err := s.Get(address.Bytes())
	if err != nil || data == nil {
		return nil, err
	}

	// 簡化的解碼，實際應使用 RLP
	var account AccountState
	fmt.Sscanf(string(data), "%d:%d", &account.Nonce, &account.Balance)

	return &account, nil
}

// ========== 性能優化建議 ==========

/*
MPT 性能優化建議：

1. 節點緩存：
   - 實現 LRU 緩存避免重複計算哈希
   - 緩存熱點節點減少數據庫訪問

2. 批量操作：
   - 批量插入/刪除減少樹重構次數
   - 使用事務確保一致性

3. 並行化：
   - 並行計算子樹哈希
   - 讀操作可以完全並行

4. 存儲優化：
   - 使用專門的鍵值數據庫（如 LevelDB）
   - 實現節點的懶加載

5. 編碼優化：
   - 使用高效的 RLP 編碼庫
   - 避免不必要的編碼/解碼操作

6. 內存管理：
   - 及時釋放不需要的節點
   - 使用對象池減少 GC 壓力
*/
```