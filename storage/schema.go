package storage

import (
	"encoding/binary"
	"frizo-blockchain/common"
)

// BlockDB key prefix
var (
	headerPrefix    = []byte("h") // headerPrefix + num (8 bytes) + hash -> header
	bodyPrefix      = []byte("b") // bodyPrefix + num (8 bytes) + hash -> body
	receiptsPrefix  = []byte("r") // receiptsPrefix + num (8 bytes) + hash -> receipts
	canonicalPrefix = []byte("c") // canonicalPrefix + num (8 bytes) -> canonical hash

	// chain metadata
	headHeaderKey = []byte("LastHeader")
	headBlockKey  = []byte("LastBlock")
	headFastKey   = []byte("LastFast")
)

// StateDB key prefix
var (
	// MPT Node（using hash as key directly）
	// contract code
	codePrefix = []byte("c") // codePrefix + code hash -> code
	// account snapshot
	snapshotPrefix = []byte("s") // snapshotPrefix + account hash -> account data
)

// IndexDB key prefix
var (
	// txn lookup index
	txLookupPrefix = []byte("l") // txLookupPrefix + tx hash -> block number + index
	// Bloom filter index
	bloomBitsPrefix = []byte("B") // bloomBitsPrefix + section + hash -> bloom bits
)

// ====== key encoding =======

// headerKey = headerPrefix + num (8 bytes) + hash
func headerKey(number uint64, hash common.Hash) []byte {
	key := make([]byte, len(headerPrefix)+8+len(hash))
	copy(key, headerPrefix)
	binary.BigEndian.PutUint64(key[len(headerPrefix):], number)
	copy(key[len(headerPrefix)+8:], hash.Bytes())
	return key
}

// bodyKey = bodyPrefix + num (8 bytes) + hash
func bodyKey(number uint64, hash common.Hash) []byte {
	key := make([]byte, len(bodyPrefix)+8+len(hash))
	copy(key, bodyPrefix)
	binary.BigEndian.PutUint64(key[len(bodyPrefix):], number)
	copy(key[len(bodyPrefix)+8:], hash.Bytes())
	return key
}

// receiptsKey = receiptsPrefix + num (8 bytes) + hash
func receiptsKey(number uint64, hash common.Hash) []byte {
	key := make([]byte, len(receiptsPrefix)+8+len(hash))
	copy(key, receiptsPrefix)
	binary.BigEndian.PutUint64(key[len(receiptsPrefix):], number)
	copy(key[len(receiptsPrefix)+8:], hash.Bytes())
	return key
}

// codeKey = codePrefix + code hash
func codeKey(codeHash common.Hash) []byte {
	key := make([]byte, len(codePrefix)+len(codeHash))
	copy(key, codePrefix)
	copy(key[len(codePrefix):], codeHash.Bytes())
	return key
}

// txLookupKey = txLookupPrefix + tx hash
func txLookupKey(txHash common.Hash) []byte {
	key := make([]byte, len(txLookupPrefix)+len(txHash))
	copy(key, txLookupPrefix)
	copy(key[len(txLookupPrefix):], txHash.Bytes())
	return key
}

// txLookupKey = txLookupPrefix + tx hash
func canonicalKey(blockNum uint64) []byte {
	key := make([]byte, len(canonicalPrefix)+8)
	copy(key, canonicalPrefix)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, blockNum)
	copy(key[len(canonicalPrefix):], b)
	return key
}
