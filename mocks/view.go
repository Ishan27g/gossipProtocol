package mocks

import (
	"math/rand"
	"strings"
	"time"

	"github.com/Ishan27gOrg/gossipProtocol/gossip/sampling"
	sll "github.com/emirpasic/gods/lists/singlylinkedlist"
)

// mockPeers returns the set of all possible peers, excluding self
func mockPeers(self string) []string {
	var networkPeers []string
	var peers = []string{"localhost:1101", "localhost:1102", "localhost:1103", "localhost:1104", "localhost:1105",
		"localhost:1106", "localhost:1107", "localhost:1108", "localhost:1109", "localhost:1110"}
	for _, peer := range peers {
		if strings.Compare(self, peer) != 0 {
			networkPeers = append(networkPeers, peer)
		}
	}
	return networkPeers
}

// mock a random view
func mockView(self string) *sampling.View {
	view := &sampling.View{Nodes: sll.New()}
	rand.Seed(time.Now().Unix())
	hop := rand.Intn(len(mockPeers(self)))
	for _, peer := range mockPeers(self) {
		view.Nodes.Add(sampling.NodeDescriptor{
			Address: peer,
			Hop:     hop,
		})
	}
	view.RandomView()
	return view
}
