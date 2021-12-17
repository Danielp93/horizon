package horizon

import (
	"context"
	"sync"
	"time"
)

type Handler func(*EventCtx)

// Emitter is a function that returns a channel to
// receive messages on, and a channel with which we can
// stop the emitter. It takes a channel which should be closed
// to signal the Emitter when to stop.
type Emitter func(done <-chan struct{}) <-chan *Event

// wraps a normal function that emits an event into an emitter to be handled
func EmitterFunc(f func() *Event) Emitter {
	return func(done <-chan struct{}) <-chan *Event {
		eChan := make(chan *Event)
		go func() {
			defer close(eChan)

			select {
			case <-done:
				return
			case eChan <- f():
			}
		}()
		return eChan
	}
}

// EventCtx contains incoming events and wraps
// them in context
type EventCtx struct {
	// Incoming Event
	*Event

	// EventBus that this event is being handled on
	eb *Eventbus

	// Meta is a collection of user defined contextual data
	Values map[interface{}]interface{}
}

func (ctx *EventCtx) UUID() string {
	return ctx.Event.UUID()
}

func (ctx *EventCtx) Timestamp() time.Time {
	return ctx.Event.Timestamp()
}

func (ctx *EventCtx) Topic() string {
	return ctx.Event.Topic()
}

func (ctx *EventCtx) Origin() string {
	return ctx.Event.Origin()
}

func (ctx *EventCtx) Meta() map[string]string {
	return ctx.Event.Metadata()
}

func (ctx *EventCtx) Data() interface{} {
	return ctx.Event.Data()
}

// Following (ctx *EventCtx) functions are to implement the context.Context Interface

// Deadline returns the time when work done on behalf of this context
// should be canceled. Deadline returns ok==false when no deadline is
// set. Successive calls to Deadline return the same results.
func (ctx *EventCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done returns a channel that's closed when work done on behalf of this
// context should be canceled. Done may return nil if this context can
// never be canceled. Successive calls to Done return the same value.
func (ctx *EventCtx) Done() <-chan struct{} {
	return ctx.eb.done
}

func (ctx *EventCtx) Err() error {
	// No errors implemented except for cancellation by 'done'
	select {
	case <-ctx.eb.done:
		return context.Canceled
	default:
		return nil
	}
}

func (ctx *EventCtx) Value(key interface{}) interface{} {
	if value, ok := ctx.Values[key]; ok {
		return value
	}
	return nil
}

// compile error if EventCtx doesn't implement context.Context interface
var _ context.Context = &EventCtx{}

type Eventbus struct {
	// Mutex for when altering the topics/handlers
	mu sync.RWMutex

	// Map for storing the handlers that handle upon the Events
	topics map[string][]Handler

	emitters []Emitter
	ctxpool  sync.Pool
	done     chan struct{}
}

func NewEventbus() *Eventbus {
	return &Eventbus{
		mu:       sync.RWMutex{},
		topics:   make(map[string][]Handler),
		emitters: make([]Emitter, 0),
		done:     make(chan struct{}),
		ctxpool: sync.Pool{
			New: func() interface{} {
				return &EventCtx{}
			},
		},
	}
}

func (eb *Eventbus) Close() {
	if eb.done != nil {
		close(eb.done)
	}
}

func (eb *Eventbus) Emit(event *Event) {
	if event == nil {
		return
	}

	handlers := eb.Handlers(event.Topic())
	if handlers == nil {
		return
	}
	ctx := eb.getCtx(event)
	if ctx == nil {
		return
	}

	for _, handler := range handlers {
		handler(ctx)
	}

	eb.putCtx(ctx)
}

func (eb *Eventbus) Topics() []string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	res := make([]string, 0, len(eb.topics))
	for t := range eb.topics {
		res = append(res, t)
	}

	return res
}

func (eb *Eventbus) StartTopics(topics ...string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for _, t := range topics {
		eb.startTopic(t)
	}
}

// startTopic initializes a topic. This is internal
// Make sure to Lock mutex in calling functions
func (eb *Eventbus) startTopic(t string) {
	if _, ok := eb.topics[t]; ok {
		return
	}
	eb.topics[t] = make([]Handler, 0)
}

// addHandler adds a handler to a topic. This is
// an interal function.
// creates topic if non-existent
// Make sure to Lock mutex in calling functions
func (eb *Eventbus) addHandler(topic string, h func(*EventCtx)) {
	eb.topics[topic] = append(eb.topics[topic], h)
}

// addEmitter adds an emitter to the eventbus. This is
// an interal function.
// Make sure to Lock mutex in calling functions
func (eb *Eventbus) addEmitter(e Emitter) {
	eb.emitters = append(eb.emitters, Emitter(e))
}

// Register adds an emitter to the eventbus
// It listens to this emitter until it is closed
func (eb *Eventbus) Register(e Emitter) {
	eb.mu.Lock()
	eb.addEmitter(e)
	eb.mu.Unlock()

	go func() {
		done := make(chan struct{})
		defer close(done)

		ecChan := e(done)
		for {
			select {
			case <-eb.done:
				return
			case event := <-ecChan:
				if event == nil {
					return
				}
				eb.Emit(event)
			}
		}
	}()
}

// TopicHandleFunc constructs a simple handler from a HandlerFunc and
// subscribes it to the specified topic.
// will create topic if it doesn't exist yet
// Will return a reference to the handler to potentially remove it afterwards
func (eb *Eventbus) Handle(handler func(*EventCtx), topics ...string) {
	if handler == nil {
		return
	}
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Handle case nil (means add to all topics)
	if topics == nil {
		for topic := range eb.topics {
			eb.addHandler(topic, handler)
		}
		return
	}

	for _, topic := range topics {
		eb.addHandler(topic, handler)
	}
}

// Handlers retrieves the that should handle an event
// If there are no handlers for given event, return nil
func (eb *Eventbus) Handlers(topic string) []Handler {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if topic == "" {
		return nil
	}

	return eb.topics[topic]
}

func (eb *Eventbus) getCtx(e *Event) *EventCtx {
	var ctx *EventCtx
	v := eb.ctxpool.Get()
	if ctx == nil {
		ctx = &EventCtx{
			eb: eb,
		}
	} else {
		ctx = v.(*EventCtx)
	}
	ctx.Event = e
	return ctx
}

func (eb *Eventbus) putCtx(ctx *EventCtx) {
	ctx.Event.SetData(nil)
	ctx.Event = nil
	ctx.Values = nil

	eb.ctxpool.Put(ctx)
}
