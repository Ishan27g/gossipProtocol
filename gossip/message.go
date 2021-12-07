package gossip

import (
	"encoding/json"
	"fmt"
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
	//fmt.Printf("\ngossipToByte - %v\n", string(b))
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
	// g.GossipMessage.GossipMessageHash = hash(g.GossipMessage)
	_ = json.Unmarshal(b, &g)
	//fmt.Printf("\nByteToPacket %v\n", g)
	if g.GossipMessage.GossipMessageHash == "" {
		fmt.Println("NOOOOOOOOOOOO - " + string(b))
	}
	return g
}

// newGossipMessage creates a Gossip message with current timestamp,
// creating a unique hash for every message
func newGossipMessage(data string, from string, clock vClock.EventClock) Packet {
	g := gossipMessage{
		Data:              data,
		CreatedAt:         time.Now().UTC(),
		GossipMessageHash: "",
		Version:           0,
	}
	g.GossipMessageHash = hash(g)
	return *gossipToPacket(g, from, clock)
}
