package crypto

import (
	"crypto/rand"
	"fmt"
	"frizo-blockchain/common"
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

func testCreateWallet(t *testing.T) (common.Address, string) {
	addr, privKeyStr, err := CreateWallet()
	assert.Nil(t, err)
	return addr, privKeyStr
}

func TestCreateWallet(t *testing.T) {
	addr, _ := testCreateWallet(t)
	assert.NotNil(t, addr)
	fmt.Println(addr)
}

func TestSign(t *testing.T) {
	addr, privStr := testCreateWallet(t)
	assert.NotNil(t, addr)
	fmt.Println(addr)

	priv, _ := ImportPrivateKey(privStr)
	message := randomHash(t)
	sign, err := SignMessage(priv, message)
	assert.Nil(t, err)
	assert.NotNil(t, sign)
	fmt.Println(sign.HexStr())

	publicKey, err := Ecrecover(message, sign)
	assert.Nil(t, err)
	ok := VerifySignature(publicKey, message, sign)
	assert.True(t, ok)
}

func TestVerifyAndReturnAddress(t *testing.T) {
	addr, privStr := testCreateWallet(t)
	assert.NotNil(t, addr)
	fmt.Println(addr)

	priv, _ := ImportPrivateKey(privStr)
	message := randomHash(t)
	sign, err := SignMessage(priv, message)
	assert.Nil(t, err)
	assert.NotNil(t, sign)
	fmt.Println(sign.HexStr())

	recAddr, err := VerifyAndReturnAddress(message, sign)
	assert.Nil(t, err)
	assert.True(t, recAddr == addr)
}
