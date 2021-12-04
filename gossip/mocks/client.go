package mocks

import (
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
)

type MockUdpClientI interface {
	// MockExchangeView mocks a random view response by the UDP server
	MockExchangeView(address string, view gossip.View) *gossip.View
	// MockSendGossip mocks a gossip response by the UDP server
	MockSendGossip(address string, data []byte) []byte
}
type mockUdpClient struct {
	client func() gossip.Client
	view   gossip.View
}

// MockExchangeView mocks a random view response by the UDP server
func (m *mockUdpClient) MockExchangeView(address string, v gossip.View) *gossip.View {
	rsp, _ := gossip.BytesToView(MockServer(address).MockReceiveView(v, address)) // todo mockServer(address=port)
	return &rsp
}

// MockSendGossip mocks a gossip response by the UDP server
func (m *mockUdpClient) MockSendGossip(address string, data []byte) []byte {
	return MockServer(address).MockReceiveGossip(gossip.ByteToPacket(data)) // todo mockServer(address=port)
}

// MockClient mock
func MockClient(self string) MockUdpClientI {
	view := mockView(self)
	return &mockUdpClient{
		client: gossip.GetClient,
		view:   *view,
	}
}
