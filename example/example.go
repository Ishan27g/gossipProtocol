package main

import (
	"fmt"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
)

func main() {
	mLogger.New("ok", "trace")
	g := gossip.Default("http://localhost", "1000", 1) // zone 1

	packetsReceivedFromPeers := make(chan map[string]string)
	packetsSentByUser := make(chan gossip.Packet)
	g.Join([]string{"localhost:1001", "localhost:2102", "localhost:4103", "localhost:2004"}, packetsReceivedFromPeers, packetsSentByUser) // across zones
	// g.StartRumour("")

	for {
		select {
		case from, id := <-packetsReceivedFromPeers:
			fmt.Println("received gossip ", id, " from  ", from)
		case packet := <-packetsSentByUser:
			fmt.Println("Sent gossip to network - ", packet.GossipMessage.GossipMessageHash, " data - ", packet.GossipMessage.Data)
		}
	}
}
