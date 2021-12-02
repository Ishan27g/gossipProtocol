package mocks

import (
	"math/rand"
	"time"

	"github.com/Ishan27gOrg/gossipProtocol/gossip"
	sll "github.com/emirpasic/gods/lists/singlylinkedlist"
)

// mock a random view
func mockView(self string) *gossip.View {
	view := &gossip.View{Nodes: sll.New()}
	rand.Seed(time.Now().Unix())
	hop := rand.Intn(len(gossip.NetworkPeers(self)))
	for _, peer := range gossip.NetworkPeers(self) {
		view.Nodes.Add(gossip.NodeDescriptor{
			Address: peer,
			Hop:     hop,
		})
	}
	view.RandomView()
	return view
}
