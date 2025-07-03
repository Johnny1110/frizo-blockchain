package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

// Hash is a 32 len bytes array
type Hash [32]uint8

func (h Hash) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func HashFromByte(b []byte) Hash {
	if len(b) != 32 {
		panic(fmt.Sprintf("wrong hash size: %d", len(b)))
	}
	var h Hash
	copy(h[:], b)
	return h
}

// RandomHash generate a random Hash, basically for unit test case.
func RandomHash() Hash {
	var h Hash
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 32; i++ {
		h[i] = uint8(r.Intn(256))
	}
	return h
}

func HashSha256(buf *bytes.Buffer) Hash {
	return sha256.Sum256(buf.Bytes())
}
