package gossip

import (
	"math/rand"
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
	JoinWithSampling(peers []string, newGossip chan Packet)
	// JoinWithoutSampling starts the gossip protocol with these initial peers. Peers are iteratively selected.
	// Data is sent of the channel when gossip is received from a peer or from the user (StartRumour)
	JoinWithoutSampling(peers []string, newGossip chan Packet)
	// StartRumour is the equivalent of receiving a gossip message from the user. This is sent to peers
	StartRumour(data string)
	ReceiveGossip() chan Packet // gossip from user/peer : packet
	// RemovePacket will return the packet and its latest event clock after removing it from memory
	// Should be called after Any new packet with this id will initiate a new clock
	RemovePacket(id string) (*Packet, vClock.EventClock)
	CurrentPeers() []string
}
type gossip struct {
	c                 *Config       // gossip protocol configuration
	env               EnvCfg        // env
	udp               client.Client // udp client
	mutex             sync.Mutex
	logger            hclog.Logger
	peerSelector      func() string
	peers             []string
	viewCb            func(view sampling2.View, from string) []byte
	newGossipPacket   chan Packet                   // gossip from user/peer : packet
	receivedGossipMap map[string]*Packet            // map of all gossip id & gossip data received
	eventClock        map[string]vClock.VectorClock // map of gossip id & event vector clock
}

func (g *gossip) CurrentPeers() []string {
	return g.peers
}

func (g *gossip) ReceiveGossip() chan Packet {
	return g.newGossipPacket
}

// RemovePacket will return the packet and its latest event clock after removing it from memory
// Should be called after Any new packet with this id will initiate a new clock
func (g *gossip) RemovePacket(id string) (*Packet, vClock.EventClock) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	packet, clock := g.receivedGossipMap[id], g.eventClock[id].Get(id)
	delete(g.receivedGossipMap, id)
	delete(g.eventClock, id)
	return packet, clock
}

// gossipCb is called when the udp server receives a gossip message. This is sent by peers and gossiped to peers
// response for a gossip message sent to a peer is not used, returns anything
func (g *gossip) gossipCb(gossip Packet, from string) []byte {
	id := gossip.GossipMessage.GossipMessageHash

	g.mutex.Lock()
	if g.eventClock[id] == nil {
		g.eventClock[id] = vClock.Init(g.SelfAddress())
	}
	// from peer, clock receive event
	g.eventClock[id].ReceiveEvent(id, gossip.VectorClock)
	g.mutex.Unlock()

	// if new id -> clock send event
	go func() {
		if g.startRumour(gossip) {
			g.newGossipPacket <- gossip
		}
	}()
	return []byte("OKAY")
}

// StartRumour is the equivalent of receiving a gossip message from the user. This is gossiped to peers
func (g *gossip) StartRumour(data string) {
	gP := newGossipMessage(data, g.SelfAddress(), nil)
	gP.GossipMessage.Version++
	g.startRumour(gP) // clock send event
	g.newGossipPacket <- Packet{
		AvailableAt:   gP.AvailableAt,
		GossipMessage: gP.GossipMessage,
		VectorClock:   g.eventClock[gP.GetId()].Get(gP.GetId()), // return updated clock
	}
}
func (g *gossip) startRumour(gP Packet) bool {
	newGossip := false
	if g.receivedGossipMap[gP.GetId()] == nil {
		newGossip = true
		gP.AvailableAt = append(gP.AvailableAt, g.SelfAddress())
		g.logger.Trace("Received new gossip [version-" + strconv.Itoa(gP.GetVersion()) + "] " + gP.GetId() + " gossiping....✅")
		g.mutex.Lock()
		g.receivedGossipMap[gP.GetId()] = &gP
		g.mutex.Unlock()

		g.beginGossipRounds(gP.GossipMessage)
	} else {
		g.logger.Trace("Received existing gossip [version-" + strconv.Itoa(gP.GetVersion()) + "] " + gP.GetId() +
			"checking version, not gossipping....❌")
		if g.receivedGossipMap[gP.GetId()].GetVersion() < gP.GetVersion() {
			g.mutex.Lock()
			g.receivedGossipMap[gP.GetId()] = &gP
			g.mutex.Unlock()
			g.logger.Trace("gossip [version-" + strconv.Itoa(gP.GetVersion()) + "] " + gP.GetId() + " Updated locally")
		}
	}
	return newGossip
}

