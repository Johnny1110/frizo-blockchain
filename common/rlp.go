package common

import (
	"github.com/ethereum/go-ethereum/rlp"
)

func RlpEncodeToBytes(val interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(val)
}

// RlpDecodeBytes parses RLP data from b into val.
func RlpDecodeBytes(b []byte, val interface{}) error {
	return rlp.DecodeBytes(b, val)
}

// RlpSplit parse RLP encoded
// return：type, content, remaining data, error
func RlpSplit(data []byte) (typ rlp.Kind, content []byte, remaining []byte, err error) {
	return rlp.Split(data)
}
