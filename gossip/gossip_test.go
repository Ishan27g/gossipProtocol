package gossip

import (
	"testing"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip/peer"
	"github.com/Ishan27gOrg/gossipProtocol/gossip/sampling"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

const hostname = "http://localhost"

var udpPort = []string{":50001", ":50002", ":50003", ":50004", ":50005", ":50006", ":50007"}
var httpPort = []string{":1001", ":1002", ":1003", ":1004", ":1005", ":1006", ":1007"}
var httpPorts = func() []string {
	return httpPort
}
var udpPorts = func(exclude int) []peer.Peer {
	var ports []peer.Peer
	for i, s := range udpPort {
		if i == exclude {
			continue
		}
		ports = append(ports, peer.Peer{
			UdpAddress:        s,
			ProcessIdentifier: "PID-" + s,
		})
	}
	return ports
}

var network = func() []string {
	return []string{"1001", "1002", "1003", "1004"}
}

type mockGossip struct {
	g              Gossip
	newGossipEvent chan Packet
}

func remove(this string, from []string) []string {
	index := 0
	for i, s := range from {
		if s == this {
			index = i
			break
		}
	}
	return append(from[:index], from[index+1:]...)
}
func TestRemove(t *testing.T) {
	this := "this"
	from := []string{"some", "slice", "with", "this", "in", "it"}
	removed := remove(this, from)
	assert.Equal(t, len(from)-1, len(removed))
	for _, s := range removed {
		assert.NotEqual(t, "this", s)
	}
}

//
//func mockGossipDefaultConfig(hostname, port string) mockGossip {
//	options := Options{
//		Logger(true),
//		Env(hostname, port, "PID-"+port),
//		Hash(func(in interface{}) string {
//			return in.(string)
//		}),
//	}
//
//	m := mockGossip{
//		g:              Apply(options).New(),
//		newGossipEvent: make(chan Packet),
//	}
//	self := port
//	var nw []peer.Peer
//	for _, s := range remove(self, network()) {
//		nw = append(nw, peer.Peer{
//			UdpAddress:        s,
//			ProcessIdentifier: s + "-ID",
//		})
//	}
//	m.g.JoinWithoutSampling(func() []peer.Peer {
//		return nw
//	}, m.newGossipEvent) // across zones
//	// g.StartRumour("")
//
//	return m
//}
//
//func TestGossip(t *testing.T) {
//	var mg []mockGossip
//	events := make(chan Packet, 4)
//	var wg sync.WaitGroup
//	for i := 0; i < len(network()); i++ {
//		m := mockGossipDefaultConfig("", network()[i])
//		mg = append(mg, m)
//		wg.Add(1)
//		go func(i int) {
//			defer wg.Done()
//			packet := <-m.newGossipEvent
//			fmt.Printf("\n["+network()[i]+"]received gossip %v\n", packet)
//			events <- packet
//		}(i)
//
//	}
//	go func() {
//		<-time.After(1 * time.Second)
//		mg[1].g.StartRumour("hello")
//	}()
//	wg.Wait()
//	close(events)
//	assert.Equal(t, 4, len(events))
//	for event := range events {
//		assert.Equal(t, "hello", event.GossipMessage.Data)
//	}
//}

func setupPeers() []sampling.Sampling {
	var peers []sampling.Sampling
	for i, port := range httpPorts() {
		peer := sampling.Init(hostname+port, "pid-"+hostname+port, true)
		go Listen(udpPort[i], nil, peer.ReceiveView)
		peers = append(peers, peer)
	}
	time.Sleep(1 * time.Second)
	return peers
}

func TestGossip_JoinWithSampling(t *testing.T) {
	t.Parallel()
	mLogger.Apply(mLogger.Color(false), mLogger.Level(hclog.Trace))
	peers := setupPeers()
	for i := 0; i < len(peers); i++ {
		p := udpPorts(i)
		peers[i].Start(p)
	}
	time.Sleep(20 * time.Second)
	for i := 0; i < len(peers); i++ {
		//	fmt.Println(sampling.PrintView(*peers[i].GetView()))
		view := *peers[i].GetView()
		view.Nodes.Each(func(index int, value interface{}) {
			n := value.(sampling.NodeDescriptor)
			assert.NotEqual(t, hostname+httpPort[i], n.Address)
		})
	}
}
