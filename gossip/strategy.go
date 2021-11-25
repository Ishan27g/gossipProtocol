package gossip

const (
	Random = iota
	Head
	Tail
	Push
	Pull
	PushPull
)

func DefaultStrategy() PeerSamplingStrategy {
	return PeerSamplingStrategy{
		PeerSelectionStrategy:   Random,
		ViewPropagationStrategy: Push,
		ViewSelectionStrategy:   Random,
	}
}

type PeerSamplingStrategy struct {
	PeerSelectionStrategy   int
	ViewPropagationStrategy int
	ViewSelectionStrategy   int
}
