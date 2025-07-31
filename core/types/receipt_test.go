package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Bloom(t *testing.T) {
	bloom := &Bloom{}
	t1 := "Test123"
	t2 := "Test124"
	t3 := "Test125"
	bloom.Add([]byte(t1))
	bloom.Add([]byte(t2))
	bloom.Add([]byte(t3))

	assert.True(t, bloom.Contains([]byte(t1)))
	assert.True(t, bloom.Contains([]byte(t2)))
	assert.True(t, bloom.Contains([]byte(t3)))
	assert.False(t, bloom.Contains([]byte("Not-You")))
}
