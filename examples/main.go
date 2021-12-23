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

	// Create a full Emitter
	ev.Register(func(done <-chan struct{}) <-chan horizon.Event {
		evChan := make(chan horizon.Event)
		go func() {
			defer close(evChan)
			evChan <- horizon.NewEvent(topicName, "Emitter", fmt.Sprintf("%s\n", "Message From Emitter"))
		}()
		return evChan
	})

	ev.Register(horizon.EmitterFunc(func() horizon.Event {
		return horizon.NewEvent(topicName, "EmitOnceFunc", 1)
	}).Once())

	ev.Register(horizon.EmitterFunc(func() horizon.Event {
		return horizon.NewEvent(topicName, "EmitEvery1Second", time.Now())
	}).Interval(1 * time.Second))

	// very ugly block until ctrl-C
	select {}
}
