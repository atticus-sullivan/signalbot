package perioder

// signalbot
// Copyright (C) 2024  Lukas Heindl
// 
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import (
	"context"
	"fmt"
	"math"
	"time"

	"log/slog"
)

// implements ReocEvent. Primary reason for public members are to be able to
// marshal/unmarshal from yaml! Modification is not intended.
// Should be created via `NewReocEventImpl`
type ReocEventImpl[T any] struct {
	Description string `yaml:"desc"`
	// Can be used for reconstructing Foo after (de)serialization
	Metadata_store T             `yaml:"meta"`
	Start          time.Time     `yaml:"start"`
	Interval       time.Duration `yaml:"interval"`

	Foo          func(time.Time, ReocEvent[T]) `yaml:"-"`
	log          *slog.Logger                  `yaml:"-"`
	checkStopped func() bool                   `yaml:"-"`
	cancel_      context.CancelFunc
}

func NewReocEventImpl[T any](start time.Time, interval time.Duration, desc string, meta T, foo func(time.Time, ReocEvent[T])) *ReocEventImpl[T] {
	e := ReocEventImpl[T]{
		Start:          start.UTC(),
		Interval:       interval,
		Description:    desc,
		Metadata_store: meta,
		Foo:            foo,
	}
	if foo != nil {
		e.Foo = foo
	} else {
		e.Foo = func(_ time.Time, _ ReocEvent[T]) {}
	}

	return &e
}

// run this event-loop
func (event *ReocEventImpl[T]) run(ctx context.Context) {
	event.run_(ctx)
}

// cancel the event-loop of this event
func (event *ReocEventImpl[T]) cancel() {
	// only run if cancel func is set
	if event.cancel_ != nil {
		event.cancel_()
		// avoid running this function multiple times
		event.cancel_ = nil
	}
}

// run this event-loop asynchronously in context `ctx`
func (event *ReocEventImpl[T]) runAsync(ctx context.Context) (context.Context, context.CancelFunc) {
	c, cFun := context.WithCancel(ctx)
	event.cancel_ = cFun
	go event.run(c)
	return c, cFun
}

// most basic run function. You most probably don't want to shadow/override
// this in a type embedding this one
func (event *ReocEventImpl[T]) run_(ctx context.Context) {
	event.checkStopped = func() bool {
		return ctx.Err() != nil
	}
	// calculate how long until the event occurs the next time
	start_in := time.Until(event.Start)
	if start_inS := start_in.Seconds(); start_inS < 0 {
		// if event already happened the first time, calculate how long until the next time
		x := math.Ceil(-start_inS/event.Interval.Seconds())*event.Interval.Seconds() + start_inS
		start_in = time.Duration(x) * time.Second
	}
	event.log.LogAttrs(context.TODO(), slog.LevelInfo, "initial event", slog.Duration("triggering in", start_in))
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
			event.log.LogAttrs(context.TODO(), slog.LevelInfo, "event triggering", slog.String("desc", event.Description))
			event.Foo(t, event)
		case <-ctx.Done():
			event.log.LogAttrs(context.TODO(), slog.LevelInfo, "event cancel", slog.String("desc", event.Description))
			running = false
		}
	}
}

// return if the event-loop was stopped
func (event *ReocEventImpl[T]) Stopped() bool {
	if event.checkStopped != nil {
		return event.checkStopped()
	}
	return false
}

// set log member (needed to be able to work with ReocEventImpl through an
// interface).
func (event *ReocEventImpl[T]) setLog(log *slog.Logger) {
	event.log = log
}

// get Metadata member (needed to be able to work with ReocEventImpl through an
// interface).
func (event *ReocEventImpl[T]) Metadata() T {
	return event.Metadata_store
}

func (r ReocEventImpl[T]) String() string {
	// return fmt.Sprintf("{id: %v, start: %v, int: %v, desc: %v}", r.Id, r.Start, r.Interval, r.Description)
	return fmt.Sprintf("{desc: %v, start: %v, int: %v}", r.Description, r.Start.Format(time.RFC3339), r.Interval)
}
