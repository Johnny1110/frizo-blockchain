# Frizo Blockchain State 模組完整架構設計

## 一、模組架構總覽

```
core/state/
├── interfaces.go         # 介面定義
├── state_db.go           # 主要的 StateDB 實現
├── state_object.go       # 狀態對象（賬戶）
├── journal.go            # 變更日誌（用於快照和回滾）
├── snapshot.go           # 快照管理
├── access_list.go        # EIP-2929 訪問列表
├── trie_prefetcher.go   # 預取優化
└── cache.go              # 緩存層
```

## 二、核心介面定義

### 2.1 StateDB 介面

```go
// core/state/interfaces.go
package state

import (
    "frizo-blockchain/common"
    "frizo-blockchain/core/types"
    "math/big"
)

// StateDB 是以太坊狀態數據庫的核心介面
type StateDB interface {
    // 賬戶基本操作
    CreateAccount(common.Address)
    SubBalance(common.Address, *big.Int)
    AddBalance(common.Address, *big.Int)
    GetBalance(common.Address) *big.Int
    
    GetNonce(common.Address) uint64
    SetNonce(common.Address, uint64)
    
    GetCodeHash(common.Address) common.Hash
    GetCode(common.Address) []byte
    SetCode(common.Address, []byte)
    GetCodeSize(common.Address) int
    
    // 合約存儲操作
    GetState(common.Address, common.Hash) common.Hash
    SetState(common.Address, common.Hash, common.Hash)
    GetCommittedState(common.Address, common.Hash) common.Hash
    
    // 賬戶管理
    HasSuicided(common.Address) bool
    Suicide(common.Address) bool
    Exist(common.Address) bool
    Empty(common.Address) bool
    
    // 快照和回滾
    Snapshot() int
    RevertToSnapshot(int)
    
    // 提交和持久化
    Commit(deleteEmptyObjects bool) (common.Hash, error)
    IntermediateRoot(deleteEmptyObjects bool) common.Hash
    Finalise(deleteEmptyObjects bool)
    
    // 日誌相關
    AddLog(*types.Log)
    GetLogs(hash common.Hash) []*types.Log
    
    // 預編譯和退款
    AddRefund(uint64)
    SubRefund(uint64)
    GetRefund() uint64
    
    // 訪問列表（EIP-2929）
    AddAddressToAccessList(addr common.Address)
    AddSlotToAccessList(addr common.Address, slot common.Hash)
    IsAddressInAccessList(addr common.Address) bool
    IsSlotInAccessList(addr common.Address, slot common.Hash) (addressOk, slotOk bool)
    
    // 調試和工具
    ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) error
    Copy() StateDB
    Database() Database
}

// Database 包裝了狀態和存儲 Trie 操作
type Database interface {
    OpenTrie(root common.Hash) (Trie, error)
    OpenStorageTrie(addrHash, root common.Hash) (Trie, error)
    CopyTrie(Trie) Trie
    ContractCode(addrHash, codeHash common.Hash) ([]byte, error)
    ContractCodeSize(addrHash, codeHash common.Hash) (int, error)
    TrieDB() *trie.Database
}

// Trie 是 Merkle Patricia Trie 的介面
type Trie interface {
    TryGet(key []byte) ([]byte, error)
    TryUpdate(key, value []byte) error
    TryDelete(key []byte) error
    Commit(onleaf trie.LeafCallback) (common.Hash, error)
    Hash() common.Hash
    NodeIterator(startKey []byte) trie.NodeIterator
    Prove(key []byte, fromLevel uint, proofDb ethdb.KeyValueWriter) error
}
```

## 三、核心數據結構

### 3.1 StateDB 結構體

```go
// core/state/state_db.go
type StateDB struct {
    db           Database               // 底層數據庫
    trie         Trie                   // 賬戶 trie
    
    // 狀態對象緩存
    stateObjects        map[common.Address]*stateObject
    stateObjectsPending map[common.Address]struct{} // 待 finalise 的對象
    stateObjectsDirty   map[common.Address]struct{} // 待 commit 的對象
    
    // 日誌和快照
    journal        *journal
    validRevisions []revision
    nextRevisionId int
    
    // 交易執行上下文
    thash, bhash common.Hash           // 當前交易和區塊哈希
    txIndex      int                    // 當前交易索引
    logs         map[common.Hash][]*types.Log
    logSize      uint
    
    // 性能指標
    AccountReads   time.Duration
    AccountHashes  time.Duration
    AccountUpdates time.Duration
    AccountCommits time.Duration
    StorageReads   time.Duration
    StorageHashes  time.Duration
    StorageUpdates time.Duration
    StorageCommits time.Duration
    
    // 預取器
    prefetcher *triePrefetcher
    
    // 訪問列表
    accessList *accessList
    
    // 緩存配置
    cache *stateCache
}
```

### 3.2 State Object（賬戶對象）

```go
// core/state/state_object.go
type stateObject struct {
    address  common.Address
    addrHash common.Hash // 地址的哈希（用於存儲）
    data     Account     // 賬戶數據
    db       *StateDB
    
    // 緩存
    trie Trie // 存儲 trie
    code Code // 合約代碼
    
    // 存儲緩存
    originStorage  Storage // 原始存儲（從數據庫讀取）
    pendingStorage Storage // 待處理存儲（當前更改）
    dirtyStorage   Storage // 髒存儲（需要刷新到磁盤）
    
    // 標記
    dirtyCode bool // 代碼是否被修改
    suicided  bool // 是否被銷毀
    deleted   bool // 是否被刪除
}

type Account struct {
    Nonce    uint64
    Balance  *big.Int
    Root     common.Hash // 存儲 trie 的根
    CodeHash []byte
}
```

