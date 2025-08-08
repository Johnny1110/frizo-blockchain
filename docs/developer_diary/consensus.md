# Frizo 通用共識機制介面設計

<br>

## 一、核心共識介面架構

```go
// consensus/interface.go
package consensus

import (
    "frizo-blockchain/common"
    "frizo-blockchain/core/state"
    "frizo-blockchain/core/types"
    "math/big"
    "time"
)

// Engine 是所有共識機制必須實現的核心介面
type Engine interface {
    // 基礎驗證介面
    Verifier
    
    // 區塊生產介面
    BlockProducer
    
    // 共識特定介面
    ConsensusCore
    
    // 生命週期管理
    Lifecycle
}

// Verifier 負責區塊和交易驗證
type Verifier interface {
    // VerifyHeader 驗證區塊頭（快速驗證）
    VerifyHeader(chain ChainHeaderReader, header *types.Header, seal bool) error
    
    // VerifyHeaders 批量驗證區塊頭（用於同步）
    VerifyHeaders(chain ChainHeaderReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error)
    
    // VerifyBody 驗證區塊體（交易列表等）
    VerifyBody(chain ChainReader, block *types.Block) error
    
    // VerifySeal 驗證區塊密封（PoW的nonce, PoS的簽名等）
    VerifySeal(chain ChainHeaderReader, header *types.Header) error
}

// BlockProducer 負責區塊生產
type BlockProducer interface {
    // Prepare 準備區塊頭（設置難度、時間戳等）
    Prepare(chain ChainHeaderReader, header *types.Header) error
    
    // Finalize 完成區塊（計算獎勵、更新狀態等）
    Finalize(chain ChainReader, header *types.Header, state *state.StateDB, 
             txs []*types.Transaction, receipts []*types.Receipt) (*types.Block, error)
    
    // FinalizeAndAssemble 完成並組裝區塊
    FinalizeAndAssemble(chain ChainReader, header *types.Header, state *state.StateDB,
                        txs []*types.Transaction, receipts []*types.Receipt) (*types.Block, error)
    
    // Seal 密封區塊（挖礦、簽名等）
    Seal(chain ChainReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error
    
    // SealHash 返回用於密封的哈希
    SealHash(header *types.Header) common.Hash
}

// ConsensusCore 共識核心功能
type ConsensusCore interface {
    // Author 返回區塊生產者地址
    Author(header *types.Header) (common.Address, error)
    
    // CalcDifficulty 計算難度（PoW）或權重（PoS）
    CalcDifficulty(chain ChainHeaderReader, time uint64, parent *types.Header) *big.Int
    
    // APIs 返回共識特定的RPC APIs
    APIs(chain ChainReader) []rpc.API
    
    // Protocol 返回共識協議信息
    Protocol() Protocol
}

// Lifecycle 生命週期管理
type Lifecycle interface {
    // Start 啟動共識引擎
    Start(chain ChainReader, currentBlock *types.Block, hasBadBlock func(common.Hash) bool) error
    
    // Stop 停止共識引擎
    Stop() error
}
```

## 二、支援不同共識的擴展介面

```go
// consensus/extensions.go

// PoWEngine 工作量證明擴展
type PoWEngine interface {
    Engine
    
    // Hashrate 返回當前算力
    Hashrate() float64
    
    // SetThreads 設置挖礦線程數
    SetThreads(threads int)
    
    // SetCoinbase 設置挖礦地址
    SetCoinbase(coinbase common.Address)
}

// PoSEngine 權益證明擴展
type PoSEngine interface {
    Engine
    
    // Validators 返回當前驗證者集合
    Validators() ValidatorSet
    
    // Stake 質押操作
    Stake(validator common.Address, amount *big.Int) error
    
    // Unstake 解除質押
    Unstake(validator common.Address, amount *big.Int) error
    
    // Slash 懲罰作惡節點
    Slash(validator common.Address, reason SlashReason) error
}

// PoAEngine 權威證明擴展（Clique）
type PoAEngine interface {
    Engine
    
    // Authorize 授權簽名者
    Authorize(signer common.Address, signFn SignerFn)
    
    // Signers 返回授權簽名者列表
    Signers() []common.Address
}

// BFTEngine 拜占庭容錯擴展（PBFT, Tendermint, HotStuff）
type BFTEngine interface {
    Engine
    
    // Round 當前共識輪次
    Round() uint64
    
    // Step 當前共識階段
    Step() ConsensusStep
    
    // Propose 提議區塊
    Propose(block *types.Block) error
    
    // Vote 投票
    Vote(block *types.Block, voteType VoteType) error
    
    // Commit 提交區塊
    Commit(block *types.Block) error
}

// DPoSEngine 委託權益證明擴展
type DPoSEngine interface {
    Engine
    
    // Delegates 返回當前代表列表
    Delegates() []common.Address
    
    // Vote 投票給代表
    VoteDelegate(voter common.Address, delegate common.Address, amount *big.Int) error
    
    // GetVotes 獲取代表的票數
    GetVotes(delegate common.Address) *big.Int
}

// DAGEngine DAG共識擴展（用於IOTA, Nano等）
type DAGEngine interface {
    Engine
    
    // SelectTips 選擇DAG尖端
    SelectTips() ([]*types.Transaction, error)
    
    // ValidateDAG 驗證DAG結構
    ValidateDAG(tx *types.Transaction) error
    
    // GetConfirmationLevel 獲取確認級別
    GetConfirmationLevel(tx common.Hash) uint64
}
```

