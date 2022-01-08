package gossipProtocol

import (
	"crypto/sha1"
	"fmt"
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
var defaultStrategy = sampling.DefaultStrategy()
var defaultHashMethod = func(obj interface{}) string {
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%v", obj)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// default env
const gossipDelay = 250 * time.Millisecond
const fanOut = 2
const minimumPeersInNetwork = 10

type envConfig struct {
	Hostname              string `env:"HOST_NAME"`
	UdpPort               string `env:"UDP_PORT,required"`
	ProcessIdentifier     string
	RoundDelay            time.Duration // timeout between each round for a gossipMessage
	FanOut                int           // num of peers to gossip a message to
	MinimumPeersInNetwork int           // number of rounds a message is gossiped = log(minPeers/FanOut)
}

type gossipI struct {
	g Gossip
}

// Apply a set of options to the gossip listener
func Apply(options Options) Listener {
	g := newGossipI("", "", "")
	for _, option := range options {
		option(&g)
	}
	return g
}
func Env(hostname, port, selfAddress string) Option {
	return func(gI *Listener) {
		*gI = newGossipI(hostname, port, selfAddress)
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

func (g *gossipI) New() Gossip {
	return g.g
}

func newGossipI(hostname, port, selfAddress string) Listener {
	return &gossipI{g: withConfig(hostname, port, selfAddress)}
}