## 四、實現細節

### 4.1 快照和日誌機制

```go
// core/state/journal.go
type journal struct {
    entries []journalEntry // 當前變更記錄
    dirties map[common.Address]int // 髒賬戶索引
}

type journalEntry interface {
    revert(*StateDB)
    dirtied() *common.Address
}

// 各種日誌條目類型
type (
    createObjectChange struct {
        account *common.Address
    }
    resetObjectChange struct {
        prev *stateObject
    }
    suicideChange struct {
        account     *common.Address
        prev        bool
        prevBalance *big.Int
    }
    balanceChange struct {
        account *common.Address
        prev    *big.Int
    }
    nonceChange struct {
        account *common.Address
        prev    uint64
    }
    storageChange struct {
        account       *common.Address
        key, prevalue common.Hash
    }
    codeChange struct {
        account  *common.Address
        prevcode []byte
        prevhash common.Hash
    }
)
```

### 4.2 快照實現

```go
// core/state/snapshot.go
func (s *StateDB) Snapshot() int {
    id := s.nextRevisionId
    s.nextRevisionId++
    s.validRevisions = append(s.validRevisions, revision{id, s.journal.length()})
    return id
}

func (s *StateDB) RevertToSnapshot(revid int) {
    idx := s.validRevisions.length() - 1
    for idx >= 0 {
        if s.validRevisions[idx].id == revid {
            break
        }
        idx--
    }
    if idx < 0 {
        panic("revision id not found")
    }
    
    snapshot := s.validRevisions[idx].journalIndex
    s.journal.revert(s, snapshot)
    s.validRevisions = s.validRevisions[:idx]
}
```

### 4.3 Commit 流程

```go
func (s *StateDB) Commit(deleteEmptyObjects bool) (common.Hash, error) {
    // 1. Finalise 所有待處理的狀態對象
    for addr := range s.stateObjectsPending {
        obj := s.stateObjects[addr]
        if obj.suicided || (deleteEmptyObjects && obj.empty()) {
            s.deleteStateObject(obj)
        } else {
            obj.updateRoot(s.db)
            s.updateStateObject(obj)
        }
    }
    
    // 2. 提交所有髒對象到 trie
    for addr := range s.stateObjectsDirty {
        obj := s.stateObjects[addr]
        if obj.deleted {
            s.trie.TryDelete(addr[:])
        } else {
            s.trie.TryUpdate(addr[:], obj.data.Encode())
        }
    }
    
    // 3. 提交 trie 到數據庫
    root, err := s.trie.Commit(nil)
    if err != nil {
        return common.Hash{}, err
    }
    
    // 4. 清理緩存
    s.stateObjectsPending = make(map[common.Address]struct{})
    s.stateObjectsDirty = make(map[common.Address]struct{})
    
    return root, nil
}
```

## 五、性能優化策略

### 5.1 緩存層設計

```go
// core/state/cache.go
type stateCache struct {
    accountCache *lru.Cache // LRU 緩存賬戶
    storageCache *lru.Cache // LRU 緩存存儲
    codeCache    *lru.Cache // LRU 緩存代碼
    
    accountCacheSize int
    storageCacheSize int
    codeCacheSize    int
}
```

### 5.2 預取優化

```go
// core/state/trie_prefetcher.go
type triePrefetcher struct {
    db       Database
    root     common.Hash
    fetches  map[common.Hash]Trie
    
    deliveryMissMeter metrics.Meter
    accountLoadMeter  metrics.Meter
    storageLoadMeter  metrics.Meter
}
```

## 六、實現路線圖

### 第一階段：基礎功能（2-3天）
1. 實現核心介面定義
2. 實現 StateDB 基本結構
3. 實現 stateObject 賬戶管理
4. 整合現有的 MPT 實現

### 第二階段：快照和日誌（2天）
1. 實現 journal 日誌系統
2. 實現 snapshot/revert 機制
3. 添加單元測試

### 第三階段：存儲和提交（2天）
1. 實現合約存儲管理
2. 實現 Commit 流程
3. 整合 storage 層

### 第四階段：優化和測試（1天）
1. 添加 LRU 緩存
2. 實現預取優化
3. 性能測試和基準測試
4. 集成測試

## 七、與現有代碼的整合

### 需要修改的現有模組：

1. **MPT (trie/mpt.go)**
    - 添加 Iterator 支持
    - 優化 Commit 方法
    - 添加 Prove 方法

2. **Storage (storage/)**
    - 擴展 StateStore 介面
    - 優化批量操作

3. **Blockchain (core/blockchain/)**
    - 使用新的 StateDB 替換 SimpleStateDB
    - 更新 state_processor.go

## 八、測試計劃

### 單元測試
- 賬戶 CRUD 操作
- 快照和回滾
- 存儲操作
- Commit 流程

### 集成測試
- 與 blockchain 整合
- 交易執行
- 區塊處理
- 狀態同步

### 性能測試
- 大量賬戶操作
- 深度存儲樹
- 並發訪問
- 內存使用

## 九、生產級質量檢查清單

- [ ] 完整的錯誤處理
- [ ] 並發安全（必要時使用鎖）
- [ ] 內存洩漏檢查
- [ ] 性能基準測試
- [ ] 代碼覆蓋率 > 80%
- [ ] 完整的文檔
- [ ] 日誌和監控
- [ ] 配置管理