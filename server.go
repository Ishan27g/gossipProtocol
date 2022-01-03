package gossipProtocol

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/peer"
	"github.com/Ishan27gOrg/gossipProtocol/sampling"
	"github.com/hashicorp/go-hclog"
)

type udpServer struct {
	logger   hclog.Logger
	address  string
	gossipCb func(Packet, string) []byte
	viewCb   func(sampling.View, peer.Peer) []byte
}

// Listen starts the udp server that listens for an incoming view or gossip from peers.
// Responds with the current view / gossip as per strategy.
func Listen(ctx context.Context, port string, gossipCb func(Packet, string) []byte, viewCb func(sampling.View, peer.Peer) []byte) {
	server := udpServer{
		logger:   nil,
		address:  "",
		gossipCb: gossipCb,
		viewCb:   viewCb,
	}
	server.logger = mLogger.Get(port + "-udp-server")
	if !loggerOn {
		server.logger.SetLevel(hclog.Error)
	}
	s, err := net.ResolveUDPAddr("udp4", port)
	if err != nil {
		server.logger.Error(err.Error())
		return
	}
	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		server.logger.Error(err.Error())
		return
	}
	server.logger.Info("UDP server listening on " + connection.LocalAddr().String())
	server.address = connection.LocalAddr().String()

	go func() {
		<-ctx.Done()
		connection.Close()
		fmt.Println("\n\n\nCLOSED")
	}()

	for {
		buffer := make([]byte, 1024)
		readLen, addr, e := connection.ReadFromUDP(buffer)
		if e != nil {
			return
		}
		buffer = buffer[:readLen]
		view, from, err := sampling.BytesToView(buffer)
		if from.UdpAddress != "" {
			server.logger.Trace("udpServer received view " + " from: " + from.UdpAddress)
			rsp := server.viewCb(view, from)
			_, err = connection.WriteToUDP(rsp, addr)
			if err != nil {
				server.logger.Error(err.Error())
				// return
			} else {
				server.logger.Trace("udpServer sent view response to: " + from.UdpAddress)
			}
		} else {
			server.logger.Trace("udpServer received gossip " + " from: " + addr.IP.String() + ":" + strconv.Itoa(addr.Port))
			rsp := server.gossipCb(ByteToPacket(buffer), addr.IP.String()+":"+strconv.Itoa(addr.Port))
			server.logger.Trace("udpServer sending gossip response to: " + addr.String())
			_, err = connection.WriteToUDP(rsp, addr)
			if err != nil {
				server.logger.Error(err.Error())
				// return
			} else {
				server.logger.Trace("udpServer sent gossip response to: " + addr.String())
			}
		}
	}
}
