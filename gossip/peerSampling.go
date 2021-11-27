package gossip

/**
Peer Sampling Service based on http://lpdwww.epfl.ch/upload/documents/publications/neg--1184036295all.pdf
*/
import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	reg "github.com/Ishan27gOrg/registry/package"

	"github.com/Ishan27g/go-utils/mLogger"
	sll "github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/hashicorp/go-hclog"
)

// NetworkPeers returns the set of all possible peers
func NetworkPeers() []string {
	self := Env.envCfg.Hostname + ":" + Env.envCfg.UdpPort
	var networkPeers []string
	for _, peer := range minPeers {
		if strings.Compare(self, peer) != 0 {
			networkPeers = append(networkPeers, peer)
		}
	}
	return networkPeers
}

type nodeDescriptor struct {
	address string
	hop     int
}

type PeerSampling interface {
	// getPeer returns a random peer from a partial view of the entire network
	getPeer() string
	receivedAView(view view, address string) view
}
type passiveView struct {
	view     view
	fromPeer string
}
type pSampling struct {
	logger         hclog.Logger
	wait           time.Duration
	strategy       PeerSamplingStrategy
	view           view
	selfDescriptor nodeDescriptor
	receivedView   chan passiveView
}

// receivedAView from a peer over UDP (udpServer)
func (p *pSampling) receivedAView(v view, address string) view {
	p.receivedView <- passiveView{
		view:     v,
		fromPeer: address,
	}
	if p.strategy.ViewPropagationStrategy == Push {
		return view{nodes: sll.New()} // no response needed
	} else {
		return p.view
	}
}

// GetPeer returns a random peer from a partial view of the entire network
func (p *pSampling) getPeer() string {
	rand.Seed(time.Now().Unix())
	node, _ := p.view.nodes.Get(rand.Intn(p.view.nodes.Size()))
	return node.(nodeDescriptor).address
}

func initPeerSampling(st PeerSamplingStrategy, peers reg.PeerResponse) PeerSampling {
	ps := &pSampling{
		logger:   mLogger.Get("peer-sampling"),
		wait:     ViewExchangeDelay,
		strategy: st,
		view: view{
			nodes: sll.New(),
		},
		selfDescriptor: nodeDescriptor{
			address: Env.envCfg.Hostname + ":" + Env.envCfg.UdpPort,
			hop:     0,
		},
		receivedView: make(chan passiveView),
	}
	for _, peer := range peers {
		ps.view.nodes.Add(nodeDescriptor{
			address: peer.Address,
			hop:     0,
		})
	}

	ps.view.sortNodes()
	// select a subset of whole view
	ps.selectView(&ps.view)
	go ps.active()
	go ps.passive()
	return ps
}

func (p *pSampling) passive() {
	for {
		receivedView := new(view)
		<-time.After(p.wait)
		p.logger.Info("[PT]:ViewPropagationStrategy is - [" + strconv.Itoa(p.strategy.ViewPropagationStrategy) + "]")
		peer := p.selectPeer()
		p.logger.Debug("[PT]:Propagating view to - [" + peer + "]")
		if p.strategy.ViewPropagationStrategy == Push {
			mergedView := mergeView(p.view, view{nodes: sll.New(p.selfDescriptor)})
			// send mergedView to peer, receive Ok or view
			p.logger.Debug("[PT]:PUSH : sending current view merged with self descriptor to peer - " + peer)
			receivedView = udpSendView(peer, mergedView)
			if receivedView != nil {
				p.logger.Debug("[PT]:PUSH : receivedView from peer - " + peer)
				// printView(*receivedView)
			}
		} else {
			p.logger.Debug("[PT]:PULL/PUSH-PULL :sending empty view to peer - " + peer)
			// send emptyView to peer to trigger response
			receivedView = udpSendView(peer, view{nodes: sll.New()})
			if receivedView != nil {
				p.logger.Debug("[PT]:PULL/PUSH-PULL : receivedView from peer - " + peer)
				//printView(*receivedView)
			}
		}
		if p.strategy.ViewPropagationStrategy == Pull {
			p.logger.Info("[PT]:ViewPropagationStrategy is still- [" + strconv.Itoa(p.strategy.ViewPropagationStrategy) + "]")

			if receivedView != nil {
				p.logger.Debug("[PT]:PULL : increasing hop count for received view")
				increaseHopCount(receivedView)
				p.logger.Trace("[PT]:PULL : increased hop count for received view")
				p.logger.Trace("[PT]:PULL : merging received view with current view")
				mergedView := mergeViewExcludeNode(*receivedView, p.view, p.selfDescriptor)
				p.logger.Debug("[PT]:PULL : selected subset of mergedView")
				p.selectView(&mergedView)
				// p.view = mergedView // set inside selectView
				//printView(p.view)
			} else {
				p.logger.Debug("[PT]PULL : received nil from " + peer)
			}
		}
	}
}

func (p *pSampling) active() {
	for {
		// wait to receive a view
		receivedView := <-p.receivedView
		p.logger.Info("[AT]:Active thread received view, incrementing hop count")
		//printView(receivedView.view)
		increaseHopCount(&receivedView.view)
		//printView(receivedView.view)
		if p.strategy.ViewPropagationStrategy == Pull {
			p.logger.Trace("[AT]:PULL - merging current view with self")
			mergedView := mergeView(p.view, selfDescriptor(p.selfDescriptor))
			// send mergedView to peer, receive Ok or view
			p.logger.Debug("[AT]:PULL : sending current view merged with self descriptor to peer - " + receivedView.fromPeer)
			receivedView.view = *udpSendView(receivedView.fromPeer, mergedView)
			p.logger.Debug("[AT]:PULL : response view from peer - " + receivedView.fromPeer)
			//printView(receivedView.view)
		}
		p.logger.Debug("[AT] merging received view with current view ")
		receivedView.view = mergeViewExcludeNode(receivedView.view, p.view, p.selfDescriptor)
		p.logger.Debug(printView(receivedView.view))
		p.logger.Debug("[AT] selecting subset of view as current view")
		p.selectView(&receivedView.view)
		// p.logger.Debug(printView(receivedView.view))
		// p.view = receivedView.view  // set inside selectView
		p.logger.Debug(printView(p.view))
	}
}

func selfDescriptor(n nodeDescriptor) view {
	return view{nodes: sll.New(n)}
}

// returns a peer address from the network group
func (p *pSampling) selectPeer() string {
	address := ""
	switch p.strategy.PeerSelectionStrategy {
	case Random: // select random peer
		address = p.view.randomNode()
	case Head: // select peer with the lowest hop
		address = p.view.headNode()
	case Tail: // select peer with the highest hop
		address = p.view.tailNode()
	}
	p.logger.Info("SELECTED PEER " + address)
	//if strings.Compare(address, p.selfDescriptor.address) == 0{
	//	return p.selectPeer()
	//}
	return address
}

// selectView updates the view to be a subset of at most MaxNodesInView
func (p *pSampling) selectView(view *view) {
	// remove self if present in view
	/*
		if i := view.nodes.IndexOf(p.selfDescriptor); i != 1{
			view.nodes.Remove(i)
		}
		if exists, _ := view.checkExists(p.selfDescriptor.address); exists{
			if i := view.nodes.IndexOf(p.selfDescriptor); i != 1{
				view.nodes.Remove(i)
			}
		}
	*/
	switch p.strategy.ViewSelectionStrategy {
	case Random: // select random MaxNodesInView nodes
		view.randomView()
	case Head: // select first MaxNodesInView nodes
		view.headView()
	case Tail: // select last MaxNodesInView nodes
		view.tailView()
	}
	p.view = *view
}
