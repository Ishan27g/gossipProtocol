package gossip

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	sll "github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/hashicorp/go-hclog"
)

type Sampling interface {
	Start()
	GetPeer() string
	ReceivedView(view View, from string) []byte

	setUdp(udp Client)
	fillView(peers ...string) // view to select peers from based on strategy
	passive()
	active()
	getView() *View
	selectView(view *View)

	selectPeer() string
}

type peerSampling struct {
	logger         hclog.Logger
	wait           time.Duration
	strategy       PeerSamplingStrategy
	view           View
	selfDescriptor NodeDescriptor // udp port
	receivedView   chan passiveView
	udp            Client
}

func (p *peerSampling) Start() {
	// fill view from possible peers
	p.fillView(NetworkPeers(p.selfDescriptor.Address)...)
	// select a subset of whole view
	p.selectView(p.getView())
	go p.active()
	go p.passive()
}

type passiveView struct {
	view     View
	fromPeer string
}

// GetPeer returns a random peer from a partial view of the entire network
func (p *peerSampling) GetPeer() string {
	rand.Seed(time.Now().Unix())
	node, _ := p.view.Nodes.Get(rand.Intn(p.view.Nodes.Size()))
	return node.(NodeDescriptor).Address
}

func (p *peerSampling) ReceivedView(view View, from string) []byte {
	p.receivedView <- passiveView{
		view:     view,
		fromPeer: from,
	}
	if p.strategy.ViewPropagationStrategy == Push {
		return ViewToBytes(View{Nodes: sll.New()}) // no response needed
	} else {
		return ViewToBytes(p.view)
	}
}

// fillView fills the current view as these peers
func (p *peerSampling) fillView(peers ...string) {
	for _, peer := range peers {
		p.view.Nodes.Add(NodeDescriptor{
			Address: peer,
			Hop:     0,
		})
	}
	p.view.sortNodes()
}

func (p *peerSampling) passive() {
	for {
		receivedView := new(View)
		<-time.After(p.wait)
		p.logger.Info("[PT]:ViewPropagationStrategy is - [" + strconv.Itoa(p.strategy.ViewPropagationStrategy) + "]")
		peer := p.selectPeer()
		p.logger.Debug("[PT]:Propagating view to - [" + peer + "]")
		if p.strategy.ViewPropagationStrategy == Push {
			mergedView := mergeView(p.view, View{Nodes: sll.New(p.selfDescriptor)})
			// send mergedView to peer, receive Ok or view
			p.logger.Debug("[PT]:PUSH : sending current view merged with self descriptor to peer - " + peer)
			p.logger.Info(printView(mergedView))
			rspView, err := BytesToView(p.udp.ExchangeView(peer, ViewToBytes(mergedView)))
			if err == nil {
				receivedView := &rspView
				p.logger.Debug("[PT]:PUSH : receivedView from peer - " + peer + receivedView.Nodes.String()) // []byte("OKAY")
				// printView(*receivedView)
			}
		} else {
			p.logger.Debug("[PT]:PULL/PUSH-PULL :sending empty view to peer - " + peer)
			// send emptyView to peer to trigger response
			rspView, err := BytesToView(p.udp.ExchangeView(peer, ViewToBytes(View{Nodes: sll.New()})))
			if err == nil {
				receivedView := &rspView
				p.logger.Debug("[PT]:PULL/PUSH-PULL : receivedView from peer - " + peer + receivedView.Nodes.String()) // []byte("OKAY")
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

func (p *peerSampling) active() {
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
			receivedView.view, _ = BytesToView(p.udp.ExchangeView(receivedView.fromPeer, ViewToBytes(mergedView)))
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

func (p *peerSampling) getView() *View {
	return &p.view
}

func (p *peerSampling) selectView(view *View) {
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
		view.RandomView()
	case Head: // select first MaxNodesInView nodes
		view.headView()
	case Tail: // select last MaxNodesInView nodes
		view.tailView()
	}
	p.view = *view
}

// returns a peer address from the network group
func (p *peerSampling) selectPeer() string {
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

func (p *peerSampling) setUdp(udp Client) {
	p.udp = udp
}
func selfDescriptor(n NodeDescriptor) View {
	return View{Nodes: sll.New(n)}
}
func Init(strategy PeerSamplingStrategy, self string) Sampling {
	ps := peerSampling{
		logger:   mLogger.Get("peer-sampling"),
		wait:     ViewExchangeDelay,
		strategy: strategy,
		view: View{
			Nodes: sll.New(),
		},
		selfDescriptor: NodeDescriptor{
			Address: self,
			Hop:     0,
		},
		receivedView: make(chan passiveView),
		udp:          nil,
	}
	return &ps
}
