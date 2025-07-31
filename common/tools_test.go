package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_RLP(t *testing.T) {
	rawData := []interface{}{
		"Test_1",
		"Test_2",
		"Test_3",
		uint64(123),
	}

	bytes, err := RlpEncodeToBytes(rawData)
	assert.Nil(t, err)
	fmt.Println(bytes)

	assert.Equal(t, []byte{214, 134, 84, 101, 115, 116, 95, 49, 134, 84, 101, 115, 116, 95, 50, 134, 84, 101, 115, 116, 95, 51, 123}, bytes)
}
