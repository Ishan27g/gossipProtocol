package main

import (
	"fmt"
	"time"

	"github.com/Ishan27gOrg/gossipProtocol/gossip"
	"github.com/Ishan27gOrg/gossipProtocol/gossip/peer"
)

var hostname = "localhost"
var network = []string{"1001", "1002", "1003", "1004"}
var peers = []peer.Peer{
	{"localhost:" + network[0], "p1"},
	{"localhost:" + network[1], "p2"},
	{"localhost:" + network[2], "p3"},
	{"localhost:" + network[3], "p4"},
}

func exampleCustomStrategy(index int, hostname, udp string) gossip.Gossip {

	options := gossip.Options{
		gossip.Env(hostname, udp, peers[index].ProcessIdentifier),
		gossip.Logger(false),
	}

	g := gossip.Apply(options).New()

	newGossipEvent := make(chan gossip.Packet)
	g.JoinWithoutSampling(func() []peer.Peer {
		var p []peer.Peer
		for i := 0; i < len(peers); i++ {
			if i != index {
				p = append(p, peers[i])
			}
		}
		return p
	}, newGossipEvent) // across zones
	// g.JoinWithSampling(peers, newGossipEvent)
	// g.StartRumour("")

	go func(g gossip.Gossip) {
		for {
			packet := <-g.ReceiveGossip()
			fmt.Printf("At [%s] received gossip %v\n", udp, packet)
			return
		}
	}(g)
	return g
}

func main() {

	for i := len(network) - 1; i >= 1; i-- {
		exampleCustomStrategy(i, hostname, network[i])
	}
	<-time.After(3 * time.Second)
	g := exampleCustomStrategy(0, hostname, network[0])
	g.StartRumour("hello")
	<-make(chan bool)

}
