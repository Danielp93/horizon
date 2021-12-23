# (Event) Horizon
<img align="right" src="docs/images/horizon_round_650x650.png" alt="Event Horizon" width="250" height="250" >

Horizon is my personal project for understanding and implementing a library that enables event-driven design.
The goal of this package is to provide a bus or communication method that consists of entities that emit events to certain topics, and handlers that act upon those events.


## Installation

```bash
$ go get github.com/danielp93/horizon
```

## Usage

There are a few key components to this library

* Handlers: Functions of signature `func (*EventCtx)` they take events and act upon those
* Emitter: Functions of signature `func(done) <-chan struct{}`, the `done` input is a channel that signals when to close the emitter (automatically closed if an emitter closes itself)
* Eventbus: place where handles and emitters are registered

```go
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
		return horizon.NewEvent(topicName, "EmitEvery1Second", time.Now().String())
	}).Interval(1 * time.Second))

	// very ugly block until ctrl-C
	select {}
}
```

```bash
$ go run examples/main.go 
Topic: TopicName
Source: EmitEvery1Second
Topic: TopicName
Source: Emitter
Data: Message From Emitter

Topic: TopicName
Source: EmitOnceFunc
Data: 1
Data: 2021-12-23 14:48:20.4377732 +0100 CET m=+0.000097001
```

## Considerations

* Be sure to close the returned event channel in an emitter when you're done, otherwise it will linger until the eventbus is closed
* No Acking/nAcking is implemented yet, possibly in the future. Closing the eventbus/handlers/Emitters will not wait for undelivered events 
