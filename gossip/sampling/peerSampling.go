package sampling

import (
	"strconv"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip/client"
	sll "github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/hashicorp/go-hclog"
)

const (
	ViewExchangeDelay = 5 * time.Second // timeout after which a View  is exchanged with a peer
	MaxNodesInView    = 6               // max peers kept in local View TODO MaxNodesInView=6
)

type Sampling interface {
	Start(initialPeers []string)

	SelectPeer() string
	ReceiveView(view View, from string) []byte
	fillView(peers ...string) // view to select peers from based on strategy
	passive()
	active()
	GetView() *View
	selectView(view *View)
}

type peerSampling struct {
	logger         hclog.Logger
	wait           time.Duration
	strategy       PeerSamplingStrategy
	view           View
	selfDescriptor NodeDescriptor // udp port
	receivedView   chan passiveView
	udp            client.Client
}

func (p *peerSampling) Start(initialPeers []string) {
	p.logger.Trace("Starting sampling on - " + p.selfDescriptor.Address)
	// fill view from possible peers
	p.fillView(initialPeers...)
	// select a subset of whole view
	p.selectView(p.GetView())
	go p.active()
	go p.passive()
	p.logger.Trace("Started sampling on - " + p.selfDescriptor.Address)
}

type passiveView struct {
	view     View
	fromPeer string
}

func (p *peerSampling) ReceiveView(view View, from string) []byte {
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
		p.logger.Info("[PT]:waiting for - [" + p.wait.String() + "]")
		<-time.After(p.wait)
		p.logger.Info("[PT]:ViewPropagationStrategy is - [" + strconv.Itoa(p.strategy.ViewPropagationStrategy) + "]")
		peer := p.SelectPeer()
		p.logger.Debug("[PT]:Propagating view to - [" + peer + "]")
		if p.strategy.ViewPropagationStrategy == Push {
			mergedView := MergeView(p.view, View{Nodes: sll.New(p.selfDescriptor)})
			// send mergedView to peer, receive Ok or view
			p.logger.Debug("[PT]:PUSH : sending current view merged with self descriptor to peer - " + peer)
			p.logger.Info(PrintView(mergedView))
			rspView, err := BytesToView(p.udp.ExchangeView(peer, ViewToBytes(mergedView)))
			if err == nil {
				receivedView := &rspView
				p.logger.Debug("[PT]:PUSH : receivedView from peer - " + peer + receivedView.Nodes.String()) // []byte("OKAY")
				// PrintView(*receivedView)
			}
		} else {
			p.logger.Debug("[PT]:PULL/PUSH-PULL :sending empty view to peer - " + peer)
			// send emptyView to peer to trigger response
			rspView, err := BytesToView(p.udp.ExchangeView(peer, ViewToBytes(View{Nodes: sll.New()})))
			if err == nil {
				receivedView := &rspView
				p.logger.Debug("[PT]:PULL/PUSH-PULL : receivedView from peer - " + peer + receivedView.Nodes.String()) // []byte("OKAY")
				//PrintView(*receivedView)
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
				//PrintView(p.view)
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
		PrintView(receivedView.view)
		increaseHopCount(&receivedView.view)
		//PrintView(receivedView.view)
		if p.strategy.ViewPropagationStrategy == Pull {
			p.logger.Trace("[AT]:PULL - merging current view with self")
			mergedView := MergeView(p.view, selfDescriptor(p.selfDescriptor))
			// send mergedView to peer, receive Ok or view
			p.logger.Debug("[AT]:PULL : sending current view merged with self descriptor to peer - " + receivedView.fromPeer)
			receivedView.view, _ = BytesToView(p.udp.ExchangeView(receivedView.fromPeer, ViewToBytes(mergedView)))
			p.logger.Debug("[AT]:PULL : response view from peer - " + receivedView.fromPeer)
			//PrintView(receivedView.view)
		}
		p.logger.Debug("[AT] merging received view with current view ")
		receivedView.view = mergeViewExcludeNode(receivedView.view, p.view, p.selfDescriptor)
		p.logger.Debug(PrintView(receivedView.view))
		p.logger.Debug("[AT] selecting subset of view as current view")
		p.selectView(&receivedView.view)
		// p.logger.Debug(PrintView(receivedView.view))
		// p.view = receivedView.view  // set inside selectView
		p.logger.Debug(PrintView(p.view))
	}
}

func (p *peerSampling) GetView() *View {
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

// SelectPeer returns a peer address from the network group
// based on strategy and current view
func (p *peerSampling) SelectPeer() string {
	address := ""
	switch p.strategy.PeerSelectionStrategy {
	case Random: // select random peer
		address = p.view.randomNode()
	case Head: // select peer with the lowest hop
		address = p.view.headNode()
	case Tail: // select peer with the highest hop
		address = p.view.tailNode()
	}
	p.logger.Debug("SELECTED PEER " + address)
	//if strings.Compare(address, p.selfDescriptor.address) == 0{
	//	return p.SelectPeer()
	//}
	return address
}

func selfDescriptor(n NodeDescriptor) View {
	return View{Nodes: sll.New(n)}
}
func Init(self string) Sampling {
	ps := peerSampling{
		logger:   mLogger.Get("ps-" + self),
		wait:     ViewExchangeDelay,
		strategy: DefaultStrategy(),
		view: View{
			Nodes: sll.New(),
		},
		selfDescriptor: NodeDescriptor{
			Address: self,
			Hop:     0,
		},
		receivedView: make(chan passiveView),
		udp:          client.GetClient(self),
	}
	return &ps
}
