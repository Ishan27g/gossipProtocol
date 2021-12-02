package gossip

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	reg "github.com/Ishan27gOrg/registry/package"
	"github.com/hashicorp/go-hclog"
)

const RegistryUrl = "https://bootstrap-registry.herokuapp.com"

type Gossip interface {
	// StartRumour is the equivalent of receiving a gossip message from the user. This is sent to peers
	StartRumour(data string)
}

// StartRumour is the equivalent of receiving a gossip message from the user. This is gossiped to peers
func (g *gossip) StartRumour(data string) {
	g.startRumour(newGossipMessage(data, g.selfAddress()))
}
func (g *gossip) startRumour(gP Packet) bool {
	newGossip := false
	if g.receivedGossipMap[gP.GossipMessage.GossipMessageHash] == nil {
		newGossip = true
		g.logger.Debug("Received new gossip - " + gP.GossipMessage.GossipMessageHash + " from - " + gP.AvailableAt[0] + " gossiping....✅")
		g.receivedGossipMap[gP.GossipMessage.GossipMessageHash] = &gP
		go g.beginGossipRounds(gP.GossipMessage)
	} else {
		gPExisting := g.receivedGossipMap[gP.GossipMessage.GossipMessageHash]
		gPExisting.AvailableAt = append(gPExisting.AvailableAt, gP.AvailableAt[0])
		g.logger.Debug("Received existing gossip - " + gP.GossipMessage.GossipMessageHash + " adding new download address, not gossipping....❌")
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
		peer := g.peerSampling.GetPeer()
		g.logger.Debug("[Gossip " + gm.GossipMessageHash + " ] to peer - " + peer)
		if gossipRsp := g.udp.SendGossip(peer, gossipToByte(gm, g.selfAddress())); gossipRsp != nil {
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
// todo send to raft->leader if newGossip
func (g *gossip) gossipCb(gossip Packet) []byte {
	_ = g.startRumour(gossip)
	return []byte("OKAY")
}
func (g *gossip) selfAddress() string {
	self := g.env.Hostname + ":" + g.env.UdpPort
	return self
}

type gossip struct {
	gossipChan        chan Packet        // gossip from user or udp server, to be sent to peers
	receivedGossipMap map[string]*Packet // map of all gossip received
	udp               Client             // udp client
	env               EnvCfg             // env

	mutex        sync.Mutex
	logger       hclog.Logger
	peerSampling Sampling
}

type EnvCfg struct {
	Hostname     string `env:"HOST_NAME"`
	Zone         int    `env:"ZONE"`
	UdpPort      string `env:"UDP_PORT,required"`
	HttpFilePort string `env:"HTTP_PORT_FILE,required"` // todo data download endpoint
}

func Default(env EnvCfg) Gossip {
	g := gossip{
		mutex:             sync.Mutex{},
		logger:            mLogger.Get("gossip"),
		receivedGossipMap: make(map[string]*Packet),
		udp:               nil,
		env:               env,
		peerSampling:      nil,
	}

	peers := RegisterAndGetZonePeers(g.env.Zone, g.selfAddress(), nil)
	g.logger.Info(fmt.Sprintf("ZONE-PEERS -> %v", peers))

	g.set(GetClient(), env, Init(DefaultStrategy(), g.selfAddress()))
	return &g
}

func (g *gossip) set(udp Client, env EnvCfg, ps Sampling) Gossip {
	g.peerSampling = ps
	g.env = env
	g.udp = udp
	g.peerSampling.setUdp(udp)

	// listen calls the udp server to start and registers callbacks for incoming gossip or views from peers
	go Listen(g.env.UdpPort, g.gossipCb, g.peerSampling.ReceivedView)
	// start peer sampling and exchange views
	go g.peerSampling.Start()

	return g
}

func RegisterAndGetZonePeers(zone int, address string, meta reg.MetaData) []string {
	var raft_peers []string
	for _, p := range reg.RegistryClient(RegistryUrl).Register(zone, address, meta) {
		raft_peers = append(raft_peers, p.Address)
	}
	return raft_peers
}
