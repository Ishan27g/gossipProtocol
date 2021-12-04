package gossip

import (
	"strconv"
	"sync"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/hashicorp/go-hclog"
)

const RegistryUrl = "https://bootstrap-registry.herokuapp.com"

type Gossip interface {
	// Join starts the gossip protocol with these initial peers. Peer sampling is done to periodically
	// maintain a partial view (subset) of the gossip network. Data is sent of the channel when gossip
	// is received from a peer or from the user (StartRumour)
	Join(peers []string, fromPeer chan<- map[string]string, fromUser chan<- Packet)
	// StartRumour is the equivalent of receiving a gossip message from the user. This is sent to peers
	StartRumour(data string)

	AllGossip() map[string]*Packet
}

// StartRumour is the equivalent of receiving a gossip message from the user. This is gossiped to peers
func (g *gossip) StartRumour(data string) {
	gP := newGossipMessage(data, g.selfAddress())
	go g.startRumour(gP)
	g.gossipFromUser <- gP
}
func (g *gossip) startRumour(gP Packet) bool {
	newGossip := false
	if g.receivedGossipMap[gP.GossipMessage.GossipMessageHash] == nil {
		newGossip = true
		g.logger.Debug("Received new gossip - " + gP.GossipMessage.GossipMessageHash +
			" from - " + gP.AvailableAt[0] + " gossiping....✅")
		g.receivedGossipMap[gP.GossipMessage.GossipMessageHash] = &gP
		go g.beginGossipRounds(gP.GossipMessage)
	} else {
		gPExisting := g.receivedGossipMap[gP.GossipMessage.GossipMessageHash]
		gPExisting.AvailableAt = append(gPExisting.AvailableAt, gP.AvailableAt[0])
		g.logger.Debug("Received existing gossip - " + gP.GossipMessage.GossipMessageHash +
			" adding new download address, not gossipping....❌")
	}
	return newGossip
}

//
func (g *gossip) beginGossipRounds(gsp gossipMessage) {
	rounds := numRounds()
	g.logger.Trace("num of gossip rounds - " + strconv.Itoa(rounds))
	g.logger.Debug("[Gossip started] - " + gsp.GossipMessageHash)
	for i := 0; i < rounds; i++ {
		<-time.After(RoundDelay)
		g.sendGossip(gsp)
	}
	g.logger.Debug("[Gossip ended] - " + gsp.GossipMessageHash)
}

// sendGossip sends the gossip message to FanOut number of peers
func (g *gossip) sendGossip(gm gossipMessage) {
	gsp := make(chan []byte)
	for i := 0; i < FanOut; i++ {
		peer := g.peerSampling.getPeer()
		g.logger.Debug("[Gossip " + gm.GossipMessageHash + " ] to peer - " + peer)
		if gossipRsp := g.udp.sendGossip(peer, gossipToByte(gm, g.selfAddress())); gossipRsp != nil {
			gsp <- gossipRsp
		}
	}
	go func() {
		for i := 0; i < FanOut; i++ {
			gs := <-gsp
			g.logger.Debug("Received on main " + string(gs))
			ByteToPacket(gs)
		}
	}()
}

// gossipCb is called when the udp server receives a gossip message. This is sent by peers and gossiped to peers
// todo not used -> gossip response
func (g *gossip) gossipCb(gossip Packet, from string) []byte {
	if g.startRumour(gossip) { // if new id, send to raft -> mark as new receive event
		g.gossipEventFromPeer <- map[string]string{
			from: gossip.GossipMessage.GossipMessageHash,
		}
	}
	return []byte("OKAY")
}
func (g *gossip) selfAddress() string {
	self := g.env.Hostname + ":" + g.env.UdpPort
	return self
}

type gossip struct {
	env                 EnvCfg // env
	udp                 Client // udp client
	mutex               sync.Mutex
	logger              hclog.Logger
	peerSampling        Sampling
	gossipFromUser      chan<- Packet            // gossip from user === (sendEvent)
	gossipEventFromPeer chan<- map[string]string // gossip Id from peer === (receiveEvent)
	receivedGossipMap   map[string]*Packet       // map of all gossip id & gossip data received
}

func (g *gossip) AllGossip() map[string]*Packet {
	return g.receivedGossipMap
}

type EnvCfg struct {
	Hostname string `env:"HOST_NAME"`
	Zone     int    `env:"ZONE"`
	UdpPort  string `env:"UDP_PORT,required"`
}

// Default returns the default gossip protocol interface after
// sets up a default peer sampling strategy
func Default(hostname, port string, zone int) Gossip {
	g := gossip{
		mutex:  sync.Mutex{},
		logger: mLogger.Get("gossip"),
		udp:    nil,
		env: EnvCfg{
			Hostname: hostname,
			Zone:     zone,
			UdpPort:  port,
		},
		receivedGossipMap: make(map[string]*Packet),
		peerSampling:      nil,
	}
	g.peerSampling = InitPs(g.selfAddress())
	g.peerSampling.setUdp(GetClient())
	return &g
}

// Join starts the gossip protocol with these initial peers. Peer sampling is done to periodically
// maintain a partial view (subset) of the gossip network. Data is sent of the channel when gossip
// is received from a peer or from the user (StartRumour)
func (g *gossip) Join(peers []string, fromPeer chan<- map[string]string, fromUser chan<- Packet) {
	g.gossipEventFromPeer = fromPeer
	g.gossipFromUser = fromUser
	// listen calls the udp server to start and registers callbacks for incoming gossip or views from peers
	go Listen(g.env.UdpPort, g.gossipCb, g.peerSampling.ReceivedView)
	// start peer sampling and exchange views
	go g.peerSampling.start(peers)
}
