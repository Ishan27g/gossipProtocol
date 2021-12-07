package client

import (
	"net"

	"github.com/Ishan27g/go-utils/mLogger"

	"github.com/hashicorp/go-hclog"
)

// Client is the UDP Client interface
type Client interface {
	// ExchangeView sends a view to this peer and returns the view of the peer
	ExchangeView(address string, data []byte) []byte
	// SendGossip sends the gossip message to FanOut number of peers
	// no return is expected (todo)
	SendGossip(address string, data []byte) []byte
}

func GetClient(todo string) Client {
	return &udpClient{
		logger: mLogger.Get(todo + "-udp-client"),
	}
}

type udpClient struct {
	logger hclog.Logger
}

// SendGossip sends the gossip message to FanOut number of peers
// no return is expected (todo)
func (u *udpClient) SendGossip(address string, data []byte) []byte {
	s, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		u.logger.Error(err.Error())
		return nil
	}
	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		u.logger.Error(err.Error())
		return nil
	}

	u.logger.Trace("Sending gossip to - " + c.RemoteAddr().String())
	defer c.Close()

	_, err = c.Write(data)

	if err != nil {
		u.logger.Error("this - " + err.Error())
		return nil
	}

	buffer := make([]byte, 1024)
	readLen, _, err := c.ReadFromUDP(buffer)
	if err != nil {
		u.logger.Error("This?? " + err.Error())
		return nil
	}
	buffer = buffer[:readLen]

	u.logger.Trace("Received gossip response from - " + c.RemoteAddr().String())
	return buffer
}

// ExchangeView sends a view to this peer and returns the view of the peer
func (u *udpClient) ExchangeView(address string, data []byte) []byte {
	s, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		u.logger.Error(err.Error())
		return nil
	}
	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		u.logger.Error(err.Error())
		return nil
	}

	u.logger.Trace("Sending view UDP to - " + c.RemoteAddr().String())
	defer c.Close()
	_, err = c.Write(data)

	if err != nil {
		u.logger.Error(err.Error())
		return nil
	}

	buffer := make([]byte, 1024)
	readLen, _, err := c.ReadFromUDP(buffer)
	if err != nil {
		u.logger.Error(err.Error())
		return nil
	}
	buffer = buffer[:readLen]
	u.logger.Trace("Received view UDP from - " + c.RemoteAddr().String())
	return buffer
}
