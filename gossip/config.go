package gossip

import (
	"crypto/sha1"
	"fmt"
	"time"
)

type envConfig struct {
	Hostname          string `env:"HOST_NAME"`
	UdpPort           string `env:"UDP_PORT,required"`
	ProcessIdentifier string
}
type config struct {
	RoundDelay            time.Duration // timeout between each round for a gossipMessage
	FanOut                int           // num of peers to gossip a message to
	MinimumPeersInNetwork int           // number of rounds a message is gossiped = log(minPeers/FanOut)
}

func defaultEnv(hostname string, port string, address string) envConfig {
	return envConfig{
		Hostname:          hostname,
		UdpPort:           ":" + port,
		ProcessIdentifier: address,
	}
}
func defaultConfig() *config {
	return &config{
		RoundDelay:            500 * time.Millisecond,
		FanOut:                2,
		MinimumPeersInNetwork: 10,
	}
}
func hash(obj interface{}) string {
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%v", obj)))
	return fmt.Sprintf("%x", h.Sum(nil))
}
