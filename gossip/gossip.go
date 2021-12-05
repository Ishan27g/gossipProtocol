package gossip

import (
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip/client"
	sampling2 "github.com/Ishan27gOrg/gossipProtocol/gossip/sampling"
	"github.com/Ishan27gOrg/vClock"
	"github.com/hashicorp/go-hclog"
)

type Gossip interface {
	// JoinWithSampling starts the gossip protocol with these initial peers. Peer sampling is done to periodically
	// maintain a partial view (subset) of the gossip network. Data is sent of the channel when gossip
	// is received from a peer or from the user (StartRumour)
	JoinWithSampling(peers []string, newGossip chan<- map[string]Packet)
	// JoinWithoutSampling starts the gossip protocol with these initial peers. Peers are iteratively selected.
	// Data is sent of the channel when gossip is received from a peer or from the user (StartRumour)
	JoinWithoutSampling(peers []string, newGossip chan<- map[string]Packet)
	// StartRumour is the equivalent of receiving a gossip message from the user. This is sent to peers
	StartRumour(data string)

	// RemovePacket will return the packet and its latest event clock after removing it from memory
	// Should be called after Any new packet with this id will initiate a new clock
	RemovePacket(id string) (*Packet, vClock.EventClock)
}
type gossip struct {
	c                 *Config       // gossip protocol configuration
	env               EnvCfg        // env
	udp               client.Client // udp client
	mutex             sync.Mutex
	logger            hclog.Logger
	peerSelector      func() string
	viewCb            func(view sampling2.View, from string) []byte
	newGossipPacket   chan<- map[string]Packet      // gossip from user/peer : packet
	receivedGossipMap map[string]*Packet            // map of all gossip id & gossip data received
	eventClock        map[string]vClock.VectorClock // map of gossip id & event vector clock
}

// RemovePacket will return the packet and its latest event clock after removing it from memory
// Should be called after Any new packet with this id will initiate a new clock
func (g *gossip) RemovePacket(id string) (*Packet, vClock.EventClock) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	packet, clock := g.receivedGossipMap[id], g.eventClock[id].Get()
	delete(g.receivedGossipMap, id)
	delete(g.eventClock, id)
	return packet, clock
}

// StartRumour is the equivalent of receiving a gossip message from the user. This is gossiped to peers
func (g *gossip) StartRumour(data string) {
	gP := newGossipMessage(data, g.selfAddress(), nil)
	g.startRumour(gP) // clock send event
	g.newGossipPacket <- map[string]Packet{
		g.selfAddress(): { // return updated clock
			AvailableAt:   gP.AvailableAt,
			GossipMessage: gP.GossipMessage,
			VectorClock:   g.eventClock[gP.GossipMessage.GossipMessageHash].Get(), // return updated clock
		},
	}
}
func (g *gossip) startRumour(gP Packet) bool {
	newGossip := false
	if g.receivedGossipMap[gP.GossipMessage.GossipMessageHash] == nil {
		g.mutex.Lock()
		newGossip = true
		g.logger.Debug("Received new gossip - " + gP.GossipMessage.GossipMessageHash +
			" from - " + gP.AvailableAt[0] + " gossiping....✅")
		g.receivedGossipMap[gP.GossipMessage.GossipMessageHash] = &gP
		g.mutex.Unlock()
		g.beginGossipRounds(gP.GossipMessage)
	} else {
		g.mutex.Lock()
		gPExisting := g.receivedGossipMap[gP.GossipMessage.GossipMessageHash]
		gPExisting.AvailableAt = append(gPExisting.AvailableAt, gP.AvailableAt[0])
		g.logger.Debug("Received existing gossip - " + gP.GossipMessage.GossipMessageHash +
			" adding new download address, not gossipping....❌")
		g.mutex.Unlock()
	}
	return newGossip
}

// beginGossipRounds sends the gossip message for M rounds
// where M is calculated from the config provided.
// https://flopezluis.github.io/gossip-simulator/
func (g *gossip) beginGossipRounds(gsp gossipMessage) {
	rounds := int(math.Ceil(math.Log10(float64(g.c.MinimumPeersInNetwork)) /
		math.Log10(float64(g.c.FanOut))))
	g.logger.Trace("num of gossip rounds - " + strconv.Itoa(rounds))
	g.logger.Debug("[Gossip started] - " + gsp.GossipMessageHash)
	for i := 0; i < rounds; i++ {
		<-time.After(g.c.RoundDelay)
		g.sendGossip(gsp)
	}
	g.logger.Debug("[Gossip ended] - " + gsp.GossipMessageHash)
}

