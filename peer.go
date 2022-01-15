package gossipProtocol

type Peer struct {
	UdpAddress        string
	ProcessIdentifier string
	Hop               int
}

type passiveView struct {
	view View
	from Peer
}
