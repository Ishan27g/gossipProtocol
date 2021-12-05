package main

import (
	"fmt"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
)

func main() {
	mLogger.New("ok", "trace")
	g := gossip.DefaultConfig("http://localhost", "1000") // zone 1

	newGossipEvent := make(chan map[string]gossip.Packet)
	g.JoinWithSampling([]string{"localhost:1001", "localhost:2102", "localhost:4103", "localhost:2004"}, newGossipEvent) // across zones
	// g.StartRumour("")

	for {
		select {
		case gossips := <-newGossipEvent:
			for from, packet := range gossips {
				fmt.Println("received gossip ", packet.GossipMessage.GossipMessageHash, " from  ", from)
				go func(id string) {
					time.After(5 * time.Second)
					_, clock := g.RemovePacket(id)
					fmt.Printf("\nvector-clock for %s after receive-event : %v\n", id, clock)
				}(packet.GossipMessage.GossipMessageHash)
			}
		}
	}
}
