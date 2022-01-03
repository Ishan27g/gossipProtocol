package main

import (
	"fmt"
	"time"

	"github.com/Ishan27gOrg/gossipProtocol"
	"github.com/Ishan27gOrg/gossipProtocol/peer"
	"github.com/Ishan27gOrg/gossipProtocol/sampling"
)

var hostname = "localhost"
var network = []string{"1001", "1002", "1003", "1004"}
var peers = []peer.Peer{
	{"localhost:" + network[0], "p1"},
	{"localhost:" + network[1], "p2"},
	{"localhost:" + network[2], "p3"},
	{"localhost:" + network[3], "p4"},
}

func exampleCustomStrategy(index int, hostname, udp string) gossipProtocol.Gossip {

	options := gossipProtocol.Options{
		gossipProtocol.Env(hostname, udp, peers[index].ProcessIdentifier),
		gossipProtocol.Logger(false),
		gossipProtocol.Strategy(sampling.Random, sampling.Push, sampling.Random),
	}

	g := gossipProtocol.Apply(options).New()

	newGossipEvent := make(chan gossipProtocol.Packet)
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

	go func(g gossipProtocol.Gossip) {
		for {
			packet := <-g.ReceiveGossip()
			fmt.Printf("At [%s] received gossip %v\n", udp, packet)
			p, vc := g.RemovePacket(packet.GetId())
			fmt.Println(*p)
			fmt.Println(vc)
			fmt.Println(g.RemovePacket(packet.GetId()))
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

	g.JoinWithSampling(peers, newGossipEvent)
	g.StartRumour("")

}
