package gossipProtocol

import (
	"context"
	"math/rand"
	"time"

	sll "github.com/emirpasic/gods/lists/singlylinkedlist"
)

const (
	ViewExchangeDelay = 3 * time.Second // timeout after which a View  is exchanged with a peer
	MaxNodesInView    = 6               // max peers kept in local View TODO MaxNodesInView=6
)

type iSampling interface {
	SetInitialPeers(...Peer)
	Start()
	AddPeer(...Peer)
	GetPeer(exclude Peer) Peer
	ViewFromPeer(View, Peer) []byte
	Size() int
	removePeer(peer Peer)
	printView() string
	getView() View
}
type sampling struct {
	ctx            context.Context
	cancel         context.CancelFunc
	strategy       PeerSamplingStrategy
	view           View
	selfDescriptor Peer
	knownPeers     map[string]Peer
	viewFromPeer   chan passiveView
	udpClient      client

	previousPeer Peer
}

func (s *sampling) getView() View {
	return s.view
}
func (s *sampling) printView() string {
	return PrintView(s.view)
}
func (s *sampling) Size() int {
	return s.view.Nodes.Size()
}
func (s *sampling) Start() {
	go s.active()
	go s.passive()
}
func (s *sampling) SetInitialPeers(initialPeers ...Peer) {
	s.fillView(initialPeers...)
	s.selectView(&s.view)
}

// fillView fills the current view as these peers
func (s *sampling) fillView(peers ...Peer) {
	for _, peer := range peers {
		s.addPeerToView(peer)
	}
	s.view.sortNodes()
}

func (s *sampling) addPeerToView(peer Peer) {
	s.view.add(peer)
	s.knownPeers[peer.ProcessIdentifier] = peer
}

func (s *sampling) removePeer(peer Peer) {
	// s.lock.Lock()
	// defer s.lock.Unlock()
	s.view.remove(peer)
	delete(s.knownPeers, peer.ProcessIdentifier)
	println(s.selfDescriptor.ProcessIdentifier, " - Removed peer from view - ", peer.ProcessIdentifier)
	println(s.selfDescriptor.ProcessIdentifier, PrintView(s.view))
}
func (s *sampling) selectView(view *View) {
	switch s.strategy.ViewSelectionStrategy {
	case Random: // select random MaxNodesInView nodes
		view.RandomView()
	case Head: // select first MaxNodesInView nodes
		view.headView()
	case Tail: // select last MaxNodesInView nodes
		view.tailView()
	}

	s.view = mergeViewExcludeNode(*view, *view, s.selfDescriptor)
}

func (s *sampling) active() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case receivedView := <-s.viewFromPeer: // wait to receive a view

			increaseHopCount(&receivedView.view)
			if s.strategy.ViewPropagationStrategy == Pull {
				mergedView := MergeView(receivedView.view, selfDescriptor(s.selfDescriptor))
				// send mergedView to peer, receive Ok or view
				view, from, err := BytesToView(s.udpClient.send(receivedView.from.UdpAddress,
					ViewToBytes(mergedView, s.knownPeers[s.selfDescriptor.ProcessIdentifier])))
				if err == nil {
					s.knownPeers[from.ProcessIdentifier] = from
					receivedView.view = view
				}
			}
			receivedView.view = mergeViewExcludeNode(s.view, receivedView.view, s.selfDescriptor)
			s.selectView(&receivedView.view)

			//println(s.selfDescriptor.ProcessIdentifier, " [PT]  : ", PrintView(s.view))
		}
	}
}

