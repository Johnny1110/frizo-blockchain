# Frizo Blockchain Phase-1 開發計劃
<br>

---

<br>

## 一、已完成的主要功能

### 1. 密碼學相關功能 ✅

* 公私鑰 (Wallet) 生成: crypto/keys.go
* 簽署與驗證: crypto/signature.go
* 哈希函數: crypto/hash.go - Keccak256 實現

<br>

### 2. 區塊鏈核心資料結構定義 ✅

* 區塊結構: core/types/block.go

    * Header、Body 結構完整
    * 區塊驗證邏輯實現
    * RLP 編碼支援


* 交易結構: core/types/transaction.go

    * 支援簽名驗證
    * From `Ecrecover()` 地址恢復
    * Gas 計算機制
    * 交易緩存優化


* 收據結構: core/types/receipt.go

    * 交易執行結果記錄
    * Gas 使用追蹤



<br>

### 3. Merkle Tree 與 MPT 實現 ✅

* Merkle Tree: /trie/merkle.go

    * 完整的樹構建邏輯
    * 證明生成與驗證


* Modified Merkle Patricia Tree (MPT):

    * /trie/mpt.go - 核心樹結構，支援 RLP 編碼
    * /trie/mpt_proof.go - 完整的證明功能
    * /trie/mpt_tool.go - Debug 工具



<br>

### 4. Storage 持久化層 ✅

* DB 介面定義: /storage/interfaces/interfaces.go
* LevelDB 實現: /storage/leveldb/leveldb.go
* 記憶體 DB: /storage/memory/memdb.go（測試用）
* 鏈資料庫管理: /storage/database.go

<br>

三層分離架構（blockDB, stateDB, indexDB）

<br>

* 區塊存儲: /storage/block_store.go

    * 區塊讀寫功能
    * 交易索引管理


* 狀態存儲: /storage/state_store.go

    * MPT 節點持久化
    * 合約代碼存儲


* 存儲架構: /storage/schema.go

    * 鍵值編碼規範



<br>
5. 區塊鏈核心邏輯（待完成）⚠️

* Blockchain 主體: core/blockchain/blockchain.go

    * 區塊插入與驗證
    * 狀態轉換執行
    * 鏈狀態管理


* 創世區塊: core/blockchain/genesis.go

    * 讀取 genesis.json
    * 預分配賬戶設置
    * 默認創世配置


* 狀態處理器: core/blockchain/state_processor.go

    * 交易執行邏輯
    * 收據生成



<br>

### 6. 簡化版狀態管理(單元測試用) ✅

* SimpleStateDB: core/state/simple_state.go

    * 賬戶餘額管理
    * Nonce 追蹤
    * Snapshot/Revert 機制
    * 狀態根計算



<br>

### 7. 通用工具與常量 ✅

* 通用類型: common/types.go

    * Address、Hash 類型定義
    * RLP 編碼工具


* 錯誤定義: common/errors.go

    * 完整的錯誤類型系統


* 常量定義: common/constants.go

    * 鏈參數配置
    * Gas 費用設置
    * 網路常量



<br>

### 8. 構建與測試基礎設施 ✅

* Makefile: 完整的構建系統
* 測試框架: 單元測試覆蓋主要模組

<br>
<br>

---

<br>
<br>

## 二、未完成但必需的功能（Phase-1）

### 1. MPT 與 State 深度整合 🔴

* 將 SimpleStateDB 與 MPT 完全整合
  實現真正的狀態樹管理
* 支援狀態證明生成

<br>

### 2. 交易池（TxPool）🔴

* 交易驗證與排序
* Gas Price 排序機制
* 交易替換策略
* Pending/Queued 狀態管理
* 內存限制與驅逐策略

<br>

### 3. 共識機制 (要做成後續可置換實作，未來要可靈活抽換不同共識算法實作) 🔴

* 實現彈性共識機制框架 [參考文件](consensus.md)

* 先實現 Clique (PoA) 共識

  實現要點：

    * 授權節點管理
    * 輪流出塊機制
    * 簽名驗證

<br>

### 4. P2P 網路層 🔴

基礎協議實現：

* 節點發現（可先用靜態節點）
* TCP 連接管理
* 消息編解碼（RLP）


同步協議：

* 握手協議
* 區塊頭同步
* 區塊體請求/響應
* 交易廣播



<br>

### 5. RPC API 接口 🔴

基礎 JSON-RPC：

* eth_blockNumber
* eth_getBlockByNumber
* eth_getBalance
* eth_sendRawTransaction
* eth_getTransactionReceipt



