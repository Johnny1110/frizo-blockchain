# 🌲 Merkle Tree 核心原理

<br>

---

<br>

## Merkle Tree 的基本概念

結構組成：

* 葉節點（Leaf Nodes）：存放實際的資料項目（或其雜湊值）
* 內部節點（Internal Nodes）：存放子節點雜湊值的組合雜湊
* 根節點（Root Node）：整棵樹的頂端，稱為 Merkle Root

## 工作原理：

* 將所有資料項目進行雜湊運算，形成葉節點
* 將相鄰的兩個葉節點雜湊值合併後再次雜湊，形成父節點
* 重複此過程直到產生唯一的根節點
* __如果資料項目數量為奇數，最後一個節點會與自己合併__

## 主要優點：

* 只需要比較 Merkle Root 就能快速驗證整個資料集是否被篡改
* 可以在不下載完整資料的情況下驗證特定資料項目的存在性
* 驗證複雜度為 O(log n)，非常高效

<br>

## 區塊結構中的三棵 Merkle Tree

__每個以太坊區塊都包含三個 Merkle Root：__

### 交易樹（Transaction Tree）：

* 包含該區塊中的所有交易
* 交易按順序排列並構建 Merkle Tree
* 區塊標頭中的 transactionRoot 就是這棵樹的根

### 收據樹（Receipt Tree）：

* 包含每筆交易執行後的收據信息
* 收據包含執行狀態、Gas 使用量、事件日誌等
* 區塊標頭中的 receiptRoot 是收據樹的根

### 狀態樹（State Tree）：

* 最複雜的一棵，使用 Modified Merkle Patricia Tree
* 存儲所有帳戶的狀態信息（餘額、nonce、合約代碼等）
* 區塊標頭中的 stateRoot 是狀態樹的根

<br>
<br>

## 實際應用場景

### 輕節點驗證：

* 輕節點只需下載區塊標頭（約 500 bytes）
* 通過 Merkle Proof 可以驗證特定交易是否包含在區塊中
* 無需下載完整區塊（可能數 MB）

### 狀態證明：

* 可以證明某個帳戶在特定區塊的餘額
* 只需提供從該帳戶到狀態根的 Merkle Path
* 大大減少了需要傳輸的資料量

擴容解決方案：

Layer 2 解決方案（如 Optimistic Rollups）使用 Merkle Tree 來批量提交交易
將多筆交易打包成一個 Merkle Root 提交到主網，大幅降低成本


## Modified Merkle Patricia Tree

### 以太坊的狀態樹使用了一種特殊的變體：

* 結合了 Merkle Tree 的安全性和 Patricia Tree 的效率
* 支持高效的插入、刪除和查詢操作
* 每個節點的路徑對應帳戶地址或存儲位置
* 空的分支會被壓縮，節省存儲空間

<br>
<br>

## 細節

1. 結構

   ```
           Root
           /    \
        H12      H34
        /  \     /  \
       H1   H2  H3   H4
       /|   /|   /|   /|
      L1L2 L3L4 L5L6 L7L8
   ```

* 葉子節點：實際數據的哈希
* 內部節點：左右子節點哈希的組合哈希
* 根節點：整個數據集的唯一指紋

核心結構說明：

1. Node 結構

每個節點包含左右子節點的指針、自己的哈希值和原始數據
葉子節點有數據，非葉子節點只有哈希值

2. MerkleTree 結構

包含根節點、所有葉子節點的引用和哈希函數
可以自定義哈希函數，默認使用 Keccak-256

3. Merkle Tree 高度如何算：

* Merkle Tree 是每層都把資料對半合併（兩個 hash 合一）：
	* 所以每上一層節點數會變成原來的一半。
	* 這種「每次減半」的特性，就是 對數以 2 為底的情境。
	* `height = log2(leafCount)`

4. 核心功能

NewMerkleTree: 從數據創建 Merkle Tree
GetMerkleRoot: 獲取樹根哈希
GetMerkleProof: 生成某個葉子節點的證明路徑
VerifyProof: 驗證證明是否有效

<br>

