package types

import (
	"fmt"
	"frizo-blockchain/crypto"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func Test_Create_Txn(t *testing.T) {
	wallet, _, err := crypto.CreateWallet()
	assert.Nil(t, err)

	txn := NewTransaction(13, &wallet, big.NewInt(10000), big.NewInt(6000), big.NewInt(10000), []byte{})
	fmt.Println("value:", txn.Value())
	fmt.Println("data:", txn.Data())
	fmt.Println("nonce:", txn.Nonce())
	fmt.Println("hash:", txn.Hash())
	fmt.Println("size:", txn.Size())

	assert.Equal(t, big.NewInt(10000), txn.Value())
	assert.Equal(t, []byte{}, txn.Data())
	assert.Equal(t, uint64(13), txn.Nonce())
}

func Test_Sign_Txn_then_Verify(t *testing.T) {
	wallet, privStr, err := crypto.CreateWallet()
	assert.Nil(t, err)

	txn := NewTransaction(13, &wallet, big.NewInt(10000), big.NewInt(6000), big.NewInt(10000), []byte{})
	fmt.Println("value:", txn.Value())
	fmt.Println("data:", txn.Data())
	fmt.Println("nonce:", txn.Nonce())
	fmt.Println("hash:", txn.Hash())

	privKey, err := crypto.ImportPrivateKey(privStr)
	assert.Nil(t, err)

	err = txn.SignTx(privKey)
	assert.Nil(t, err)
	assert.NotNil(t, txn.data.signature)
	fmt.Println("signature:", txn.data.signature)
	sender, err := txn.Sender()
	assert.Nil(t, err)
	fmt.Println("sender:", sender)
	assert.Equal(t, wallet, sender)

	assert.True(t, txn.VerifySignature())

	fmt.Println("txn:", txn)
}
