package mocks

//
//import (
//	"context"
//	"fmt"
//	"sync"
//	"testing"
//	"time"
//
//	"github.com/Ishan27gOrg/gossipProtocol/gossip"
//	"github.com/Ishan27gOrg/gossipProtocol/gossip/peer"
//	"github.com/stretchr/testify/assert"
//)
//
//var hostname = "localhost"
//var network = []string{"1001", "1002", "1003", "1004"}
//var peers = []peer.Peer{
//	{network[0], "localhost:1001"},
//	{network[1], "localhost:1002"},
//	//{network[2], "p3"},
//	//{network[3], "p4"},
//}
//
//func remove(this peer.Peer, from []peer.Peer) []peer.Peer {
//	var peers []peer.Peer
//	for _, p := range from {
//		if p.UdpAddress != this.UdpAddress {
//			p.UdpAddress = ":" + p.UdpAddress
//			peers = append(peers, p)
//		}
//	}
//	return peers
//}
//func TestGossip_StartRumour(t *testing.T) {
//	var mGs []MockGossip
//
//	ctx, cancel := context.WithCancel(context.Background())
//
//	for index, p := range peers {
//		mg := SetupGossipMock(p)
//		go func(mGs MockGossip) {
//			<-ctx.Done()
//			newGossipEvent := make(chan gossip.Packet)
//			mGs.G.JoinWithoutSampling(func() []peer.Peer {
//				fmt.Println("index - ", index)
//				var nwp []peer.Peer
//				nwPeers := remove(p, peers)
//				for i := 0; i < len(nwPeers); i++ {
//					if i != index {
//						nwp = append(nwp, nwPeers[i])
//					}
//				}
//				fmt.Println("peers - ", nwp)
//				return nwp
//			}, newGossipEvent)
//			//mGs.G.JoinWithSampling(remove(p, peers), newGossipEvent)
//			mGs.NewGossipEvent = newGossipEvent
//		}(mg)
//		mGs = append(mGs, mg)
//	}
//	<-time.After(3000)
//	cancel()
//	var wg sync.WaitGroup
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		<-time.After(5 * time.Second)
//		mGs[0].G.StartRumour("Test")
//		<-mGs[0].G.ReceiveGossip()
//	}()
//	for i, mg := range mGs {
//		if i == 0 {
//			continue
//		}
//		wg.Add(1)
//		go func(mg MockGossip) {
//			defer wg.Done()
//			fmt.Println("Waiting....")
//			g := <-mg.G.ReceiveGossip()
//			fmt.Println("rec-", g)
//			assert.Equal(t, "Test", g.GossipMessage.Data)
//		}(mg)
//	}
//
//	wg.Wait()
//}
