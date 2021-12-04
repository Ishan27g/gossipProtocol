package main

import (
	"fmt"

	"github.com/Ishan27gOrg/gossipProtocol/gossip"
)

func exampleCustomStrategy(hostname, udp string, zone int) {
	g := gossip.Default(hostname, udp, zone) // zone 1

	packetsReceivedFromPeers := make(chan map[string]string)
	packetsSentByUser := make(chan gossip.Packet)
	g.Join(network, packetsReceivedFromPeers, packetsSentByUser) // across zones
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

var hostname = "http://localhost"
var network = []string{"1001", "2102", "4103", "2003", "1000"}

func main() {
	for i := len(network) - 1; i >= 1; i-- {
		go exampleCustomStrategy(hostname, network[i], i)
	}
	exampleCustomStrategy(hostname, network[0], 1)
}
