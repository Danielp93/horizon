package main

import (
	"fmt"
	"time"

	"github.com/danielp93/horizon"
)

func main() {
	topicName := "TopicName"
	ev := horizon.NewEventbus()
	// Propagates to emitters/handlers
	defer ev.Close()

	ev.StartTopics(topicName)

	ev.Handle(func(ctx *horizon.EventCtx) {
		fmt.Println("Topic:", ctx.Topic())
		fmt.Println("Source:", ctx.Origin())
		fmt.Println("Data:", ctx.Data())
	}, topicName)

	ev.Register(func(done <-chan struct{}) <-chan horizon.Event {
		evChan := make(chan horizon.Event)
		go func() {
			defer close(evChan)
			evChan <- horizon.NewEvent(topicName, "Emitter", fmt.Sprintf("%s\n", "Message From Emitter"))
		}()
		return evChan
	})

	ev.Register(horizon.EmitterFunc(func() horizon.Event {
		return horizon.NewEvent(topicName, "horizon", 1)
	}))

	// Sleep for a sec to give event time to be handled
	time.Sleep(time.Second)
}
