package gossipProtocol

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/stretchr/testify/assert"
)

type gArgs struct {
	self         Peer
	initialPeers []Peer
	gossip       Gossip
	rcvGossip    <-chan Packet
}

var Reset = "\033[0m"
var Purple = "\033[35m"

func (g *gArgs) printView() {
	tr := table.NewWriter()
	tr.SetOutputMirror(os.Stdout)
	tr.SetStyle(table.StyleBold)
	tr.Style().Options.DrawBorder = false
	tr.AppendHeader(table.Row{"View at " + g.self.ProcessIdentifier, "Peer hop"})

	var logs []table.Row
	for _, peer := range g.gossip.CurrentView() {
		logs = append(logs, table.Row{peer.ProcessIdentifier, peer.Hop})
	}
	tr.AppendRows(logs)
	tr.AppendSeparator()
	tr.Render()

	fmt.Println()
}
func setupGossipProcesses(base string, numProcesses int) []gArgs {
	var processes = make(chan gArgs, numProcesses)
	var wg sync.WaitGroup

	for i := 0; i < numProcesses; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, processes chan gArgs, i int) {
			defer wg.Done()
			self := network(base, -1, numProcesses)[i] // all peers, this index
			peers := network(base, i, numProcesses)    // all peers except this index
			gossip, rcvGossip := Config("localhost", self.UdpAddress, self.ProcessIdentifier)
			gossip.Join(peers...)
			processes <- gArgs{
				self:         self,
				initialPeers: peers,
				gossip:       gossip,
				rcvGossip:    rcvGossip,
			}
		}(&wg, processes, i)
	}
	wg.Wait()
	close(processes)
	var pro []gArgs
	for p := range processes {
		pro = append(pro, p)
	}
	return pro
}
func matchGossip(wg *sync.WaitGroup, r <-chan Packet, data string) bool {
	defer wg.Done()
	g := <-r
	return g.GetData() == data
}
func Test_Gossip(t *testing.T) {
	t.Parallel()

	var data = "some data"
	var numProcesses = 25

	processes := setupGossipProcesses("90", numProcesses)

	t.Run("Gossip protocol", func(t *testing.T) {
		var wg sync.WaitGroup

		processes[0].gossip.SendGossip(data)
		<-time.After(1 * time.Second)
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
	})
	t.Cleanup(func() {
		<-time.After(ViewExchangeDelay * 2)
		for _, p := range processes {
			p.printView()
		}
		<-time.After(ViewExchangeDelay)
		for _, p := range processes {
			p.printView()
		}
	})
}
