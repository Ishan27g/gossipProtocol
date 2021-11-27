package gossip

import (
	"net"
	"os"
	"strconv"

	"github.com/Ishan27g/go-utils/mLogger"
	reg "github.com/Ishan27gOrg/registry/package"
	"github.com/Ishan27gOrg/vClock"
	"github.com/caarlos0/env/v6"
	"github.com/hashicorp/go-hclog"
)

type EnvCfg struct {
	envCfg     envConfig
	logger     hclog.Logger
	ps         PeerSampling
	gossipChan chan map[string]gossipMessage // fromPeer : gossipMessage
}

type envConfig struct {
	Registry string `env:"Registry"`
	ZoneId   int    `env:"ZoneId"`
	Hostname string `env:"HOST_NAME"`
	UdpPort  string `env:"UDP_PORT,required"`
}

var Env EnvCfg

// Config gossip protocol with default PeerSamplingStrategy
func Config() Gossip {
	defSt := DefaultStrategy()
	return ConfigWithStrategy(&defSt)
}

// ConfigWithStrategy gossip protocol with the provided PeerSamplingStrategy.
func ConfigWithStrategy(st *PeerSamplingStrategy) Gossip {
	gossipChan := make(chan map[string]gossipMessage)
	envCfg := envConfig{}
	if err := env.Parse(&envCfg); err != nil {
		mLogger.Get("env").Error(err.Error())
		os.Exit(1)
	}
	Env = EnvCfg{
		envCfg:     envCfg,
		logger:     mLogger.Get("udp"),
		ps:         nil,
		gossipChan: gossipChan,
	}

	go udpServer(envCfg.UdpPort)

	// no use once register, as peers will exchange view with each other
	registryClient := reg.RegistryClient(Env.envCfg.Registry)

	self := "http://" + Env.envCfg.Hostname + Env.envCfg.UdpPort
	zonePeers := registryClient.Register(Env.envCfg.ZoneId, self, nil)

	Env.ps = initPeerSampling(*st, zonePeers)

	vClock := vClock.Init(self, zonePeers.GetPeerAddr(self))
	return ListenForGossip(Env.gossipChan, vClock)

}

// udpSendView sends a view to this peer and returns the view of the peer
func udpSendView(address string, view view) *view {
	s, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		Env.logger.Error(err.Error())
		return nil
	}
	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		Env.logger.Error(err.Error())
		return nil
	}

	Env.logger.Debug("Sending view UDP to - " + c.RemoteAddr().String())
	defer c.Close()
	_, err = c.Write(viewToBytes(view))

	if err != nil {
		Env.logger.Error(err.Error())
		return nil
	}

	buffer := make([]byte, 1024)
	readLen, _, err := c.ReadFromUDP(buffer)
	if err != nil {
		Env.logger.Error(err.Error())
		return nil
	}
	buffer = buffer[:readLen]

	v, e := bytesToView(buffer)
	if e == nil {
		Env.logger.Debug("Received view UDP from - " + c.RemoteAddr().String())
		Env.logger.Debug(printView(v))
		return &v
	} else {
		Env.logger.Debug("Received err UDP from - " + c.RemoteAddr().String())
		return nil
	}
}

// udpSendGossip sends a gossip message to this peer and returns the response of this peer
func udpSendGossip(address string, data []byte) []byte {
	s, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		Env.logger.Error(err.Error())
		return nil
	}
	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		Env.logger.Error(err.Error())
		return nil
	}

	Env.logger.Debug("Sending gossip UDP to - " + c.RemoteAddr().String())
	defer c.Close()

	_, err = c.Write(data)

	if err != nil {
		Env.logger.Error(err.Error())
		return nil
	}

	buffer := make([]byte, 1024)
	readLen, _, err := c.ReadFromUDP(buffer)
	if err != nil {
		Env.logger.Error(err.Error())
		return nil
	}
	buffer = buffer[:readLen]

	Env.logger.Debug("Received gossip UDP response from - " + c.RemoteAddr().String())
	return buffer
}

// udpServer listens for an incoming view/gossip from a peer. Responds with the current view as per strategy
func udpServer(port string) {
	s, err := net.ResolveUDPAddr("udp4", ":"+port)
	if err != nil {
		Env.logger.Error(err.Error())
		return
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		Env.logger.Error(err.Error())
		return
	}

	Env.logger.Info("UDP server listening on " + connection.LocalAddr().String())

	defer connection.Close()
	buffer := make([]byte, 1024)

	for {
		readLen, addr, err := connection.ReadFromUDP(buffer)
		buffer = buffer[:readLen]

		view, err := bytesToView(buffer)
		if err == nil {
			Env.logger.Debug("Server received view " + " from: " + addr.IP.String() + ":" + strconv.Itoa(addr.Port))
			responseView := Env.ps.receivedAView(view, addr.String())
			_, err = connection.WriteToUDP(viewToBytes(responseView), addr)
			if err != nil {
				Env.logger.Error(err.Error())
				// return
			} else {
				Env.logger.Debug("Server sent view response to: " + addr.String())
			}
		} else {
			gsp := byteToGossip(buffer)
			from := addr.String()
			gs := make(map[string]gossipMessage)
			gs[from] = gsp
			Env.gossipChan <- gs

			receivedGossip := string(buffer)
			Env.logger.Debug("Server received gossip " + receivedGossip + " from: " + addr.String())
			_, err = connection.WriteToUDP([]byte("sneaky"), addr)
			if err != nil {
				Env.logger.Error(err.Error())
				return
			} else {
				Env.logger.Debug("Server sent gossip " + "sneaky" + " to: " + addr.String())
			}
		}
	}
}
