package core

import (
	"bytes"
	"fmt"
	"frizo-blockchain/internal/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHeader_Encode_Decode(t *testing.T) {
	header := &Header{
		Version:   1,
		PrevBlock: types.RandomHash(),
		Timestamp: time.Now().UnixNano(),
		Height:    10,
		Nonce:     901238,
	}
	buf := &bytes.Buffer{}
	assert.Nil(t, header.EncodeBinary(buf))

	headerDecode := &Header{}
	assert.Nil(t, headerDecode.DecodeBinary(buf))

	assert.Equal(t, header, headerDecode)
}

func TestBlock_Encode_Decode(t *testing.T) {
	block := &Block{
		Header: Header{
			Version:   1,
			PrevBlock: types.RandomHash(),
			Timestamp: time.Now().UnixNano(),
			Height:    10,
			Nonce:     901238,
		},
		Transactions: make([]Transaction, 2),
	}

	block.Transactions[0] = Transaction{}
	block.Transactions[1] = Transaction{}

	buf := &bytes.Buffer{}
	assert.Nil(t, block.EncodeBinary(buf))

	blockDecode := &Block{}
	assert.Nil(t, blockDecode.DecodeBinary(buf))

	fmt.Printf("blockDecode: %+v", blockDecode)

	assert.Equal(t, block, blockDecode)
}

func TestBlock_Hash(t *testing.T) {
	block := &Block{
		Header: Header{
			Version:   1,
			PrevBlock: types.RandomHash(),
			Timestamp: time.Now().UnixNano(),
			Height:    10,
			Nonce:     901238,
		},
		Transactions: make([]Transaction, 2),
	}

	block.Transactions[0] = Transaction{}
	block.Transactions[1] = Transaction{}

	blockHash := block.Hash()
	fmt.Println(&blockHash)
	assert.False(t, false, blockHash.IsZero())
}
