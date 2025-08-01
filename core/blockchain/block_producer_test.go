package blockchain

import (
	"fmt"
	"frizo-blockchain/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func Test_ProduceBlock(t *testing.T) {
	test_txns := generate_test_txns(t)
	bc := test_CreateBlockchain(t)

	txPool := NewTxPool(bc)

	for _, txn := range test_txns {
		err := txPool.AddTx(txn)
		assert.Nil(t, err)
	}

	processor := &StateProcessor{bc}

	bp := &BlockProducer{
		bc, txPool, processor,
	}

	block, err := bp.ProduceBlock()
	assert.Nil(t, err)
	fmt.Println("produce block: ================================================")
	fmt.Println(block)

	err = bc.InsertBlock(block)
	debugBlockchain(bc)
	assert.Nil(t, err)

	toAddress := common.HexToAddress("0xa2d969e82524001cb6a2357dbf5922b04ad2fcd8")
	expectETHBalance := big.NewInt(101)
	assert.Equal(t, new(big.Int).Mul(expectETHBalance, common.Ether), bc.currentState.GetBalance(toAddress))
}

func debugBlockchain(bc *Blockchain) {
	state := bc.currentState
	fmt.Println("current state: ================================================")
	fmt.Println(state)

	block := bc.CurrentBlock()
	fmt.Println("current block: ================================================")
	fmt.Println("txns:")
	for _, tx := range block.Transactions() {
		fmt.Println(tx)
	}

	err := block.Validate()
	if err != nil {
		panic(err)
	}

	fmt.Println("txn tree:")

	block.DebugTxnTree()

	fmt.Println("rec tree:")
	block.DebugReceiptTree()
}
