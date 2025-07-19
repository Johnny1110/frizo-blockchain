# insertIntoExtension() 規範與實作指南

## 🎯 EXTENSION 節點特性回顧

- **用途**: 路徑壓縮，存儲共同的路徑前綴
- **結構**: `[共享路徑, 子節點]`
- **特點**: 只有一個子節點 (`Child`)，沒有自己的值

## 📋 核心處理邏輯

### 步驟 1: 計算公共前綴
```go
commonLen := commonPrefixLen(ext.Path, path)
```

### 步驟 2: 根據公共前綴長度分情況處理

---

## 🔄 情況分類與處理

### 情況 1: 路徑完全匹配擴展節點路徑
**條件**: `commonLen == len(ext.Path)`

**說明**: 插入路徑的前綴完全匹配擴展節點的路徑

**處理**:
- 移除已匹配的路徑部分
- 向子節點遞歸插入剩餘路徑
- 更新子節點引用

```go
if commonLen == len(ext.Path) {
// 計算剩餘路徑
remainingPath := path[commonLen:]

// 遞歸向子節點插入
newChild, err := t.insert(ext.Child, remainingPath, value)
if err != nil {
return nil, err
}

// 更新子節點
ext.Child = newChild
return ext, nil
}
```

---

### 情況 2: 需要分裂擴展節點
**條件**: `commonLen < len(ext.Path)`

**說明**: 插入路徑與擴展節點路徑只有部分匹配，需要分裂擴展節點

#### 子情況 2.1: 插入路徑在公共前綴後結束
**條件**: `commonLen == len(path)`

**結構變化**:
```
原來: EXT[abcd] -> Child
分裂後: EXT[ab] -> BRANCH -> EXT[cd] -> Child
-> VALUE = value
```

**處理邏輯**:
```go
if commonLen == len(path) {
// 創建新的分支節點，存儲插入值
newBranch := &MPTNode{
NodeType: BRANCH,
Value:    value,  // 值存在分支節點
Dirty:    true,
}

// 創建新的擴展節點存儲剩餘路徑
if len(ext.Path[commonLen+1:]) > 0 {
// 還有剩餘路徑，創建新擴展節點
remainingExt := &MPTNode{
NodeType: EXTENSION,
Path:     ext.Path[commonLen+1:],
Child:    ext.Child,
Dirty:    true,
}
branchIndex := ext.Path[commonLen]
newBranch.Children[branchIndex] = remainingExt
} else {
// 沒有剩餘路徑，直接連接原子節點
branchIndex := ext.Path[commonLen]
newBranch.Children[branchIndex] = ext.Child
}

// 更新當前擴展節點或創建新的
if commonLen > 0 {
ext.Path = ext.Path[:commonLen]
ext.Child = newBranch
return ext, nil
} else {
return newBranch, nil
}
}
```

#### 子情況 2.2: 插入路徑在公共前綴後繼續
**條件**: `commonLen < len(path)`

**結構變化**:
```
原來: EXT[abcd] -> Child
插入: path[abef] = value
分裂後: EXT[ab] -> BRANCH -> EXT[cd] -> Child
-> EXT[ef] -> LEAF[value]
```

**處理邏輯**:
```go
if commonLen < len(path) {
// 創建新的分支節點
newBranch := &MPTNode{
NodeType: BRANCH,
Dirty:    true,
}

// 處理原擴展節點的剩餘部分
if len(ext.Path[commonLen+1:]) > 0 {
// 創建新擴展節點存儲原路徑剩餘部分
remainingExt := &MPTNode{
NodeType: EXTENSION,
Path:     ext.Path[commonLen+1:],
Child:    ext.Child,
Dirty:    true,
}
branchIndex := ext.Path[commonLen]
newBranch.Children[branchIndex] = remainingExt
} else {
// 沒有剩餘路徑，直接連接原子節點
branchIndex := ext.Path[commonLen]
newBranch.Children[branchIndex] = ext.Child
}

// 處理新插入路徑的剩餘部分
newPathRemaining := path[commonLen+1:]
newBranchIndex := path[commonLen]

if len(newPathRemaining) > 0 {
// 創建新的葉子節點
newLeaf := &MPTNode{
NodeType: LEAF,
Path:     newPathRemaining,
Value:    value,
Dirty:    true,
}
newBranch.Children[newBranchIndex] = newLeaf
} else {
// 新路徑剛好在分支節點結束
newBranch.Value = value
}

// 更新當前擴展節點或返回分支節點
if commonLen > 0 {
ext.Path = ext.Path[:commonLen]
ext.Child = newBranch
return ext, nil
} else {
return newBranch, nil
}
}
```

---

## 🎭 完整實作模板

