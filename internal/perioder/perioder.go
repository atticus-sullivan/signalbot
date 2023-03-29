// package main // only for testing so that this can be run with go run
package perioder

// TODO pass pointers to interfaces? e.g. *ReocEvent?

import (
	"context"
	"fmt"
	"strings"

	// https://gogoapps.io/blog/passing-loggers-in-go-golang-logging-best-practices/ -> pass loggers as struct arguments
	// check for news on testing after comment https://github.com/golang/go/issues/56345
	// https://go.dev/play/p/9APBgQuvoo9 maybe for testing handler otherwise just disabel slog when testing
	"math"
	"os"
	"sync"
	"time"

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
	Metadata() *T
	// get string representation
	String() string
	Cancel()
}

type ReocEventImpl[T any] struct {
	// Id          uint // TODO should it be this way or should the event know its ID?
	Description  string
	Foo          func(time.Time)
	metadata     *T // can be used for reconstructing Foo after (de)serialization
	Start        time.Time
	Interval     time.Duration
	Log          *slog.Logger
	checkStopped func() bool
	// should store context as well to be able to tell if the event was stopped with the context
	cancel context.CancelFunc
}

func NewReocEventImpl[T any](start time.Time, interval time.Duration, desc string, meta *T, foo func(time.Time)) *ReocEventImpl[T] {
	e := ReocEventImpl[T]{
		Start:       start,
		Interval:    interval,
		Description: desc,
		metadata:    meta,
		Foo:         foo,
	}
	if foo != nil {
		e.Foo = foo
	} else {
		e.Foo = func(_ time.Time) {}
	}

	return &e
}
func (event *ReocEventImpl[T]) Run(ctx context.Context) {
	event.run(ctx)
}
func (event *ReocEventImpl[T]) Cancel() {
	event.cancel()
}
func (event *ReocEventImpl[T]) RunAsync(ctx context.Context) (context.Context, context.CancelFunc) {
	c, cFun := context.WithCancel(ctx)
	event.cancel = cFun
	go event.Run(c)
	return c, cFun
}

// most basic run function. You most probably don't want to shadow/override
// this in a type embedding this one
func (event *ReocEventImpl[T]) run(ctx context.Context) {
	event.checkStopped = func() bool {
		return ctx.Err() == nil
	}
	// calculate how long until the event occurs the next time
	start_in := time.Until(event.Start)
	if start_inS := start_in.Seconds(); start_inS < 0 {
		// if event already happened the first time, calculate how long until the next time
		x := math.Ceil(-start_inS/event.Interval.Seconds())*event.Interval.Seconds() + start_inS
		start_in = time.Duration(x) * time.Second
	}
	event.Log.LogAttrs(context.TODO(), slog.LevelInfo, "initial event", slog.Duration("triggering in", start_in))
	ticker := time.NewTicker(start_in)
	first := true
	running := true
	for running {
		select {
		case t := <-ticker.C:
			if first {
				first = false
				ticker.Reset(event.Interval)
			}
			event.Log.LogAttrs(context.TODO(), slog.LevelInfo, "event triggering", slog.String("desc", event.Description))
			event.Foo(t)
		case <-ctx.Done():
			event.Log.LogAttrs(context.TODO(), slog.LevelInfo, "event cancel", slog.String("desc", event.Description))
			running = false
		}
	}
}
func (event *ReocEventImpl[T]) Stopped() bool {
	if event.checkStopped != nil {
		return event.checkStopped()
	}
	return false
}
func (event *ReocEventImpl[T]) SetLog(log *slog.Logger) {
	event.Log = log
}
func (event *ReocEventImpl[T]) Metadata() *T {
	return event.metadata
}
func (r ReocEventImpl[T]) String() string {
	// return fmt.Sprintf("{id: %v, start: %v, int: %v, desc: %v}", r.Id, r.Start, r.Interval, r.Description)
	return fmt.Sprintf("{start: %v, int: %v, desc: %v}", r.Start, r.Interval, r.Description)
}

type ReocEventImplDeadline[T any] struct {
	ReocEventImpl[T]
	Stop time.Time
}

func NewReocEventImplDeadline[T any](start time.Time, interval time.Duration, stop time.Time, desc string, meta *T, foo func(time.Time)) *ReocEventImplDeadline[T] {
	e := ReocEventImplDeadline[T]{
		ReocEventImpl: *NewReocEventImpl(start, interval, desc, meta, foo),
		Stop:          stop,
	}
	return &e
}
func (event *ReocEventImplDeadline[T]) Run(ctx context.Context) {
	if time.Now().Compare(event.Stop) == 1 {
		event.checkStopped = func() bool {
			return true
		}
		return // do not start if deadline already exceeded
	}
	event.run(ctx)
}
func (event *ReocEventImplDeadline[T]) RunAsync(ctx context.Context) (context.Context, context.CancelFunc) {
	c, cFun := context.WithDeadline(ctx, event.Stop)
	event.cancel = cFun
	go event.Run(c)
	return c, cFun
}
func (r ReocEventImplDeadline[T]) String() string {
	return fmt.Sprintf("{start: %v, stop: %v, int: %v, desc: %v}", r.Start, r.Stop, r.Interval, r.Description)
}

type Perioder[T any] interface {
	// synchronously starts the Perioder. You may want to call this with go
	Start(ctx context.Context)
	// add an event to the Perioder
	Add(ReocEvent[T])
	// getter for registered events
	Events() map[uint]ReocEvent[T]
	// get string representation
	String() string
}

// todo add a remove channel?
type PerioderImpl[T any] struct {
	add_s       chan<- ReocEvent[T] // used by Add()
	add_r       <-chan ReocEvent[T] // used for receiving the added Events
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
		r[id] = event
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

func main() {
	var per Perioder[any] = NewPerioderImpl[any](slog.New(slog.NewTextHandler(os.Stderr)))
	ctx, p_cancel := context.WithCancel(context.Background())
	go per.Start(ctx)

	e1 := NewReocEventImpl[any](
		time.Date(2023, 03, 15, 22, 40, 00, 00, time.Now().Local().Location()),
		time.Duration(20*time.Second),
		"Hello reocc",
		nil, // no metadata
		nil, // no foo
	)
	e2 := NewReocEventImplDeadline[any](
		time.Date(2023, 03, 15, 22, 42, 00, 00, time.Now().Local().Location()),
		time.Duration(10*time.Second),
		time.Now().Add(15*time.Second), // Deadline
		"Hello fastly reocc",
		nil, // no metadata
		nil, // no foo
	)
	per.Add(e1)
	per.Add(e2)

	// time.Sleep(30000)
	stop := make(chan struct{})
	fmt.Println("You'll need to stop this via CTRL+C")
	<-stop
	p_cancel()
}
