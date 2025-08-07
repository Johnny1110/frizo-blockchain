package tests

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

func Test_BlockStore(t *testing.T) {
	config := &storage.DatabaseConfig{
		DataDir:   "/Users/johnny.wang/projects",
		Cache:     2,
		Handles:   3,
		Namespace: "test",
	}
	chainDB, err := storage.NewChainDatabaseInMemory(config)
	assert.Nil(t, err)
	blockStore := storage.NewBlockStore(chainDB)

	writeBlock, _ := generate_test_block(t)
	fmt.Println(">>>>>>>>>>>> writeBlock: ", writeBlock)
	fmt.Println("writeBlock Header: ", writeBlock.Header())
	fmt.Println("writeBlock Number: ", writeBlock.Number())
	fmt.Println("writeBlock HASH: ", writeBlock.Hash())
	fmt.Println("writeBlock GasUsed: ", writeBlock.GasUsed())
	for i, tx := range writeBlock.Transactions() {
		fmt.Println("index:", i, "tx:", tx)
	}

	err = blockStore.Write(writeBlock)
	assert.Nil(t, err)

	blockHash := writeBlock.Hash()
	blockNumber := writeBlock.NumberU64()

	readBlock, err := blockStore.ReadBlock(blockHash, blockNumber)

	assert.Nil(t, err)
	assert.NotNil(t, readBlock)
	fmt.Println(">>>>>>>>>>>> readBlock: ", readBlock)
	fmt.Println("readBlock Header: ", readBlock.Header())
	fmt.Println("readBlock Number: ", readBlock.Number())
	fmt.Println("readBlock HASH: ", readBlock.Hash())
	fmt.Println("readBlock GasUsed: ", readBlock.GasUsed())
	for i, tx := range readBlock.Transactions() {
		fmt.Println("index:", i, "tx:", tx)
	}

	assert.Equal(t, writeBlock.Hash(), readBlock.Hash())

	fmt.Println("****************** Receipts ******************")
	for idx, rec := range readBlock.Receipts() {
		fmt.Println("index:", idx, "receipt:", rec)
	}
}

func generate_test_block(t *testing.T) (*types.Block, types.Receipts) {
	txns := generate_test_txns(t)
	recis := generate_text_recis(t)
	header := types.NewHeader(common.Hash{}, big.NewInt(100), uint64(time.Now().Nanosecond()),
		big.NewInt(50000), big.NewInt(50000), []byte{}, common.Hash{}, 0)

	block := types.NewBlock(header, txns, recis)
	fmt.Println("generate mock block>: ", block)
	return block, recis
}

func generate_text_recis(t *testing.T) types.Receipts {
	var receipts types.Receipts

	rec_1 := &types.Receipt{
		PostState:         common.Keccak256Hash([]byte("State")),
		Status:            1,
		CumulativeGasUsed: big.NewInt(2500),
		Bloom:             types.CreateBloom(nil),
		Logs:              make([]*types.Log, 0),
		TxHash:            common.Keccak256Hash([]byte("TxHash_1")),
		GasUsed:           big.NewInt(2500),
		BlockHash:         common.Keccak256Hash([]byte("BlockHash_1")),
		BlockNumber:       big.NewInt(100),
		TransactionIndex:  1,
	}

	rec_2 := &types.Receipt{
		PostState:         common.Keccak256Hash([]byte("State")),
		Status:            1,
		CumulativeGasUsed: big.NewInt(5000),
		Bloom:             types.CreateBloom(nil),
		Logs:              make([]*types.Log, 0),
		TxHash:            common.Keccak256Hash([]byte("TxHash_2")),
		GasUsed:           big.NewInt(2500),
		BlockHash:         common.Keccak256Hash([]byte("BlockHash_1")),
		BlockNumber:       big.NewInt(100),
		TransactionIndex:  2,
	}

	receipts = append(receipts, rec_1, rec_2)
	return receipts
}

func generate_test_txns(t *testing.T) types.Transactions {
	wallet := common.HexToAddress("0x521147b68d948f24A341E5C0Da50Fa5AA39A06D7")
	privStr := "cec0e19f7ab452764324201077f9c5b9d4921ca85ba56c1f48af48efb2263b85"

	toAddress := common.HexToAddress("0xA2D969E82524001Cb6a2357dBF5922B04aD2FCD8")

	txn_1 := types.NewTransaction(1, &toAddress, common.Ether, big.NewInt(6000), big.NewInt(10000), []byte{})
	fmt.Println("value:", txn_1.Value())
	fmt.Println("data:", txn_1.Data())
	fmt.Println("nonce:", txn_1.Nonce())
	fmt.Println("hash:", txn_1.Hash())

	privKey, err := crypto.ImportPrivateKey(privStr)
	assert.Nil(t, err)

	err = txn_1.SignTx(privKey)
	assert.Nil(t, err)
	sender, err := txn_1.Sender()
	assert.Nil(t, err)
	fmt.Println("sender:", sender)
	assert.Equal(t, wallet, sender)
	assert.True(t, txn_1.VerifySignature())
	fmt.Println("txn:", txn_1)

	txn_2 := types.NewTransaction(2, &toAddress, common.Ether, big.NewInt(6000), big.NewInt(10000), []byte{})
	fmt.Println("value:", txn_2.Value())
	fmt.Println("data:", txn_2.Data())
	fmt.Println("nonce:", txn_2.Nonce())
	fmt.Println("hash:", txn_2.Hash())

	assert.Nil(t, err)

	err = txn_2.SignTx(privKey)
	assert.Nil(t, err)
	assert.Nil(t, err)
	fmt.Println("sender:", sender)
	assert.Equal(t, wallet, sender)
	assert.True(t, txn_2.VerifySignature())
	fmt.Println("txn:", txn_2)

	return []*types.Transaction{txn_1, txn_2}
}
