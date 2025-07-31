# Frizo Blockchain Phase-1 開發計劃

## 一、MPT 完善計劃（1-2 週）

### 1. 實現哈希計算
```go
// 在 crypto/mpt/mpt.go 中添加
func (n *MPTNode) Hash() common.Hash {
// 實現節點的 RLP 編碼和哈希計算
}
```

### 2. 實現 RLP 編碼
```go
// 新建 crypto/rlp/encode.go
func Encode(val interface{}) ([]byte, error) {
// 實現 RLP 編碼邏輯
}
```

### 3. 實現 Merkle Proof
```go
func (t *MPT) GetProof(key []byte) ([][]byte, error) {
// 實現 Merkle proof 生成
}

func VerifyProof(root common.Hash, key []byte, proof [][]byte) ([]byte, error) {
// 實現 proof 驗證
}
```

## 二、區塊鏈核心實現（2-3 週）

### 1. 創建 core/blockchain/blockchain.go
```go
type BlockChain struct {
db          storage.Database
genesisBlock *types.Block
currentBlock *types.Block

mu          sync.RWMutex
blocks      map[common.Hash]*types.Block
blockIndex  map[uint64]common.Hash
}

// 核心方法
func (bc *BlockChain) InsertBlock(block *types.Block) error
func (bc *BlockChain) GetBlock(hash common.Hash) *types.Block
func (bc *BlockChain) GetBlockByNumber(number uint64) *types.Block
func (bc *BlockChain) ValidateBlock(block *types.Block) error
```

### 2. 實現區塊驗證邏輯
- 驗證區塊頭（時間戳、父區塊哈希等）
- 驗證交易列表
- 驗證默克爾根
- 驗證區塊哈希

### 3. 實現狀態管理
```go
// core/state/statedb.go
type StateDB struct {
trie *crypto.MPT
db   storage.Database

accounts map[common.Address]*Account
}
```

## 三、存儲層實現（1 週）

### 1. 定義資料庫接口
```go
// storage/database.go
type Database interface {
Put(key []byte, value []byte) error
Get(key []byte) ([]byte, error)
Delete(key []byte) error
NewBatch() Batch
Close() error
}
```

### 2. 實現 LevelDB 包裝器
```go
// storage/leveldb/leveldb.go
type LevelDB struct {
db *leveldb.DB
}

func NewLevelDB(path string) (*LevelDB, error)
```

## 四、網路層基礎（2 週）

### 1. 定義消息類型
```go
// network/p2p/message.go
const (
MsgTypeHandshake = iota
MsgTypeBlock
MsgTypeTransaction
MsgTypeGetBlocks
)
```

### 2. 實現節點管理
```go
// network/p2p/peer.go
type Peer struct {
id       string
conn     net.Conn
version  string
}
```

### 3. 實現基本的同步協議
- 握手協議
- 區塊請求/響應
- 交易廣播

## 五、測試計劃

### 1. 單元測試（每個模組）
- MPT 操作測試
- 區塊驗證測試
- 存儲層測試
- 網路消息測試

### 2. 整合測試
- 創建測試鏈
- 多節點同步測試
- 狀態一致性測試

## 六、里程碑檢查點

### 第一個檢查點（2 週）
- [ ] MPT 完整實現
- [ ] 基本的區塊結構和驗證
- [ ] LevelDB 存儲層

### 第二個檢查點（4 週）
- [ ] 區塊鏈核心功能完成
- [ ] 狀態管理基本實現
- [ ] 單節點可以創建和存儲區塊

### 第三個檢查點（6 週）
- [ ] P2P 網路基礎完成
- [ ] 兩節點可以同步區塊
- [ ] 測試覆蓋率達到 80%

## 七、開發小貼士

1. **先寫測試**：採用 TDD 方式，先寫測試再實現功能
2. **漸進式開發**：從最簡單的情況開始，逐步增加複雜度
3. **參考 go-ethereum**：遇到困難時參考官方實現，但要理解原理
4. **持續重構**：保持代碼整潔，及時重構
5. **文檔同步**：實現功能的同時更新文檔

## 八、下一步具體行動

1. **今天**：完成 MPT 的 Hash() 方法實現
2. **本週**：完成 MPT 的所有核心功能並通過測試
3. **下週**：開始實現區塊鏈核心邏輯