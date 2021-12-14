package horizon

import (
	"testing"
)

func BenchmarkHandle10kEvents(b *testing.B) {
	HandleEvents(1e4)
}

func BenchmarkHandle50kEvents(b *testing.B) {
	HandleEvents(5e4)
}

func BenchmarkHandle100kEvents(b *testing.B) {
	HandleEvents(1e5)
}

func TestHandle10kEvents(t *testing.T) {
	HandleEvents(1e4)
}

func HandleEvents(n int) {
	topicName := "topicname"

	ev := NewEventbus()
	ev.startTopic(topicName)

	ev.Handle(topicName, func(ctx *EventCtx) {})

	for i := 0; i < n; i++ {
		e := NewEvent(topicName, "testOrigin", i)
		ev.Emit(e)
	}
}
