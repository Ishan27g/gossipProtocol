package gossip

import (
	"net"
	"strconv"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/hashicorp/go-hclog"
)


type Server struct {
	logger hclog.Logger
	address string
	gossipCb func (Packet) []byte
	viewCb func (View, string) []byte
}
// Listen starts the udp server that listens for an incoming view or gossip from peers.
// Responds with the current view / gossip as per strategy.
func Listen(port string, gossipCb func(Packet) []byte, viewCb func(View, string) []byte ){
	server := Server{
		logger:      mLogger.Get("udp-server"),
		address:     "",
		gossipCb: gossipCb,
		viewCb: viewCb,
	}
	s, err := net.ResolveUDPAddr("udp4", ":"+port)
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

	defer connection.Close()
	buffer := make([]byte, 1024)

	for {
		readLen, addr, err := connection.ReadFromUDP(buffer)
		buffer = buffer[:readLen]
		view, err := BytesToView(buffer)
		if err == nil {
			server.logger.Debug("Server received view " + " from: " + addr.IP.String() + ":" + strconv.Itoa(addr.Port))
			rsp := server.viewCb(view, addr.IP.String())
			_, err = connection.WriteToUDP(rsp, addr)
			if err != nil {
				server.logger.Error(err.Error())
				// return
			} else {
				server.logger.Debug("Server sent view response to: " + addr.String())
			}
		}else {
			server.logger.Debug("Server received gossip " + " from: " + addr.IP.String() + ":" + strconv.Itoa(addr.Port))
			rsp := server.gossipCb(ByteToPacket(buffer))
			_, err = connection.WriteToUDP(rsp, addr)
			if err != nil {
				server.logger.Error(err.Error())
				// return
			} else {
				server.logger.Debug("Server sent gossip response to: " + addr.String())
			}
		}
	}
}

