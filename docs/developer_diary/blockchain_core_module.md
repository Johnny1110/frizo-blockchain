# Blockchain 模組

<br>

---

<br>

## 一、架構設計

```
core/blockchain/
├── interfaces.go           # 核心介面定義
├── blockchain.go           # 主鏈管理（精簡）
├── chain_manager.go        # 鏈狀態管理
├── block_validator.go      # 區塊驗證邏輯
├── state_processor.go      # 狀態轉換處理 ✅
├── block_processor.go      # 區塊處理流程
├── tx_processor.go         # 交易處理邏輯
├── chain_indexer.go        # 鏈索引管理
├── chain_maker.go          # 區塊生產
├── genesis.go              # 創世區塊 ✅
├── events.go               # 事件系統
├── metrics.go              # 性能指標
│
├── rawdb/                  # 底層數據庫操作
│   ├── database.go         # 數據庫介面
│   ├── accessors_chain.go  # 鏈數據存取
│   ├── accessors_indexes.go# 索引存取
│   ├── accessors_state.go  # 狀態存取
│   ├── schema.go           # 數據庫模式
│   └── ancient.go          # 歷史數據歸檔
│
├── txpool/                 # 交易池（獨立子模組）
│   ├── txpool.go           # 交易池主體
│   ├── txlist.go           # 交易列表管理
│   ├── txpricedlist.go     # 價格排序列表
│   ├── txnoncer.go         # Nonce 管理
│   └── txjournal.go        # 交易日誌
│
├── bloombits/              # 布隆過濾器（用於日誌檢索）
│   ├── bloombits.go
│   └── matcher.go
│
└── tests/                  # 測試套件
    ├── blockchain_test.go
    ├── chain_makers_test.go
    └── mock_consensus.go
```

<br>
<br>
<br>
<br>

## 二、核心介面設計

### 2.1 主要介面定義

```go
// interfaces.go
package blockchain

import (
    "frizo-blockchain/common"
    "frizo-blockchain/core/types"
    "frizo-blockchain/core/state"
    "frizo-blockchain/event"
    "math/big"
)

// ChainReader 定義了讀取區塊鏈的介面
type ChainReader interface {
    // Config retrieves the chain's configuration
    Config() *ChainConfig
    
    // Current block
    CurrentBlock() *types.Block
    CurrentHeader() *types.Header
    CurrentState() state.StateDB
    
    // Block retrieval
    GetBlock(hash common.Hash, number uint64) *types.Block
    GetBlockByHash(hash common.Hash) *types.Block
    GetBlockByNumber(number uint64) *types.Block
    GetHeader(hash common.Hash, number uint64) *types.Header
    GetHeaderByHash(hash common.Hash) *types.Header
    GetHeaderByNumber(number uint64) *types.Header
    
    // Transaction retrieval
    GetTransaction(hash common.Hash) (*types.Transaction, common.Hash, uint64, uint64)
    GetReceipt(hash common.Hash) *types.Receipt
    
    // Canonical chain
    GetCanonicalHash(number uint64) common.Hash
    HasBlock(hash common.Hash, number uint64) bool
}

// ChainWriter 定義了寫入區塊鏈的介面
type ChainWriter interface {
    // Block insertion
    InsertBlock(block *types.Block) error
    InsertBlocks(blocks []*types.Block) error
    InsertChain(blocks []*types.Block) (int, error)
    
    // State commit
    CommitState(root common.Hash) error
    
    // Rollback
    Rollback(hash common.Hash) error
}

// ChainProcessor 處理區塊和交易
type ChainProcessor interface {
    // Process processes the state changes for a block
    Process(block *types.Block, statedb state.StateDB) (types.Receipts, []*types.Log, uint64, error)
    
    // ValidateBody validates the body of a block
    ValidateBody(block *types.Block) error
    
    // ValidateState validates the state after execution
    ValidateState(block *types.Block, statedb state.StateDB, receipts types.Receipts, usedGas uint64) error
}

// ChainEventPublisher 發布鏈事件
type ChainEventPublisher interface {
    SubscribeChainEvent(ch chan<- ChainEvent) event.Subscription
    SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription
    SubscribeChainSideEvent(ch chan<- ChainSideEvent) event.Subscription
    SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription
    SubscribeRemovedLogsEvent(ch chan<- RemovedLogsEvent) event.Subscription
}

// BlockChain 主介面
type BlockChain interface {
    ChainReader
    ChainWriter
    ChainProcessor
    ChainEventPublisher
    
    // Additional methods
    Stop()
    Reset() error
    Export(w io.Writer) error
    ExportN(w io.Writer, first, last uint64) error
    Snapshot() error
}
```

