package trie

import (
	"fmt"
	"sort"
)

// ========== MPT Debug & Print Tools ==========

// PrintTree 打印整個 MPT 結構（改進版）
func (t *ModifiedMerklePatriciaTree) PrintTree() {
	t.GetRoot()
	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║         Modified Merkle Patricia Tree Structure        ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	if t.root == nil {
		fmt.Println("  [Empty Tree]")
		return
	}

	fmt.Println("\nRoot:")
	t.printNodeImproved(t.root, "", []byte{}, 0)
	fmt.Println()
}

// printNodeImproved 改進的節點打印方法
func (t *ModifiedMerklePatriciaTree) printNodeImproved(node *MPTNode, indent string, currentPath []byte, depth int) {
	if node == nil {
		return
	}

	// 根據節點類型打印
	switch node.NodeType {
	case LEAF:
		t.printLeafNodeImproved(node, indent, currentPath)

	case EXTENSION:
		t.printExtensionNodeImproved(node, indent, currentPath, depth)

	case BRANCH:
		t.printBranchNodeImproved(node, indent, currentPath, depth)

	default:
		fmt.Printf("%s[UNKNOWN NODE TYPE]\n", indent)
	}
}

// printLeafNodeImproved 改進的葉子節點打印
func (t *ModifiedMerklePatriciaTree) printLeafNodeImproved(node *MPTNode, indent string, currentPath []byte) {
	fullPath := append(currentPath, node.Path...)
	key := hexToKey(fullPath)

	fmt.Printf("%s", indent)
	fmt.Printf("🍃 LEAF")

	// 顯示路徑（如果有）
	if len(node.Path) > 0 {
		fmt.Printf(" [path: %s]", formatCompactHexPath(node.Path))
	}

	// 顯示完整的鍵和值
	fmt.Printf(" → '%s' = '%s'", string(key), string(node.Value))

	// 顯示哈希
	if node.Hash != nil {
		fmt.Printf(" %s", shortHash(node))
	}

	// 顯示狀態
	if node.Dirty {
		fmt.Printf(" 🔴")
	}
	fmt.Println()
}

// printExtensionNodeImproved 改進的擴展節點打印
func (t *ModifiedMerklePatriciaTree) printExtensionNodeImproved(node *MPTNode, indent string, currentPath []byte, depth int) {
	fullPath := append(currentPath, node.Path...)

	fmt.Printf("%s", indent)
	fmt.Printf("📐 EXTENSION [path: %s]", formatCompactHexPath(node.Path))

	// 顯示哈希
	if node.Hash != nil {
		fmt.Printf(" %s", shortHash(node))
	}

	if node.Dirty {
		fmt.Printf(" 🔴")
	}
	fmt.Println()

	// 打印子節點（Extension 只有一個子節點）
	if node.Children[0] != nil {
		newIndent := indent + "    "
		t.printNodeImproved(node.Children[0], newIndent, fullPath, depth+1)
	}
}

// printBranchNodeImproved 改進的分支節點打印
func (t *ModifiedMerklePatriciaTree) printBranchNodeImproved(node *MPTNode, indent string, currentPath []byte, depth int) {
	fmt.Printf("%s", indent)
	fmt.Printf("🌿 BRANCH")

	// 顯示分支節點的值（如果有）
	if node.Value != nil {
		key := hexToKey(currentPath)
		fmt.Printf(" [value: '%s' = '%s']", string(key), string(node.Value))
	}

	// 顯示哈希
	if node.Hash != nil {
		fmt.Printf(" %s", shortHash(node))
	}

	if node.Dirty {
		fmt.Printf(" 🔴")
	}

	// 計算非空子節點
	var nonNilChildren []int
	for i, child := range node.Children {
		if child != nil {
			nonNilChildren = append(nonNilChildren, i)
		}
	}

	fmt.Printf(" (%d children)", len(nonNilChildren))
	fmt.Println()

	// 打印所有非空子節點
	for _, i := range nonNilChildren {
		newPath := append(currentPath, byte(i))
		newIndent := indent + "    "

		// 打印分支索引
		fmt.Printf("%s[%X] → ", newIndent, i)

		// 在同一行開始打印子節點
		t.printNodeInline(node.Children[i], newIndent+"    ", newPath, depth+1)
	}
}

// printNodeInline 內聯打印節點（用於分支的子節點）
func (t *ModifiedMerklePatriciaTree) printNodeInline(node *MPTNode, indent string, currentPath []byte, depth int) {
	if node == nil {
		fmt.Println("nil")
		return
	}

	switch node.NodeType {
	case LEAF:
		fullPath := append(currentPath, node.Path...)
		key := hexToKey(fullPath)

		fmt.Printf("🍃 LEAF")
		if len(node.Path) > 0 {
			fmt.Printf(" [+%s]", formatCompactHexPath(node.Path))
		}
		fmt.Printf(" → '%s' = '%s'", string(key), string(node.Value))
		if node.Hash != nil {
			fmt.Printf(" %s", shortHash(node))
		}
		if node.Dirty {
			fmt.Printf(" 🔴")
		}
		fmt.Println()

	case EXTENSION:
		fmt.Printf("📐 EXTENSION [+%s]", formatCompactHexPath(node.Path))
		if node.Hash != nil {
			fmt.Printf(" %s", shortHash(node))
		}
		if node.Dirty {
			fmt.Printf(" 🔴")
		}
		fmt.Println()

		// Extension 的子節點需要換行打印
		if node.Children[0] != nil {
			fullPath := append(currentPath, node.Path...)
			t.printNodeImproved(node.Children[0], indent, fullPath, depth+1)
		}

	case BRANCH:
		// Branch 節點需要換行打印
		fmt.Println()
		t.printBranchNodeImproved(node, indent, currentPath, depth)

	default:
		fmt.Println("[UNKNOWN]")
	}
}

// formatCompactHexPath 更緊湊的十六進制路徑格式化
func formatCompactHexPath(path []byte) string {
	if len(path) == 0 {
		return "∅"
	}

	result := ""
	for _, b := range path {
		result += fmt.Sprintf("%x", b)
	}
	return result
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
func (t *ModifiedMerklePatriciaTree) mergeStats(target, source *MPTStats) {
	target.totalNodes += source.totalNodes
	target.leafNodes += source.leafNodes
	target.extensionNodes += source.extensionNodes
	target.branchNodes += source.branchNodes
	target.totalValues += source.totalValues
	target.dirtyNodes += source.dirtyNodes

	if source.maxDepth > target.maxDepth {
		target.maxDepth = source.maxDepth
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

	// 按鍵排序
	sort.Slice(pairs, func(i, j int) bool {
		return string(pairs[i].key) < string(pairs[j].key)
	})

	// 打印所有鍵值對
	for i, pair := range pairs {
		fmt.Printf("  %d. key='%s' => value='%s'\n", i+1, string(pair.key), string(pair.value))
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
}

// shortHash 返回簡短的哈希表示
func shortHash(node *MPTNode) string {
	if node == nil || node.Hash == nil {
		return "<nil>"
	}

	hashBytes := node.Hash.Bytes()
	if len(hashBytes) == 0 {
		return "<empty>"
	}

	hashStr := fmt.Sprintf("%x", hashBytes)

	if len(hashStr) <= 10 {
		return "0x" + hashStr
	}

	return fmt.Sprintf("<HASH: 0x%s...%s>", hashStr[:4], hashStr[len(hashStr)-4:])
}
