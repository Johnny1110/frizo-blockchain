# Patricia Tree（Practical Algorithm To Retrieve Information Coded In Alphanumeric）


<br>

---

<br>

## 📌 1. Patricia Tree 是什麼？

Patricia Tree 是一種壓縮版的 Trie 樹，它是一種用來儲存字串集合（通常是二進位字串）的資料結構，常被用在：

* 區塊鏈（如 Ethereum 的 Merkle Patricia Tree）

* 路由表（IP prefix matching）

* 文字搜尋系統

<br>
<br>

## 🧠 2. 為什麼需要 Patricia Tree？

普通 Trie 在每個節點只儲存一個字元，但如果你有大量相同開頭的 key，就會很浪費空間。

Patricia Tree 的目的就是：

* 將只有單一路徑的節點壓縮成一個節點，提升空間效率。

<br>
<br>

## 🧩 3. 結構與特性

* 資料儲存方式

    * 每個節點不只儲存一個字元，而是儲存一段字串（prefix path）。

    * 路徑以共享前綴來合併節點。

*  節點欄位（簡化版）

    ```go
    type Node struct {
        label string     // prefix path
        isLeaf bool      // 是否為終點
        children map[rune]*Node  // 依據下一個字元路徑的子節點
    }
    ```


* 舉例

    假設要儲存以下字串集合：
    
    ```
    ["apple", "app", "apt", "bat", "bath"]
    ```
    
    對應的 Patricia Tree 為：
    
    ```
    ┌── "app" ──► [le]
    │
    root ──┤
    └── "apt"
    └── "bat" ──► [h]
    ```
    
    >> 這裡 "app" 是共享前綴，"apple" 與 "app" 共用 "app"，但 "le" 是 apple 的延伸。

<br>
<br>

## 🛠️ 4. 插入邏輯（簡化版）

插入一個字串時：

* 從 root 開始，尋找與目前節點 label 的最長共同前綴。

* 若完全相同，則往下走。

* 若有部分相同：

    * 將當前節點分裂成兩個節點

    * 插入新的 label 到新節點

* 若沒有相同，則建立新的子節點。

<br>
<br>

## 💡 5. 使用場景

* Ethereum（以太坊） 中使用「Merkle Patricia Trie」來：

  * 儲存帳戶狀態（State）

  * 儲存交易（Transaction）

* IP 路由表

* DNS

* 字首查詢（Autocomplete）