### 2.2 事件系統

```go
// events.go
package blockchain

import (
    "frizo-blockchain/common"
    "frizo-blockchain/core/types"
)

// ChainEvent 區塊插入事件
type ChainEvent struct {
    Block *types.Block
    Hash  common.Hash
    Logs  []*types.Log
}

// ChainHeadEvent 鏈頭更新事件
type ChainHeadEvent struct {
    Block *types.Block
}

// ChainSideEvent 側鏈事件（分叉）
type ChainSideEvent struct {
    Block *types.Block
}

// RemovedLogsEvent 日誌移除事件（鏈重組時）
type RemovedLogsEvent struct {
    Logs []*types.Log
}

// TxPoolEvent 交易池事件
type TxPoolEvent struct {
    Type string // "add", "remove", "drop"
    Tx   *types.Transaction
}
```

<br>
<br>
<br>
<br>

## 三、核心模組實現

### 3.1 區塊鏈主體（精簡版）

```go
// blockchain.go
package blockchain

import (
    "sync"
    "sync/atomic"
    "frizo-blockchain/common"
    "frizo-blockchain/core/types"
    "frizo-blockchain/core/state"
    "frizo-blockchain/core/rawdb"
    "frizo-blockchain/event"
    "frizo-blockchain/consensus"
)

type BlockChain struct {
    chainConfig *ChainConfig      // 鏈配置
    db          rawdb.Database    // 底層數據庫
    
    // Caches
    headerCache  *lru.Cache[common.Hash, *types.Header]
    blockCache   *lru.Cache[common.Hash, *types.Block]
    receiptCache *lru.Cache[common.Hash, types.Receipts]
    stateCache   state.Database
    
    // Current state
    currentBlock     atomic.Pointer[types.Block]  // 當前區塊
    currentState     atomic.Pointer[state.StateDB] // 當前狀態
    currentFastBlock atomic.Pointer[types.Block]  // 當前快速同步區塊
    
    // Components
    processor   Processor         // 區塊處理器
    validator   Validator         // 區塊驗證器
    prefetcher  *StatePrefetcher  // 狀態預取器
    indexer     *ChainIndexer     // 鏈索引器
    
    // Consensus
    engine      consensus.Engine  // 共識引擎
    
    // Events
    scope       event.SubscriptionScope
    chainFeed   event.Feed
    chainHeadFeed event.Feed
    chainSideFeed event.Feed
    logsFeed    event.Feed
    rmLogsFeed  event.Feed
    
    // Channels
    quit        chan struct{}      // 退出信號
    
    // Mutexes
    chainmu     sync.RWMutex       // 鏈操作鎖
    procmu      sync.Mutex         // 處理鎖
    
    // Metrics
    metrics     *ChainMetrics
}

// NewBlockChain 創建新的區塊鏈實例
func NewBlockChain(db rawdb.Database, cacheConfig *CacheConfig, chainConfig *ChainConfig, engine consensus.Engine) (*BlockChain, error) {
    bc := &BlockChain{
        chainConfig:  chainConfig,
        db:           db,
        quit:         make(chan struct{}),
        engine:       engine,
        headerCache:  lru.NewCache[common.Hash, *types.Header](headerCacheLimit),
        blockCache:   lru.NewCache[common.Hash, *types.Block](blockCacheLimit),
        receiptCache: lru.NewCache[common.Hash, types.Receipts](receiptCacheLimit),
    }
    
    // Initialize components
    bc.processor = NewStateProcessor(chainConfig, bc, engine)
    bc.validator = NewBlockValidator(chainConfig, bc, engine)
    bc.prefetcher = NewStatePrefetcher(bc.stateCache)
    
    // Load blockchain state from database
    if err := bc.loadLastState(); err != nil {
        return nil, err
    }
    
    // Initialize genesis if necessary
    if err := bc.initGenesis(); err != nil {
        return nil, err
    }
    
    // Start indexer
    bc.indexer = NewChainIndexer(db, bc)
    bc.indexer.Start()
    
    return bc, nil
}
```

