package perioder

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/exp/slog"
)

// event which reoccurs (manages its own loop via `Run`. Can contain arbitrary
// metadata (T).
type ReocEvent[T any] interface {
	// take the context as it is and run the reoccurring event
	run(ctx context.Context)
	// branches off the given context and runs the reoccurring event in this new
	// context in a new goroutine. The new context is returned.
	runAsync(ctx context.Context) (context.Context, context.CancelFunc)
	// check if the event was stopped by the context
	Stopped() bool
	setLog(*slog.Logger)
	Metadata() T
	cancel()
	// get string representation
	String() string
}

// manages reoccurring events
type Perioder[T any] interface {
	// synchronously starts the Perioder. You may want to call this with go
	Start(ctx context.Context)
	// add an event to the Perioder
	Add(ReocEvent[T])
	// remove an event from the Perioder
	Remove(uint)
	// getter for registered events
	Events() map[uint]ReocEvent[T]
	// get string representation
	String() string
}

// implements the perioder interface. Should be created with `NewPerioderImpl`
type PerioderImpl[T any] struct {
	add_s       chan<- ReocEvent[T] // used by Add()
	add_r       <-chan ReocEvent[T] // used for receiving the added Events
	rem_s       chan<- uint         // used by Remove()
	rem_r       <-chan uint         // used for receiving the to be remove Events
	events      map[uint]ReocEvent[T]
	eventsMutex sync.RWMutex
	log         *slog.Logger
}

// Creates a new perioder.
func NewPerioderImpl[T any](log *slog.Logger) *PerioderImpl[T] {
	cAdd := make(chan ReocEvent[T], 3)
	cRem := make(chan uint)
	p := PerioderImpl[T]{
		add_r: cAdd, add_s: cAdd,
		rem_r: cRem, rem_s: cRem,
		log:    log,
		events: make(map[uint]ReocEvent[T]),
	}

	return &p
}

// add events to the perioder
func (p *PerioderImpl[T]) Add(event ReocEvent[T]) {
	p.add_s <- event
}

// remove the event with `id` from the perioder and stop it.
func (p *PerioderImpl[T]) Remove(id uint) {
	p.rem_s <- id
}

// start the perioder in a context (can be canceled). Listens for add/remove
// calls. This function Won't return until the perioder is stopped -> should be
// most probable called in a new goroutine.
func (p *PerioderImpl[T]) Start(ctx context.Context) {
	id := uint(0)
	for {
		select {
		case event := <-p.add_r:
			// event.Id = id
			event.setLog(p.log.With())
			p.eventsMutex.Lock()
			p.events[id] = event
			p.eventsMutex.Unlock()
			id++

			event.runAsync(ctx)
		case id := <-p.rem_r:
			p.eventsMutex.Lock()
			p.events[id].cancel()
			delete(p.events, id)
			p.eventsMutex.Unlock()
		case <-ctx.Done():
			break
		}
	}
}

// returns a map id -> event of all still running events
func (p *PerioderImpl[T]) Events() map[uint]ReocEvent[T] {
	p.eventsMutex.RLock()
	defer p.eventsMutex.RUnlock()

	r := make(map[uint]ReocEvent[T], len(p.events))
	for id, event := range p.events {
		if !event.Stopped() {
			r[id] = event
		}
	}
	return r
}

// stringer
func (p *PerioderImpl[T]) String() string {
	builder := strings.Builder{}
	builder.WriteByte('{')
	events := p.Events()
	first := true
	for id, event := range events {
		if !first {
			builder.WriteString(", ")
		} else {
			first = false
		}
		builder.WriteString(fmt.Sprint(id))
		builder.WriteString(": ")
		builder.WriteString(event.String())
	}
	builder.WriteByte('}')
	return builder.String()
}
