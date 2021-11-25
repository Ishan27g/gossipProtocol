package main

import (
	"os"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
)

func exampleCustomStrategy() {
	if os.Getenv("HOST_NAME") == "" || os.Getenv("UDP_PORT") == "" {
		os.Exit(1)
	}

	processName := os.Getenv("ProcessName")
	// set a global level to print gossip protocol logs
	mLogger.New(processName, "debug")

	// define a peer sampling strategy
	strategy := gossip.PeerSamplingStrategy{
		PeerSelectionStrategy:   gossip.Random, // Random, Head, Tail
		ViewPropagationStrategy: gossip.Push,   // Push, Pull, PushPull
		ViewSelectionStrategy:   gossip.Random, // Random, Head, Tail
	}

	// Initialise with provided strategy
	_ = gossip.ConfigWithStrategy(&strategy)

	// send a message to network
	// g.StartRumour("message1")
	// send another message to network
	// g.StartRumour("message2")

	<-make(chan bool)
}

func main() {
	exampleCustomStrategy()
}
