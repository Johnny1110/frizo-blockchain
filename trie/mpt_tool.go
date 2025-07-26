package trie

import (
	"fmt"
	"frizo-blockchain/common"
	"sort"
)

// ========== MPT Debug & Print Tools ==========

// PrintTree 打印整個 MPT 結構
func (t *ModifiedMerklePatriciaTree) PrintTree() {
	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║         Modified Merkle Patricia Tree Structure        ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	if t.root == nil {
		fmt.Println("  [Empty Tree]")
		return
	}

	fmt.Println("\nRoot:")
	t.printNode(t.root, "", true, []byte{})
	fmt.Println()
}

// printNode 遞歸打印節點
func (t *ModifiedMerklePatriciaTree) printNode(node *MPTNode, prefix string, isLast bool, currentPath []byte) {
	if node == nil {
		return
	}

	// 打印連接線
	fmt.Print(prefix)
	if isLast {
		fmt.Print("└── ")
		prefix += "    "
	} else {
		fmt.Print("├── ")
		prefix += "│   "
	}

	// 根據節點類型打印
	switch node.NodeType {
	case LEAF:
		t.printLeafNode(node, currentPath)

	case EXTENSION:
		t.printExtensionNode(node, prefix, currentPath)

	case BRANCH:
		t.printBranchNode(node, prefix, currentPath)

	default:
		fmt.Printf("[UNKNOWN NODE TYPE]\n")
	}
}

// printLeafNode 打印葉子節點
func (t *ModifiedMerklePatriciaTree) printLeafNode(node *MPTNode, currentPath []byte) {
	fullPath := append(currentPath, node.Path...)

	// 將十六進制路徑轉換為可讀字符串
	key := hexToKey(fullPath)

	fmt.Printf("[LEAF] ")
	fmt.Printf("path=%s ", formatHexPath(node.Path))
	fmt.Printf("key='%s' ", string(key))
	fmt.Printf("value='%s'", string(node.Value))

	if node.Hash != nil {
		fmt.Printf(" hash=%s", node.Hash.Hex()[:10]+"...")
	}
	if node.Dirty {
		fmt.Printf(" *dirty*")
	}
	fmt.Println()
}

// printExtensionNode 打印擴展節點
func (t *ModifiedMerklePatriciaTree) printExtensionNode(node *MPTNode, prefix string, currentPath []byte) {
	fullPath := append(currentPath, node.Path...)

	fmt.Printf("[EXTENSION] ")
	fmt.Printf("path=%s ", formatHexPath(node.Path))

	if node.Hash != nil {
		fmt.Printf("hash=%s ", node.Hash.Hex()[:10]+"...")
	}
	if node.Dirty {
		fmt.Printf("*dirty*")
	}
	fmt.Println()

	// 打印子節點
	if node.Children[0] != nil {
		t.printNode(node.Children[0], prefix, true, fullPath)
	}
}

// printBranchNode 打印分支節點
func (t *ModifiedMerklePatriciaTree) printBranchNode(node *MPTNode, prefix string, currentPath []byte) {
	fmt.Printf("[BRANCH]")

	if node.Value != nil {
		fmt.Printf(" value='%s'", string(node.Value))
	}

	if node.Hash != nil {
		fmt.Printf(" hash=%s", node.Hash.Hex()[:10]+"...")
	}
	if node.Dirty {
		fmt.Printf(" *dirty*")
	}
	fmt.Println()

	// 計算非空子節點
	var nonNilChildren []int
	for i, child := range node.Children {
		if child != nil {
			nonNilChildren = append(nonNilChildren, i)
		}
	}

	// 打印所有非空子節點
	for idx, i := range nonNilChildren {
		isLastChild := idx == len(nonNilChildren)-1
		newPath := append(currentPath, byte(i))

		fmt.Print(prefix)
		if isLastChild {
			fmt.Printf("└─[%X]─", i)
		} else {
			fmt.Printf("├─[%X]─", i)
		}

		t.printNode(node.Children[i], prefix, isLastChild, newPath)
	}
}

// ========== Helper Functions ==========

// formatHexPath 格式化十六進制路徑為可讀字符串
func formatHexPath(path []byte) string {
	if len(path) == 0 {
		return "[]"
	}

	result := "["
	for i, b := range path {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%X", b)
	}
	result += "]"
	return result
}

// hexToKey 將十六進制路徑轉換回原始鍵
func hexToKey(hexPath []byte) []byte {
	if len(hexPath)%2 != 0 {
		// 處理奇數長度（不應該發生在正常情況下）
		return []byte{}
	}

	key := make([]byte, len(hexPath)/2)
	for i := 0; i < len(hexPath); i += 2 {
		key[i/2] = hexPath[i]<<4 | hexPath[i+1]
	}
	return key
}

