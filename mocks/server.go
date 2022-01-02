package mocks

import (
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
	"github.com/Ishan27gOrg/gossipProtocol/gossip/peer"
	"github.com/Ishan27gOrg/gossipProtocol/gossip/sampling"
)

type MockServerI interface {
	// MockReceiveGossip returns the gossip response from a peer('s udp server)
	MockReceiveGossip(g gossip.Packet) []byte
	// MockReceiveView returns a view that a udp peer can return
	MockReceiveView(view sampling.View, s string) []byte
}
type mockServer struct {
	self string
}

// MockReceiveGossip returns the gossip response from a peer('s udp server)
func (ms *mockServer) MockReceiveGossip(g gossip.Packet) []byte {
	return []byte("OKAY")
}

// MockReceiveView returns a view that a peer('s udp server) can return
func (ms *mockServer) MockReceiveView(view sampling.View, s string) []byte {
	return sampling.ViewToBytes(sampling.MergeView(*mockView(ms.self), view), peer.Peer{
		UdpAddress:        "",
		ProcessIdentifier: "",
	})
}

// MockServer returns an interface that mocks the udp server at this address
func MockServer(address string) MockServerI {
	ms := mockServer{
		self: address,
	}
	return &ms
}
