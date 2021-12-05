package gossip

import (
	"crypto/sha1"
	"fmt"
	"sync"
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