<br>
<br>

---

<br>
<br>

## 三、優化後的開發計劃

### Phase 1-A: 核心功能完善

<br>

1. State 與 MPT 整合
```go
// 需要實現的介面
type StateDB interface {
    GetState(addr Address, key Hash) Hash
    SetState(addr Address, key Hash, value Hash)
    GetCode(addr Address) []byte
    SetCode(addr Address, code []byte)
    Commit() (Hash, error)
    Database() Database
}
```

<br>

2. 完善 blockchain.go 相關產塊功能

<br>

3. 交易池實現
```go
3. type TxPool struct {
    pending map[Address]TxList  // 待打包交易
    queue   map[Address]TxList  // 排隊交易
    config  TxPoolConfig       // 配置參數
}
```

<br>

### Phase 1-B: 共識機制

設計成後續可抽換共識的框架。

目前實作先採用 Clique PoA 共識

實現步驟：

* 定義授權節點列表
* 實現簽名者輪換邏輯
* 區塊簽名與驗證
* 投票機制（可選）



<br>

### Phase 1-C: 網路層實現

1. P2P 基礎架構
```go
type P2PServer struct {
    peers     map[NodeID]*Peer
    protocols []Protocol
    listener  net.Listener
}
```

2. 以太坊 Wire Protocol

必需消息類型：

```
Status (0x00)：握手
NewBlockHashes (0x01)：新區塊通知
Transactions (0x02)：交易廣播
GetBlockHeaders (0x03)：請求區塊頭
BlockHeaders (0x04)：區塊頭響應
GetBlockBodies (0x05)：請求區塊體
BlockBodies (0x06)：區塊體響應
```


3. 同步策略

Fast Sync 簡化版：

* 同步區塊頭鏈
* 下載最新狀態
* 同步最近區塊體
* 驗證狀態根

<br>

### Phase 1-D: 整合測試

測試場景

* 單節點測試：

    * 創世區塊初始化
    * 交易執行與狀態更新
    * 區塊生成與持久化


* 多節點測試：

    * 3 節點 PoA 網路
    * 交易廣播與同步
    * 分叉處理


壓力測試：

* TPS 測試（目標 100+ TPS）
* 狀態膨脹測試
* 網路分區恢復



<br>
<br>

---

<br>
<br>

## 四、技術架構

1. 模組化設計
```
frizo-blockchain/
├── consensus/        # 共識介面
│   ├── clique/      # PoA 實現
│   └── interface.go # 共識介面定義
├── eth/             # 以太坊協議
│   ├── protocols/   # Wire Protocol
│   └── sync/        # 同步邏輯
├── rpc/            # RPC 服務
│   ├── server.go   
│   └── endpoints/  # API 端點
└── txpool/         # 交易池
    ├── txpool.go
    └── txlist.go
```

2. 介面抽象

```go
// Consensus 介面
type Engine interface {
    Author(header *Header) (Address, error)
    VerifyHeader(chain ChainReader, header *Header) error
    Prepare(chain ChainReader, header *Header) error
    Finalize(chain ChainReader, header *Header, state *StateDB, txs []*Transaction) error
    Seal(chain ChainReader, block *Block, results chan<- *Block, stop <-chan struct{}) error
}
```

3. 配置管理
```go
type Config struct {
    ChainID     *big.Int
    Consensus   ConsensusConfig
    Network     NetworkConfig
    TxPool      TxPoolConfig
    Database    DatabaseConfig
}
```

<br>
<br>

---

<br>
<br>

## 五、里程碑與驗收標準

### Milestone 1: 本地區塊鏈

* 區塊創建與驗證
* 狀態管理與持久化
* 交易池基本功能
* 單節點出塊

<br>

### Milestone 2: 網路同步

* P2P 連接建立
* 區塊同步協議
* 交易廣播
* 3 節點測試網

<br>

### Milestone 3: Phase-1 發布

* RPC API 完整實現
* 性能優化（100+ TPS）
* 文檔完善
* Docker 部署支援

<br>
<br>

---

<br>
<br>

## 六、風險與挑戰

* 狀態同步複雜度：建議先實現全節點同步
* 網路穩定性：需要完善的錯誤處理與重連機制
* 共識安全性：PoA 需要可信的初始節點集

<br>
<br>

---

<br>
<br>

## 七、後續發展路線

* 升級到 PoS 共識
* Casper FFG 簡化版
* Validator 管理
* Slashing 機制
