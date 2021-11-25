```go
package main

import (
	"os"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
)

func exampleDefaultStrategy(){
	if os.Getenv("HOST_NAME") == "" || os.Getenv("UDP_PORT") == "" {
		os.Exit(1)
	}

	// Initialise with default strategy
	g := gossip.Config()

	// send a message to network
	// g.StartRumour("message1")
	// send another message to network
	// g.StartRumour("message2")

	<- make(chan bool)
}

func exampleCustomStrategy(){
	if os.Getenv("HOST_NAME") == "" || os.Getenv("UDP_PORT") == "" {
		os.Exit(1)
	}

	// set a global level to print gossip protocol logs
	mLogger.New("gossiper", "debug")

	// select a peer sampling strategy
	peerSamplingStrategy := gossip.PeerSamplingStrategy{
		PeerSelectionStrategy:   gossip.Random, // Random, Head, Tail
		ViewPropagationStrategy: gossip.Push,	// Push, Pull, PushPull
		ViewSelectionStrategy:   gossip.Random, // Random, Head, Tail
	}

	// Initialise with provided strategy
	g := gossip.ConfigWithStrategy(&peerSamplingStrategy)

	// send a message to network
	// g.StartRumour("message1")
	// send another message to network
	// g.StartRumour("message2")

	<- make(chan bool)
}
```