// PrintStats 打印 MPT 統計信息
func (t *ModifiedMerklePatriciaTree) PrintStats() {
	stats := t.collectStats(t.root, 0)

	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║                   MPT Statistics                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Printf("  Total Nodes:      %d\n", stats.totalNodes)
	fmt.Printf("  Leaf Nodes:       %d\n", stats.leafNodes)
	fmt.Printf("  Extension Nodes:  %d\n", stats.extensionNodes)
	fmt.Printf("  Branch Nodes:     %d\n", stats.branchNodes)
	fmt.Printf("  Max Depth:        %d\n", stats.maxDepth)
	fmt.Printf("  Total Values:     %d\n", stats.totalValues)
	fmt.Printf("  Dirty Nodes:      %d\n", stats.dirtyNodes)
	fmt.Println()
}

// MPTStats 統計信息結構
type MPTStats struct {
	totalNodes     int
	leafNodes      int
	extensionNodes int
	branchNodes    int
	maxDepth       int
	totalValues    int
	dirtyNodes     int
}

// collectStats 收集統計信息
func (t *ModifiedMerklePatriciaTree) collectStats(node *MPTNode, depth int) *MPTStats {
	if node == nil {
		return &MPTStats{}
	}

	stats := &MPTStats{
		totalNodes: 1,
		maxDepth:   depth,
	}

	if node.Dirty {
		stats.dirtyNodes = 1
	}

	switch node.NodeType {
	case LEAF:
		stats.leafNodes = 1
		stats.totalValues = 1

	case EXTENSION:
		stats.extensionNodes = 1
		childStats := t.collectStats(node.Children[0], depth+1)
		t.mergeStats(stats, childStats)

	case BRANCH:
		stats.branchNodes = 1
		if node.Value != nil {
			stats.totalValues = 1
		}

		for _, child := range node.Children {
			if child != nil {
				childStats := t.collectStats(child, depth+1)
				t.mergeStats(stats, childStats)
			}
		}
	}

	return stats
}

// mergeStats 合併統計信息
func (t *ModifiedMerklePatriciaTree) mergeStats(stats, childStats *MPTStats) {
	stats.totalNodes += childStats.totalNodes
	stats.leafNodes += childStats.leafNodes
	stats.extensionNodes += childStats.extensionNodes
	stats.branchNodes += childStats.branchNodes
	stats.totalValues += childStats.totalValues
	stats.dirtyNodes += childStats.dirtyNodes

	if childStats.maxDepth > stats.maxDepth {
		stats.maxDepth = childStats.maxDepth
	}
}

// PrintAllKeys 打印所有鍵值對
func (t *ModifiedMerklePatriciaTree) PrintAllKeys() {
	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║                  All Key-Value Pairs                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	if t.root == nil {
		fmt.Println("  [Empty Tree]")
		return
	}

	pairs := t.collectAllPairs(t.root, []byte{})

	// 排序鍵以便更好地顯示
	sort.Slice(pairs, func(i, j int) bool {
		return string(pairs[i].key) < string(pairs[j].key)
	})

	for i, pair := range pairs {
		fmt.Printf("  %3d. key='%s' => value='%s'\n", i+1, string(pair.key), string(pair.value))
	}
	fmt.Printf("\n  Total: %d entries\n", len(pairs))
}

// keyValuePair 鍵值對結構
type keyValuePair struct {
	key   []byte
	value []byte
}

// collectAllPairs 收集所有鍵值對
func (t *ModifiedMerklePatriciaTree) collectAllPairs(node *MPTNode, currentPath []byte) []keyValuePair {
	if node == nil {
		return nil
	}

	var pairs []keyValuePair

	switch node.NodeType {
	case LEAF:
		fullPath := append(currentPath, node.Path...)
		key := hexToKey(fullPath)
		pairs = append(pairs, keyValuePair{key: key, value: node.Value})

	case EXTENSION:
		fullPath := append(currentPath, node.Path...)
		pairs = append(pairs, t.collectAllPairs(node.Children[0], fullPath)...)

	case BRANCH:
		// Branch 節點可能有值
		if node.Value != nil {
			key := hexToKey(currentPath)
			pairs = append(pairs, keyValuePair{key: key, value: node.Value})
		}

		// 遍歷所有子節點
		for i, child := range node.Children {
			if child != nil {
				newPath := append(currentPath, byte(i))
				pairs = append(pairs, t.collectAllPairs(child, newPath)...)
			}
		}
	}

	return pairs
}

// Size returns the number of key-value pairs in the MPT
func (t *ModifiedMerklePatriciaTree) Size() int {
	return t.countValues(t.root)
}

// countValues recursively counts the number of values in the tree
func (t *ModifiedMerklePatriciaTree) countValues(node *MPTNode) int {
	if node == nil {
		return 0
	}

	count := 0

	switch node.NodeType {
	case LEAF:
		return 1

	case EXTENSION:
		return t.countValues(node.Children[0])

	case BRANCH:
		if node.Value != nil {
			count = 1
		}
		for _, child := range node.Children {
			if child != nil {
				count += t.countValues(child)
			}
		}
		return count

	default:
		return 0
	}
}

// Clear removes all key-value pairs from the MPT
func (t *ModifiedMerklePatriciaTree) Clear() {
	t.root = nil
	t.db = make(map[common.Hash][]byte)
}
