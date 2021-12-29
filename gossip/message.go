package gossip

import (
	"encoding/json"
	"time"

	"github.com/Ishan27gOrg/vClock"
)

// Packet exchanged between peers
type Packet struct {
	AvailableAt   []string // at which addresses the data is available
	GossipMessage gossipMessage
	VectorClock   vClock.EventClock
}

func (p *Packet) GetId() string {
	return p.GossipMessage.GossipMessageHash
}

func (p *Packet) GetVersion() int {
	return p.GossipMessage.Version
}

type gossipMessage struct {
	Data              string
	CreatedAt         time.Time
	GossipMessageHash string
	Version           int
}

func gossipToByte(g gossipMessage, from string, clock vClock.EventClock) []byte {
	b, _ := json.Marshal(gossipToPacket(g, from, clock))
	return b
}

func gossipToPacket(g gossipMessage, from string, clock vClock.EventClock) *Packet {
	return &Packet{
		AvailableAt:   []string{from},
		GossipMessage: g,
		VectorClock:   clock,
	}
}
func ByteToPacket(b []byte) Packet {
	g := Packet{}
	_ = json.Unmarshal(b, &g)
	return g
}

// NewGossipMessage creates a Gossip message with current timestamp,
// creating a unique hash for every message
func NewGossipMessage(data string, from string, clock vClock.EventClock) Packet {
	g := gossipMessage{
		Data:              data,
		CreatedAt:         time.Now().UTC(),
		GossipMessageHash: "",
		Version:           0,
	}
	g.GossipMessageHash = hash(g)
	return *gossipToPacket(g, from, clock)
}
