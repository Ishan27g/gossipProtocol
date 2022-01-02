package gossip

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip/client"
	"github.com/Ishan27gOrg/gossipProtocol/gossip/peer"
	sampling2 "github.com/Ishan27gOrg/gossipProtocol/gossip/sampling"
	"github.com/Ishan27gOrg/vClock"
	"github.com/hashicorp/go-hclog"
)

var loggerOn bool
var defaultHashMethod = hash

type Gossip interface {
	// JoinWithSampling starts the gossip protocol with these initial peers. Peer sampling is done to periodically
	// maintain a partial view (subset) of the gossip network. Data is sent of the channel when gossip
	// is received from a peer or from the user (StartRumour)
	JoinWithSampling(peers []peer.Peer, newGossip chan Packet)
	// JoinWithoutSampling starts the gossip protocol with these initial peers. Peers are iteratively selected.
	// Data is sent of the channel when gossip is received from a peer or from the user (StartRumour)
	JoinWithoutSampling(peers func() []peer.Peer, newGossip chan Packet)
	// StartRumour is the equivalent of receiving a gossip message from the user. This is sent to peers
	StartRumour(data string)
	ReceiveGossip() chan Packet // gossip from user/peer : packet
	// RemovePacket will return the packet and its latest event clock after removing it from memory
	// Should be called after Any new packet with this id will initiate a new clock
	RemovePacket(id string) (*Packet, vClock.EventClock)
	CurrentPeers() []string
}

type gossip struct {
	c                 *config       // gossip protocol configuration
	env               envConfig     // env
	udp               client.Client // udp client
	logger            hclog.Logger
	peerSelector      func() peer.Peer
	peers             func() []peer.Peer
	viewCb            func(view sampling2.View, from peer.Peer) []byte
	newGossipPacket   chan Packet                   // gossip from user/peer : packet
	receivedGossipMap map[string]*Packet            // map of all gossip id & gossip data received
	eventClock        map[string]vClock.VectorClock // map of gossip id & event vector clock
}

func (g *gossip) CurrentPeers() []string {
	var udpAddresses []string
	for _, p := range g.peers() {
		udpAddresses = append(udpAddresses, p.UdpAddress)
	}
	return udpAddresses
}

func (g *gossip) ReceiveGossip() chan Packet {
	return g.newGossipPacket
}

// RemovePacket will return the packet and its latest event clock after removing it from memory
// Should be called after Any new packet with this id will initiate a new clock
func (g *gossip) RemovePacket(id string) (*Packet, vClock.EventClock) {
	packet, clock := g.receivedGossipMap[id], g.eventClock[id].Get(id)
	delete(g.receivedGossipMap, id)
	delete(g.eventClock, id)
	return packet, clock
}

// gossipCb is called when the udp server receives a gossip message. This is sent by peers and gossiped to peers
// response for a gossip message sent to a peer is not used, returns anything
func (g *gossip) gossipCb(gossip Packet, from string) []byte {
	id := gossip.GossipMessage.GossipMessageHash

	if g.eventClock[id] == nil {
		g.eventClock[id] = vClock.Init(g.processIdentifier())
	}
	// from peer, clock receive event
	g.eventClock[id].ReceiveEvent(id, gossip.VectorClock)

	// if new id -> clock send event
	go func() {
		if g.startRumour(gossip) {
			gossip.VectorClock = g.eventClock[id].Get(id)
			g.newGossipPacket <- gossip
			fmt.Println("Sent to channel---------")
		}
	}()
	return []byte("OKAY")
}

