# 關於以太坊錢包地址以及公私鑰如何實現以及簽署驗證原理

<br>

以太坊的錢包生成籠統一點講是用非對稱式加密技術實現，具體在實現上又有很多門道。以下透過邊實戰邊講解錢包生成規則以及如何利用私鑰簽屬訊息並透過公鑰驗證
首先來看一下如何生成地址:


###  地址生成過程（不可逆）

* 私鑰 → 公鑰：使用橢圓曲線加密算出（secp256k1）
* 公鑰 → 地址：對公鑰進行 Keccak-256 HASH，取後 20 bytes 當作地址表達
  * Keccak-256 是甚麼?
  * Keccak256 is a kind of cryptographic hash function, input any data and output 256 bit (32 bytes).
* 地址格式：加上"0x"前綴，得到42 bytes ("0x" + "20碼") 的以太坊地址
  ex: `0xa087BCe13bc6d82A0Da93Af61259e5Eb04b2CdD1`


<br>


第一步:生成私鑰和公鑰

```golang
func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	// 橢圓曲線原理：
	// secp256k1是比特幣和以太坊使用的橢圓曲線
	// 曲線方程：y² = x³ + 7 (mod p)
	// 私鑰是一個隨機的256位整數
	// 公鑰是私鑰與曲線生成點 G 的標量乘法結果
	
	fmt.Println("🔑 生成橢圓曲線密鑰對...")
	fmt.Println("📘 原理：使用 secp256k1 曲線 y² = x³ + 7")
	fmt.Println("📘 私鑰：256 位隨機數")
	fmt.Println("📘 公鑰：私鑰 × 生成點G")
	
	privateKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成私鑰失敗: %v", err)
	}
	
	fmt.Printf("✅ 私鑰生成成功: %x\n", privateKey.D.Bytes())
	fmt.Printf("✅ 公鑰座標: x=%x, y=%x\n", privateKey.X.Bytes(), privateKey.Y.Bytes())
	
	return privateKey, nil
}
```

第二步:從公鑰生成以太坊地址

```go
func PublicKeyToAddress(publicKey *ecdsa.PublicKey) common.Address {
// 地址生成原理：
// 1. 將公鑰的x,y座標串聯（去掉前綴04）
// 2. 對串聯結果進行 Keccak-256 哈希 (Keccak-256 是可以將任意數量的 bytes 單向 hash 成 256 bit (32 bytes)
// 3. 取 hash 結果的後 20 bytes 作為地址

fmt.Println("\n🏠 生成以太坊地址...")
fmt.Println("📘 原理：Keccak256(公鑰座標串聯) 取後20字節")

// 串聯公鑰的 x,y 座標
pubKeyBytes := elliptic.Marshal(secp256k1.S256(), publicKey.X, publicKey.Y)
// 去掉前綴04
pubKeyBytes = pubKeyBytes[1:]

fmt.Printf("📘 公鑰串聯: %x\n", pubKeyBytes)

// 對公鑰進行 Keccak-256，得到一個 32 bytes 的資料
hash := sha3.NewLegacyKeccak256()
hash.Write(pubKeyBytes)
hashBytes := hash.Sum(nil)

fmt.Printf("📘 Keccak256哈希: %x\n", hashBytes)

// 取後20 bytes 作為地址
address := common.BytesToAddress(hashBytes[12:])
fmt.Printf("✅ 以太坊地址: %s\n", address.Hex())

return address
}
```

<br>
<br>

###  簽屬與驗證

以上我們就可以得到一組可用的 WalletAddress + Public Key + Private Key 了，接下來看一下簽屬與驗證。

<br>

使用私鑰簽署消息後，會得到一組名為 __VRS__ 且長度為 64 的 bytes。

VRS 組成:

* V (Recovery ID)

  作用：恢復識別碼，用於確定正確的公鑰 `VRS[64]`
  值域：通常是 27 或 28
  用途：因為橢圓曲線簽名可能對應兩個不同的公鑰，V 幫助確定使用哪一個

* R (簽名的 R 值)

  作用：ECDSA簽名的第一部分 `VRS[:32]`
  性質：32 bytes 的 big.Int
  來源：橢圓曲線上隨機點的 x 坐標