## 三、共識相關的輔助類型

```go
// consensus/types.go

// Protocol 共識協議信息
type Protocol struct {
    Name      string
    Version   uint
    NetworkID uint64
}

// ValidatorSet 驗證者集合
type ValidatorSet interface {
    // List 返回驗證者列表
    List() []common.Address
    
    // Get 獲取驗證者信息
    Get(addr common.Address) (*Validator, error)
    
    // Add 添加驗證者
    Add(validator *Validator) error
    
    // Remove 移除驗證者
    Remove(addr common.Address) error
    
    // TotalPower 總投票權
    TotalPower() *big.Int
}

// Validator 驗證者信息
type Validator struct {
    Address     common.Address
    VotingPower *big.Int
    PublicKey   []byte
    Status      ValidatorStatus
}

// ValidatorStatus 驗證者狀態
type ValidatorStatus uint8

const (
    Active ValidatorStatus = iota
    Pending
    Jailed
    Unbonding
)

// ConsensusStep BFT共識階段
type ConsensusStep uint8

const (
    StepPropose ConsensusStep = iota
    StepPrevote
    StepPrecommit
    StepCommit
)

// VoteType 投票類型
type VoteType uint8

const (
    Prevote VoteType = iota
    Precommit
)

// SlashReason 懲罰原因
type SlashReason uint8

const (
    DoubleSign SlashReason = iota
    Downtime
    InvalidProposal
)

// ChainReader 鏈讀取介面
type ChainReader interface {
    ChainHeaderReader
    
    // GetBlock 獲取區塊
    GetBlock(hash common.Hash, number uint64) *types.Block
    
    // StateAt 獲取某個區塊的狀態
    StateAt(root common.Hash) (*state.StateDB, error)
}

// ChainHeaderReader 區塊頭讀取介面
type ChainHeaderReader interface {
    // Config 鏈配置
    Config() *params.ChainConfig
    
    // CurrentHeader 當前區塊頭
    CurrentHeader() *types.Header
    
    // GetHeader 獲取區塊頭
    GetHeader(hash common.Hash, number uint64) *types.Header
    
    // GetHeaderByNumber 通過高度獲取區塊頭
    GetHeaderByNumber(number uint64) *types.Header
    
    // GetHeaderByHash 通過哈希獲取區塊頭
    GetHeaderByHash(hash common.Hash) *types.Header
}
```

## 四、共識工廠模式

