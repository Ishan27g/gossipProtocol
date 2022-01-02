package main

import (
	"fmt"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
	"github.com/Ishan27gOrg/gossipProtocol/gossip/peer"
	"github.com/hashicorp/go-hclog"
)

func main() {
	mLogger.Apply(mLogger.Color(true), mLogger.Level(hclog.Trace))

	mLogger.New("ok")
	options := gossip.Options{
		gossip.Logger(true),
		gossip.Env("http://localhost", ":1000", "http://localhost:8001"),
	}
	var peers = []peer.Peer{
		{"localhost:1001", "p1"},
		{"localhost:1002", "p2"},
		{"localhost:1003", "p3"},
		{"localhost:1004", "p4"},
	}
	g := gossip.Apply(options).New()
	newGossipEvent := make(chan gossip.Packet)
	g.JoinWithSampling(peers, newGossipEvent) // across zones
	// g.StartRumour("")

	for {
		select {
		case packet := <-newGossipEvent:
			fmt.Printf("\nreceived gossip %v\n", packet)
		}
	}
}
