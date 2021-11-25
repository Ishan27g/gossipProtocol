package gossip

import (
	"crypto/sha1"
	"fmt"
	"math"
	"time"
)

const (
	ViewExchangeDelay = 5 * time.Second // timeout after which a view  is exchanged with a peer
	RoundDelay        = 5 * time.Second // timeout between each round for a gossipMessage
	MaxNodesInView    = 2               // max peers kept in local view TODO MaxNodesInView=6
	FanOut            = 2               // num of peers to gossip a message to
)

var minPeers = []string{"localhost:1101", "localhost:1102", "localhost:1103"}//, "localhost:1104"}//, "localhost:1105",
// "localhost:1106"}//, "localhost:1107", "localhost:1108", "localhost:1109", "localhost:1110"}
var numRounds = func() int {
	return int(math.Ceil(math.Log10(float64(len(minPeers))) / math.Log10(FanOut)))
}

func hash(obj interface{}) string {
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%v", obj)))
	return fmt.Sprintf("%x", h.Sum(nil))
}
