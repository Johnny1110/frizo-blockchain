# Frizo Blockchain Phase-1 開發計劃

<br>

---

<br>

## 已完成的主要功能

###  密碼學相關功能

* 公私鑰 (Wallet) 生成: crypto/keys.go
* 簽署與驗證: crypto/signature.go

<br>

### 區塊鏈核心資料結構定義

* core/types/block.go
* core/types/receipt.go
* core/types/transaction.go

<br>

### Merkle Tree 與 MPT

區塊的資料結構核心實現

* Merkle Tree: /trie/merkle.go
* Modified Merkle Patricia Tree (MPT):
    * /trie/mpt.go (樹的主體，已實現 rlp 編碼)
    * /trie/mpt_proof.go (證明相關功能，生成與驗證 proof 皆以完備)
    * /trie/mpt_tool.go (debug 工具)

<br>

### Storage 相關

state block 等持久化相關功能

* DB 介面: /storage/interfaces/interfaces.go
* levelDB 實現: /storage/leveldb/leveldb.go
* in-memory 實現(供單元測試使用): /storage/leveldb/memdb.go
* DB 實例化方法: /storage/leveldb/database.go
* block 存取轉接類別: /storage/block_store.go, /storage/schema.go
* state 存取轉接類別: /storage/state_store.go

<br>
<br>
<br>
<br>

## 未來計劃

### 一、MPT 與 state 整合

* core/simple_state 與實際 MPT 整合．變成完全體 state 管理，實現 snapshot. commit 與 revert 功能．

<br>
<br>

### 二、區塊鏈核心實現

* 重構 core (blockchain 核心相關) 目標是達成 Phase-1 驗收標準的完成度(完整實現轉帳，節點共識並同步區塊)
* 整合 block，state 與 storage 的功能，提供持久化儲存與恢復
* 完成 txnPool 功能，目標是達成 Phase-1 驗收標準的完成度
* 完成簽署交易 -> 入池 -> 收集交易 -> 打包交易生產區塊 -> 驗證區塊 -> 執行交易帳戶狀態轉化 -> 完成持久化  整套流程的整合測試


<br>
<br>

### 三、實現共識算法

* 未知（目前只知道拜占庭將軍共識概念，不具備現代化區塊鏈共識算法概念經驗）
* 支援 POS，由共識選取出指定節點產出區塊(預計先以12秒產一個區塊為階段性目標)，可提供節點運行者 POS "挖礦"

<br>
<br>

### 四、網路層基礎

* 1. 定義消息類型

* 2. 實現節點管理

* 3. 實現基本的同步協議
    * 握手協議
    * 區塊請求/響應
    * 交易廣播

<br>
<br>

### 五、測試計劃

* 1. 單元測試（每個模組）
    * MPT 操作測試
    * 區塊驗證測試
    * 存儲層測試
    * 網路消息測試

* 2. 整合測試
    * 創建測試鏈
    * 多節點同步測試
    * 狀態一致性測試


<br>
<br>

### 六、Phase-1 發表

* 開啟 3 台 GCP 虛擬機，並啟動三台節點，穩定運行 1 個月並不斷測試交易與挖礦功能，保證交易與產塊穩定．
* 嘗試修改第四台節點，攻擊 frizo-blockchain．用偽造交易等手法，驗證交易安全性以及節點資料同步穩定性．