func (s *sampling) passive() {
	wait := 5 * time.Second
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-time.After((wait) + time.Duration(rand.Intn(int(wait.Milliseconds()/2)))): // add random delay

			receivedView := new(View)
			nwPeer := s.getPeer()
			if nwPeer.UdpAddress == "" {
				continue
			}
			if s.strategy.ViewPropagationStrategy == Push || s.strategy.ViewPropagationStrategy == PushPull {
				mergedView := MergeView(s.view, selfDescriptor(s.selfDescriptor))
				buffer := ViewToBytes(mergedView, s.knownPeers[s.selfDescriptor.ProcessIdentifier])
				rspView, from, err := BytesToView(s.udpClient.send(nwPeer.UdpAddress, buffer))
				if err == nil {
					s.knownPeers[from.ProcessIdentifier] = from
					_ = &rspView // for PUSH -> []byte("OKAY")
				} else {
					s.removePeer(nwPeer)
				}
			} else {
				// send emptyView to nwPeer to trigger response
				rspView, from, err := BytesToView(s.udpClient.send(nwPeer.UdpAddress,
					ViewToBytes(View{Nodes: sll.New()}, s.knownPeers[s.selfDescriptor.ProcessIdentifier])))
				if err == nil {
					s.knownPeers[from.ProcessIdentifier] = from
					receivedView = &rspView
				} else {
					s.removePeer(nwPeer)
				}
			}
			if s.strategy.ViewPropagationStrategy == Pull || s.strategy.ViewPropagationStrategy == PushPull {
				if receivedView.Nodes != nil {
					increaseHopCount(receivedView)
					mergedView := mergeViewExcludeNode(s.view, *receivedView, s.selfDescriptor)
					s.selectView(&mergedView)
				}
			}
			println(s.selfDescriptor.ProcessIdentifier, "END - ", PrintView(s.view))
		}
	}
}

func (s *sampling) ViewFromPeer(receivedView View, peer Peer) []byte {
	s.knownPeers[peer.ProcessIdentifier] = peer
	var rsp []byte
	increaseHopCount(&receivedView)
	if s.strategy.ViewPropagationStrategy == Pull || s.strategy.ViewPropagationStrategy == PushPull {
		mergedView := MergeView(s.view, selfDescriptor(s.selfDescriptor))
		rsp = ViewToBytes(mergedView, s.knownPeers[s.selfDescriptor.ProcessIdentifier])
	}
	merged := mergeViewExcludeNode(s.view, receivedView, s.selfDescriptor)
	s.selectView(&merged)
	println("After receiving view, current view is - ", PrintView(s.view))
	return rsp // empty incase of PUSH
}

func (s *sampling) AddPeer(peer ...Peer) {
	for _, p := range peer {
		s.addPeerToView(p)
	}
	s.selectView(&s.view)
}

// getPeer returns a peer from the current view except self
func (s *sampling) GetPeer(exclude Peer) Peer {
	// if s.view.Nodes.Size() <= 1 { // only self in view
	// 	return Peer{}
	// }
	rand.Seed(rand.Int63n(100000))

	node := Peer{}
	goto selectPeer
selectPeer:
	{
		node = s.view.randomNode()
	}

	if node.ProcessIdentifier == s.selfDescriptor.ProcessIdentifier {
		goto selectPeer
	}
	if s.previousPeer.ProcessIdentifier == Peer(node).ProcessIdentifier && s.Size() != 1 {
		goto selectPeer
	}
	if exclude.ProcessIdentifier == Peer(node).ProcessIdentifier && s.Size() != 1 {
		goto selectPeer
	}

	s.previousPeer = Peer(node)
	return s.previousPeer
}

// getPeer returns a peer from the current view based on applied strategy
func (s *sampling) getPeer() Peer {
	node := Peer{}
	switch s.strategy.PeerSelectionStrategy {
	case Random: // select random peer
		node = s.view.randomNode()
	case Head: // select peer with the lowest hop
		node = s.view.headNode()
	case Tail: // select peer with the highest hop
		node = s.view.tailNode()
	}
	// println(s.selfDescriptor.ProcessIdentifier, " - SELECTED PEER ", node.UdpAddress)
	return Peer(node)
}

func initSampling(udpAddress string, identifier string, strategy PeerSamplingStrategy) iSampling {
	ctx, cancel := context.WithCancel(context.Background())
	s := sampling{
		ctx:      ctx,
		cancel:   cancel,
		strategy: strategy,
		view: View{
			Nodes: sll.New(),
		},
		selfDescriptor: Peer{
			UdpAddress:        udpAddress,
			Hop:               0,
			ProcessIdentifier: identifier,
		},
		knownPeers:   make(map[string]Peer),
		viewFromPeer: make(chan passiveView),
		udpClient:    getClient(identifier),
		previousPeer: Peer{},
	}
	s.knownPeers[s.selfDescriptor.ProcessIdentifier] = Peer(s.selfDescriptor)
	return &s
}