2. 哈希計算
   ```go
   // 以太坊使用 Keccak256
   parent_hash = Keccak256(left_hash + right_hash)
   ```

    <br>   

   🔍 Merkle 證明原理
   證明組成

    目標葉子哈希
    兄弟節點哈希路徑
    方向信息（左/右）
    
    驗證過程
    ```go
    // 從葉子向根重新計算
    current_hash = leaf_hash
    for each sibling in proof_path:
    if direction == left:
    current_hash = hash(sibling + current_hash)
    else:
    current_hash = hash(current_hash + sibling)
   ```
   
<br>
<br>
<br>
<br>

## 🏛️ 以太坊 PoS 特定功能

<br>

1. 驗證者集合

    每個驗證者包含：地址、公鑰、質押金額、狀態
    驗證者集合的 Merkle Root 存儲在區塊頭中

<br>

2. 高效驗證

    無需下載整個驗證者集合
    通過 Merkle 證明驗證特定驗證者
    O(log n) 複雜度

<br>

3. 區塊頭結構

   ```go
   type BlockHeader struct {
       ValidatorsRoot   []byte  // 驗證者集合根哈希
       TransactionsRoot []byte  // 交易根哈希
       BlockNumber      uint64  // 區塊號
       Timestamp        uint64  // 時間戳
   }
   ```

<br>

🚀 主要優勢

* 效率：O(log n) 驗證複雜度
* 安全性：任何數據修改都會改變根哈希
* 可擴展性：適用於大型數據集
* 輕量級：證明大小與數據量對數成正比

💡 實際應用場景

* 驗證者集合驗證：證明某個驗證者在活躍集合中
* 交易包含證明：證明交易被包含在區塊中
* 狀態根驗證：驗證帳戶狀態
* 輕客戶端同步：無需下載完整區塊鏈

<br>

---

<br>

