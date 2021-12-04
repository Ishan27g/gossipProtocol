package gossip

import (
	"crypto/sha1"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

const (
	ViewExchangeDelay = 3 * time.Second // timeout after which a View  is exchanged with a peer
	RoundDelay        = 3 * time.Second // timeout between each round for a gossipMessage
	MaxNodesInView    = 4               // max peers kept in local View TODO MaxNodesInView=6
	FanOut            = 2               // num of peers to gossip a message to
)

/*
	refactor to get info from raft : leader sends back a subset of all peers (during heartbeat?)
*/

// NetworkPeers returns the set of all possible peers, excluding self
func NetworkPeers(self string) []string {
	var networkPeers []string
	for _, peer := range minPeers {
		if strings.Compare(self, peer) != 0 {
			networkPeers = append(networkPeers, peer)
		}
	}
	return networkPeers
}

var minPeers = []string{"localhost:1101", "localhost:1102", "localhost:1103", "localhost:1104", "localhost:1105",
	"localhost:1106", "localhost:1107", "localhost:1108", "localhost:1109", "localhost:1110"}

/*
	https://flopezluis.github.io/gossip-simulator/
*/
var numRounds = func() int {
	return int(math.Ceil(math.Log10(float64(len(minPeers))) / math.Log10(FanOut)))
}

func hash(obj interface{}) string {
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%v", obj)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

type versions struct {
	mutex sync.Mutex
	v     map[string]*versionAbleS
}

func (v *versions) GetVersion(id string) int {
	vs := v.v[id]
	if vs != nil {
		return vs.version
	}
	return -1
}

func (v *versions) UpdateVersion(id string) {
	if v.v[id] != nil {
		v.v[id].version = v.v[id].version + 1
	} else {
		v.v[id] = &versionAbleS{
			id:      id,
			version: 1,
		}
	}
}

type versionAbleI interface {
	// GetVersion returns the version (>= 1) for this identifier or -1 if not found
	GetVersion(id string) int
	// UpdateVersion increments the version for this identifier, starting from 1. Creates a new entry if not found
	UpdateVersion(id string)
}
type versionAbleS struct {
	id      string
	version int
}

func NewVersions() versionAbleI {
	return &versions{
		mutex: sync.Mutex{},
		v:     make(map[string]*versionAbleS),
	}
}
