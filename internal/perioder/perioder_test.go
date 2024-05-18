package perioder_test

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
	"io"
	"signalbot_go/internal/perioder"
	"testing"
	"time"

	"log/slog"
)

func nopLog() *slog.Logger {
	return slog.New(slog.HandlerOptions{Level: slog.LevelError}.NewTextHandler(io.Discard))
}

func TestPerioder(t *testing.T) {
	log := nopLog()
	p := perioder.NewPerioderImpl[any](log)

	ctx, cancel := context.WithCancel(context.Background())
	go p.Start(ctx)
	defer cancel()

	if len(p.Events()) != 0 {
		t.Fatalf("Invalid amount (%d should be %d) of events before adding anything", len(p.Events()), 0)
	}

	e1_run := make(chan time.Time, 1)
	e1_time := time.Now().Add(-500 * time.Millisecond)
	e1_int := 4 * time.Second
	e1_exec := e1_time.Add(e1_int)
	e1 := perioder.NewReocEventImpl(e1_time, e1_int, "testingA", nil, func(t time.Time, p perioder.ReocEvent[any]) {
		select {
		case e1_run <- time.Now():
			// value was sent
		default:
			// drop it
		}
	})

	p.Add(e1)
	time.Sleep(500 * time.Millisecond)
	if len(p.Events()) != 1 {
		t.Fatalf("Invalid amount (%d should be %d) of events", len(p.Events()), 1)
	}

	e2_run := make(chan time.Time, 1)
	e2_time := time.Now().Add(-500 * time.Millisecond)
	e2_int := 4 * time.Second
	e2_exec := e2_time.Add(e2_int)
	e2_stop := make(chan time.Time, 1)
	e2_s := e2_time.Add(7 * time.Second)
	e2 := perioder.NewReocEventImplDeadline(e2_time, e2_int, e2_s, "testingB", nil, func(t time.Time, p perioder.ReocEvent[any]) {
		select {
		case e2_run <- time.Now():
			// value was sent
		default:
			// drop it
		}
	})

	// check when e2 is being canceled
	go func() {
		for !e2.Stopped() {
			time.Sleep(1 * time.Second)
		}
		e2_stop <- time.Now()
	}()

	p.Add(e2)
	time.Sleep(500 * time.Millisecond)
	if len(p.Events()) != 2 {
		t.Fatalf("Invalid amount (%d should be %d) of events", len(p.Events()), 2)
	}

	for i := 0; i < 3; i++ {
		select {
		case ti := <-e1_run:
			e1_run = nil
			d := ti.Sub(e1_exec)
			if d > 2*time.Second {
				t.Fatalf("e1 run too early %v should: %v diff: %v (started: %v)", ti, e1_exec, d, e1_time)
			}
			if d < -2*time.Second {
				t.Fatalf("e1 run too late %v should: %v", ti, e1_exec)
			}
			if e1.Stopped() {
				t.Fatalf("e1 was stopped too early")
			}

		case ti := <-e2_run:
			e2_run = nil
			d := ti.Sub(e2_exec)
			if d > 2*time.Second {
				t.Fatalf("e2 run too early %v should: %v", ti, e2_exec)
			}
			if d < -2*time.Second {
				t.Fatalf("e2 run too late %v should: %v", ti, e2_exec)
			}
			if e2.Stopped() {
				t.Fatalf("e2 was stopped too early")
			}

		case ti := <-e2_stop:
			e2_stop = nil
			if ti.Sub(e2_s) > time.Second {
				t.Fatalf("e2 stopped too early")
			}
			if ti.Sub(e2_s) < -time.Second {
				t.Fatalf("e2 stopped too late")
			}
		}
	}

	p.Remove(0)
	time.Sleep(1 * time.Second)
	if len(p.Events()) != 0 {
		t.Fatalf("Invalid amount (%d should be %d) of events", len(p.Events()), 0)
	}
	time.Sleep(1 * time.Second)
	if !e1.Stopped() {
		t.Fatalf("e1 wasn't stopped when it should be")
	}
}
