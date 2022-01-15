package gossipProtocol

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

type udpGossip struct {
	Packet Packet
	From   Peer
}

func packetToUdp(packet Packet, from Peer) udpGossip {
	return udpGossip{
		Packet: packet, From: from,
	}
}

type gossipMessage struct {
	Data              string
	CreatedAt         time.Time
	GossipMessageHash string
	Version           int
}

func (p *Packet) GetId() string {
	return p.GossipMessage.GossipMessageHash
}
func (p *Packet) GetVersion() int {
	return p.GossipMessage.Version
}
func (p *Packet) GetData() string {
	return p.GossipMessage.Data
}

func gossipToByte(g gossipMessage, from Peer, clock vClock.EventClock) []byte {
	packet := gossipToPacket(g, from.ProcessIdentifier, clock)
	udp := packetToUdp(*packet, from)
	b, _ := json.Marshal(&udp)
	return b
}

func gossipToPacket(g gossipMessage, from string, clock vClock.EventClock) *Packet {
	return &Packet{
		AvailableAt:   []string{from},
		GossipMessage: g,
		VectorClock:   clock,
	}
}
func ByteToPacket(b []byte) (Packet, Peer) {
	udp := udpGossip{
		Packet: Packet{
			AvailableAt:   []string{},
			GossipMessage: gossipMessage{},
			VectorClock:   vClock.EventClock{},
		},
		From: Peer{
			UdpAddress:        "",
			ProcessIdentifier: "",
			Hop:               -1,
		},
	}
	_ = json.Unmarshal(b, &udp)
	return udp.Packet, udp.From
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
	g.GossipMessageHash = defaultHashMethod(g.Data + g.CreatedAt.String())
	return *gossipToPacket(g, from, clock)
}
