// package main // only for testing so that this can be run with go run
package perioder

import (
	"context"
	"fmt"
	"strings"

	// https://gogoapps.io/blog/passing-loggers-in-go-golang-logging-best-practices/ -> pass loggers as struct arguments
	// check for news on testing after comment https://github.com/golang/go/issues/56345
	// https://go.dev/play/p/9APBgQuvoo9 maybe for testing handler otherwise just disabel slog when testing

	"sync"

	"golang.org/x/exp/slog"
)

type ReocEvent[T any] interface {
	// take the context as it is and run the reoccurring event
	Run(ctx context.Context)
	// branches off the given context and runs the reoccuring event in this new
	// context in a new goroutine. The new context is returned.
	RunAsync(ctx context.Context) (context.Context, context.CancelFunc)
	// check if the event was stopped by the context
	Stopped() bool
	SetLog(*slog.Logger)
	Metadata() T
	// get string representation
	String() string
	Cancel()
}

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

// todo add a remove channel?
type PerioderImpl[T any] struct {
	add_s       chan<- ReocEvent[T] // used by Add()
	add_r       <-chan ReocEvent[T] // used for receiving the added Events
	rem_s       chan<- uint // used by Remove()
	rem_r       <-chan uint // used for receiving the to be removec Events
	events      map[uint]ReocEvent[T]
	eventsMutex sync.RWMutex
	log         *slog.Logger
}

func NewPerioderImpl[T any](log *slog.Logger) *PerioderImpl[T] {
	c := make(chan ReocEvent[T], 3)
	p := PerioderImpl[T]{add_r: c, add_s: c, log: log, events: make(map[uint]ReocEvent[T])}

	return &p
}
func (p *PerioderImpl[T]) Add(event ReocEvent[T]) {
	p.add_s <- event
}
func (p *PerioderImpl[T]) Remove(id uint) {
	p.rem_s <- id
}
func (p *PerioderImpl[T]) Start(ctx context.Context) {
	id := uint(0)
	for {
		select {
		case event := <-p.add_r:
			// event.Id = id
			event.SetLog(p.log.With())
			p.eventsMutex.Lock()
			p.events[id] = event
			p.eventsMutex.Unlock()
			id++

			event.RunAsync(ctx)
		case id := <-p.rem_r:
			p.eventsMutex.Lock()
			delete(p.events, id)
			p.eventsMutex.Unlock()
		case <-ctx.Done():
			break
		}
	}
}

// todo should this filter if the event is still running?
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
