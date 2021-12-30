package gossip

import (
	"sync"
	"time"
)

type versionAbleI interface {
	// GetVersion returns the version (>= 1) for this identifier or -1 if not found
	GetVersion(id string) int
	// SetVersion sets the version for this identifier. Creates a new entry if not found
	SetVersion(id string, version int)
	// Delete the unused ids
	Delete(ids ...string)
}
type versions struct {
	id      string
	version int
}
type versionAbleS struct {
	mutex  sync.Mutex
	v      map[string]*versions
	remove chan string
}

func (v *versionAbleS) Delete(ids ...string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	for _, id := range ids {
		delete(v.v, id)
	}
}
func (v *versionAbleS) GetVersion(id string) int {
	v.mutex.Lock()
	vs := v.v[id]
	defer v.mutex.Unlock()
	if vs != nil {
		return vs.version
	}
	return -1
}

func (v *versionAbleS) SetVersion(id string, version int) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if v.v[id] != nil {
		v.v[id].version = version
	} else {
		v.v[id] = &versions{
			id:      id,
			version: version,
		}
		go func() { // delete after 60s from creation
			v.remove <- id
		}()

	}
}

func versionMap() versionAbleI {
	vM := versionAbleS{
		mutex:  sync.Mutex{},
		v:      make(map[string]*versions),
		remove: make(chan string),
	}
	go func(vM *versionAbleS) {
		for {
			id := <-vM.remove
			go func(vM *versionAbleS, id string) {
				<-time.After(60 * time.Second)
				vM.Delete(id)
			}(vM, id)
		}
	}(&vM)
	return &vM

}
