package perioder

import (
	"context"
	"fmt"
	"time"
)

type ReocEventImplDeadline[T any] struct {
	ReocEventImpl[T] `yaml:",inline"`
	Stop             time.Time `yaml:"stop"`
}

func NewReocEventImplDeadline[T any](start time.Time, interval time.Duration, stop time.Time, desc string, meta T, foo func(time.Time, ReocEvent[T])) *ReocEventImplDeadline[T] {
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