* S (簽名的 S 值)

  作用：ECDSA簽名的第二部分 `VRS[32:64]`
  性質：32 bytes 的 big.Int
  限制：必須在特定範圍內以防止簽名延展性攻擊

<br>

簽署實現:
```go
func SignMessage(privateKey *ecdsa.PrivateKey, message []byte) ([]byte, error) {
// 簽屬原理：
// 1. 對消息進行 Keccak256
// 2. 使用ECDSA算法簽名
// 3. 生成 R,S 值和 Recover-ID

fmt.Println("\n✍️ 簽署消息...")
fmt.Printf("📘 原始消息: %s\n", string(message))

// 對消息進行哈希
messageHash := crypto.Keccak256(message)
fmt.Printf("📘 消息 HASH: %x\n", messageHash)

// 簽屬
signature, err := crypto.Sign(messageHash, privateKey)
if err != nil {
return nil, fmt.Errorf("簽屬失敗: %v", err)
}

// 解析VRS
r := new(big.Int).SetBytes(signature[:32])
s := new(big.Int).SetBytes(signature[32:64])
v := signature[64]

fmt.Printf("📘 簽屬組成:\n")
fmt.Printf("   R: %x\n", r.Bytes())
fmt.Printf("   S: %x\n", s.Bytes())
fmt.Printf("   V: %d\n", v)
fmt.Printf("✅ 完整簽屬: %x\n", signature)

return signature, nil
}
```

<br>

驗證簽名:
```go
func VerifySignature(publicKey *ecdsa.PublicKey, message, signature []byte) bool {
// 驗證原理：
// 1. 重新計算消息哈希
// 2. 使用ECDSA驗證算法
// 3. 檢查簽名是否由對應私鑰生成

fmt.Println("\n🔍 驗證簽名...")

messageHash := crypto.Keccak256(message)

// 移除 Recover-ID（64 bytes RSV 的最後一位）
signatureNoRecoveryID := signature[:len(signature)-1]

valid := crypto.VerifySignature(
crypto.FromECDSAPub(publicKey),
messageHash,
signatureNoRecoveryID,
)

fmt.Printf("📘 驗證結果: %t\n", valid)
return valid
}
```

###  回推公鑰

接下來是以太坊核心的交易驗證機制，如何透過消息與簽名回推公鑰。

我們現在知道如果得知了公鑰就可以對公鑰進行 Keccak-256 HASH，取後 20 bytes 後算出該公鑰對應的地址，但是我們無法從地址回推公鑰是甚麼，
畢竟光是截斷後 20 bytes 這一點就不可能有機會再回推了。

那我們如何能夠得知一筆交易究竟是不是某個地址簽署的呢?


看實做: 從簽名恢復公鑰 (ecrecover)
```go
func Ecrecover(message, signature []byte) (*ecdsa.PublicKey, error) {
// ecrecover 原理：
// 1. 從簽名的 r,s,v 值重建簽名點
// 2. 使用數學計算恢復原始公鑰

fmt.Println("\n🔄 恢復公鑰 (ecrecover)...")
fmt.Println("📘 原理：從簽名的 VRS 值數學計算恢復公鑰")

// hash 消息
messageHash := crypto.Keccak256(message)
// 恢復公鑰
publicKey, err := crypto.SigToPub(messageHash, signature)
if err != nil {
return nil, fmt.Errorf("恢復公鑰失敗: %v", err)
}

fmt.Printf("✅ 恢復的公鑰: x=%x, y=%x\n", publicKey.X.Bytes(), publicKey.Y.Bytes())

return publicKey, nil
}
```

當用戶發送交易時，會用私鑰對交易進行簽名。以太坊可以從簽名中恢復公鑰：

* 每筆交易都包含簽名信息（v, r, s）
* 節點可以從這些信息恢復發送者的公鑰
* 然後驗證恢復的公鑰對應的地址是否與交易發送者一致

這種設計確保了：

* 地址本身不洩露公鑰信息
* 只有在用戶主動簽名交易時才會暴露公鑰
* 提供了良好的隱私保護