package vClock

import (
	"encoding/json"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/emirpasic/gods/lists/arraylist"
	avl "github.com/emirpasic/gods/trees/avltree"
)

// Events : provides interface for a process with lease/leader-role
// receives `new-events` from followers with eventId and followers vector clock
type Events interface {
	// MergeEvents merges the current event clocks with received event clocks.
	// New events are added to current list
	MergeEvents(es ...Event)
	// MergeEvent takes an eventId & clock and merges with existing clock
	// New events are added to current list
	MergeEvent(e Event)
	// GetCurrentEvents returns the events currently saved. Not in order
	GetCurrentEvents() []Event
	// GetEventsOrder returns the eventIds ordered according to vector clock for the events
	GetEventsOrder() (eventIdsOrHashes []string)
}

func cloudEvent(id string, data EventClock) cloudevents.Event {
	return convertToCloud(Event{
		EventId:    id,
		EventClock: data,
	})
}

func convertToLocal(c cloudevents.Event) Event {
	var e Event
	_ = json.Unmarshal(c.DataEncoded, &e.EventClock)
	e.EventId = c.ID()
	return e
}
func convertToCloud(e Event) cloudevents.Event {
	c := cloudevents.NewEvent()
	c.SetID(e.EventId)
	c.SetSource("oko	k")
	err := c.SetData(cloudevents.ApplicationJSON, e.EventClock)
	if err != nil {
		return cloudevents.Event{}
	}
	return c
}

// Event contains value for eventClocks tree
type Event struct {
	EventId    string     `json:"event_id"`
	EventClock EventClock `json:"event_clock"`
}

// all events
type events struct {
	eventClocks *avl.Tree // key = eventIdOrHash, value = EventClock
}

func (e *events) GetCurrentEvents() []Event {
	var events []Event
	for it := e.eventClocks.Iterator(); it.Next(); {
		clock := it.Value().(Event)
		events = append(events, Event{
			EventId:    it.Key().(string),
			EventClock: clock.EventClock,
		})
	}
	return events
}

func (e *events) MergeEvents(ev ...Event) {
	for _, c := range ev {
		e.MergeEvent(c)
	}
}

// merge entries in v1 with those found in v2
func merge(v1, v2 EventClock) EventClock {
	v := make(EventClock)
	for s, i := range v1 {
		if v2[s] == 0 && i != 0 { // in v1 and not in v2
			v[s] = i
		}
		if v2[s] < i { // in v1 and not in v2
			v[s] = i
		} else {
			v[s] = v2[s]
		}
	}
	return v
}

// MergeClocks merges the current event clock with the provided event clock.
// unique entries from both clocks are kept
func MergeClocks(v1 EventClock, v2 EventClock) *EventClock {
	v := merge(v1, v2)
	v = merge(v2, v)
	return &v
}

func newEvent(eventIdOrHash string, v2 EventClock) Event {
	return Event{
		EventId:    eventIdOrHash,
		EventClock: v2,
	}
}

func (e *events) MergeEvent(ev Event) {
	// check if present with another vectorClock
	v1, found := e.eventClocks.Get(ev.EventId)
	if !found {
		// new entry
		e.eventClocks.Put(ev.EventId, newEvent(ev.EventId, ev.EventClock))
		return
	}
	// get existing EventClock
	v := v1.(Event)
	// merge v1 and v2
	v.EventClock = *v.EventClock.mergeWith(ev.EventClock)
	// update eventClocks
	e.eventClocks.Put(ev.EventId, v)
}

func (e *events) GetEventsOrder() []string {
	k := e.eventClocks.Values()
	a := arraylist.New()
	for i := 0; i < len(k); i++ {
		ec := k[i].(Event)
		a.Add(ec)
	}
	a.Sort(eventComparator)
	var eventIdsOrHashes []string
	a.Each(func(_ int, value interface{}) {
		ec := value.(Event)
		eventIdsOrHashes = append(eventIdsOrHashes, ec.EventId)
	})

	return eventIdsOrHashes
}

var eventComparator = func(a, b interface{}) int {
	v1 := a.(Event)
	v2 := b.(Event)
	c1 := compareClock(v1.EventClock, v2.EventClock)
	c2 := compareClock(v2.EventClock, v1.EventClock)
	if c1 && c2 { // both are same
		return 0
	} else if c1 && !c2 { // e1 happened before
		return -1
	} else {
		return 1
	}
}

func NewEventVector() Events {
	e := events{
		eventClocks: avl.NewWithStringComparator(),
	}
	return &e
}

// compareClock returns true if v1 is before or concurrent to v2
func compareClock(v1, v2 EventClock) bool {
	v1IsBefore := true

	v1 = *v1.mergeWith(v2)

	for addr, v1Clock := range v1 {
		v2Clock := v2[addr]
		if v2Clock < v1Clock {
			v1IsBefore = false
		}
	}
	for addr, v2Clock := range v2 {
		v1Clock := v1[addr]
		if v2Clock < v1Clock {
			v1IsBefore = false
		}
	}
	return v1IsBefore
}
