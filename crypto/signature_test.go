package crypto

import (
	"crypto/rand"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func randomHash(t *testing.T) []byte {
	t.Helper()
	hash := make([]byte, 32)
	if _, err := rand.Read(hash); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return hash
}

func TestSign(t *testing.T) {

	priv, err := GenerateKey()
	privBytes := FromECDSA(priv)
	assert.Nil(t, err)

	fmt.Println("priv:", priv)

	// 2. 取得公鑰並換算位址
	addr := PubkeyToAddress(priv.PublicKey)
	fmt.Println("addr:", addr)

	// make priv bytes to priv key
	priv_2, _ := ToECDSA(privBytes)
	addr_2 := PubkeyToAddress(priv_2.PublicKey)
	fmt.Println("addr_2:", addr_2)
	assert.Equal(t, addr_2, addr)

	// sign and verify
	hash := randomHash(t)
	// sign 用私鑰簽署
	sig, _ := Sign(hash, priv)
	// public key bytes
	pubBytes := FromECDSAPub(&priv.PublicKey)
	// 驗證 私要簽署的 sign，使用公鑰可以解開
	ok := VerifySignature(pubBytes, hash, sig)
	assert.True(t, ok)

	fmt.Println("公鑰:", pubBytes)
	fmt.Println("私鑰:", privBytes)
	fmt.Println("地址:", addr)
}

// 私鑰簽署的
func TestSign_2(t *testing.T) {
	// 建立一組地址:
	priv, _ := GenerateKey()

	privBytes := FromECDSA(priv)
	pubBytes := FromECDSAPub(&priv.PublicKey)
	addr := PubkeyToAddress(priv.PublicKey)

	fmt.Println("addr:", addr)
	fmt.Println("pubBytes:", pubBytes)
	fmt.Println("privBytes:", privBytes)

	// 使用私鑰簽署一個 sign
	hash := randomHash(t)
	decodePriv, _ := ToECDSA(privBytes)
	sig, _ := Sign(hash, decodePriv)

	// 使用公鑰驗證 sign 是否正確
	ok := VerifySignature(pubBytes, hash, sig)
	assert.True(t, ok)

	// 竄改 sign
	sig[0] = 0
	ok = VerifySignature(pubBytes, hash, sig)
	assert.False(t, ok)
}
