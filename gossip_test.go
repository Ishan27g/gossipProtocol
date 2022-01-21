package gossipProtocol

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type gArgs struct {
	self         Peer
	initialPeers []Peer
	gossip       Gossip
	rcvGossip    <-chan Packet
}

func setupGossipProcesses(base string, numProcesses int) []gArgs {
	var processes []gArgs
	for i := 0; i < numProcesses; i++ {
		self := network(base, -1, numProcesses)[i] // all peers, this index
		peers := network(base, i, numProcesses)    // all peers except this index
		time.Sleep(100 * time.Millisecond)
		gossip, rcvGossip := Config("localhost", self.UdpAddress, self.ProcessIdentifier)
		gossip.Join(peers...)
		processes = append(processes, gArgs{
			self:         self,
			initialPeers: peers,
			gossip:       gossip,
			rcvGossip:    rcvGossip,
		})
	}
	return processes
}
func matchGossip(wg *sync.WaitGroup, r <-chan Packet, data string) bool {
	defer wg.Done()
	g := <-r
	return g.GetData() == data
}
func Test_Gossip(t *testing.T) {
	t.Parallel()
	var data = "some data"
	var numProcesses = 10
	processes := setupGossipProcesses("90", numProcesses)
	processes[0].gossip.SendGossip(data)

	<-time.After(1 * time.Second)
	var wg sync.WaitGroup
	for _, p := range processes {
		wg.Add(1)
		go func(p gArgs) {
			assert.True(t, matchGossip(&wg, p.rcvGossip, data))
		}(p)
	}

	processes[numProcesses-1].gossip.SendGossip(data + data)
	<-time.After(1 * time.Second)

	for _, p := range processes {
		wg.Add(1)
		go func(p gArgs) {
			assert.True(t, matchGossip(&wg, p.rcvGossip, data+data))
		}(p)
	}
	processes[4].gossip.SendGossip(data + data + data)
	<-time.After(1 * time.Second)

	for _, p := range processes {
		wg.Add(1)
		go func(p gArgs) {
			assert.True(t, matchGossip(&wg, p.rcvGossip, data+data+data))
		}(p)
	}
	wg.Wait()

}
