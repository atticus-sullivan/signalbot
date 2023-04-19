package perioder

import (
	"context"
	"fmt"
	"time"
)

// adds deadline to ReocEvent. Should be created with `NewReocEventImplDeadline`
type ReocEventImplDeadline[T any] struct {
	ReocEventImpl[T] `yaml:",inline"`
	Stop             time.Time `yaml:"stop"`
}

func NewReocEventImplDeadline[T any](start time.Time, interval time.Duration, stop time.Time, desc string, meta T, foo func(time.Time, ReocEvent[T])) *ReocEventImplDeadline[T] {
	e := ReocEventImplDeadline[T]{
		ReocEventImpl: *NewReocEventImpl(start, interval, desc, meta, foo),
		Stop:          stop.UTC(),
	}
	return &e
}

// start the event-loop synchronously
func (event *ReocEventImplDeadline[T]) run(ctx context.Context) {
	if time.Now().Compare(event.Stop) == 1 {
		event.checkStopped = func() bool {
			return true
		}
		return // do not start if deadline already exceeded
	}
	event.run_(ctx)
}

// start the event-loop asynchronously in the context ctx
func (event *ReocEventImplDeadline[T]) runAsync(ctx context.Context) (context.Context, context.CancelFunc) {
	c, cFun := context.WithDeadline(ctx, event.Stop)
	event.cancel_ = cFun
	go event.run(c)
	return c, cFun
}

func (r ReocEventImplDeadline[T]) String() string {
	return fmt.Sprintf("{start: %v, stop: %v, int: %v, desc: %v}", r.Start.Format(time.RFC3339), r.Stop.Format(time.RFC3339), r.Interval, r.Description)
}
