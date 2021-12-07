package vClock

import (
	"fmt"
	"sync"

	"github.com/emirpasic/gods/lists/arraylist"
)

// VectorClock provides interface to implement vector clock
type VectorClock interface {
	// Get returns the current vector clock
	Get(eventIdOrHash string) EventClock

	// SendEvent returns the current vector clock after first updating self clock and then
	// updating the individual clocks for corresponding addresses
	SendEvent(eventIdOrHash string, addresses []string) EventClock

	// ReceiveEvent updates the current vector clock using element wise maximum with the passed vector clock
	ReceiveEvent(eventIdOrHash string, v EventClock)

	Clear(eventIdOrHash string)
	print()
}

// EventClock is vector-clock of peer-address and its individual clock
type EventClock map[string]int

func (v *EventClock) mergeWith(v2 EventClock) *EventClock {
	return MergeClocks(*v, v2)
}

type vClock struct {
	lock        sync.Mutex
	self        string
	vectorClock map[string]EventClock      // clock per event
	addressList map[string]*arraylist.List // clock-peers for each event
}

func (v *vClock) Clear(eventIdOrHash string) {
	v.lock.Lock()
	v.vectorClock[eventIdOrHash] = make(EventClock)
	v.addressList[eventIdOrHash].Clear()
	v.initClock(eventIdOrHash, v.self)
	v.lock.Unlock()
}

func (v *vClock) print() {
	for a, c := range v.vectorClock {
		fmt.Println(a, c)
	}
}

// Get returns the current vector clock
func (v *vClock) Get(eventIdOrHash string) EventClock {
	return v.vectorClock[eventIdOrHash]
}

// SendEvent returns the current vector clock after updating the individual clocks for these entries
func (v *vClock) SendEvent(eventIdOrHash string, address []string) EventClock {
	v.lock.Lock()
	defer v.lock.Unlock()
	// update the individual clock entry for self
	v.event(eventIdOrHash, v.self)

	for _, a := range address {
		v.event(eventIdOrHash, a)
	}
	return v.Get(eventIdOrHash)
}

// ReceiveEvent updates the current vector clock using element wise maximum with the passed vector clock
func (v *vClock) ReceiveEvent(eventIdOrHash string, v1 EventClock) {
	v.lock.Lock()
	defer v.lock.Unlock()
	// update local clock
	v.event(eventIdOrHash, v.self)

	if v.vectorClock[eventIdOrHash] == nil {
		v.vectorClock[eventIdOrHash] = make(EventClock)
	}
	// merge with received clock
	for address, newClock := range v1 {
		if v.addressList[eventIdOrHash].Contains(address) {
			v.updateClock(eventIdOrHash, address, newClock)
		} else { // if new address
			// v.initClock(eventIdOrHash, address)
			v.addressList[eventIdOrHash].Add(address)
			v.updateClock(eventIdOrHash, address, newClock)
		}
	}
}

// Event updates the individual clock entry for this entry
func (v *vClock) event(eventIdOrHash, address string) {
	if v.vectorClock[eventIdOrHash] == nil {
		v.initClock(eventIdOrHash, address)
		v.addressList[eventIdOrHash].Add(address)
	}
	currentClock := v.vectorClock[eventIdOrHash]
	v.vectorClock[eventIdOrHash][address] = currentClock[address] + 1
}

// updateClock updates the individual clock if it is lower than the new clock
func (v *vClock) updateClock(eventIdOrHash, address string, newClock int) {
	if v.vectorClock[eventIdOrHash][address] < newClock {
		v.vectorClock[eventIdOrHash][address] = newClock
	}
}

func Init(self string) VectorClock {
	v := vClock{
		lock:        sync.Mutex{},
		vectorClock: make(map[string]EventClock),
		self:        self,
		addressList: make(map[string]*arraylist.List),
	}
	//v.initClock(v.self)
	return &v
}

func (v *vClock) initClock(event, peer string) {
	v.vectorClock[event] = EventClock{
		peer: 0,
	}
	v.addressList[event] = arraylist.New()
}