```go
package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"

	"golang.org/x/crypto/sha3"
)

// MerkleNode Merkle樹節點
type MerkleNode struct {
	Hash   []byte      // 節點哈希值
	Left   *MerkleNode // 左子節點
	Right  *MerkleNode // 右子節點
	Data   []byte      // 葉子節點數據（只有葉子節點有）
	Index  int         // 節點索引
	IsLeaf bool        // 是否為葉子節點
}

// MerkleTree Merkle樹結構
type MerkleTree struct {
	Root        *MerkleNode   // 根節點
	Leaves      []*MerkleNode // 所有葉子節點
	hashFunc    func([]byte) []byte // 哈希函數
	TreeHeight  int           // 樹高度
	LeafCount   int           // 葉子節點數量
}

// MerkleProof Merkle證明
type MerkleProof struct {
	LeafHash   []byte   // 葉子哈希
	LeafIndex  int      // 葉子索引
	Proof      [][]byte // 證明路徑
	Directions []bool   // 方向（true=右，false=左）
	RootHash   []byte   // 根哈希
}

// 以太坊標準哈希函數 (Keccak256)
func ethereumHash(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

// SHA256哈希函數（用於比較）
func sha256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// 創建新的Merkle樹
func NewMerkleTree(data [][]byte, useEthereumHash bool) *MerkleTree {
	fmt.Println("🌲 創建 Merkle Tree...")
	fmt.Printf("📘 原理：Merkle Tree 是一種二叉樹，葉子節點是數據，內部節點是子節點哈希的組合\n")
	fmt.Printf("📘 以太坊使用 Keccak256 哈希函數\n")
	fmt.Printf("📘 輸入數據量: %d 項\n", len(data))
	
	if len(data) == 0 {
		return &MerkleTree{}
	}
	
	// 選擇哈希函數
	var hashFunc func([]byte) []byte
	if useEthereumHash {
		hashFunc = ethereumHash
		fmt.Printf("📘 使用哈希函數: Keccak256 (以太坊標準)\n")
	} else {
		hashFunc = sha256Hash
		fmt.Printf("📘 使用哈希函數: SHA256\n")
	}
	
	tree := &MerkleTree{
		hashFunc:   hashFunc,
		LeafCount:  len(data),
		TreeHeight: int(math.Ceil(math.Log2(float64(len(data))))),
	}
	
	// 創建葉子節點
	leaves := make([]*MerkleNode, len(data))
	fmt.Printf("\n📋 創建葉子節點:\n")
	for i, datum := range data {
		hash := hashFunc(datum)
		leaves[i] = &MerkleNode{
			Hash:   hash,
			Data:   datum,
			Index:  i,
			IsLeaf: true,
		}
		fmt.Printf("  葉子 %d: %s -> %x\n", i, string(datum), hash)
	}
	
	tree.Leaves = leaves
	
	// 構建樹
	tree.Root = tree.buildTree(leaves)
	
	fmt.Printf("\n✅ Merkle Tree 構建完成\n")
	fmt.Printf("📊 樹高度: %d\n", tree.TreeHeight)
	fmt.Printf("📊 葉子數量: %d\n", tree.LeafCount)
	fmt.Printf("📊 根哈希: %x\n", tree.Root.Hash)
	
	return tree
}

// 構建樹的遞歸函數
func (mt *MerkleTree) buildTree(nodes []*MerkleNode) *MerkleNode {
	if len(nodes) == 1 {
		return nodes[0]
	}
	
	var parentNodes []*MerkleNode
	
	// 兩兩配對創建父節點
	for i := 0; i < len(nodes); i += 2 {
		left := nodes[i]
		var right *MerkleNode
		
		if i+1 < len(nodes) {
			right = nodes[i+1]
		} else {
			// 奇數個節點，複製最後一個節點
			right = &MerkleNode{
				Hash:   make([]byte, len(left.Hash)),
				IsLeaf: false,
			}
			copy(right.Hash, left.Hash)
		}
		
		// 創建父節點
		combinedHash := append(left.Hash, right.Hash...)
		parentHash := mt.hashFunc(combinedHash)
		
		parent := &MerkleNode{
			Hash:   parentHash,
			Left:   left,
			Right:  right,
			IsLeaf: false,
		}
		
		parentNodes = append(parentNodes, parent)
	}
	
	return mt.buildTree(parentNodes)
}

// 獲取根哈希
func (mt *MerkleTree) GetRootHash() []byte {
	if mt.Root == nil {
		return nil
	}
	return mt.Root.Hash
}

// 生成 Merkle 證明
func (mt *MerkleTree) GenerateProof(leafIndex int) (*MerkleProof, error) {
	fmt.Printf("\n🔍 生成 Merkle 證明 (葉子索引: %d)...\n", leafIndex)
	fmt.Printf("📘 原理：Merkle 證明包含從葉子到根的路徑上所有兄弟節點的哈希\n")
	fmt.Printf("📘 驗證者可以使用這些哈希重新計算根哈希來驗證數據\n")
	
	if leafIndex < 0 || leafIndex >= len(mt.Leaves) {
		return nil, fmt.Errorf("葉子索引超出範圍: %d", leafIndex)
	}
	
	leaf := mt.Leaves[leafIndex]
	proof := &MerkleProof{
		LeafHash:   leaf.Hash,
		LeafIndex:  leafIndex,
		RootHash:   mt.Root.Hash,
		Proof:      [][]byte{},
		Directions: []bool{},
	}
	
	// 從葉子向根遍歷，收集兄弟節點
	currentIndex := leafIndex
	nodes := mt.Leaves
	
	fmt.Printf("📋 證明路徑:\n")
	level := 0
	
	for len(nodes) > 1 {
		var nextLevelNodes []*MerkleNode
		
		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			var right *MerkleNode
			
			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				// 處理奇數個節點
				right = &MerkleNode{Hash: make([]byte, len(left.Hash))}
				copy(right.Hash, left.Hash)
			}
			
			// 檢查當前節點是否在目標路徑上
			if currentIndex == i {
				// 當前節點是左子節點，添加右兄弟
				proof.Proof = append(proof.Proof, right.Hash)
				proof.Directions = append(proof.Directions, true) // 兄弟在右邊
				fmt.Printf("  層 %d: 左節點 %x，右兄弟 %x\n", level, left.Hash, right.Hash)
				currentIndex = i / 2
			} else if currentIndex == i+1 {
				// 當前節點是右子節點，添加左兄弟
				proof.Proof = append(proof.Proof, left.Hash)
				proof.Directions = append(proof.Directions, false) // 兄弟在左邊
				fmt.Printf("  層 %d: 右節點 %x，左兄弟 %x\n", level, right.Hash, left.Hash)
				currentIndex = i / 2
			}
			
			// 創建父節點
			combinedHash := append(left.Hash, right.Hash...)
			parentHash := mt.hashFunc(combinedHash)
			parent := &MerkleNode{Hash: parentHash}
			nextLevelNodes = append(nextLevelNodes, parent)
		}
		
		nodes = nextLevelNodes
		level++
	}
	
	fmt.Printf("✅ 證明生成完成，包含 %d 個兄弟節點\n", len(proof.Proof))
	return proof, nil
}

// 驗證 Merkle 證明
func (mt *MerkleTree) VerifyProof(proof *MerkleProof) bool {
	fmt.Printf("\n✅ 驗證 Merkle 證明...\n")
	fmt.Printf("📘 原理：使用證明路徑重新計算根哈希，與已知根哈希比較\n")
	
	if proof == nil {
		fmt.Printf("❌ 證明為空\n")
		return false
	}
	
	currentHash := proof.LeafHash
	fmt.Printf("📋 驗證步驟:\n")
	fmt.Printf("  起始葉子哈希: %x\n", currentHash)
	
	// 沿著證明路徑向上計算
	for i, siblingHash := range proof.Proof {
		var combinedHash []byte
		
		if proof.Directions[i] {
			// 兄弟在右邊
			combinedHash = append(currentHash, siblingHash...)
			fmt.Printf("  步驟 %d: %x + %x (兄弟在右)\n", i+1, currentHash, siblingHash)
		} else {
			// 兄弟在左邊
			combinedHash = append(siblingHash, currentHash...)
			fmt.Printf("  步驟 %d: %x + %x (兄弟在左)\n", i+1, siblingHash, currentHash)
		}
		
		currentHash = mt.hashFunc(combinedHash)
		fmt.Printf("    結果: %x\n", currentHash)
	}
	
	// 比較計算得到的根哈希與原始根哈希
	isValid := compareHashes(currentHash, proof.RootHash)
	fmt.Printf("  計算根哈希: %x\n", currentHash)
	fmt.Printf("  期望根哈希: %x\n", proof.RootHash)
	fmt.Printf("  驗證結果: %t\n", isValid)
	
	return isValid
}

// 比較兩個哈希
func compareHashes(hash1, hash2 []byte) bool {
	if len(hash1) != len(hash2) {
		return false
	}
	for i := 0; i < len(hash1); i++ {
		if hash1[i] != hash2[i] {
			return false
		}
	}
	return true
}

// 獲取葉子節點數據
func (mt *MerkleTree) GetLeafData(index int) ([]byte, error) {
	if index < 0 || index >= len(mt.Leaves) {
		return nil, fmt.Errorf("索引超出範圍: %d", index)
	}
	return mt.Leaves[index].Data, nil
}

// 檢查數據是否存在（生成證明並驗證）
func (mt *MerkleTree) ContainsData(data []byte) (bool, int, error) {
	dataHash := mt.hashFunc(data)
	
	// 查找匹配的葉子節點
	for i, leaf := range mt.Leaves {
		if compareHashes(leaf.Hash, dataHash) {
			// 生成並驗證證明
			proof, err := mt.GenerateProof(i)
			if err != nil {
				return false, -1, err
			}
			
			isValid := mt.VerifyProof(proof)
			return isValid, i, nil
		}
	}
	
	return false, -1, nil
}

// 打印樹結構
func (mt *MerkleTree) PrintTree() {
	fmt.Printf("\n🌲 Merkle Tree 結構:\n")
	if mt.Root == nil {
		fmt.Printf("空樹\n")
		return
	}
	
	mt.printNode(mt.Root, 0, "Root")
}

// 遞歸打印節點
func (mt *MerkleTree) printNode(node *MerkleNode, level int, prefix string) {
	if node == nil {
		return
	}
	
	indent := strings.Repeat("  ", level)
	
	if node.IsLeaf {
		fmt.Printf("%s%s: %x (數據: %s)\n", indent, prefix, node.Hash, string(node.Data))
	} else {
		fmt.Printf("%s%s: %x\n", indent, prefix, node.Hash)
		if node.Left != nil {
			mt.printNode(node.Left, level+1, "L")
		}
		if node.Right != nil {
			mt.printNode(node.Right, level+1, "R")
		}
	}
}

// 以太坊 PoS 特定功能

// ValidatorMerkleTree 驗證者 Merkle Tree
type ValidatorMerkleTree struct {
	*MerkleTree
	Validators []ValidatorInfo
}

// ValidatorInfo 驗證者信息
type ValidatorInfo struct {
	Address    string  // 驗證者地址
	PublicKey  string  // 公鑰
	Stake      uint64  // 質押金額
	Index      int     // 驗證者索引
	IsActive   bool    // 是否活躍
}

// 創建驗證者 Merkle Tree
func NewValidatorMerkleTree(validators []ValidatorInfo) *ValidatorMerkleTree {
	fmt.Printf("\n🏛️ 創建驗證者 Merkle Tree...\n")
	fmt.Printf("📘 以太坊 PoS 使用 Merkle Tree 來高效驗證驗證者集合\n")
	fmt.Printf("📘 驗證者數量: %d\n", len(validators))
	
	// 將驗證者信息序列化為字節數組
	var data [][]byte
	for _, validator := range validators {
		validatorData := fmt.Sprintf("%s:%s:%d:%t", 
			validator.Address, 
			validator.PublicKey, 
			validator.Stake, 
			validator.IsActive)
		data = append(data, []byte(validatorData))
	}
	
	// 創建 Merkle Tree
	merkleTree := NewMerkleTree(data, true) // 使用以太坊哈希
	
	return &ValidatorMerkleTree{
		MerkleTree: merkleTree,
		Validators: validators,
	}
}

// 驗證驗證者是否在集合中
func (vmt *ValidatorMerkleTree) VerifyValidator(validator ValidatorInfo) (bool, error) {
	fmt.Printf("\n🔍 驗證驗證者: %s\n", validator.Address)
	
	validatorData := fmt.Sprintf("%s:%s:%d:%t", 
		validator.Address, 
		validator.PublicKey, 
		validator.Stake, 
		validator.IsActive)
	
	contains, index, err := vmt.ContainsData([]byte(validatorData))
	if err != nil {
		return false, err
	}
	
	if contains {
		fmt.Printf("✅ 驗證者在集合中，索引: %d\n", index)
	} else {
		fmt.Printf("❌ 驗證者不在集合中\n")
	}
	
	return contains, nil
}

// 生成驗證者證明
func (vmt *ValidatorMerkleTree) GenerateValidatorProof(validatorIndex int) (*MerkleProof, error) {
	if validatorIndex < 0 || validatorIndex >= len(vmt.Validators) {
		return nil, fmt.Errorf("驗證者索引超出範圍: %d", validatorIndex)
	}
	
	fmt.Printf("\n📋 生成驗證者證明: %s\n", vmt.Validators[validatorIndex].Address)
	return vmt.GenerateProof(validatorIndex)
}

// 模擬以太坊區塊頭中的 Merkle Root
type BlockHeader struct {
	ValidatorsRoot []byte    // 驗證者集合根哈希
	TransactionsRoot []byte  // 交易根哈希
	BlockNumber    uint64    // 區塊號
	Timestamp      uint64    // 時間戳
}

// 創建區塊頭
func CreateBlockHeader(validators []ValidatorInfo, transactions [][]byte) *BlockHeader {
	fmt.Printf("\n📦 創建區塊頭...\n")
	
	// 創建驗證者 Merkle Tree
	validatorTree := NewValidatorMerkleTree(validators)
	
	// 創建交易 Merkle Tree
	transactionTree := NewMerkleTree(transactions, true)
	
	return &BlockHeader{
		ValidatorsRoot:   validatorTree.GetRootHash(),
		TransactionsRoot: transactionTree.GetRootHash(),
		BlockNumber:      12345,
		Timestamp:        1640995200, // 示例時間戳
	}
}

// 主函數示例
func main() {
	fmt.Println("🚀 以太坊 PoS Merkle Tree 完整示例")
	fmt.Println("=" + strings.Repeat("=", 50))
	
	// 1. 基本 Merkle Tree 示例
	fmt.Println("\n📋 基本 Merkle Tree 示例:")
	data := [][]byte{
		[]byte("transaction1"),
		[]byte("transaction2"),
		[]byte("transaction3"),
		[]byte("transaction4"),
	}
	
	tree := NewMerkleTree(data, true)
	tree.PrintTree()
	
	// 2. 生成和驗證證明
	fmt.Println("\n🔍 Merkle 證明示例:")
	proof, err := tree.GenerateProof(1)
	if err != nil {
		log.Fatal(err)
	}
	
	isValid := tree.VerifyProof(proof)
	fmt.Printf("證明驗證結果: %t\n", isValid)
	
	// 3. 驗證者 Merkle Tree 示例
	fmt.Println("\n🏛️ 驗證者 Merkle Tree 示例:")
	validators := []ValidatorInfo{
		{Address: "0x1234", PublicKey: "pubkey1", Stake: 32000000000000000000, IsActive: true},
		{Address: "0x5678", PublicKey: "pubkey2", Stake: 32000000000000000000, IsActive: true},
		{Address: "0x9abc", PublicKey: "pubkey3", Stake: 64000000000000000000, IsActive: true},
		{Address: "0xdef0", PublicKey: "pubkey4", Stake: 32000000000000000000, IsActive: false},
	}
	
	validatorTree := NewValidatorMerkleTree(validators)
	
	// 驗證特定驗證者
	testValidator := ValidatorInfo{
		Address: "0x5678", 
		PublicKey: "pubkey2", 
		Stake: 32000000000000000000, 
		IsActive: true,
	}
	
	isValidValidator, err := validatorTree.VerifyValidator(testValidator)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("驗證者驗證結果: %t\n", isValidValidator)
	
	// 4. 生成驗證者證明
	validatorProof, err := validatorTree.GenerateValidatorProof(1)
	if err != nil {
		log.Fatal(err)
	}
	
	isValidProof := validatorTree.VerifyProof(validatorProof)
	fmt.Printf("驗證者證明驗證結果: %t\n", isValidProof)
	
	// 5. 創建區塊頭
	fmt.Println("\n📦 區塊頭示例:")
	transactions := [][]byte{
		[]byte("tx1: Alice -> Bob 1 ETH"),
		[]byte("tx2: Bob -> Charlie 0.5 ETH"),
		[]byte("tx3: Charlie -> Dave 2 ETH"),
	}
	
	blockHeader := CreateBlockHeader(validators, transactions)
	fmt.Printf("區塊頭:\n")
	fmt.Printf("  驗證者根哈希: %x\n", blockHeader.ValidatorsRoot)
	fmt.Printf("  交易根哈希: %x\n", blockHeader.TransactionsRoot)
	fmt.Printf("  區塊號: %d\n", blockHeader.BlockNumber)
	fmt.Printf("  時間戳: %d\n", blockHeader.Timestamp)
	
	// 6. 性能測試
	fmt.Println("\n⚡ 性能測試:")
	performanceTest()
}

// 性能測試
func performanceTest() {
	sizes := []int{100, 1000, 10000}
	
	for _, size := range sizes {
		fmt.Printf("\n測試大小: %d 個數據項\n", size)
		
		// 生成測試數據
		var testData [][]byte
		for i := 0; i < size; i++ {
			testData = append(testData, []byte(fmt.Sprintf("data_%d", i)))
		}
		
		// 創建樹
		tree := NewMerkleTree(testData, true)
		
		// 測試證明生成
		proof, err := tree.GenerateProof(size / 2)
		if err != nil {
			log.Fatal(err)
		}
		
		// 測試證明驗證
		isValid := tree.VerifyProof(proof)
		
		fmt.Printf("  樹高度: %d\n", tree.TreeHeight)
		fmt.Printf("  證明長度: %d\n", len(proof.Proof))
		fmt.Printf("  驗證結果: %t\n", isValid)
	}
}

// 工具函數：十六進制字符串轉換
func HashToHex(hash []byte) string {
	return hex.EncodeToString(hash)
}

// 工具函數：從十六進制字符串創建哈希
func HexToHash(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

// 工具函數：比較兩個 Merkle Tree 的根哈希
func CompareTrees(tree1, tree2 *MerkleTree) bool {
	if tree1.Root == nil || tree2.Root == nil {
		return tree1.Root == tree2.Root
	}
	return compareHashes(tree1.Root.Hash, tree2.Root.Hash)
}
```