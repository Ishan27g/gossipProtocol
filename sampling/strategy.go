package sampling

const (
	Random   = iota // *selection strategies
	Head            // *selection strategies
	Tail            // *selection strategies
	Push            // *propagation strategies
	Pull            // *propagation strategies
	PushPull        // *propagation strategies
)

type PeerSamplingStrategy struct {
	PeerSelectionStrategy   int
	ViewPropagationStrategy int
	ViewSelectionStrategy   int
}

func With(ps, vp, vs int) PeerSamplingStrategy {
	if ps > 2 || vs > 2 {
		ps = Random
		vs = Random
	}
	if vp < 3 {
		vp = Push // todo PushPull
	}
	return PeerSamplingStrategy{
		PeerSelectionStrategy:   ps,
		ViewPropagationStrategy: vp,
		ViewSelectionStrategy:   vs,
	}
}
func DefaultStrategy() PeerSamplingStrategy {
	return PeerSamplingStrategy{
		PeerSelectionStrategy:   Random,
		ViewPropagationStrategy: Push,
		ViewSelectionStrategy:   Random,
	}
}