// StartRumour is the equivalent of receiving a gossip message from the user. This is gossiped to peers
func (g *gossip) StartRumour(data string) {
	gP := NewGossipMessage(data, g.processIdentifier(), nil)
	gP.GossipMessage.Version++
	g.startRumour(gP) // clock send event
	clock := g.eventClock[gP.GetId()].Get(gP.GetId())
	g.newGossipPacket <- Packet{
		AvailableAt:   gP.AvailableAt,
		GossipMessage: gP.GossipMessage,
		VectorClock:   clock, // return updated clock
	}
}
func (g *gossip) startRumour(gP Packet) bool {
	newGossip := false

	if g.receivedGossipMap[gP.GetId()] == nil {
		newGossip = true
		gP.AvailableAt = append(gP.AvailableAt, g.processIdentifier())
		g.logger.Trace("Received new gossip [version-" + strconv.Itoa(gP.GetVersion()) + "] " + gP.GetId() + " gossiping....✅")
		g.receivedGossipMap[gP.GetId()] = &gP
		g.beginGossipRounds(gP.GossipMessage)
	} else {
		g.logger.Trace("Received existing gossip [version-" + strconv.Itoa(gP.GetVersion()) + "] " + gP.GetId() +
			"checking version, not gossipping....❌")
		if g.receivedGossipMap[gP.GetId()].GetVersion() < gP.GetVersion() {
			g.receivedGossipMap[gP.GetId()] = &gP
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
	rounds := 1
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
	var peers []peer.Peer
	goto selectPeers
selectPeers:
	{
		for i := 1; i <= g.c.FanOut; i++ {
			peer := g.peerSelector()
			if peer.UdpAddress != g.udpAddr() {
				fmt.Println("Added peer - ", peer.UdpAddress, " at ", g.env.Hostname+g.env.UdpPort)
				peers = append(peers, peer)
			} else {
			}
			if len(peers) == g.c.FanOut {
				break
			}
		}
	}
	if len(peers) != g.c.FanOut && g.peerSelector().UdpAddress != g.udpAddr() { // only 1 peer, self
		goto selectPeers
	}

	id := gm.GossipMessageHash
	if g.eventClock[id] == nil {
		g.eventClock[id] = vClock.Init(g.processIdentifier())
	}
	if len(peers) == 0 {
		return
	}
	for _, peer := range peers {
		tmp := g.eventClock[id]
		clock := tmp.SendEvent(id, []string{peer.ProcessIdentifier})
		g.logger.Info("Gossipping Id - [ " + id + " ] to peer - " + peer.UdpAddress)
		if g.udp.SendGossip(peer.UdpAddress, gossipToByte(gm, g.processIdentifier(), clock)) != nil {
			g.eventClock[id] = tmp
			fmt.Println("Sent")
		}
	}
}

func (g *gossip) udpAddr() string {
	return g.env.Hostname + g.env.UdpPort
}

// processIdentifier identifies packets sent by this process
func (g *gossip) processIdentifier() string {
	return g.env.ProcessIdentifier
}

// JoinWithSampling starts the gossip protocol with these initial peers. Peer sampling is done to periodically
// maintain a partial view (subset) of the gossip network. Data is sent of the channel when gossip
// is received from a peer or from the user (StartRumour)
func (g *gossip) JoinWithSampling(peers []peer.Peer, newGossip chan Packet) {

	ps := sampling2.Init(g.udpAddr(), g.processIdentifier(), loggerOn)
	g.peers = func() []peer.Peer {
		return peers
	}
	g.peerSelector = func() peer.Peer {
		return ps.SelectPeer()
	}
	g.viewCb = ps.ReceiveView

	g.newGossipPacket = newGossip
	// listen calls the udp server to Start and registers callbacks for incoming gossip or views from peers
	go Listen(g.env.UdpPort, g.gossipCb, g.viewCb)
	// Start peer sampling and exchange views
	go ps.Start(g.peers())
}

func (g *gossip) JoinWithoutSampling(peersFn func() []peer.Peer, newGossip chan Packet) {
	rand.Seed(time.Now().Unix())
	g.peers = peersFn
	var randomPeer = func() peer.Peer {
		peers := g.peers()
		fmt.Println("PEERS - ", peers)
		if len(peers) == 0 {
			return peers[0]
		}
		return peers[rand.Intn(len(peers))]
	}
	var noViewExchange = func(view sampling2.View, s peer.Peer) []byte {
		return sampling2.ViewToBytes(view, peer.Peer{
			UdpAddress:        g.udpAddr(),
			ProcessIdentifier: g.processIdentifier(),
		})
	}
	g.peerSelector = randomPeer
	g.viewCb = noViewExchange

	g.newGossipPacket = newGossip
	// listen calls the udp server to Start and registers callbacks for incoming gossip or views from peers
	go Listen(g.env.UdpPort, g.gossipCb, noViewExchange)

	fmt.Println("STARTED - ", g.udpAddr(), g.processIdentifier())
}
func withConfig(hostname, port, selfAddress string, c *config) Gossip {

	g := gossip{
		logger:            nil,
		udp:               client.GetClient(port),
		env:               defaultEnv(hostname, port, selfAddress),
		receivedGossipMap: make(map[string]*Packet),
		peerSelector:      nil,
		peers:             nil,
		eventClock:        make(map[string]vClock.VectorClock),
		c:                 c,
		newGossipPacket:   make(chan Packet),
	}
	g.logger = mLogger.New(port)
	if !loggerOn {
		fmt.Println("logger off")
		g.logger.SetLevel(hclog.Off)
	} else {
		fmt.Println("logger on")
	}
	return &g
}
