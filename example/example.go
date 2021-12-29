package main

import (
	"fmt"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
)

func main() {
	mLogger.New("ok", "trace")
	options := gossip.Options{
		gossip.Logger(true),
		gossip.Env("http://localhost", ":1000", "http://localhost:8001"),
	}

	g := gossip.Apply(options).New()
	newGossipEvent := make(chan gossip.Packet)
	g.JoinWithSampling([]string{"localhost:1001", "localhost:2102", "localhost:4103", "localhost:2004"}, newGossipEvent) // across zones
	// g.StartRumour("")

	for {
		select {
		case packet := <-newGossipEvent:
			fmt.Printf("\nreceived gossip %v\n", packet)
		}
	}
}
