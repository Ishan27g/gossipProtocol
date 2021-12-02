package mocks

import (
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
)

type MockServerI interface {
	// MockReceiveGossip returns the gossip response from a peer('s udp server)
	MockReceiveGossip(g gossip.Packet) []byte
	// MockReceiveView returns a view that a udp peer can return
	MockReceiveView(view gossip.View, s string) []byte
}
type mockServer struct {
	self string
}

// MockReceiveGossip returns the gossip response from a peer('s udp server)
func (ms *mockServer) MockReceiveGossip(g gossip.Packet) []byte {
	return []byte("OKAY")
}

// MockReceiveView returns a view that a peer('s udp server) can return
func (ms *mockServer) MockReceiveView(view gossip.View, s string) []byte {
	return gossip.ViewToBytes(*mockView(ms.self))
}

// MockServer returns an interface that mocks the udp server at this address
func MockServer(port string) MockServerI {
	ms := mockServer{
		self: "http://localhost:" + port,
	}
	return &ms
}
