package main

import (
	"fmt"
	"time"

	"github.com/Ishan27gOrg/gossipProtocol"
	"github.com/Ishan27gOrg/gossipProtocol/peer"
	"github.com/Ishan27gOrg/gossipProtocol/sampling"
)

var hostname = "localhost"
var network = []string{"1001", "1002", "1003", "1004", "1005", "1006", "1007", "1008",
	"1009", "1010", "1011", "1012"}
var peers = []peer.Peer{
	{hostname + ":" + network[0], "p1"},
	{hostname + ":" + network[1], "p2"},
	{hostname + ":" + network[2], "p3"},
	{hostname + ":" + network[3], "p4"},
	{hostname + ":" + network[4], "p5"},
	{hostname + ":" + network[5], "p6"},
	{hostname + ":" + network[6], "p7"},
	{hostname + ":" + network[7], "p8"},
	{hostname + ":" + network[8], "p9"},
	{hostname + ":" + network[9], "p10"},
	{hostname + ":" + network[10], "p11"},
	{hostname + ":" + network[11], "p12"},
}

func exampleCustomStrategy(index int, hostname, udp string) gossipProtocol.Gossip {

	options := gossipProtocol.Options{
		gossipProtocol.Env(hostname, udp, peers[index].ProcessIdentifier),
		gossipProtocol.Logger(false),
		gossipProtocol.Strategy(sampling.Random, sampling.Push, sampling.Random),
	}

	g := gossipProtocol.Apply(options).New()

	newGossipEvent := make(chan gossipProtocol.Packet)
	g.JoinWithSampling(peers, newGossipEvent)
	// g.StartRumour("")

	go func(g gossipProtocol.Gossip) {
		for {
			packet := <-g.ReceiveGossip()
			fmt.Printf("[%s] received gossip %v\n", udp, packet)
			//p, vc := g.RemovePacket(packet.GetId())
			//fmt.Printf("[%s] Packet - %v", udp, *p)
			//fmt.Printf("[%s] VectorClock - %v", udp, vc)
			<-time.After(1 * time.Second)
		}
	}(g)
	return g
}

func main() {

	options := gossipProtocol.Options{
		gossipProtocol.Env(hostname, network[0], peers[0].ProcessIdentifier),
		gossipProtocol.Logger(false),
	}
	g := gossipProtocol.Apply(options).New()
	newGossipEvent := make(chan gossipProtocol.Packet)

	g.JoinWithoutSampling(func() []peer.Peer {
		var p []peer.Peer
		for i := 1; i < len(peers); i++ {
			p = append(p, peers[i])
		}
		return p
	}, newGossipEvent)

	for i := len(network) - 1; i >= 1; i-- {
		go exampleCustomStrategy(i, hostname, network[i])
	}
	<-time.After(3 * time.Second)

	g.StartRumour("ok")
	<-make(chan bool)
}
