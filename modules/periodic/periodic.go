package periodic

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
	"os"
	"path/filepath"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/internal/perioder"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules"
	"signalbot_go/signalcli"
	"strings"
	"time"

	"log/slog"

	"github.com/alexflint/go-arg"
	"gopkg.in/yaml.v3"
)

type Periodic struct {
	modules.Module
	perioder perioder.Perioder[signalcli.Message] `yaml:"-"`
	stop     context.CancelFunc                    `yaml:"-"`
}

func NewPeriodic(log *slog.Logger, cfgDir string) (*Periodic, error) {
	r := Periodic{
		Module:   modules.NewModule(log, cfgDir),
		perioder: perioder.NewPerioderImpl[signalcli.Message](log.With()),
	}

	// validation
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if err := r.Module.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (p *Periodic) Validate() error {
	return nil
}

type Args struct {
	Add *addArgs `arg:"subcommand:add|a"`
	Ls  *lsArgs  `arg:"subcommand:list|ls|l"`
	Rm  *rmArgs  `arg:"subcommand:remove|rm|r"`
}

type addArgs struct {
	Start  time.Time     `arg:"--time"`
	Until  time.Time     `arg:"--until"`
	Every  time.Duration `arg:"--every"`
	EveryD uint          `arg:"--everyD" default:"0"`
	Desc   string        `arg:"--desc"`
	Msg    string        `arg:"positional"`
}
type lsArgs struct{}
type rmArgs struct {
	Id uint `arg:"--id,-i,required"`
}

// handle a signalmessage
func (r *Periodic) Handle(m *signalcli.Message, signal signalsender.SignalSender, virtRcv func(*signalcli.Message)) {
	var args Args
	parser, err := arg.NewParser(arg.Config{}, &args)
	if err != nil {
		r.Log.Error(fmt.Sprintf("newParser -> %v", err))
		return
	}

	if err := r.Module.Handle(m, signal, virtRcv, parser); err != nil {
		errMsg := err.Error()
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	switch {
	case args.Add != nil:
		args.Add.Every += time.Duration(24*time.Hour) * time.Duration(args.Add.EveryD)
		args.Add.EveryD = 0
		if args.Add.Start.IsZero() {
			args.Add.Start = time.Now()
		}
		r.Add(args.Add, *m, signal, virtRcv)
	case args.Ls != nil:
		r.Ls(args.Ls, *m, signal, virtRcv)
	case args.Rm != nil:
		r.Rm(args.Rm, *m, signal, virtRcv)
	}
}

func (r *Periodic) Add(add *addArgs, m signalcli.Message, signal signalsender.SignalSender, virtRcv func(*signalcli.Message)) {
	if add.Every == time.Duration(0) {
		errMsg := fmt.Sprintf("Invalid duration: %v", add.Every)
		r.Log.Info(errMsg)
		r.SendError(&m, signal, errMsg)
		return
	}
	var err error
	m.Message, err = cmdsplit.Unescape(add.Msg)
	if err != nil {
		errMsg := fmt.Sprintf("Error on unescaping message: %v", err)
		r.Log.Error(errMsg)
		r.SendError(&m, signal, errMsg)
		return
	}
	if add.Desc == "" {
		add.Desc = m.Message
	}
	var event perioder.ReocEvent[signalcli.Message]
	if add.Until.IsZero() {
		event = perioder.NewReocEventImpl(add.Start, add.Every, add.Desc, m, func(time time.Time, event perioder.ReocEvent[signalcli.Message]) {
			meta := event.Metadata()
			virtRcv(&meta)
		})
	} else {
		event = perioder.NewReocEventImplDeadline(add.Start, add.Every, add.Until, add.Desc, m, func(time time.Time, event perioder.ReocEvent[signalcli.Message]) {
			meta := event.Metadata()
			virtRcv(&meta)
		})
	}
	r.perioder.Add(event)
	if _, err := signal.Respond(fmt.Sprintf("Added %v\n", event.String()), nil, &m, true); err != nil {
		r.Log.Error(fmt.Sprintf("error sending add success msg: %v", err))
	}
}

func (r *Periodic) Ls(ls *lsArgs, m signalcli.Message, signal signalsender.SignalSender, virtRcv func(*signalcli.Message)) {
	eventsAll := r.perioder.Events()
	events := make(map[uint]perioder.ReocEvent[signalcli.Message])
	for k, v := range eventsAll {
		if v.Metadata().Sender == m.Sender {
			events[k] = v
		}
	}
	builder := strings.Builder{}
	first := true
	for i,j := range events {
		if !first {
			builder.WriteRune('\n')
		}
		builder.Write([]byte(fmt.Sprintf("%d: %v", i, j)))
		first = false
	}
	if _, err := signal.Respond(builder.String(), nil, &m, true); err != nil {
		r.Log.Error(fmt.Sprintf("error sending ls output: %v", err))
	}
}

func (r *Periodic) Rm(rm *rmArgs, m signalcli.Message, signal signalsender.SignalSender, virtRcv func(*signalcli.Message)) {
	event, ok := r.perioder.Events()[rm.Id]
	if !ok || event.Metadata().Sender != m.Sender {
		errMsg := fmt.Sprintf("Error: Event with ID %d does not exist or you don't added this event", rm.Id)
		r.Log.Error(errMsg)
		r.SendError(&m, signal, errMsg)
		return
	}
	r.Log.Info(fmt.Sprintf("canceling event with ID: %d (%s)", rm.Id, event.String()))
	r.perioder.Remove(rm.Id)
	if _, err := signal.Respond(fmt.Sprintf("Removed %v\n", event.String()), nil, &m, true); err != nil {
		r.Log.Error(fmt.Sprintf("error sending rm success msg: %v", err))
	}
}

func (r *Periodic) Start(virtRcv func(*signalcli.Message)) error {
	if err := r.Module.Start(virtRcv); err != nil {
		return err
	}
	// start perioder
	var ctx context.Context
	ctx, r.stop = context.WithCancel(context.Background())
	go r.perioder.Start(ctx)

	// read saved events
	f, err := os.Open(filepath.Join(r.ConfigDir, "events.yaml"))
	if !os.IsNotExist(err) {
		if err != nil {
			// p.Log.Error(fmt.Sprintf("Error opening 'events.yaml': %v", err))
			return err
		}
		d := yaml.NewDecoder(f)
		events := make(map[uint]perioder.ReocEventImplDeadline[signalcli.Message])
		err = d.Decode(&events)
		if err != nil {
			// p.Log.Error(fmt.Sprintf("Error decoding to 'events.yaml': %v", err))
			return err
		}
		// add events
		for _, vIter := range events {
			v := vIter // force copy
			v.Foo = func(time time.Time, event perioder.ReocEvent[signalcli.Message]) {
				meta := event.Metadata()
				virtRcv(&meta)
			}
			if v.Stop.IsZero() {
				r.perioder.Add(&v.ReocEventImpl)
			} else {
				r.perioder.Add(&v)
			}
		}
	}
	return nil
}

func (r *Periodic) Close(virtRcv func(*signalcli.Message)) {
	r.Module.Close(virtRcv)

	r.Log.Info("closing periodic stuff")
	f, err := os.Create(filepath.Join(r.ConfigDir, "events.yaml"))
	if err != nil {
		r.Log.Error(fmt.Sprintf("Error opening 'events.yaml': %v", err))
	}
	defer f.Close()

	e := yaml.NewEncoder(f)
	defer e.Close()
	events := r.perioder.Events()
	err = e.Encode(events)
	if err != nil {
		r.Log.Error(fmt.Sprintf("Error endcoding to 'events.yaml': %v", err))
	}

	r.stop()
}
