package gossip

import (
	"encoding/json"
	"time"

	"github.com/Ishan27gOrg/vClock"
)

type gossipMessage struct {
	Data              string
	CreatedAt         time.Time
	GossipMessageHash string
	vClock            vClock.EventClock
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
		vClock:            make(vClock.EventClock),
	}
	g.GossipMessageHash = hash(g)
	_ = json.Unmarshal(b, &g)
	return g
}

// newGossipMessage creates a Gossip message
func newGossipMessage(data string, vClock vClock.EventClock) gossipMessage {
	g := gossipMessage{
		Data:              data,
		CreatedAt:         time.Now().UTC(),
		GossipMessageHash: "",
		vClock:            vClock,
	}
	g.GossipMessageHash = hash(g)
	return g
}
