package gossip

type Listener interface {
	New() Gossip
}
type Option func(gI *Listener)
type Options []Option

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

type gossipI struct {
	g Gossip
}

func (g *gossipI) New() Gossip {
	return g.g
}

func newGossipI(hostname, port, selfAddress string, c *config) Listener {
	return &gossipI{g: withConfig(hostname, port, selfAddress, c)}
}
