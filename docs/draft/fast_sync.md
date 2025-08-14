# Fast Sync 機制理解與實作指南

<br>

---

<br>

## 一、Fast Sync 核心概念

Fast Sync 是一種快速同步策略，允許新節點在不下載完整歷史狀態的情況下，快速追趕到網路的最新狀態並成為合格的驗證節點。

### 1.1 關鍵概念：Pivot Point（樞軸點）

**定義**：選擇一個距離最新區塊一定距離的歷史區塊作為狀態同步的起點。

**選擇原則**：
- 距離最新區塊 64-128 個區塊的位置
- 確保該點的狀態已被網路廣泛確認
- 避免選擇可能被回滾的區塊

```
創世區塊 ────────────────── Pivot Point ─── 最新區塊
│                           │              │
└── 只下載區塊頭              └── 完整狀態     └── 完整驗證
```

### 1.2 分階段同步策略

```go
type SyncPhase int

const (
PhaseHeaders    SyncPhase = iota  // 歷史區塊頭同步
PhaseState                        // Pivot Point 狀態下載
PhaseBodies                       // 近期區塊體下載
PhaseExecution                    // 交易執行與狀態重建
PhaseFullVerify                   // 完整驗證模式
)
```

<br>

## 二、Fast Sync 完整流程

### 動作1：選擇 Pivot Point 並分段下載

假設當前網路最高區塊為 10000，選擇第 9000 個區塊作為 Pivot Point：

```go
func (fs *FastSync) planDownload() SyncPlan {
pivot := 9000
latest := 10000

return SyncPlan{
// 歷史區塊：只下載區塊頭（1-9000）
HeadersOnly: Range{1, pivot},

// Pivot點：下載完整狀態樹
FullState: pivot,

// 近期區塊：下載區塊體用於執行（9001-10000）
BodiesNeeded: Range{pivot+1, latest},

// 實時區塊：完整同步（10000+）
FullSync: Range{latest+1, current},
}
}
```

**下載內容：**
1. **區塊頭 1-9000**：用於驗證鏈的完整性
2. **第 9000 區塊的完整狀態**：包含所有賬戶餘額、合約存儲等
3. **區塊體 9001-10000**：包含交易數據，用於狀態重建

### 動作2：狀態驗證與交易執行

```go
func (fs *FastSync) executeFromPivot() error {
// 驗證下載的 Pivot 狀態
if err := fs.verifyPivotState(); err != nil {
return err
}

currentState := fs.pivotState  // 第 9000 區塊的狀態

// 從 9001 開始逐區塊執行交易
for height := 9001; height <= 10000; height++ {
block := fs.getBlock(height)

// 執行區塊中的所有交易
receipts, newState, err := fs.executeBlock(block, currentState)
if err != nil {
return fmt.Errorf("execution failed at block %d: %v", height, err)
}

// 關鍵：驗證計算出的狀態根是否匹配區塊頭
if newState.Root() != block.Header.StateRoot {
return fmt.Errorf("state root mismatch at block %d", height)
}

// 驗證收據根
if receipts.Root() != block.Header.ReceiptRoot {
return fmt.Errorf("receipt root mismatch at block %d", height)
}

currentState = newState
}

return nil
}
```

**關鍵驗證點：**
- **Pivot 狀態驗證**：確保下載的狀態根與第 9000 區塊頭匹配
- **狀態轉換驗證**：每執行一個區塊都驗證狀態根
- **收據驗證**：確保交易執行結果正確

### 動作3：完成同步並升級為驗證節點

```go
func (fs *FastSync) completeSync() error {
// 3.1 追趕到網路最新區塊
if err := fs.catchUpToLatest(); err != nil {
return err
}

// 3.2 執行最終安全性檢查
if err := fs.finalSecurityChecks(); err != nil {
return err
}

// 3.3 切換到實時同步模式
fs.switchToLiveSync()

// 3.4 升級節點能力
fs.node.upgradeToFullValidator()

// 3.5 開始提供網路服務
fs.node.startServingPeers()

log.Info("Fast Sync completed - Node is now a full validator")
return nil
}

func (fs *FastSync) finalSecurityChecks() error {
// 1. 多節點狀態一致性檢查
if err := fs.verifyStateConsistency(); err != nil {
return err
}

// 2. 硬編碼檢查點驗證（如果有）
if err := fs.verifyCheckpoints(); err != nil {
return err
}

// 3. 隨機狀態抽樣驗證
return fs.randomStateVerification()
}
```

