package blockchain

import (
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/core/types"
	"frizo-blockchain/crypto"
	"frizo-blockchain/storage"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func generate_test_txns(t *testing.T) []*types.Transaction {
	wallet := common.HexToAddress("0x521147b68d948f24A341E5C0Da50Fa5AA39A06D7")
	privStr := "cec0e19f7ab452764324201077f9c5b9d4921ca85ba56c1f48af48efb2263b85"

	toAddress := common.HexToAddress("0xA2D969E82524001Cb6a2357dBF5922B04aD2FCD8")

	txn := types.NewTransaction(0, &toAddress, common.Ether, big.NewInt(6000), big.NewInt(10000), []byte{})
	fmt.Println("value:", txn.Value())
	fmt.Println("data:", txn.Data())
	fmt.Println("nonce:", txn.Nonce())
	fmt.Println("hash:", txn.Hash())

	privKey, err := crypto.ImportPrivateKey(privStr)
	assert.Nil(t, err)

	err = txn.SignTx(privKey)
	assert.Nil(t, err)
	sender, err := txn.Sender()
	assert.Nil(t, err)
	fmt.Println("sender:", sender)
	assert.Equal(t, wallet, sender)
	assert.True(t, txn.VerifySignature())
	fmt.Println("txn:", txn)
	return []*types.Transaction{txn}
}

func generate_test_block(t *testing.T, previous *types.Block) *types.Block {
	parentHash := previous.Hash()
	header := types.NewHeader(parentHash, big.NewInt(1),
		uint64(time.Now().UnixMilli()), big.NewInt(1), big.NewInt(1), []byte{}, common.Hash{}, 0)
	fmt.Println("gen header (測試 - 高度): ", header.Number)
	block := types.NewBlock(header, generate_test_txns(t), []*types.Receipt{})
	fmt.Println("gen block (測試 - 高度): ", block.Number())
	return block
}

func test_CreateBlockchain(t *testing.T) *Blockchain {
	chain, err := NewBlockchain(storage.NewMockDatabase(), DefaultGenesis())
	assert.Nil(t, err)

	fmt.Println(chain.CurrentBlock())
	state := chain.CurrentState()
	genesisAddress := common.HexToAddress("0x521147b68d948f24A341E5C0Da50Fa5AA39A06D7")
	balance := state.GetBalance(genesisAddress)
	nonce := state.GetNonce(genesisAddress)
	fmt.Println("genesisAddress:", genesisAddress)
	fmt.Println("balance:", balance)
	fmt.Println("nonce:", nonce)
	return chain
}

func Test_Block(t *testing.T) {
	chain := test_CreateBlockchain(t)
	genesisBlock := chain.CurrentBlock()
	block := generate_test_block(t, genesisBlock)
	assert.NotNil(t, block)
}