```go
func (t *ModifiedMerklePatriciaTree) insertIntoExtension(ext *MPTNode, path []byte, value []byte) (*MPTNode, error) {
// 計算公共前綴長度
commonLen := commonPrefixLen(ext.Path, path)

// ---------------------------------------------------------------------------
// S-1: 路徑完全匹配擴展節點路徑
// ---------------------------------------------------------------------------
if commonLen == len(ext.Path) {
// 向子節點遞歸插入剩餘路徑
remainingPath := path[commonLen:]
newChild, err := t.insert(ext.Child, remainingPath, value)
if err != nil {
return nil, err
}
ext.Child = newChild
return ext, nil
}

// ---------------------------------------------------------------------------
// S-2: 需要分裂擴展節點
// ---------------------------------------------------------------------------

// 創建新的分支節點
newBranch := &MPTNode{
NodeType: BRANCH,
Dirty:    true,
}

// 2.1: 插入路徑在公共前綴後結束
if commonLen == len(path) {
// 值存儲在分支節點
newBranch.Value = value

// 處理原擴展節點剩餘路徑
if len(ext.Path[commonLen+1:]) > 0 {
// 還有剩餘路徑，創建新擴展節點
remainingExt := &MPTNode{
NodeType: EXTENSION,
Path:     ext.Path[commonLen+1:],
Child:    ext.Child,
Dirty:    true,
}
branchIndex := ext.Path[commonLen]
newBranch.Children[branchIndex] = remainingExt
} else {
// 沒有剩餘路徑，直接連接原子節點
branchIndex := ext.Path[commonLen]
newBranch.Children[branchIndex] = ext.Child
}
} else {
// 2.2: 插入路徑在公共前綴後繼續

// 處理原擴展節點剩餘路徑
if len(ext.Path[commonLen+1:]) > 0 {
remainingExt := &MPTNode{
NodeType: EXTENSION,
Path:     ext.Path[commonLen+1:],
Child:    ext.Child,
Dirty:    true,
}
branchIndex := ext.Path[commonLen]
newBranch.Children[branchIndex] = remainingExt
} else {
branchIndex := ext.Path[commonLen]
newBranch.Children[branchIndex] = ext.Child
}

// 處理新插入路徑剩餘部分
newPathRemaining := path[commonLen+1:]
newBranchIndex := path[commonLen]

if len(newPathRemaining) > 0 {
// 創建新葉子節點
newLeaf := &MPTNode{
NodeType: LEAF,
Path:     newPathRemaining,
Value:    value,
Dirty:    true,
}
newBranch.Children[newBranchIndex] = newLeaf
} else {
// 新路徑在分支節點結束
newBranch.Value = value
}
}

// 根據公共前綴長度決定返回結果
if commonLen > 0 {
// 有公共前綴，更新當前擴展節點
ext.Path = ext.Path[:commonLen]
ext.Child = newBranch
return ext, nil
} else {
// 沒有公共前綴，直接返回分支節點
return newBranch, nil
}
}
```

---

## 🧪 測試用例

### 測試 1: 完全匹配
```go
// 原樹: EXT["abc"] -> LEAF["def"]
// 插入: key="abcxyz", value="newvalue"
// 結果: EXT["abc"] -> BRANCH -> LEAF["def"] (原值)
//                          -> LEAF["xyz"] (新值)
```

### 測試 2: 部分匹配 + 路徑結束
```go
// 原樹: EXT["abcd"] -> LEAF["value1"]
// 插入: key="ab", value="value2"
// 結果: EXT["ab"] -> BRANCH -> EXT["cd"] -> LEAF["value1"]
//                          -> VALUE="value2"
```

### 測試 3: 部分匹配 + 路徑繼續
```go
// 原樹: EXT["abcd"] -> LEAF["value1"]
// 插入: key="abef", value="value2"
// 結果: EXT["ab"] -> BRANCH -> EXT["cd"] -> LEAF["value1"]
//                          -> EXT["ef"] -> LEAF["value2"]
```

### 測試 4: 無公共前綴
```go
// 原樹: EXT["abc"] -> LEAF["value1"]
// 插入: key="def", value="value2"
// 結果: BRANCH -> EXT["abc"] -> LEAF["value1"]
//              -> EXT["def"] -> LEAF["value2"]
```

---

## ⚠️ 注意事項

1. **邊界檢查**: 確保數組索引不越界
2. **節點複用**: 盡量複用現有節點，避免不必要的複製
3. **錯誤處理**: 處理遞歸插入可能的錯誤
4. **路徑壓縮**: 正確處理路徑的切分和重組
5. **子節點更新**: 確保正確更新擴展節點的子節點引用