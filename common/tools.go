package common

import (
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

// CalcGasLimit calculate new gas limit for new block
// based on parent block gas limit (dynamic adjustment)
func CalcGasLimit(parentGasLimit, desiredLimit uint64) uint64 {
	limit := parentGasLimit

	// calculate delta changing
	delta := parentGasLimit / 1024

	if desiredLimit < parentGasLimit {
		limit = parentGasLimit - delta
	} else {
		limit = parentGasLimit + delta
	}

	// setup min limit
	if limit < BlockMinGasLimit {
		limit = BlockMinGasLimit
	} else if limit > BlockMaxGasLimit {
		limit = BlockMaxGasLimit
	}

	return limit
}

// IntrinsicGas calculate txn gas consume (basic fee + data len fee)
func IntrinsicGas(data []byte, contractCreation bool) (uint64, error) {
	var gas uint64

	if contractCreation {
		gas = TxGasContractCreation
	} else {
		gas = TxGas
	}

	// calculate data len gas
	if len(data) > 0 {
		var nonZeroCount uint64
		for _, byt := range data {
			if byt != 0 {
				nonZeroCount++
			}
		}

		gas += nonZeroCount * TxDataNonZeroGas
		gas += (uint64(len(data)) - nonZeroCount) * TxDataZeroGas
	}

	return gas, nil
}

func RlpEncodeToBytes(val interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(val)
}

func RlpDecodeBytes(b []byte, val interface{}) error {
	return rlp.DecodeBytes(b, val)
}

// Keccak256 is a kind of cryptographic hash function, input any data and output 256 bit (32 bytes).
// What Keccak256 do in ethereum?
// 1. generate address
// 2. calculate txn ID
// 3. calculate smart contract storage mapping slot
// 4. functionSelector = keccak256("transfer(address,uint256)")[0:4]
// 5. calculate contract bytecode hash verify
// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	h := sha3.NewLegacyKeccak256()
	for _, b := range data {
		h.Write(b)
	}
	return h.Sum(nil)
}

// Keccak256Hash calculates and returns the Keccak256 hash of the input data,
// converting it to a common.Hash.
func Keccak256Hash(data ...[]byte) Hash {
	return BytesToHash(Keccak256(data...))
}