<br>

## 三、節點能力的演進

### 3.1 同步期間的節點狀態

```go
type NodeCapability struct {
CanValidateNewBlocks bool  // Fast Sync期間為false
CanProvideState      bool  // Fast Sync期間為false
CanMine              bool  // Fast Sync期間為false
CanSync              bool  // 只能接收，不能提供數據
SyncPhase           SyncPhase
}

func (n *Node) GetCapability() NodeCapability {
switch n.syncPhase {
case PhaseHeaders, PhaseState, PhaseBodies:
return NodeCapability{
CanValidateNewBlocks: false,
CanProvideState:      false,
CanMine:              false,
CanSync:              true,
SyncPhase:           n.syncPhase,
}
case PhaseExecution:
return NodeCapability{
CanValidateNewBlocks: true,  // 可以驗證新區塊
CanProvideState:      false, // 還不能提供歷史狀態
CanMine:              false, // 還不能參與共識
CanSync:              true,
SyncPhase:           n.syncPhase,
}
case PhaseFullVerify:
return NodeCapability{
CanValidateNewBlocks: true,
CanProvideState:      true,
CanMine:              true,
CanSync:              true,
SyncPhase:           n.syncPhase,
}
}
}
```

### 3.2 完整驗證節點的能力

同步完成後，節點具備完整的驗證能力：

```go
type FullNodeCapabilities struct {
// 區塊驗證
ValidateNewBlocks    func(*Block) error
ValidateBlockHeaders func([]*Header) error

// 狀態查詢
GetAccountBalance    func(Address) *big.Int
GetAccountNonce      func(Address) uint64
GetContractStorage   func(Address, Hash) Hash
GetContractCode      func(Address) []byte

// 交易處理
ValidateTransaction  func(*Transaction) error
ExecuteTransaction   func(*Transaction) (*Receipt, error)

// 共識參與
ProposeBlock         func() *Block
ValidateProposal     func(*Block) error

// 網路服務
ServeBlockHeaders    func(start, count uint64) []*Header
ServeBlockBodies     func(hashes []Hash) []*Body
ServeStateData       func(hash Hash) []byte
ServeReceipts        func(hashes []Hash) []*Receipt
}
```

<br>

## 四、安全性保證機制

### 4.1 多重驗證策略

```go
func (fs *FastSync) securityChecks() error {
// 1. Pivot 狀態多節點一致性檢查
pivotHash := fs.pivotState.Root()
confirmations := 0

for _, peer := range fs.peers {
if peer.GetStateRoot(fs.pivotHeight) == pivotHash {
confirmations++
}
}

if confirmations < len(fs.peers)*2/3 {
return errors.New("insufficient state confirmations")
}

// 2. 隨機狀態抽樣驗證
return fs.randomStateVerification()
}

func (fs *FastSync) randomStateVerification() error {
// 隨機選擇賬戶進行 Merkle Proof 驗證
accounts := fs.selectRandomAccounts(100)

for _, addr := range accounts {
proof, err := fs.stateDB.GetProof(addr)
if err != nil {
return err
}

// 向其他節點驗證 proof
if !fs.verifyProofWithPeers(addr, proof) {
return fmt.Errorf("state verification failed for %s", addr.Hex())
}
}
return nil
}
```

### 4.2 檢查點機制

```go
// 硬編碼的可信檢查點
var TrustedCheckpoints = map[uint64]common.Hash{
1000:  common.HexToHash("0x1234..."),
5000:  common.HexToHash("0x5678..."),
10000: common.HexToHash("0xabcd..."),
}

func (fs *FastSync) verifyCheckpoints() error {
for height, expectedHash := range TrustedCheckpoints {
if height <= fs.currentHeight {
actualHash := fs.getBlockHash(height)
if actualHash != expectedHash {
return fmt.Errorf("checkpoint mismatch at height %d", height)
}
}
}
return nil
}
```