### 3.2 區塊驗證器

```go
// block_validator.go
package blockchain

import (
    "fmt"
    "frizo-blockchain/core/types"
    "frizo-blockchain/core/state"
    "frizo-blockchain/consensus"
)

// BlockValidator 負責驗證區塊
type BlockValidator struct {
    config  *ChainConfig      // 鏈配置
    bc      *BlockChain       // 區塊鏈引用
    engine  consensus.Engine  // 共識引擎
}

// ValidateBody 驗證區塊體
func (v *BlockValidator) ValidateBody(block *types.Block) error {
    // 檢查區塊是否已知
    if v.bc.HasBlock(block.Hash(), block.NumberU64()) {
        return ErrKnownBlock
    }
    
    // 檢查父區塊
    parent := v.bc.GetBlock(block.ParentHash(), block.NumberU64()-1)
    if parent == nil {
        return ErrUnknownAncestor
    }
    
    // 驗證區塊時間戳
    if block.Time() <= parent.Time() {
        return ErrInvalidTimestamp
    }
    
    // 驗證交易
    if err := v.validateTransactions(block); err != nil {
        return err
    }
    
    // 使用共識引擎驗證
    if err := v.engine.VerifyHeader(v.bc, block.Header(), true); err != nil {
        return err
    }
    
    return nil
}

// ValidateState 驗證狀態轉換
func (v *BlockValidator) ValidateState(block *types.Block, statedb state.StateDB, receipts types.Receipts, usedGas uint64) error {
    // 驗證收據根
    receiptRoot := types.DeriveReceiptsMerkleRoot(receipts)
    if receiptRoot != block.ReceiptHash() {
        return fmt.Errorf("invalid receipt root: have %x, want %x", receiptRoot, block.ReceiptHash())
    }
    
    // 驗證狀態根
    stateRoot := statedb.IntermediateRoot(true)
    if stateRoot != block.StateRoot() {
        return fmt.Errorf("invalid state root: have %x, want %x", stateRoot, block.StateRoot())
    }
    
    // 驗證 Gas 使用量
    if usedGas != block.GasUsed() {
        return fmt.Errorf("invalid gas used: have %d, want %d", usedGas, block.GasUsed())
    }
    
    return nil
}
```

### 3.3 鏈管理器

```go
// chain_manager.go
package blockchain

import (
    "frizo-blockchain/common"
    "frizo-blockchain/core/types"
)

// ChainManager 管理鏈狀態和重組
type ChainManager struct {
    bc *BlockChain
}

// InsertChain 插入區塊鏈
func (cm *ChainManager) InsertChain(blocks []*types.Block) (int, error) {
    cm.bc.chainmu.Lock()
    defer cm.bc.chainmu.Unlock()
    
    n := 0
    for _, block := range blocks {
        // 驗證區塊
        if err := cm.bc.validator.ValidateBody(block); err != nil {
            return n, err
        }
        
        // 處理區塊
        receipts, logs, usedGas, err := cm.bc.processor.Process(block, statedb)
        if err != nil {
            return n, err
        }
        
        // 驗證狀態
        if err := cm.bc.validator.ValidateState(block, statedb, receipts, usedGas); err != nil {
            return n, err
        }
        
        // 寫入數據庫
        if err := cm.writeBlockWithState(block, receipts, statedb); err != nil {
            return n, err
        }
        
        // 檢查是否需要重組
        if err := cm.reorg(block); err != nil {
            return n, err
        }
        
        n++
        
        // 發送事件
        cm.bc.sendChainEvent(block, receipts, logs)
    }
    
    return n, nil
}

// reorg 處理鏈重組
func (cm *ChainManager) reorg(newBlock *types.Block) error {
    // 獲取當前頭
    oldBlock := cm.bc.CurrentBlock()
    
    // 如果新區塊是當前頭的子區塊，直接更新
    if oldBlock.Hash() == newBlock.ParentHash() {
        return cm.bc.writeHeadBlock(newBlock)
    }
    
    // 找到共同祖先
    commonAncestor := cm.findCommonAncestor(oldBlock, newBlock)
    
    // 回滾舊鏈
    if err := cm.rollback(oldBlock, commonAncestor); err != nil {
        return err
    }
    
    // 應用新鏈
    if err := cm.insertBlocks(commonAncestor, newBlock); err != nil {
        return err
    }
    
    return nil
}
```

