package horizon

import (
	"context"
	"sync"
	"time"
)

type Handler func(*EventCtx)

// Emitter is a channel that emits
// *Events to be handled upon
type Emitter func() <-chan *Event

// EmitFunc is the single instance
// of an emitter. It wil return an event.
type EmitFunc func() *Event

// EventCtx contains incoming events and wraps
// them in context
type EventCtx struct {
	// Incoming Event
	*Event

	// EventBus that this event is being handled on
	eb *Eventbus

	// Meta is a collection of user defined contextual data
	Meta map[interface{}]interface{}
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
	if value, ok := ctx.Meta[key]; ok {
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
func (eb *Eventbus) addEmitter(e func() <-chan *Event) {
	eb.emitters = append(eb.emitters, Emitter(e))
}

// ListenTo adds emitters to the eventbus
func (eb *Eventbus) AddEmitter(e func() <-chan *Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.addEmitter(e)

	go func() {
		for {
			select {
			case <-eb.done:
				return
			case event := <-e():
				if event == nil {
					return
				}
				eb.Emit(event)
			}
		}
	}()
}

func (eb *Eventbus) IntervalEmitterFunc(timer time.Duration, f func() *Event) func() <-chan *Event {
	return func() <-chan *Event {
		ev := make(chan *Event)

		go func(chan *Event) {
			interval := time.NewTicker(timer)
			muChan := make(chan struct{})
			for {
				select {
				case <-eb.done:
					return
				case <-muChan:
					if e := f(); e != nil {
						ev <- e
					}
				case <-interval.C:
					muChan <- struct{}{}
				}
			}
		}(ev)
		return ev
	}
}

// TopicHandleFunc constructs a simple handler from a HandlerFunc and
// subscribes it to the specified topic.
// will create topic if it doesn't exist yet
// Will return a reference to the handler to potentially remove it afterwards
func (eb *Eventbus) Handle(topic string, handler func(*EventCtx)) {
	if topic == "" {
		return
	}
	if handler == nil {
		return
	}
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.addHandler(topic, handler)
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
	ctx.Meta = nil

	eb.ctxpool.Put(ctx)
}