// beginGossipRounds sends the gossip message for M rounds
// where M is calculated from the config provided.
// https://flopezluis.github.io/gossip-simulator/
func (g *gossip) beginGossipRounds(gsp gossipMessage) {
	// todo revert back
	//rounds := int(math.Ceil(math.Log10(float64(g.c.MinimumPeersInNetwork)) /
	//	math.Log10(float64(g.c.FanOut))))
	rounds := 2
	g.logger.Trace("Gossip rounds - " + strconv.Itoa(rounds) + " for id - " + gsp.GossipMessageHash)
	for i := 0; i < rounds; i++ {
		<-time.After(g.c.RoundDelay)
		g.sendGossip(gsp)
		gsp.Version++
	}
	g.logger.Trace("[Gossip ended] - " + gsp.GossipMessageHash)
}

// sendGossip sends the gossip message to FanOut number of peers
func (g *gossip) sendGossip(gm gossipMessage) {
	var peers []string
	for i := 1; i <= g.c.FanOut; i++ {
		peer := g.peerSelector()
		if peer != "" {
			peers = append(peers, g.peerSelector())
		}
	}
	id := gm.GossipMessageHash
	g.mutex.Lock()
	if g.eventClock[id] == nil {
		g.eventClock[id] = vClock.Init(g.SelfAddress())
	}
	clock := g.eventClock[id].SendEvent(gm.GossipMessageHash, peers)
	g.mutex.Unlock()

	if len(peers) == 0 {
		return
	}
	for _, peer := range peers {
		g.logger.Debug("Gossipping Id - [ " + gm.GossipMessageHash + " ] to peer - " + peer)
		go g.udp.SendGossip(peer, gossipToByte(gm, g.SelfAddress(), clock))
	}
}

func (g *gossip) SelfAddress() string {
	// self := g.env.Hostname + g.env.UdpPort
	return g.env.ProcessAddress
}

// JoinWithSampling starts the gossip protocol with these initial peers. Peer sampling is done to periodically
// maintain a partial view (subset) of the gossip network. Data is sent of the channel when gossip
// is received from a peer or from the user (StartRumour)
func (g *gossip) JoinWithSampling(peers []string, newGossip chan Packet) {

	ps := sampling2.Init(g.SelfAddress())
	g.peers = peers
	g.peerSelector = ps.SelectPeer
	g.viewCb = ps.ReceiveView

	g.newGossipPacket = newGossip
	// listen calls the udp server to Start and registers callbacks for incoming gossip or views from peers
	go Listen(g.env.UdpPort, g.gossipCb, g.viewCb)
	// Start peer sampling and exchange views
	go ps.Start(g.peers)
}

func (g *gossip) JoinWithoutSampling(peers []string, newGossip chan Packet) {
	rand.Seed(time.Now().Unix())
	var randomPeer = func() string {
		if len(peers) == 0 {
			return ""
		}
		return peers[rand.Intn(len(peers))]
	}
	var noViewExchange = func(view sampling2.View, s string) []byte {
		return []byte("never called")
	}
	g.peers = peers
	g.peerSelector = randomPeer
	g.viewCb = noViewExchange

	g.newGossipPacket = newGossip
	// listen calls the udp server to Start and registers callbacks for incoming gossip or views from peers
	go Listen(g.env.UdpPort, g.gossipCb, noViewExchange)
}
func withConfig(hostname, port, selfAddress string, c *Config) Gossip {

	g := gossip{
		mutex:  sync.Mutex{},
		logger: mLogger.Get(port),
		udp:    client.GetClient(port),
		env: EnvCfg{
			Hostname:       hostname,
			UdpPort:        ":" + port,
			ProcessAddress: selfAddress,
		},
		receivedGossipMap: make(map[string]*Packet),
		peerSelector:      nil,
		peers:             []string{},
		eventClock:        make(map[string]vClock.VectorClock),
		c:                 c,
		newGossipPacket:   make(chan Packet),
	}
	return &g
}

// DefaultConfig returns the default gossip protocol interface after
// sets up a default gossip config with a default peer sampling strategy
func DefaultConfig(hostname, port, selfAddress string) Gossip {
	return withConfig(hostname, port, selfAddress, &Config{
		RoundDelay:            1 * time.Second,
		FanOut:                3,
		MinimumPeersInNetwork: 10,
	})
}
