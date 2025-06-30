package types

import (
	"fmt"
	"math/rand"
	"time"
)

// Hash is a 32 len bytes array
type Hash [32]uint8

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
