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

	ev.Handle(topicName, func(ctx *horizon.EventCtx) {
		fmt.Println("Topic:", ctx.Topic())
		fmt.Println("Source:", ctx.Origin())
		fmt.Println("Data:", ctx.Data())
	})
	ev.Register(func(done <-chan struct{}) <-chan *horizon.Event {
		evChan := make(chan *horizon.Event)
		go func() {
			defer close(evChan)
			evChan <- horizon.NewEvent(topicName, "Emitter", fmt.Sprintf("%s\n", "Message From Emitter"))
		}()
		return evChan
	})

	// Sleep for a sec to give event time to be handled
	time.Sleep(time.Second)
}
```

```bash
$ go run examples/main.go 
Topic: TopicName
Source: Emitter
Data: Message From Emitter
```

## Considerations

* Be sure to close the returned event channel in an emitter when you're done, otherwise it will linger until the eventbus is closed