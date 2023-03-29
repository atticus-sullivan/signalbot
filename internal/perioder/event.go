package perioder

import (
	"context"
	"fmt"
	"math"
	"time"

	"golang.org/x/exp/slog"
)

type ReocEventImpl[T any] struct {
	// Id          uint // TODO should it be this way or should the event know its ID?
	Description  string `yaml:"desc"`
	Foo          func(time.Time) `yaml:"-"`
	Metadata_store     *T `yaml:"meta"` // can be used for reconstructing Foo after (de)serialization
	Start        time.Time `yaml:"start"`
	Interval     time.Duration `yaml:"interval"`
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
		Metadata_store:    meta,
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
	return event.Metadata_store
}
func (r ReocEventImpl[T]) String() string {
	// return fmt.Sprintf("{id: %v, start: %v, int: %v, desc: %v}", r.Id, r.Start, r.Interval, r.Description)
	return fmt.Sprintf("{start: %v, int: %v, desc: %v}", r.Start, r.Interval, r.Description)
}

