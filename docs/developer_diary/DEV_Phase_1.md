# Phase-1 開發計劃：基礎架構

## 專案根目錄結構

```
frizo-blockchain/
├── cmd/                      # 主程式入口
│   ├── frizo/               # 主節點程式
│   │   └── main.go
│   └── utils/               # 工具程式
│       ├── genesis/         # 創世區塊生成工具
│       └── keygen/          # 密鑰生成工具
│
├── core/                     # 核心區塊鏈邏輯
│   ├── types/               # 核心數據結構
│   │   ├── block.go         # 區塊結構定義
│   │   ├── transaction.go   # 交易結構定義
│   │   ├── receipt.go       # 交易收據
│   │   └── common.go        # 通用類型定義
│   │
│   ├── blockchain/          # 區塊鏈管理
│   │   ├── blockchain.go    # 區塊鏈主邏輯
│   │   ├── genesis.go       # 創世區塊處理
│   │   ├── validation.go    # 區塊驗證邏輯
│   │   └── chain_state.go   # 鏈狀態管理
│   │
│   └── state/               # 狀態管理
│       ├── statedb.go       # 狀態資料庫
│       └── account.go       # 帳戶狀態
│
├── crypto/                   # 加密相關
│   ├── hash.go              # 哈希函數 (Keccak256)
│   ├── signature.go         # 簽名相關 (secp256k1)
│   ├── merkle.go            # 默克爾樹實現
│   └── keys.go              # 密鑰管理
│
├── storage/                  # 存儲層
│   ├── database.go          # 資料庫接口定義
│   ├── leveldb/             # LevelDB 實現
│   │   ├── leveldb.go       # LevelDB 包裝器
│   │   └── batch.go         # 批量操作
│   └── memory/              # 內存資料庫（測試用）
│       └── memdb.go
│
├── network/                  # 網路層
│   ├── p2p/                 # P2P 網路
│   │   ├── peer.go          # 節點管理
│   │   ├── protocol.go      # 通信協議定義
│   │   ├── message.go       # 消息類型定義
│   │   └── server.go        # P2P 服務器
│   │
│   └── discovery/           # 節點發現
│       ├── discovery.go     # 節點發現接口
│       └── static.go        # 靜態節點列表
│
├── common/                   # 通用工具
│   ├── types.go             # 基礎類型定義
│   ├── utils.go             # 工具函數
│   ├── errors.go            # 錯誤定義
│   └── constants.go         # 常量定義
│
├── config/                   # 配置管理
│   ├── config.go            # 配置結構
│   ├── genesis.json         # 創世區塊配置
│   └── default.go           # 默認配置
│
├── api/                      # API 接口（階段1僅基礎）
│   └── types.go             # API 類型定義
│
├── tests/                    # 測試相關
│   ├── unit/                # 單元測試
│   ├── integration/         # 整合測試
│   └── testdata/            # 測試數據
│
├── docs/                     # 文檔
│   ├── architecture/        # 架構設計文檔
│   ├── api/                 # API 文檔
│   └── development/         # 開發指南
│
├── scripts/                  # 腳本
│   ├── build.sh             # 構建腳本
│   ├── test.sh              # 測試腳本
│   └── setup.sh             # 環境設置
│
├── go.mod                    # Go 模組定義
├── go.sum                    # Go 模組校驗
├── Makefile                  # 構建配置
├── README.md                 # 專案說明
├── .gitignore               # Git 忽略文件
└── .github/                  # GitHub 相關
└── workflows/           # CI/CD 配置
```

## 各資料夾詳細功能說明

### 1. **cmd/** - 命令行程式
- **frizo/**: 主節點程式，包含節點啟動、停止、配置載入等功能
- **utils/**: 各種工具程式
- `genesis/`: 生成創世區塊配置
- `keygen/`: 生成節點密鑰對

### 2. **core/** - 區塊鏈核心
- **types/**: 所有核心數據結構
- `Block`: 區塊頭、區塊體、區塊哈希計算
- `Transaction`: 交易結構、簽名驗證
- `Receipt`: 交易執行結果
- **blockchain/**: 區塊鏈邏輯
- 區塊添加、驗證、查詢
- 鏈重組處理
- 創世區塊初始化
- **state/**: 世界狀態管理
- 帳戶餘額、nonce 管理
- 狀態樹（簡化版，為 EVM 預留）

### 3. **crypto/** - 密碼學組件
- `hash.go`: Keccak256 哈希實現
- `signature.go`: ECDSA 簽名和驗證
- `merkle.go`: 默克爾樹構建和驗證
- `keys.go`: 公私鑰生成和管理

### 4. **storage/** - 持久化存儲
- `database.go`: 通用資料庫接口（便於切換不同實現）
- **leveldb/**: LevelDB 具體實現
- 鍵值對存儲
- 批量寫入優化
- **memory/**: 內存資料庫（用於測試）

### 5. **network/** - 網路通信
- **p2p/**: P2P 網路實現
- `peer.go`: 節點連接管理
- `protocol.go`: 自定義協議（區塊同步、交易廣播）
- `message.go`: 消息編解碼
- `server.go`: TCP 服務器
- **discovery/**: 節點發現（階段1使用靜態節點列表）

### 6. **common/** - 通用組件
- 共享的類型定義（Address, Hash 等）
- 工具函數（編碼、解碼等）
- 錯誤類型定義
- 全局常量

### 7. **config/** - 配置管理
- 節點配置（端口、資料目錄等）
- 創世區塊配置
- 網路參數配置

### 8. **tests/** - 測試套件
- **unit/**: 各模組的單元測試
- **integration/**: 跨模組整合測試
- **testdata/**: 測試用的數據文件

## 階段 1 開發重點

### 優先開發順序：
1. **common/** - 基礎類型和工具
2. **crypto/** - 加密組件
3. **core/types/** - 核心數據結構
4. **storage/** - 存儲層
5. **core/blockchain/** - 區塊鏈邏輯
6. **network/** - 基礎網路功能
7. **cmd/** - 命令行介面

### 第一個里程碑目標：
- 能夠創建和驗證區塊
- 能夠持久化存儲區塊
- 兩個節點之間能夠同步區塊

### 測試策略：
- 每個模組都有對應的 `*_test.go` 文件
- 使用 `testify` 套件進行斷言
- 使用 `mockery` 生成 mock 對象
- 目標覆蓋率 80%+