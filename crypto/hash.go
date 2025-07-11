package crypto

import (
	"frizo-blockchain/common"
	"golang.org/x/crypto/sha3"
)

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
func Keccak256Hash(data ...[]byte) common.Hash {
	return common.BytesToHash(Keccak256(data...))
}

// CreateAddress creates an Ethereum address given the bytes and the nonce
func CreateAddress(addr common.Address, nonce uint64) common.Address {
	data := make([]byte, 8)
	putUint64BE(data, nonce)
	// data contains nonce bytes.
	// create address by (address, nonce)
	return common.BytesToAddress(Keccak256(addr.Bytes(), data)[12:])
}

// CreateAddress2 creates an Ethereum address given the address bytes, initial
// contract code hash and a salt.
func CreateAddress2(addr common.Address, salt common.Hash, codeHash []byte) common.Address {
	// create address by (0xff, address, salt, codeHash)
	return common.BytesToAddress(Keccak256([]byte{0xff}, addr.Bytes(), salt.Bytes(), codeHash)[12:])
}

// putUint64BE encodes a uint64(unsigned long) as big-endian
func putUint64BE(b []byte, v uint64) {
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}
