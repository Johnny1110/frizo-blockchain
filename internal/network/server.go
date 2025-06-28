package network

import (
	"log"
	"time"
)

type ServerOpts struct {
	Transports []Transport
}

type Server struct {
	ServerOpts ServerOpts
	rpcCh      chan RPC
	quitCh     chan struct{}
}

func NewServer(opts ServerOpts) *Server {
	return &Server{
		ServerOpts: opts,
		rpcCh:      make(chan RPC),
		quitCh:     make(chan struct{}, 1),
	}
}

func (s *Server) Start() {
	s.initTransports()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

free:
	for {
		select {
		case rpc := <-s.rpcCh:
			log.Printf("[Server] received RPC: %v", rpc)
		case <-s.quitCh:
			break free
		case <-ticker.C:
			log.Printf("[Server] tick")
		}
	}

	log.Printf("[Server] stopped")
}

func (s *Server) initTransports() {
	for _, tr := range s.ServerOpts.Transports {
		go func(tr Transport) {
			// when tr.Consume() channel closed, range loop will auto break.
			for rpc := range tr.Consume() {
				// each Transport pass received rpc to s.rpcCh.
				s.rpcCh <- rpc
			}
		}(tr)
	}
}
