package gossipProtocol

import (
	"time"

	"github.com/Ishan27gOrg/gossipProtocol/sampling"
)

type Listener interface {
	New() Gossip
}
type Option func(gI *Listener)
type Options []Option

// default vars, overwritten via options
var loggerOn bool
var defaultHashMethod = hash
var defaultStrategy = sampling.DefaultStrategy()

// default env
const gossipDelay = 50 * time.Millisecond
const fanOut = 2
const minimumPeersInNetwork = 10

// Apply a set of options to the gossip listener
func Apply(options Options) Listener {
	g := newGossipI("", "", "", nil)
	for _, option := range options {
		option(&g)
	}
	return g
}
func Env(hostname, port, selfAddress string) Option {
	return func(gI *Listener) {
		*gI = newGossipI(hostname, port, selfAddress, defaultConfig())
	}
}
func Logger(on bool) Option {
	loggerOn = on
	return func(gI *Listener) {}
}
func Hash(fn func(in interface{}) string) Option {
	defaultHashMethod = fn
	return func(gI *Listener) {}
}
func Strategy(peerSelection, viewPropagation, viewSelection int) Option {
	defaultStrategy = sampling.With(peerSelection, viewPropagation, viewSelection)
	return func(gI *Listener) {}
}

type gossipI struct {
	g Gossip
}

func (g *gossipI) New() Gossip {
	return g.g
}

func newGossipI(hostname, port, selfAddress string, c *Config) Listener {
	return &gossipI{g: withConfig(hostname, port, selfAddress, c)}
}
