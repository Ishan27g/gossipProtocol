package main

import (
	"fmt"
	"os"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
)

func exampleCustomStrategy(hostname, udp string) gossip.Gossip {
	mLogger.New("ok", "trace", os.Stderr)
	g := gossip.DefaultConfig(hostname, udp) // zone 1

	newGossipEvent := make(chan gossip.Packet)
	g.JoinWithoutSampling([]string{"localhost:1001", "localhost:1002", "localhost:1003", "localhost:1004"}, newGossipEvent) // across zones
	// g.StartRumour("")

	go func() {
		for {
			select {
			case packet := <-newGossipEvent:
				fmt.Printf("\nreceived gossip %v\n", packet)
			}
		}
	}()
	return g
}

var hostname = "localhost"
var network = []string{"1001", "1002", "1003", "1004"}

func main() {
	for i := len(network) - 1; i >= 1; i-- {
		go exampleCustomStrategy(hostname, network[i])
	}
	// g := exampleCustomStrategy(hostname, network[0])
	// g.StartRumour("hello")
	// <-make(chan bool)
}
