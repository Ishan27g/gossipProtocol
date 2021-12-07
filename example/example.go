package main

import (
	"fmt"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
)

func main() {
	mLogger.New("ok", "trace")
	g := gossip.DefaultConfig("http://localhost", "1000") // zone 1

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
