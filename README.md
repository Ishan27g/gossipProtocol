
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
package gossipProtocol

type Gossip interface {
	// Join with some initial peers
	Join(...Peer)
	// Add peers
	Add(...Peer)
	// CurrentView returned as a String
	CurrentView() string
	// SendGossip  to the network
	SendGossip(data string)
}

```

### Example

```go
package main

import (
	"fmt"

	"github.com/Ishan27gOrg/gossipProtocol"
)



func main() {
	g, receive := gossipProtocol.Config("localhost", "8001", "p1")
	var peers []gossipProtocol.Peer
	peers = append(peers, gossipProtocol.Peer{UdpAddress: "localhost:8002", ProcessIdentifier: "p2"})
	peers = append(peers, gossipProtocol.Peer{UdpAddress: "localhost:8003", ProcessIdentifier: "p3"})
	g.Join(peers...)
	
	g.SendGossip("some data")
	for{
		gossipPacket := <-receive
		fmt.Println("data - ", gossipPacket.GossipMessage.Data)
		fmt.Println("event clock for this packet - ", gossipPacket.VectorClock)
	}
}
```