package gossip

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/stretchr/testify/assert"
)

var network = func() []string {
	return []string{"1001", "1002", "1003", "1004"}
}

type mockGossip struct {
	g              Gossip
	newGossipEvent chan Packet
}

func remove(this string, from []string) []string {
	index := 0
	for i, s := range from {
		if s == this {
			index = i
			break
		}
	}
	return append(from[:index], from[index+1:]...)
}
func TestRemove(t *testing.T) {
	this := "this"
	from := []string{"some", "slice", "with", "this", "in", "it"}
	removed := remove(this, from)
	assert.Equal(t, len(from)-1, len(removed))
	for _, s := range removed {
		assert.NotEqual(t, "this", s)
	}
}
func mockGossipDefaultConfig(hostname, port string) mockGossip {
	m := mockGossip{
		g:              DefaultConfig(hostname, port),
		newGossipEvent: make(chan Packet),
	}
	self := port
	var nw []string
	for _, s := range remove(self, network()) {
		nw = append(nw, ":"+s)
	}
	m.g.JoinWithoutSampling(nw, m.newGossipEvent) // across zones
	// g.StartRumour("")

	return m
}
func TestGossip(t *testing.T) {
	mLogger.New("", "info")
	var mg []mockGossip
	events := make(chan Packet, 4)
	var wg sync.WaitGroup
	for i := 0; i < len(network()); i++ {
		m := mockGossipDefaultConfig("", network()[i])
		mg = append(mg, m)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			packet := <-m.newGossipEvent
			fmt.Printf("\n["+network()[i]+"]received gossip %v\n", packet)
			events <- packet
		}(i)

	}
	go func() {
		<-time.After(1 * time.Second)
		mg[1].g.StartRumour("hello")
	}()
	wg.Wait()
	close(events)
	assert.Equal(t, 4, len(events))
	for event := range events {
		assert.Equal(t, "hello", event.GossipMessage.Data)
	}
}
