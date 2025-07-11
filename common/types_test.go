package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func Test_Hash(t *testing.T) {
	hash := HexToHash("0x1212")
	fmt.Println(hash)
	assert.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000001212", hash.Hex())

	hash = HexToHash("0x0000000000000000000000000000000000000000000000000000000000001212")
	fmt.Println(hash)
	assert.Equal(t, hash.Hex(), "0x0000000000000000000000000000000000000000000000000000000000001212")

	hash = BigToHash(big.NewInt(1121321321123))
	fmt.Println(hash)
	assert.Equal(t, "0x0000000000000000000000000000000000000000000000000000010513f582a3", hash.Hex())
}

func Test_Address(t *testing.T) {
	address := HexToAddress("0x55b09Cd9F0798233e55f3C124E18276eae850591")
	fmt.Println(address)
}