<br>
<br>
<br>
<br>

## 四、數據層重構（rawdb）

### 4.1 數據庫介面

```go
// rawdb/database.go
package rawdb

import (
    "frizo-blockchain/ethdb"
)

// Database 包裝底層數據庫並提供鏈專用方法
type Database struct {
    db ethdb.Database
}

// NewDatabase 創建新的數據庫包裝器
func NewDatabase(db ethdb.Database) *Database {
    return &Database{db: db}
}

// Ancient 訪問歷史數據
func (db *Database) Ancient(kind string, number uint64) ([]byte, error) {
    // 實現歷史數據訪問
}
```

### 4.2 數據訪問器

```go
// rawdb/accessors_chain.go
package rawdb

import (
    "frizo-blockchain/common"
    "frizo-blockchain/core/types"
)

// ReadCanonicalHash 讀取規範鏈哈希
func ReadCanonicalHash(db ethdb.Reader, number uint64) common.Hash {
    data, _ := db.Get(headerHashKey(number))
    if len(data) == 0 {
        return common.Hash{}
    }
    return common.BytesToHash(data)
}

// WriteCanonicalHash 寫入規範鏈哈希
func WriteCanonicalHash(db ethdb.KeyValueWriter, hash common.Hash, number uint64) {
    db.Put(headerHashKey(number), hash.Bytes())
}

// ReadHeader 讀取區塊頭
func ReadHeader(db ethdb.Reader, hash common.Hash, number uint64) *types.Header {
    data, _ := db.Get(headerKey(number, hash))
    if len(data) == 0 {
        return nil
    }
    header := new(types.Header)
    if err := rlp.Decode(bytes.NewReader(data), header); err != nil {
        return nil
    }
    return header
}

// WriteHeader 寫入區塊頭
func WriteHeader(db ethdb.KeyValueWriter, header *types.Header) {
    var (
        hash   = header.Hash()
        number = header.Number.Uint64()
        encoded, _ = rlp.EncodeToBytes(header)
    )
    db.Put(headerKey(number, hash), encoded)
}
```

<br>
<br>
<br>
<br>

## 五、交易池重構

### 5.1 交易池主體

