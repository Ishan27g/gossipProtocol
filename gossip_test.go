package gossipProtocol

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Ishan27gOrg/gossipProtocol/sampling"
	"github.com/stretchr/testify/assert"
)

const TestData = "hello"

var hostname = "localhost"
var network = []string{"1001", "1002", "1003", "1004"}
var peers = []sampling.Peer{
	{"localhost:" + network[0], "p1"},
	{"localhost:" + network[1], "p2"},
	{"localhost:" + network[2], "p3"},
	{"localhost:" + network[3], "p4"},
}

func setupGossip(index int, hostname string, udp string) Gossip {
	options := Options{
		Env(hostname, udp, peers[index].ProcessIdentifier),
		Logger(false),
		Strategy(sampling.Random, sampling.Push, sampling.Random),
	}
	g := Apply(options).New()
	return g
}

func receiveOnChannel(t *testing.T, wg *sync.WaitGroup, index int, g *Gossip) {
	go func(g Gossip, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			packet := <-g.ReceiveGossip()
			fmt.Printf("At [%d] received gossip %v\n", index, packet)
			p, vc := g.RemovePacket(packet.GetId())
			assert.NotNil(t, vc)
			assert.NotNil(t, *p)
			assert.Equal(t, TestData, p.GossipMessage.Data)
			return
		}
	}(*g, wg)
}
func setupWithoutSampling(index int, g *Gossip) {
	newGossipEvent := make(chan Packet)
	(*g).JoinWithoutSampling(func() []sampling.Peer {
		var p []sampling.Peer
		for i := 0; i < len(peers); i++ {
			if i != index {
				p = append(p, peers[i])
			}
		}
		return p
	}, newGossipEvent) // across zones
}

func setupWithSampling(index int, g *Gossip) {
	newGossipEvent := make(chan Packet)
	(*g).JoinWithSampling(peers, newGossipEvent)
}

func Test_Gossip_WithoutSampling(t *testing.T) {
	var wg sync.WaitGroup
	var peers []Gossip
	for i := len(network) - 1; i >= 1; i-- {
		g := setupGossip(i, hostname, network[i])
		setupWithoutSampling(i, &g)

		wg.Add(1)
		receiveOnChannel(t, &wg, i, &g)
		peers = append(peers, g)
	}
	<-time.After(3 * time.Second)
	g := setupGossip(0, hostname, network[0])
	setupWithoutSampling(0, &g)

	wg.Add(1)
	receiveOnChannel(t, &wg, 0, &g)
	peers = append(peers, g)

	<-time.After(1 * time.Second)

	g.StartRumour(TestData)
	wg.Wait()
	for _, p := range peers {
		p.Stop()
	}

}

func Test_Gossip_WithSampling(t *testing.T) {
	var wg sync.WaitGroup
	var peers []Gossip

	for i := len(network) - 1; i >= 1; i-- {
		g := setupGossip(i, hostname, network[i])
		setupWithSampling(i, &g)

		wg.Add(1)
		receiveOnChannel(t, &wg, i, &g)
		peers = append(peers, g)
	}
	<-time.After(3 * time.Second)
	g := setupGossip(0, hostname, network[0])
	setupWithSampling(0, &g)

	wg.Add(1)
	receiveOnChannel(t, &wg, 0, &g)
	peers = append(peers, g)

	<-time.After(1 * time.Second)
	g.StartRumour(TestData)
	wg.Wait()

	for _, p := range peers {
		p.Stop()
	}
}
