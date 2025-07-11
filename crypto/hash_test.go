package crypto

import (
	"fmt"
	"testing"
)

func TestKeccak256Hash(t *testing.T) {
	hash := Keccak256Hash([]byte("hello world"))
	fmt.Println(hash)
}
