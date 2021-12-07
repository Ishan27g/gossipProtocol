package gossip

import (
	"crypto/sha1"
	"fmt"
	"time"
)

type EnvCfg struct {
	Hostname string `env:"HOST_NAME"`
	UdpPort  string `env:"UDP_PORT,required"`
}
type Config struct {
	RoundDelay            time.Duration // timeout between each round for a gossipMessage
	FanOut                int           // num of peers to gossip a message to
	MinimumPeersInNetwork int           // number of rounds a message is gossiped = log(minPeers/FanOut)
}

func hash(obj interface{}) string {
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%v", obj)))
	return fmt.Sprintf("%x", h.Sum(nil))
}
