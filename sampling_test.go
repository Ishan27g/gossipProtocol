package gossipProtocol

import (
	"context"
	"testing"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/peer"
	"github.com/Ishan27gOrg/gossipProtocol/sampling"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

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

var gossipNetwork = func() []string {
	return []string{"1001", "1002", "1003", "1004"}
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

func setupPeers(ctx context.Context, strategy sampling.PeerSamplingStrategy) []sampling.Sampling {

	const hostname = "http://localhost"
	var peers []sampling.Sampling
	for i, port := range httpPorts() {
		peer := sampling.Init(hostname+port, "pid-"+hostname+port, true, strategy)
		go Listen(ctx, udpPort[i], nil, peer.ReceiveView)
		time.Sleep(1 * time.Second) // allow server to start
		peers = append(peers, peer)
	}
	time.Sleep(1 * time.Second)
	return peers
}

func TestSampling_DefaultStrategy(t *testing.T) {
	const hostname = "http://localhost"
	ctx, cancel := context.WithCancel(context.Background())
	mLogger.Apply(mLogger.Color(false), mLogger.Level(hclog.Trace))
	peers := setupPeers(ctx, sampling.DefaultStrategy())
	for i := 0; i < len(peers); i++ {
		p := udpPorts(i)
		peers[i].Start(ctx, p)
	}
	time.Sleep(20 * time.Second)
	for i := 0; i < len(peers); i++ {
		view := *peers[i].GetView()
		view.Nodes.Each(func(index int, value interface{}) {
			n := value.(sampling.NodeDescriptor)
			assert.NotEqual(t, hostname+httpPort[i], n.Address)
		})
	}

	cancel()
}