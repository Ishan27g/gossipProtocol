package sampling

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/client"
	"github.com/Ishan27gOrg/gossipProtocol/peer"
	sll "github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/hashicorp/go-hclog"
)

const (
	ViewExchangeDelay = 5 * time.Second // timeout after which a View  is exchanged with a peer
	MaxNodesInView    = 6               // max peers kept in local View TODO MaxNodesInView=6
)

type Sampling interface {
	Start(ctx context.Context, initialPeers []peer.Peer)

	SelectPeer() peer.Peer
	ReceiveView(view View, from peer.Peer) []byte
	fillView(peers ...peer.Peer) // view to select peers from based on strategy
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
	selfIdentifier string
	peers          map[string]peer.Peer // udp-address:peer
	ctx            context.Context
}

func (p *peerSampling) Start(ctx context.Context, initialPeers []peer.Peer) {
	p.ctx = ctx
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

func (p *peerSampling) ReceiveView(view View, from peer.Peer) []byte {
	p.logger.Info("Received view from - " + from.UdpAddress + " aka " + from.ProcessIdentifier)
	p.peers[from.UdpAddress] = from
	p.receivedView <- passiveView{
		view:     view,
		fromPeer: from.UdpAddress,
	}
	if p.strategy.ViewPropagationStrategy == Push {
		return ViewToBytes(View{Nodes: sll.New()}, from) // no response needed
	} else {
		return ViewToBytes(p.view, from)
	}
}

// fillView fills the current view as these peers
func (p *peerSampling) fillView(peers ...peer.Peer) {
	for _, peer := range peers {
		p.view.Nodes.Add(NodeDescriptor{
			Address: peer.UdpAddress,
			Hop:     0,
		})
		p.peers[peer.ProcessIdentifier] = peer
	}
	p.view.sortNodes()
}

func (p *peerSampling) passive() {
	rand.Seed(time.Now().Unix())
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-time.After(p.wait + time.Duration(rand.Intn(int(p.wait.Milliseconds()/2)))): // add random delay
			p.logger.Info("[PT]:ViewPropagationStrategy is - [" +
				strconv.Itoa(p.strategy.ViewPropagationStrategy) + "]")
			receivedView := new(View)
			nwPeer := p.SelectPeer()
			if nwPeer.UdpAddress == "" {
				continue
			}
			p.logger.Debug("[PT]:Propagating view to - [" + nwPeer.UdpAddress + "]")
			if p.strategy.ViewPropagationStrategy == Push {
				mergedView := MergeView(p.view, View{Nodes: sll.New(p.selfDescriptor)})
				// send mergedView to nwPeer, receive Ok or view
				p.logger.Debug("[PT]:PUSH : sending current view merged with self to - " + nwPeer.UdpAddress)
				rspView, from, err := BytesToView(p.udp.ExchangeView(nwPeer.UdpAddress,
					ViewToBytes(mergedView, p.peers[p.selfIdentifier])))
				if err == nil {
					p.peers[from.ProcessIdentifier] = from
					_ = &rspView // for PUSH -> []byte("OKAY")
				}
			} else {
				p.logger.Debug("[PT]:PULL/PUSH-PULL :sending empty view to nwPeer - " + nwPeer.UdpAddress)
				// send emptyView to nwPeer to trigger response
				rspView, from, err := BytesToView(p.udp.ExchangeView(nwPeer.UdpAddress,
					ViewToBytes(View{Nodes: sll.New()}, p.peers[p.selfIdentifier])))
				if err == nil {
					p.peers[from.ProcessIdentifier] = from
					receivedView = &rspView
					p.logger.Debug("[PT]:PULL/PUSH-PULL : receivedView from nwPeer - " +
						nwPeer.UdpAddress)
				}
			}
			if p.strategy.ViewPropagationStrategy == Pull {
				if receivedView != nil {
					p.logger.Trace("[PT]:PULL : received a view")
					increaseHopCount(receivedView)
					mergedView := mergeViewExcludeNode(*receivedView, p.view, p.selfDescriptor)
					p.selectView(&mergedView)
					p.logger.Trace("[PT] PULL : selected subset of view merged with received view as current view")
				} else {
					p.logger.Warn("[PT]PULL : received nil from " + nwPeer.UdpAddress)
				}
			}
		}
	}
}

func (p *peerSampling) active() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case receivedView := <-p.receivedView: // wait to receive a view
			p.logger.Trace("[AT]:Active thread received view, incrementing hop count")
			increaseHopCount(&receivedView.view)
			if p.strategy.ViewPropagationStrategy == Pull {
				p.logger.Trace("[AT]:PULL - merging current view with self")
				mergedView := MergeView(p.view, selfDescriptor(p.selfDescriptor))
				// send mergedView to peer, receive Ok or view
				p.logger.Trace("[AT]:PULL : sending current view merged with self descriptor to peer - " + receivedView.fromPeer)
				view, from, err := BytesToView(p.udp.ExchangeView(receivedView.fromPeer, ViewToBytes(mergedView, p.peers[p.selfIdentifier])))
				if err == nil {
					p.peers[from.ProcessIdentifier] = from
					receivedView.view = view
				}
			}
			receivedView.view = mergeViewExcludeNode(receivedView.view, p.view, p.selfDescriptor)
			p.logger.Trace("[AT] selecting subset of view merged with received view as current view")
			p.selectView(&receivedView.view)
		}
	}
}

func (p *peerSampling) GetView() *View {
	return &p.view
}

func (p *peerSampling) selectView(view *View) {
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
func (p *peerSampling) SelectPeer() peer.Peer {
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
	p.peers[address] = peer.Peer{
		UdpAddress:        address,
		ProcessIdentifier: "",
	}
	return p.peers[address]
}

func selfDescriptor(n NodeDescriptor) View {
	return View{Nodes: sll.New(n)}
}
func Init(self string, identifier string, loggerOn bool, strategy PeerSamplingStrategy) Sampling {
	ps := peerSampling{
		logger:   mLogger.Get("ps-" + self),
		wait:     ViewExchangeDelay,
		strategy: strategy,
		view: View{
			Nodes: sll.New(),
		},
		selfDescriptor: NodeDescriptor{
			Address: self,
			Hop:     0,
		},
		receivedView:   make(chan passiveView),
		udp:            client.GetClient(self),
		selfIdentifier: identifier,
		peers:          make(map[string]peer.Peer),
	}
	if !loggerOn {
		ps.logger.SetLevel(hclog.Off)
	}
	ps.peers[ps.selfIdentifier] = peer.Peer{
		UdpAddress:        self,
		ProcessIdentifier: identifier,
	}
	return &ps
}
