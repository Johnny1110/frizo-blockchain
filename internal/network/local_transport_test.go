package network

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnect(t *testing.T) {
	trA := NewLocalTransport("A")
	trB := NewLocalTransport("B")

	_ = trA.Connect(trB)
	_ = trB.Connect(trA)

	assert.Equal(t, trA.peers[trB.Addr()], trB)
	assert.Equal(t, trB.peers[trA.Addr()], trA)
}

func create2Transports() (*LocalTransport, *LocalTransport) {
	return NewLocalTransport("A"), NewLocalTransport("B")
}

func TestSendMessage(t *testing.T) {
	trA, trB := create2Transports()
	_ = trA.Connect(trB)
	_ = trB.Connect(trA)

	msg := []byte("hello world")
	assert.Nil(t, trA.SendMessage(trB.Addr(), msg))

	rpc := <-trB.Consume()
	assert.Equal(t, rpc.From, trA.Addr())
	assert.Equal(t, rpc.Payload, msg)
}
