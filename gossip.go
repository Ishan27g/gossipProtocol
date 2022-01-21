package gossipProtocol

import (
	"context"
	"sync"
	"time"

	"github.com/Ishan27gOrg/vClock"
)

type Gossip interface {
	// Join with some initial peers
	Join(...Peer)
	// Add peers
	Add(...Peer)
	// CurrentView returned as a String
	CurrentView() string
	// SendGossip  to the network
	SendGossip(data string)
}

type gossip struct {
	ctx       context.Context
	cancel    context.CancelFunc
	lock      sync.Mutex
	env       *envConfig
	sampling  iSampling
	udpClient client

	selfDescriptor Peer

	allGossip map[string]*Packet
	allEvents *vClock.VectorClock

	gossipToUser chan Packet
	// sentToPeers  []Peer // todo remove
}

func (g *gossip) CurrentView() string {
	return g.sampling.printView()
}

func (g *gossip) Join(initialPeers ...Peer) {
	g.sampling.SetInitialPeers(initialPeers...)
	g.sampling.Start()

	go Listen(g.ctx, g.env.UdpPort, g.serverCb, g.sampling.ViewFromPeer)
}
func (g *gossip) Add(peer ...Peer) {
	g.sampling.AddPeer(peer...)
}

// SendGossip from User to the network
func (g *gossip) SendGossip(data string) {
	gP := NewGossipMessage(data, g.env.ProcessIdentifier, nil)
	g.lock.Lock()
	newPacket := g.savePacket(&gP)
	g.lock.Unlock()
	if newPacket {

		g.gossip(gP.GossipMessage, g.selfDescriptor)
		g.lock.Lock()
		gP.VectorClock = (*g.allEvents).Get(gP.GetId()) // update packet's clock
		g.lock.Unlock()
		g.gossipToUser <- gP
		// println(g.selfDescriptor.ProcessIdentifier, " Returned TO USER")
	} else {
		println("\n\nnot possible")
	}
}

// from peer
func (g *gossip) serverCb(gP Packet, from Peer) []byte {
	g.lock.Lock()
	newPacket := g.savePacket(&gP)
	(*g.allEvents).ReceiveEvent(gP.GetId(), gP.VectorClock)
	g.lock.Unlock()

	go func() {
		if newPacket {
			g.gossip(gP.GossipMessage, from)
			g.lock.Lock()
			gP.VectorClock = (*g.allEvents).Get(gP.GetId()) // update packet's clock
			g.lock.Unlock()
			g.gossipToUser <- gP
			// println(g.selfDescriptor.ProcessIdentifier, " SENT TO USER")
		}
	}()

	return []byte("OKAY")
}
func (g *gossip) savePacket(gP *Packet) bool {
	newGossip := false
	if g.allGossip[gP.GetId()] == nil {
		gP.AvailableAt = append(gP.AvailableAt, g.selfDescriptor.ProcessIdentifier)
		g.allGossip[gP.GetId()] = gP
		newGossip = true
	} else {
		if g.allGossip[gP.GetId()].GetVersion() < gP.GetVersion() {
			g.allGossip[gP.GetId()] = gP
		}
	}
	return newGossip
}
func (g *gossip) gossip(gm gossipMessage, exclude Peer) {

	for i := 1; i <= rounds; i++ {
		<-time.After(g.env.RoundDelay)
		if g.sampling.Size() == 0 {
			return
		}
		//	g.sentToPeers = nil
		g.fanOut(gm, exclude)
		gm.Version++

		//for _, peer := range g.sentToPeers {
		//	fmt.Println(g.selfDescriptor.ProcessIdentifier, "Send to - ", peer.ProcessIdentifier)
		//}
	}

}

func (g *gossip) fanOut(gm gossipMessage, exclude Peer) {
	id := gm.GossipMessageHash
	if g.sampling.Size() == 0 {
		return
	}
	for i := 0; i < g.env.FanOut; i++ {
		peer := g.sampling.GetPeer(exclude)
		if peer.UdpAddress != "" && peer.ProcessIdentifier != g.selfDescriptor.ProcessIdentifier {
			g.lock.Lock()
			tmp := vClock.Copy(*g.allEvents)
			clock := tmp.SendEvent(id, []string{peer.ProcessIdentifier})
			// println("Gossippinnnng Id - [ " + id + " ] to peer - " + peer.UdpAddress)
			buffer := gossipToByte(gm, g.selfDescriptor, clock)
			g.lock.Unlock()
			if g.udpClient.send(peer.UdpAddress, buffer) != nil {
				g.lock.Lock()
				*g.allEvents = tmp
				g.lock.Unlock()
				//	g.sentToPeers = append(g.sentToPeers, peer)
			} else {
				g.sampling.removePeer(peer)
				g.lock.Lock()
				(*g.allEvents).SendEvent(id, nil)
				g.lock.Unlock()
			}

		}
	}
}

func Config(hostname string, port string, id string) (Gossip, <-chan Packet) {
	ctx, cancel := context.WithCancel(context.Background())
	g := gossip{
		lock:   sync.Mutex{},
		ctx:    ctx,
		cancel: cancel,
		selfDescriptor: Peer{
			UdpAddress:        hostname + ":" + port,
			Hop:               0,
			ProcessIdentifier: id,
		},
		env:          defaultEnv(hostname, port, id),
		udpClient:    getClient(id),
		allGossip:    make(map[string]*Packet),
		allEvents:    new(vClock.VectorClock),
		gossipToUser: make(chan Packet, 100),
		sampling:     initSampling(port, id, defaultStrategy),
		//sentToPeers:  []Peer{},
	}
	*g.allEvents = vClock.Init(id)
	return &g, g.gossipToUser
}