// sendGossip sends the gossip message to FanOut number of peers
func (g *gossip) sendGossip(gm gossipMessage) {
	gsp := make(chan []byte)
	var peers []string
	for i := 0; i < g.c.FanOut; i++ {
		peers = append(peers, g.peerSelector())
	}
	id := gm.GossipMessageHash
	g.mutex.Lock()
	if g.eventClock[id] == nil {
		g.eventClock[id] = vClock.Init(g.selfAddress())
	}
	clock := g.eventClock[id].SendEvent(gm.GossipMessageHash, peers)
	g.mutex.Unlock()

	for _, peer := range peers {
		g.logger.Debug("[Gossip " + gm.GossipMessageHash + " ] to peer - " + peer)
		if gossipRsp := g.udp.SendGossip(peer, gossipToByte(gm, g.selfAddress(), clock)); gossipRsp != nil {
			gsp <- gossipRsp
		}
	}
	go func() {
		for i := 0; i < g.c.FanOut; i++ {
			gs := <-gsp
			g.logger.Debug("Received on main " + string(gs)) // OK
		}
	}()
}

// gossipCb is called when the udp server receives a gossip message. This is sent by peers and gossiped to peers
// response for a gossip message sent to a peer is not used, returns anything
func (g *gossip) gossipCb(gossip Packet, from string) []byte {
	id := gossip.GossipMessage.GossipMessageHash
	g.mutex.Lock()
	if g.eventClock[id] == nil {
		g.eventClock[id] = vClock.Init(g.selfAddress())
	}
	// from peer, clock receive event
	g.eventClock[id].ReceiveEvent(id, gossip.VectorClock)
	g.mutex.Unlock()

	// if new id -> clock send event
	if g.startRumour(gossip) {
		g.newGossipPacket <- map[string]Packet{
			from: gossip,
		}
	}
	return []byte("OKAY")
}
func (g *gossip) selfAddress() string {
	self := g.env.Hostname + ":" + g.env.UdpPort
	return self
}
func withConfig(hostname, port string, c *Config) Gossip {

	g := gossip{
		mutex:  sync.Mutex{},
		logger: mLogger.Get("gossip"),
		udp:    nil,
		env: EnvCfg{
			Hostname: hostname,
			UdpPort:  port,
		},
		receivedGossipMap: make(map[string]*Packet),
		peerSelector:      nil,
		eventClock:        make(map[string]vClock.VectorClock),
		c:                 c,
	}
	return &g
}

// WithConfig returns the default gossip protocol interface after
// sets up the provided gossip config and a default peer sampling strategy
func WithConfig(hostname, port string, c *Config) Gossip {
	return withConfig(hostname, port, c)
}

// DefaultConfig returns the default gossip protocol interface after
// sets up a default gossip config with a default peer sampling strategy
func DefaultConfig(hostname, port string) Gossip {
	return withConfig(hostname, port, &Config{
		RoundDelay:            3 * time.Second,
		FanOut:                3,
		MinimumPeersInNetwork: 10,
	})
}

// JoinWithSampling starts the gossip protocol with these initial peers. Peer sampling is done to periodically
// maintain a partial view (subset) of the gossip network. Data is sent of the channel when gossip
// is received from a peer or from the user (StartRumour)
func (g *gossip) JoinWithSampling(peers []string, newGossip chan<- map[string]Packet) {

	ps := sampling2.Init(g.selfAddress())
	g.peerSelector = ps.SelectPeer
	g.viewCb = ps.ReceiveView

	g.newGossipPacket = newGossip
	// listen calls the udp server to Start and registers callbacks for incoming gossip or views from peers
	go Listen(g.env.UdpPort, g.gossipCb, g.viewCb)
	// Start peer sampling and exchange views
	go ps.Start(peers)
}

func (g *gossip) JoinWithoutSampling(peers []string, newGossip chan<- map[string]Packet) {
	cur := 0
	var randomPeer = func() string {
		cur++
		if cur == len(peers) {
			cur = 0
		}
		return peers[cur]
	}
	var noViewExchange = func(view sampling2.View, s string) []byte {
		return []byte("")
	}
	g.peerSelector = randomPeer
	g.newGossipPacket = newGossip
	// listen calls the udp server to Start and registers callbacks for incoming gossip or views from peers
	go Listen(g.env.UdpPort, g.gossipCb, noViewExchange)
}
