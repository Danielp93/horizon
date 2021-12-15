package horizon

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Event is the actual event that is being passed around within the eventbus
// Meta-Fields like UUID are already provided, if more are needed one can add them in the
// Meta field
type Event struct {
	// UUID is a unique identifier for this Event
	uuid string

	// Timestamp is the creation time of this event
	timestamp time.Time

	// Topic is the Topic where the event is being emitted on
	topic string

	// Origin is a textual representation of where this event originated from
	origin string

	// Meta is a key-value pair of other relevant metadata about the event
	Metadata map[string]string

	// data is the event data that is being passed around
	data interface{}
}

func NewEvent(topic string, origin string, data interface{}) *Event {
	var uuid [8]byte
	rand.Read(uuid[:])
	return &Event{
		uuid:      hex.EncodeToString(uuid[:]),
		timestamp: time.Now(),
		topic:     topic,
		origin:    origin,
		data:      data,
	}
}

func (e *Event) SetUUID(uuid string) {
	e.uuid = uuid
}

func (e *Event) UUID() string {
	return e.uuid
}

func (e *Event) SetTimestamp(timestamp time.Time) {
	e.timestamp = timestamp
}

func (e *Event) Timestamp() time.Time {
	return e.timestamp
}

func (e *Event) SetTopic(topic string) {
	e.topic = topic
}

func (e *Event) Topic() string {
	return e.topic
}

func (e *Event) SetOrigin(origin string) {
	e.origin = origin
}

func (e *Event) Origin() string {
	return e.origin
}

func (e *Event) SetData(data interface{}) {
	e.data = data
}

func (e *Event) Data() interface{} {
	return e.data
}

// CopyTo copies the Event to another newly allocated one
// BE AWARE that data is NOT copied via deepCopy
func (e *Event) CopyTo(ev *Event) {
	ev.SetTopic(e.topic)
	ev.SetOrigin(e.origin)
	ev.SetData(e.Data())
}
