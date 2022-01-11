package client

import (
	"net"

	"github.com/Ishan27g/go-utils/mLogger"

	"github.com/hashicorp/go-hclog"
)

const loggerOn = false

// Client is the UDP Client interface
type Client interface {
	// ExchangeView sends a view to this peer and returns the view of the peer
	ExchangeView(address string, data []byte) []byte
	// SendGossip sends the gossip message to FanOut number of peers, no return is expected
	SendGossip(address string, data []byte) []byte
}

func GetClient(logName string) Client {
	u := &udpClient{
		processName: logName,
		logger:      mLogger.Get(logName + "-udp-client"),
	}
	if !loggerOn {
		u.logger.SetLevel(hclog.Info)
	}
	return u
}

type udpClient struct {
	processName string
	logger      hclog.Logger
}

// SendGossip sends the gossip message to FanOut number of peers
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
	defer c.Close()
	u.logger.Trace("Sending gossip to - " + c.RemoteAddr().String())
	//println(u.processName, " Sending gossip to - "+c.RemoteAddr().String())

	_, err = c.Write(data)
	if err != nil {
		return nil
	}

	buffer := make([]byte, 1024)
	readLen, _, err := c.ReadFromUDP(buffer)
	if err != nil {
		u.logger.Trace(err.Error() + "\n for " + c.RemoteAddr().String())
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
	defer c.Close()
	u.logger.Trace("Sending view UDP to - " + c.RemoteAddr().String())
	//println(u.processName, " Sending view to - "+c.RemoteAddr().String())

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
