package gossip

import (
	"encoding/json"
	"time"
)

// Packet exchanged between peers
type Packet struct {
	AvailableAt   []string // at which addresses the data is available
	GossipMessage gossipMessage
}
type gossipMessage struct {
	Data              string
	CreatedAt         time.Time
	GossipMessageHash string
}

func gossipToByte(g gossipMessage, from string) []byte {
	b, _ := json.Marshal(gossipToPacket(g, from))
	return b
}

func gossipToPacket(g gossipMessage, from string) Packet {
	return Packet{
		AvailableAt:   []string{from},
		GossipMessage: g,
	}
}
func ByteToPacket(b []byte) Packet {
	g := Packet{
		AvailableAt: nil,
		GossipMessage: gossipMessage{
			Data:              "",
			CreatedAt:         time.Time{},
			GossipMessageHash: "",
		},
	}
	g.GossipMessage.GossipMessageHash = hash(g.GossipMessage)
	_ = json.Unmarshal(b, &g)
	return g
}

// newGossipMessage creates a Gossip message with current timestamp,
// creating a unique hash for every message
func newGossipMessage(data string, from string) Packet {
	g := gossipMessage{
		Data:              data,
		CreatedAt:         time.Now().UTC(),
		GossipMessageHash: "",
	}
	g.GossipMessageHash = hash(g)
	return gossipToPacket(g, from)
}
