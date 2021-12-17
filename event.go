package horizon

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Event is the interface that will be used to implement events
type Event interface {
	Timestamp() time.Time
	UUID() string

	Topic() string
	Origin() string
	Data() interface{}

	SetData(data interface{})
}

// Event is the actual event that is being passed around within the eventbus
// Meta-Fields like UUID are already provided, if more are needed one can add them in the
// Meta field
type DefaultEvent struct {
	// UUID is a unique identifier for this Event
	uuid string

	// Timestamp is the creation time of this event
	timestamp time.Time

	// Topic is the Topic where the event is being emitted on
	topic string

	// Origin is a textual representation of where this event originated from
	origin string

	// Metadata is a key-value pair of other relevant metadata about the event
	metadata map[string]string

	// data is the event data that is being passed around
	data interface{}
}

func NewEvent(topic string, origin string, data interface{}) Event {
	var uuid [8]byte
	rand.Read(uuid[:])
	return &DefaultEvent{
		uuid:      hex.EncodeToString(uuid[:]),
		timestamp: time.Now(),
		topic:     topic,
		origin:    origin,
		data:      data,
	}
}

func (e *DefaultEvent) SetUUID(uuid string) {
	e.uuid = uuid
}

func (e *DefaultEvent) UUID() string {
	return e.uuid
}

func (e *DefaultEvent) SetTimestamp(timestamp time.Time) {
	e.timestamp = timestamp
}

func (e *DefaultEvent) Timestamp() time.Time {
	return e.timestamp
}

func (e *DefaultEvent) SetTopic(topic string) {
	e.topic = topic
}

func (e *DefaultEvent) Topic() string {
	return e.topic
}

func (e *DefaultEvent) SetOrigin(origin string) {
	e.origin = origin
}

func (e *DefaultEvent) Origin() string {
	return e.origin
}

func (e *DefaultEvent) SetData(data interface{}) {
	e.data = data
}

func (e *DefaultEvent) Data() interface{} {
	return e.data
}

func (e *DefaultEvent) Metadata() map[string]string {
	return e.metadata
}

// CopyTo copies the Event to another newly allocated one
// BE AWARE that data is NOT copied via deepCopy
func (e *DefaultEvent) CopyTo(ev *DefaultEvent) {
	ev.SetTopic(e.topic)
	ev.SetOrigin(e.origin)
	ev.SetData(e.Data())
}
