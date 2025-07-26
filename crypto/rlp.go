package crypto

import "github.com/ethereum/go-ethereum/rlp"

func RlpEncodeToBytes(val interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(val)
}

func RlpDecodeBytes(b []byte, val interface{}) error {
	return rlp.DecodeBytes(b, val)
}