<br>

## 五、Frizo-blockchain Phase-1 實作建議

### 5.1 簡化版 Fast Sync

考慮到 Phase-1 的目標（2-3 節點私有網路），建議採用簡化策略：

```go
type SimpleFastSync struct {
recentDepth  int  // 完整同步最近的區塊數量（如100）
trustHeaders bool // 是否信任歷史區塊頭
}

func (s *SimpleFastSync) Start() error {
latest := s.getLatestBlockHeight()
pivot := latest - s.recentDepth

if pivot <= 0 {
// 區塊數量少，直接全同步
return s.doFullSync()
}

// 1. 同步歷史區塊頭（快速）
if err := s.syncHeaders(1, pivot); err != nil {
return err
}

// 2. 同步 pivot 狀態
if err := s.syncPivotState(pivot); err != nil {
return err
}

// 3. 同步並執行最近的區塊
if err := s.syncAndExecuteRecent(pivot+1, latest); err != nil {
return err
}

// 4. 升級為完整節點
s.node.upgradeToFullValidator()
return nil
}
```

### 5.2 實作優先級

**Phase 1-A: 基礎同步框架**
1. 實現基本的 P2P 同步協議
2. 實現區塊頭/區塊體下載
3. 實現基礎的狀態管理

**Phase 1-B: Fast Sync 核心**
1. 實現 Pivot Point 選擇邏輯
2. 實現狀態下載與驗證
3. 實現交易執行與狀態重建

**Phase 1-C: 安全性與優化**
1. 實現多節點驗證機制
2. 實現錯誤處理與重試邏輯
3. 實現同步進度監控

<br>

## 六、測試策略

### 6.1 單元測試重點

```go
func TestFastSyncComponents(t *testing.T) {
// 1. Pivot Point 選擇邏輯
testPivotPointSelection()

// 2. 狀態下載與驗證
testStateDownloadAndVerification()

// 3. 交易執行與狀態重建
testTransactionExecutionAndStateReconstruction()

// 4. 安全性檢查
testSecurityChecks()
}
```

### 6.2 整合測試場景

```go
func TestFastSyncIntegration(t *testing.T) {
// 場景1：2節點網路，一個新節點加入
testTwoNodeSync()

// 場景2：3節點網路，網路分區後恢復
testNetworkPartitionRecovery()

// 場景3：節點重啟後的同步恢復
testSyncResumption()
}
```

<br>

## 七、性能考量

### 7.1 並發優化

```go
type SyncScheduler struct {
headerWorkers chan struct{}  // 控制區塊頭下載併發
bodyWorkers   chan struct{}  // 控制區塊體下載併發
stateWorkers  chan struct{}  // 控制狀態下載併發
}

func (s *SyncScheduler) DownloadHeaders(requests []HeaderRequest) {
for _, req := range requests {
s.headerWorkers <- struct{}{}
go func(r HeaderRequest) {
defer func() { <-s.headerWorkers }()
s.processHeaderRequest(r)
}(req)
}
}
```

### 7.2 記憶體管理

```go
type SyncCache struct {
maxHeaders int
maxBodies  int
maxState   int

headers map[uint64]*Header
bodies  map[common.Hash]*Body
state   map[common.Hash][]byte
}

func (c *SyncCache) evictOldest() {
// LRU 淘汰策略
if len(c.headers) > c.maxHeaders {
// 淘汰最舊的區塊頭
}
}
```

<br>

---

**總結**：Fast Sync 是一個平衡安全性和效率的複雜機制。通過選擇合適的 Pivot Point，分階段下載和驗證，最終能讓新節點快速成為網路中的合格驗證者。在 frizo-blockchain 的實作中，建議從簡化版本開始，逐步完善各項安全性檢查和優化機制。