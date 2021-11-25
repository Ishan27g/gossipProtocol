package gossip

import (
	"encoding/json"
	"time"
)

type gossipMessage struct {
	Data              string
	CreatedAt         time.Time
	GossipMessageHash string
}

func gossipToByte(g gossipMessage) []byte {
	b, _ := json.Marshal(g)
	return b
}
func byteToGossip(b []byte) gossipMessage {
	g := gossipMessage{
		Data:              "",
		CreatedAt:         time.Time{},
		GossipMessageHash: "",
	}
	g.GossipMessageHash = hash(g)
	_ = json.Unmarshal(b, &g)
	return g
}

// newGossipMessage creates a Gossip message
func newGossipMessage(data string) gossipMessage {
	g := gossipMessage{
		Data:              data,
		CreatedAt:         time.Now().UTC(),
		GossipMessageHash: "",
	}
	g.GossipMessageHash = hash(g)
	return g
}
