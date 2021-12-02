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
	m.view.RandomView()
	return &m.view
}

// MockSendGossip mocks a gossip response by the UDP server
func (m *mockUdpClient) MockSendGossip(address string, data []byte) []byte {
	return []byte("Ok")
}

// MockClient mock
func MockClient(self string) MockUdpClientI {
	view := mockView(self)
	return &mockUdpClient{
		client: gossip.GetClient,
		view:   *view,
	}
}
