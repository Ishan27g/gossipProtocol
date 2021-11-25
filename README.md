
Gossip protocol to replicate - something

- a peer receives a command/log to gossip
- selects N peers and transmits this message for M rounds (https://flopezluis.github.io/gossip-simulator/) 
- various strategies for maintaining a partial view of network

With raft

1. Use raft to receive subset of network from leader. (tempUtils.go) 
2. When gossiping a message 
-> inform leader
   - each peer can that receives a new gossip message additionally sends a `new-command` to the leader
   - when leader receives majority for a `new-command`, it updates local snapshot (propagated every time with heartbeat, followers can make sure that eventually their local snapshot is consistent with received snapshot)
-> save to local snapshot
   - Snapshot : time ordered logs / DB commands / anything to replicate
   - failed peers/new peers will join raft-nw, become followers, receive the latest snapshot from leader.
   - If snapshot saves DB-commands, new peer can execute commands in order and become consistent 

```shell
cd sandbox/
make RunVendorClean
```

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