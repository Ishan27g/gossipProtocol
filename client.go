package gossipProtocol

import (
	"net"
	"strings"

	"github.com/Ishan27g/go-utils/mLogger"

	"github.com/hashicorp/go-hclog"
)

// Client is the UDP Client interface
type client interface {
	send(address string, data []byte) []byte
}

func getClient(logName string) client {
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

const hostname = "http://localhost"

func trimHost(address string) string {
	s := strings.Trim(address, hostname)
	return ":" + s
}
func (u *udpClient) send(address string, data []byte) []byte {
	s, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		address = trimHost(address)
		s, err = net.ResolveUDPAddr("udp4", address)
		if err != nil {
			return nil
		}
	}
	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		return nil
	}
	defer c.Close()
	_, err = c.Write(data)
	if err != nil {
		return nil
	}
	buffer := make([]byte, 4096)
	readLen, _, err := c.ReadFromUDP(buffer)
	if err != nil {
		u.logger.Trace(err.Error() + "\n for " + c.RemoteAddr().String())
		return nil
	}
	buffer = buffer[:readLen]
	u.logger.Trace("Received gossip response from - " + c.RemoteAddr().String())
	return buffer
}