```go
// txpool/txpool.go
package txpool

import (
    "sync"
    "frizo-blockchain/common"
    "frizo-blockchain/core/types"
    "frizo-blockchain/event"
)

type TxPool struct {
    config       TxPoolConfig
    chainconfig  *params.ChainConfig
    chain        ChainReader
    
    mu           sync.RWMutex
    
    pending      map[common.Address]*TxList  // 待打包交易
    queue        map[common.Address]*TxList  // 排隊交易
    all          *TxLookup                    // 所有交易查詢
    priced       *TxPricedList                // 價格排序列表
    
    beats        map[common.Address]time.Time // 賬戶活躍時間
    
    // Events
    scope        event.SubscriptionScope
    txFeed       event.Feed
    
    // Metrics
    metrics      *TxPoolMetrics
}

// Add 添加交易到池中
func (pool *TxPool) Add(tx *types.Transaction) error {
    pool.mu.Lock()
    defer pool.mu.Unlock()
    
    // 驗證交易
    if err := pool.validateTx(tx); err != nil {
        return err
    }
    
    // 添加到 pending 或 queue
    from, _ := tx.From()
    
    if pool.pending[from] == nil {
        pool.pending[from] = newTxList(true)
    }
    
    replaced, err := pool.pending[from].Add(tx, pool.config.PriceBump)
    if err != nil {
        return err
    }
    
    // 如果替換了舊交易，從 all 中移除
    if replaced != nil {
        pool.all.Remove(replaced.Hash())
    }
    
    // 添加到 all 和 priced
    pool.all.Add(tx)
    pool.priced.Put(tx)
    
    // 發送事件
    go pool.txFeed.Send(TxPoolEvent{Type: "add", Tx: tx})
    
    return nil
}

// Pending 返回待打包交易
func (pool *TxPool) Pending() map[common.Address]types.Transactions {
    pool.mu.RLock()
    defer pool.mu.RUnlock()
    
    pending := make(map[common.Address]types.Transactions)
    for addr, list := range pool.pending {
        pending[addr] = list.Flatten()
    }
    return pending
}

// validateTx 驗證交易
func (pool *TxPool) validateTx(tx *types.Transaction) error {
    // 1. 基本驗證
    if tx.Size() > txMaxSize {
        return ErrOversizedData
    }
    
    // 2. 簽名驗證
    from, err := tx.From()
    if err != nil {
        return ErrInvalidSender
    }
    
    // 3. Nonce 檢查
    currentNonce := pool.chain.GetNonce(from)
    if tx.Nonce() < currentNonce {
        return ErrNonceTooLow
    }
    
    // 4. 餘額檢查
    balance := pool.chain.GetBalance(from)
    if balance.Cmp(tx.Cost()) < 0 {
        return ErrInsufficientFunds
    }
    
    // 5. Gas 檢查
    if tx.Gas() < IntrinsicGas(tx.Data(), tx.To() == nil) {
        return ErrIntrinsicGas
    }
    
    return nil
}
```

<br>
<br>
<br>
<br>

## 六、整合方案

### 6.1 與其他模組的整合

```go
// 在 node 包中整合所有模組
type Node struct {
    blockchain  *blockchain.BlockChain
    txPool      *txpool.TxPool
    consensus   consensus.Engine
    p2pServer   *p2p.Server
    rpcServer   *rpc.Server
}

func (n *Node) Start() error {
    // 1. 啟動區塊鏈
    n.blockchain.Start()
    
    // 2. 啟動交易池
    n.txPool.Start()
    
    // 3. 啟動共識引擎
    n.consensus.Start(n.blockchain, n.txPool)
    
    // 4. 啟動 P2P
    n.p2pServer.Start()
    
    // 5. 啟動 RPC
    n.rpcServer.Start()
    
    return nil
}
```

### 6.2 是否合併 storage/block_store？

**建議：是的，應該整合**

理由：
1. **職責明確** - 區塊存儲是區塊鏈的核心職責
2. **減少層級** - 避免過度抽象
3. **性能優化** - 可以更好地控制緩存和批量操作

整合方式：
- 將 `storage/block_store.go` 的功能移到 `blockchain/rawdb/`
- 保留 `storage/` 作為通用數據庫層（LevelDB、MemDB）
- `blockchain/rawdb/` 專門處理區塊鏈相關的數據操作

<br>
<br>
<br>
<br>

## 七、實施計劃

### Phase 1: 基礎重構
1. 創建新的目錄結構
2. 抽取介面定義
3. 實現 rawdb 層
4. 重構 blockchain.go

### Phase 2: 功能完善
1. 實現區塊驗證器
2. 實現鏈管理器
3. 實現事件系統
4. 完善交易池

### Phase 3: 測試優化
1. 單元測試
2. 整合測試
3. 性能優化
4. 文檔編寫