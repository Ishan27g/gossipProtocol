package gossip

import (
	"strconv"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	vClock "github.com/Ishan27gOrg/vClock"
	"github.com/hashicorp/go-hclog"
)

type Gossip interface {
	StartRumour(data string)
	GetGossipChannel() chan<- gossipMessage
}
type gossip struct {
	logger            hclog.Logger
	gossipChan        chan gossipMessage
	receivedGossipMap map[string]*gossipMessage
	vClock            vClock.VectorClock
}

func ListenForGossip(gossipChannel chan map[string]gossipMessage, v vClock.VectorClock) Gossip {
	gsp := gossip{
		logger:            mLogger.Get("gossip"),
		gossipChan:        make(chan gossipMessage),
		receivedGossipMap: make(map[string]*gossipMessage),
		vClock:            v,
	}
	go func() {
		for addr, gossip := range <-gossipChannel {
			gsp.vClock.ReceiveEvent(addr, gossip.vClock)
			gsp.gossipChan <- gossip
		}
	}()
	go gsp.start()
	return &gsp
}

// start receiving gossip from the gossipChannel. After receiving, check if this is already received.
// If new, gossip to peers
func (g *gossip) start() {
	for {
		gsp := <-g.gossipChan
		if g.receivedGossipMap[gsp.GossipMessageHash] == nil {
			g.logger.Debug("Received new gossip - " + gsp.GossipMessageHash)
			g.receivedGossipMap[gsp.GossipMessageHash] = &gsp
			go g.beginGossipRounds(gsp)
		} else {
			g.logger.Debug("Already present in map - " + gsp.GossipMessageHash)
			g.logger.Debug("ignoring this message")
		}
	}
}

//
func (g *gossip) beginGossipRounds(gsp gossipMessage) {
	rounds := numRounds()
	g.logger.Debug("Rounds - " + strconv.Itoa(rounds))
	g.logger.Debug("beginGossipRounds started for - " + gsp.GossipMessageHash)
	for i := 0; i < rounds; i++ {
		<-time.After(RoundDelay)
		g.logger.Debug("RoundNum - " + strconv.Itoa(i+1))
		g.sendGossip(gsp)
	}
	g.logger.Debug("beginGossipRounds ended for- " + gsp.GossipMessageHash)
}

func (g *gossip) GetGossipChannel() chan<- gossipMessage {
	return g.gossipChan
}

// sendGossip sends the gossip message to FanOut number of peers
func (g *gossip) sendGossip(gossip gossipMessage) {
	gsp := make(chan []byte)
	var peers []string
	for i := 0; i < FanOut; i++ {
		peers = append(peers, Env.ps.getPeer())
	}

	g.logger.Info("Vector clock - send event ")
	g.vClock.SendEvent(gossip.GossipMessageHash, peers)

	data := gossipToByte(gossip)
	for _, peer := range peers {
		g.logger.Debug("gossipping to peer " + peer)
		gossipRsp := udpSendGossip(peer, data)
		if gossipRsp != nil {
			gsp <- gossipRsp
		}
	}

	go func() {
		for i := 0; i < FanOut; i++ {
			gs := <-gsp
			g.logger.Debug("Received on main " + string(gs))
		}
	}()
}

func (g *gossip) StartRumour(data string) {
	g.gossipChan <- newGossipMessage(data, g.vClock.Get())
}
