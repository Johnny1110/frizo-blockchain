# RLP (Recursive Length Prefix) 編碼詳解

<br>

---

<br>

RLP 是以太坊專門設計的編碼方法，用於序列化任意嵌套的二進制數據。它是以太坊中所有數據序列化的基礎。

<br>

## 核心特點：

* 簡單：規則少，易於實現
* 確定性：相同輸入總是產生相同輸出
* 自描述：包含長度信息，易於解碼
* 無類型：只編碼結構，不編碼類型


## RLP 編碼規則

RLP 只能編碼兩種數據類型：

* []byte
* [][]byte

<br>

### 單字節編碼規則

```go
// 規則 1：單個字節，值在 [0x00, 0x7f]
if len(input) == 1 && input[0] <= 0x7f {
    return input  // 直接返回該字節
}

// 例子：
'a' (0x61) → 0x61
'\x15' → 0x15
```

<br>

### 字符串編碼規則

```go
// 規則 2：0-55 字節的字符串
if len(input) <= 55 {
    return []byte{0x80 + len(input)} + input
}

// 例子：
"dog" → 0x83 + "dog" = [0x83, 0x64, 0x6f, 0x67]
//      0x80+3=0x83

// 規則 3：超過 55 字節的字符串
if len(input) > 55 {
    lenBytes := intToBytes(len(input))  // 長度的大端表示
    return []byte{0xb7 + len(lenBytes)} + lenBytes + input
}

// 例子：1024 字節的字符串
// 1024 = 0x0400 (2 bytes)
// 前綴 = 0xb7 + 2 = 0xb9
// 結果 = [0xb9, 0x04, 0x00] + data
```

<br>

### 列表編碼規則

```go
// 規則 4：列表總長度 0-55 字節
if totalLen <= 55 {
    return []byte{0xc0 + totalLen} + encodedItems
}

// 例子：
["cat", "dog"] → 0xc8 + 0x83"cat" + 0x83"dog"
//               0xc0+8=0xc8

// 規則 5：列表總長度超過 55 字節
if totalLen > 55 {
    lenBytes := intToBytes(totalLen)
    return []byte{0xf7 + len(lenBytes)} + lenBytes + encodedItems
}
```

<br>
<br>
<br>
<br>

## 編碼示例分析

<br>

### 示例 1：編碼字符串 "dog"

```go
輸入: "dog"
步驟:
1. 長度 = 3 (< 55)
2. 前綴 = 0x80 + 3 = 0x83
3. 結果 = [0x83, 0x64, 0x6f, 0x67]
         前綴   'd'   'o'   'g'
```

<br>

### 示例 2：編碼列表 ["cat", "dog"]

```go
輸入: ["cat", "dog"]
步驟:
1. 編碼 "cat" → [0x83, 0x63, 0x61, 0x74]
2. 編碼 "dog" → [0x83, 0x64, 0x6f, 0x67]
3. 連接 → [0x83, 0x63, 0x61, 0x74, 0x83, 0x64, 0x6f, 0x67]
4. 總長度 = 8
5. 列表前綴 = 0xc0 + 8 = 0xc8
6. 結果 = [0xc8, 0x83, 0x63, 0x61, 0x74, 0x83, 0x64, 0x6f, 0x67]
```


<br>

### 示例 3：編碼嵌套結構 [[], [[]], [[], [[]]]]

```go
輸入: [[], [[]], [[], [[]]]]
步驟:
1. [] → [0xc0]  (空列表)
2. [[]] → [0xc1, 0xc0]
3. [[], [[]]] → [0xc4, 0xc0, 0xc2, 0xc0]
4. 最終 → [0xc7, 0xc0, 0xc1, 0xc0, 0xc4, 0xc0, 0xc2, 0xc0]
```

