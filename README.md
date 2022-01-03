
# Gossip protocol to replicate - eventData

- a peer receives a command/log to gossip
- selects N peers and transmits this message for M rounds (https://flopezluis.github.io/gossip-simulator/) 
- various strategies for maintaining a partial view of network

- each peer maintains vector clocks for their partial view

# With raft

- gossip `eventData` to peers ( saved temporarily )
- send `eventId` to leader, with current vector clock. Leader determines happened before relation
- `eventData` on each peer is then ordered/sorted based on `global-event-order` received from leader
- peer can then apply the data in this order to local snapshot ( saved permanently )

Leader - 
1. Follower can request peers - sends partial view of network to followers. Followers gossip to these peers 
2. Leader can receive `new-event` from followers. Commit when received from majority of followers
3. Sends final order of gossip-events to all followers (send how? heartbeat? follower requests)

Follower -
1. Receive subset of network from leader
2. Inform leader about new gossip-event
3. Receive order of gossip-events and then apply events in order to snapshot

spread rumour - 

-> Peer gossips `eventData` to N peers and Peer informs `eventId` to leader
   - each peer can that receives a new gossip message additionally sends a `new-event` to the leader with its local vector clock
   - leader maintains a `global-vector-clock` which is updated everytime it receives `new-event` from a follower

-> Leader receives a `new-event`,
   - leader saves the event and corresponding vector clock. Each peer sends a subset of the entire vector clock.
   - when leader receives majority for a `new-event`, 
     - merge the vector clock with received vector clocks
     - sort `global-vector-clock` to get event order
     - send this `global-event-order` to peers 
   - followers can make sure that eventually their local snapshot is consistent with received snapshot

-> save to local snapshot
      - Snapshot : time ordered logs / DB commands / anything to replicate
      - failed peers/new peers will join raft-nw, become followers, receive the latest snapshot from leader.
      - If snapshot saves DB-commands, new peer can execute commands in order and become consistent 

### Gossip Listener
```go
package main

// Gossip is the interface to send and receive gossip messages.
// Created from a Listener with options 
//      minOptions := gossipProtocol.Options{
//              gossipProtocol.Env("localhost", "1001", "p1")
//             }
//      listener := gossipProtocol.Apply(minOptions)
//      gossiper := listener.New()

type Gossip interface {
    // JoinWithSampling starts the gossip protocol with these initial peers. Peer sampling is done to periodically
    // maintain a partial view (subset) of the gossip network. Data is sent of the channel when gossip
    // is received from a peer or from the user (StartRumour)
    JoinWithSampling(peers []peer.Peer, newGossip chan Packet)
    // JoinWithoutSampling starts the gossip protocol with these initial peers. Peers are iteratively selected.
    // Data is sent of the channel when gossip is received from a peer or from the user (StartRumour)
    JoinWithoutSampling(peers func() []peer.Peer, newGossip chan Packet)
    // StartRumour is the equivalent of receiving a gossip message from the user. This is sent to peers
    StartRumour(data string)
	// ReceiveGossip gossip from the network
    ReceiveGossip() chan Packet
    // RemovePacket will return the packet and its latest event clock after removing it from memory
    // Should be called a maximum of one time per packet
    RemovePacket(id string) (*Packet, vClock.EventClock)
	// Stop the listeners
    Stop()
}

```

### Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/Ishan27gOrg/gossipProtocol"
	"github.com/Ishan27gOrg/gossipProtocol/peer"
	"github.com/Ishan27gOrg/gossipProtocol/sampling"
)



func main() {
	// set options
	options := gossipProtocol.Options{
		gossipProtocol.Env("localhost", "1001", "p1"),
		gossipProtocol.Logger(false),
		gossipProtocol.Strategy(sampling.Random, sampling.Push, sampling.Random),
	}
	// init listener
	g := gossipProtocol.Apply(options).New()

	newGossipEvent := make(chan gossipProtocol.Packet)

	var initialPeers = []peer.Peer{ // other peers to gossip with
		{UdpAddress: "localhost:1002", ProcessIdentifier: "p2"},
		{UdpAddress: "localhost:1003", ProcessIdentifier: "p3"},
		{UdpAddress: "localhost:1004", ProcessIdentifier: "p4"},
		{UdpAddress: "localhost:1005", ProcessIdentifier: "p5"},
		{UdpAddress: "localhost:1006", ProcessIdentifier: "p6"},
	}
	// join either with peer sampling and view exchange
	g.JoinWithSampling(initialPeers, newGossipEvent)
	
	// or join with static peers and no view exchange
	// g.JoinWithoutSampling(func() []peer.Peer {
	//	return []peer.Peer{
	//		{UdpAddress: "localhost:1002", ProcessIdentifier: "p2"},
	//		{UdpAddress: "localhost:1003", ProcessIdentifier: "p3"},
	//		{UdpAddress: "localhost:1004", ProcessIdentifier: "p4"},
	//		{UdpAddress: "localhost:1005", ProcessIdentifier: "p5"},
	//		{UdpAddress: "localhost:1006", ProcessIdentifier: "p6"},
	//	}
	// }, newGossipEvent)

	g.StartRumour("gossip this")

    gossipPacket := <-g.ReceiveGossip()
    fmt.Println("Data - ", gossipPacket.GossipMessage.Data)
    fmt.Println("event clock for this packet - ", gossipPacket.VectorClock)
    <-time.After(5 * time.Second)
    samePacket, latestClock := g.RemovePacket(gossipPacket.GetId())
    fmt.Println("Data - ", samePacket.GossipMessage.Data)
    fmt.Println("latest event clock for this packet - ", latestClock)
}
```