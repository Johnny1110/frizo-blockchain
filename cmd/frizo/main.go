package main

import (
	"frizo-blockchain/internal/network"
	"time"
)

func main() {
	localTr := network.NewLocalTransport("LOCAL")
	remoteTr := network.NewLocalTransport("REMOTE")

	_ = localTr.Connect(remoteTr)
	_ = remoteTr.Connect(localTr)

	go func() {
		for {
			time.Sleep(1 * time.Second)
			_ = remoteTr.SendMessage("LOCAL", []byte("hello world"))
		}
	}()

	serverOpts := network.ServerOpts{
		Transports: []network.Transport{localTr},
	}

	server := network.NewServer(serverOpts)
	server.Start()
}
