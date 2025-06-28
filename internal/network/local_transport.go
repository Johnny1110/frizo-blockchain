package network

import (
	"fmt"
	"sync"
)

type LocalTransport struct {
	addr      NetAddr
	consumeCh chan RPC
	lock      sync.RWMutex
	peers     map[NetAddr]Transport
}

func NewLocalTransport(addr NetAddr) *LocalTransport {
	return &LocalTransport{
		addr:      addr,
		consumeCh: make(chan RPC, 1024),
		peers:     make(map[NetAddr]Transport),
	}
}

func (l *LocalTransport) Consume() <-chan RPC {
	return l.consumeCh
}

func (l *LocalTransport) Connect(transport Transport) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.peers[transport.Addr()] = transport
	return nil
}

func (l *LocalTransport) SendMessage(toAddr NetAddr, payload []byte) error {
	l.lock.RLock()
	defer l.lock.RUnlock()

	if peer, ok := l.peers[toAddr]; ok {
		p := peer.(*LocalTransport)
		p.consumeCh <- RPC{
			From:    l.addr,
			Payload: payload,
		}
	} else {
		return fmt.Errorf("no peer found for %s", toAddr)
	}
	return nil
}

func (l *LocalTransport) Addr() NetAddr {
	return l.addr
}
