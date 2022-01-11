package gossipProtocol

import (
	"context"
	"testing"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/sampling"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

var udpPort = []string{":50001", ":50002", ":50003", ":50004", ":50005", ":50006", ":50007"}
var httpPort = []string{":1001", ":1002", ":1003", ":1004", ":1005", ":1006", ":1007"}
var httpPorts = func() []string {
	return httpPort
}
var udpPorts = func(exclude int) []sampling.Peer {
	var ports []sampling.Peer
	for i, s := range udpPort {
		if i == exclude {
			continue
		}
		ports = append(ports, sampling.Peer{
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

func setupPeers(strategy sampling.PeerSamplingStrategy) ([]sampling.Sampling, []context.Context, []context.CancelFunc) {
	var ctxs []context.Context
	var cans []context.CancelFunc
	const hostname = "http://localhost"
	var peers []sampling.Sampling
	for i, port := range httpPort {
		ctx, can := context.WithCancel(context.Background())
		ctxs = append(ctxs, ctx)
		cans = append(cans, can)
		peer := sampling.Init(hostname+port, "pid-"+hostname+port, true, strategy)
		go Listen(ctxs[i], udpPort[i], nil, peer.ReceiveView)
		time.Sleep(1 * time.Second) // allow server to start
		peers = append(peers, peer)
	}
	time.Sleep(1 * time.Second)
	return peers, ctxs, cans
}

func TestSampling_DefaultStrategy(t *testing.T) {
	const hostname = "http://localhost"
	mLogger.Apply(mLogger.Color(false), mLogger.Level(hclog.Trace))
	peers, ctxs, cans := setupPeers(sampling.DefaultStrategy())
	for i := 0; i < len(peers); i++ {
		p := udpPorts(i)
		peers[i].Start(ctxs[i], p)
	}
	time.Sleep(20 * time.Second)
	for i := 0; i < len(peers); i++ {
		view := *peers[i].GetView()
		view.Nodes.Each(func(index int, value interface{}) {
			n := value.(sampling.NodeDescriptor)
			assert.NotEqual(t, hostname+httpPort[i], n.Address)
		})
		// fmt.Println(sampling.PrintView(view))
		cans[i]()
	}
}
func TestSampling_Strategy(t *testing.T) {
	const hostname = "http://localhost"
	mLogger.Apply(mLogger.Color(false), mLogger.Level(hclog.Trace))

	peers, ctxs, cans := setupPeers(sampling.With(sampling.Head, sampling.Push, sampling.Head))
	for i := 0; i < len(peers); i++ {
		p := udpPorts(i)
		ctxs[i], cans[i] = context.WithCancel(context.Background())
		peers[i].Start(ctxs[i], p)
	}
	time.Sleep(11 * time.Second)
	cans[0]() // stop peer 0
	time.Sleep(11 * time.Second)

	// for all others
	for i := 1; i < len(peers); i++ {
		view := *peers[i].GetView()
		view.Nodes.Each(func(index int, value interface{}) {
			n := value.(sampling.NodeDescriptor)
			assert.NotEqual(t, hostname+httpPort[i], n.Address)
		})
		// fmt.Println(sampling.PrintView(view))
		cans[i]()
	}
}
