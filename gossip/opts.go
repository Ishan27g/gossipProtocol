package gossip

type GossipI interface {
	New() Gossip
}
type gossipI struct {
	g Gossip
}

func (g *gossipI) New() Gossip {
	return g.g
}

func newGossipI(hostname, port, selfAddress string, c *config) GossipI {
	return &gossipI{g: withConfig(hostname, port, selfAddress, c)}
}

type Option func(gI *GossipI)
type Options []Option

func Apply(options Options) GossipI {
	g := newGossipI("", "", "", nil)
	for _, option := range options {
		option(&g)
	}
	return g
}
func Env(hostname, port, selfAddress string) Option {
	return func(gI *GossipI) {
		*gI = newGossipI(hostname, port, selfAddress, defaultConfig())
	}
}
func Logger(on bool) Option {
	loggerOn = on
	return func(gI *GossipI) {}
}
