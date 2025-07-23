# MPT Enhancement

<br>

---

<br>

## 需要實現的目標

1. 實現哈希計算和節點編碼
2. 完成 Merkle Proof 功能
3. 實現 RLP 編碼
4. 添加並發支持
5. 實現 LevelDB 存儲
6. 添加快照功能

<br>
<br>

## MPT 哈希計算實現指南

### Dirty 標記的作用

Dirty 標記是一個優化機制，用來追蹤哪些節點被修改過，需要重新計算哈希：

```go
// 節點被修改時的流程
修改節點 → 設置 Dirty = true → 計算哈希時檢查 Dirty → 只重算 Dirty 節點
```