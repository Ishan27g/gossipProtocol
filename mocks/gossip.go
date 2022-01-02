package mocks

import (
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
	"github.com/Ishan27gOrg/gossipProtocol/gossip/peer"
)

type MockGossip struct {
	G              gossip.Gossip
	NewGossipEvent <-chan gossip.Packet
}

func SetupGossipMock(p peer.Peer) MockGossip {
	options := gossip.Options{
		gossip.Logger(false),
		gossip.Env("localhost", p.UdpAddress, p.ProcessIdentifier),
	}
	g := gossip.Apply(options).New()
	// g.JoinWithSampling(nwPeers, newGossipEvent)
	return MockGossip{
		G:              g,
		NewGossipEvent: make(chan gossip.Packet),
	}
}
