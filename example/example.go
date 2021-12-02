package main

import (
	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27gOrg/gossipProtocol/gossip"
)

func main() {
	mLogger.New("ok", "trace")
	_ = gossip.Default(gossip.EnvCfg{
		Hostname:     "http://localhost",
		Zone:         1,
		UdpPort:      "7000",
		HttpFilePort: "7001",
	})
	//
	// g.StartRumour("")
	<-make(chan bool)
}
