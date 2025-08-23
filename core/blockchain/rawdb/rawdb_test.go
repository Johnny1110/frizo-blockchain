package rawdb

import (
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/core/types"
	"frizo-blockchain/storage"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestCanonicalHash(t *testing.T) {
	db := storage.NewInMemoryKVStore()

	hash := common.HexToHash("0x1234567890abcdef")
	number := uint64(100)

	// Write
	err := WriteCanonicalHash(db, number, hash)
	assert.NoError(t, err)

	// Read
	got := ReadCanonicalHash(db, number)
	assert.Equal(t, hash, got)

	// Delete
	err = DeleteCanonicalHash(db, number)
	assert.NoError(t, err)

	// Read after delete
	got = ReadCanonicalHash(db, number)
	assert.Equal(t, common.Hash{}, got)
}

func TestHeaderStorage(t *testing.T) {
	db := storage.NewInMemoryKVStore()

	// Create test header
	header := &types.Header{
		ParentHash:      common.HexToHash("0x1111"),
		Number:          big.NewInt(100),
		TxHashRoot:      common.HexToHash("0x1112"),
		ReceiptHashRoot: common.HexToHash("0x1113"),
		StateHashRoot:   common.HexToHash("0x1114"),
		Timestamp:       uint64(time.Now().Unix()),
		GasLimit:        big.NewInt(8000000),
		GasUsed:         big.NewInt(4000000),
	}

	// Write
	err := WriteHeader(db, header)
	assert.NoError(t, err)

	// Read
	got := ReadHeader(db, header.Hash(), header.Number.Uint64())
	assert.NotNil(t, got)
	assert.Equal(t, header.Hash(), got.Hash())
	assert.Equal(t, header.Number.Uint64(), got.Number.Uint64())
	assert.Equal(t, header.ParentHash, got.ParentHash)
	assert.Equal(t, header.TxHashRoot, got.TxHashRoot)
	assert.Equal(t, header.ReceiptHashRoot, got.ReceiptHashRoot)
	assert.Equal(t, header.StateHashRoot, got.StateHashRoot)

	// Delete
	err = DeleteHeader(db, header.Hash(), header.Number.Uint64())
	assert.NoError(t, err)

	// Read after delete
	got = ReadHeader(db, header.Hash(), header.Number.Uint64())
	assert.Nil(t, got)
}

func TestBlockStorage(t *testing.T) {
	db := storage.NewInMemoryKVStore()

	// Create test block
	header := &types.Header{
		ParentHash: common.HexToHash("0x1111"),
		Number:     big.NewInt(100),
		Timestamp:  uint64(time.Now().Unix()),
		GasLimit:   big.NewInt(8000000),
		GasUsed:    big.NewInt(4000000),
	}

	addr := common.HexToAddress("0x521147b68d948f24A341E5C0Da50Fa5AA39A06D7")

	tx1 := types.NewTransaction(
		1,
		&addr,
		big.NewInt(1000),
		big.NewInt(21000),
		big.NewInt(1),
		[]byte{},
	)

	block := types.NewBlock(header, []*types.Transaction{tx1}, []*types.Receipt{})

	fmt.Println("origin block:", block)
	fmt.Println("origin block hash:", block.Hash())
	for _, txn := range block.Transactions() {
		fmt.Println("origin txn:", txn)
	}

	// Write
	err := WriteBlock(db, block)
	assert.NoError(t, err)

	// Read
	got := ReadBlock(db, block.Hash(), block.NumberU64())

	fmt.Println("========================================================================")

	fmt.Println("read block:", got)
	fmt.Println("read block hash:", got.Hash())
	for _, txn := range got.Transactions() {
		fmt.Println("read txn:", txn)
	}

	assert.NotNil(t, got)
	assert.Equal(t, block.Hash(), got.Hash())
	assert.Equal(t, len(block.Transactions()), len(got.Transactions()))
	fmt.Println("read block:", got)

	// Delete
	err = DeleteBlock(db, block.Hash(), block.NumberU64())
	assert.NoError(t, err)

	// Read after delete
	got = ReadBlock(db, block.Hash(), block.NumberU64())
	assert.Nil(t, got)
}

func TestReceiptsStorage(t *testing.T) {
	db := storage.NewInMemoryKVStore()

	hash := common.HexToHash("0x1234")
	number := uint64(100)

	// Create test receipts
	receipts := types.Receipts{
		&types.Receipt{
			ContractAddress:   common.HexToAddress(""),
			Status:            1,
			CumulativeGasUsed: big.NewInt(21000),
			TxHash:            common.HexToHash("0xabc"),
			GasUsed:           big.NewInt(21000),
		},
		&types.Receipt{
			Status:            0,
			CumulativeGasUsed: big.NewInt(42000),
			TxHash:            common.HexToHash("0xdef"),
			GasUsed:           big.NewInt(21000),
		},
	}

	// Write
	err := WriteReceipts(db, hash, number, receipts)
	assert.NoError(t, err)

	// Read
	got := ReadReceipts(db, hash, number)
	fmt.Println("got size:", len(got))
	for _, rec := range got {
		fmt.Println("receipt:", rec)
	}

	assert.NotNil(t, got)
	assert.Equal(t, len(receipts), len(got))
	assert.Equal(t, receipts[0].TxHash, got[0].TxHash)

	// Delete
	err = DeleteReceipts(db, hash, number)
	assert.NoError(t, err)

	// Read after delete
	got = ReadReceipts(db, hash, number)
	assert.Nil(t, got)
}

func TestTxLookup(t *testing.T) {
	db := storage.NewInMemoryKVStore()

	txHash := common.HexToHash("0xabc123")
	blockHash := common.HexToHash("0xdef456")
	blockIndex := uint64(100)
	txIndex := uint64(5)

	// Write
	err := WriteTxLookupEntry(db, txHash, blockHash, blockIndex, txIndex)
	assert.NoError(t, err)

	// Read
	gotBlockHash, gotBlockIndex, gotTxIndex := ReadTxLookupEntry(db, txHash)
	assert.Equal(t, blockHash, gotBlockHash)
	assert.Equal(t, blockIndex, gotBlockIndex)
	assert.Equal(t, txIndex, gotTxIndex)

	// Delete
	err = DeleteTxLookupEntry(db, txHash)
	assert.NoError(t, err)

	// Read after delete
	gotBlockHash, _, _ = ReadTxLookupEntry(db, txHash)
	assert.Equal(t, common.Hash{}, gotBlockHash)
}