```go
// consensus/factory.go

// ConsensusType 共識類型枚舉
type ConsensusType string

const (
    ConsensusPoW   ConsensusType = "pow"
    ConsensusPoA   ConsensusType = "poa"
    ConsensusPoS   ConsensusType = "pos"
    ConsensusDPoS  ConsensusType = "dpos"
    ConsensusPBFT  ConsensusType = "pbft"
    ConsensusRaft  ConsensusType = "raft"
    ConsensusTendermint ConsensusType = "tendermint"
    ConsensusHotStuff ConsensusType = "hotstuff"
    ConsensusDAG   ConsensusType = "dag"
)

// ConsensusConfig 通用共識配置
type ConsensusConfig struct {
    Type   ConsensusType
    Config interface{} // 具體共識的配置
}

// Factory 共識工廠
type Factory struct {
    builders map[ConsensusType]Builder
}

// Builder 共識構建器
type Builder func(config interface{}) (Engine, error)

// NewFactory 創建共識工廠
func NewFactory() *Factory {
    return &Factory{
        builders: make(map[ConsensusType]Builder),
    }
}

// Register 註冊共識構建器
func (f *Factory) Register(consensusType ConsensusType, builder Builder) {
    f.builders[consensusType] = builder
}

// Create 創建共識引擎
func (f *Factory) Create(config ConsensusConfig) (Engine, error) {
    builder, exists := f.builders[config.Type]
    if !exists {
        return nil, fmt.Errorf("unknown consensus type: %s", config.Type)
    }
    return builder(config.Config)
}

// 使用示例
func InitConsensusFactory() *Factory {
    factory := NewFactory()
    
    // 註冊各種共識
    factory.Register(ConsensusPoW, func(config interface{}) (Engine, error) {
        cfg := config.(*pow.Config)
        return pow.New(cfg), nil
    })
    
    factory.Register(ConsensusPoA, func(config interface{}) (Engine, error) {
        cfg := config.(*clique.Config)
        return clique.New(cfg), nil
    })
    
    factory.Register(ConsensusPoS, func(config interface{}) (Engine, error) {
        cfg := config.(*pos.Config)
        return pos.New(cfg), nil
    })
    
    // ... 註冊其他共識
    
    return factory
}
```

## 五、具體實現示例（Clique PoA）

```go
// consensus/clique/clique.go

type Clique struct {
    config *params.CliqueConfig
    db     ethdb.Database
    
    signer common.Address // 簽名地址
    signFn SignerFn       // 簽名函數
    lock   sync.RWMutex
}

// 實現 Engine 介面
func (c *Clique) VerifyHeader(chain ChainHeaderReader, header *types.Header, seal bool) error {
    // 驗證區塊頭
    // 1. 檢查簽名者是否在授權列表
    // 2. 驗證簽名
    // 3. 檢查時間戳
    return nil
}

func (c *Clique) Seal(chain ChainReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
    // 密封區塊
    // 1. 等待輪到自己出塊
    // 2. 簽名區塊
    // 3. 返回密封後的區塊
    return nil
}

// ... 實現其他方法
```

## 六、鏈配置整合

```go
// params/config.go
type ChainConfig struct {
    ChainID *big.Int

    // 共識配置
    ConsensusType   ConsensusType
    ConsensusConfig interface{}
    
    // 共識切換（支援硬分叉切換共識）
    ConsensusTransitions []ConsensusTransition
}

type ConsensusTransition struct {
    Block      *big.Int
    Type       ConsensusType
    Config     interface{}
}

// 使用示例
config := &ChainConfig{
    ChainID:       big.NewInt(10001),
    ConsensusType: ConsensusPoA,
    ConsensusConfig: &clique.Config{
        Period: 3,
        Epoch:  30000,
    },
    ConsensusTransitions: []ConsensusTransition{
        {
            Block:  big.NewInt(1000000),
            Type:   ConsensusPoS,
            Config: &pos.Config{...},
        },
    },
}
```


## 七、實現路線圖

### 第一階段：基礎共識

* Clique (PoA) - 最簡單，適合開始
* Raft - 簡單的非拜占庭共識
* Simple PoS - 基礎權益證明

### 第二階段：經典共識

* Ethash (PoW) - 以太坊工作量證明
* PBFT - 實用拜占庭容錯
* DPoS - 委託權益證明

### 第三階段：現代共識

* Tendermint - Cosmos 的共識
* HotStuff - Facebook 的 BFT
* Casper FFG - 以太坊 PoS

### 第四階段：創新共識

* DAG-based - IOTA Tangle
* Avalanche - 雪崩共識
* Algorand - 純 PoS

<br>

## 八、測試框架

<br>

```go
// consensus/test/framework.go

type ConsensusTest struct {
    engine Engine
    chain  *core.BlockChain
}

func (ct *ConsensusTest) TestBasicConsensus(t *testing.T) {
    // 測試基本共識功能
}

func (ct *ConsensusTest) TestForkResolution(t *testing.T) {
    // 測試分叉處理
}

func (ct *ConsensusTest) TestByzantineNodes(t *testing.T) {
    // 測試拜占庭節點
}

// 為每個共識運行標準測試套件
func RunConsensusTestSuite(t *testing.T, engine Engine) {
    suite := &ConsensusTest{engine: engine}
    
    t.Run("BasicConsensus", suite.TestBasicConsensus)
    t.Run("ForkResolution", suite.TestForkResolution)
    t.Run("ByzantineNodes", suite.TestByzantineNodes)
    // ...
}
```