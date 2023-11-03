package periodic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/internal/perioder"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules"
	"signalbot_go/signaldbus"
	"time"

	"github.com/alexflint/go-arg"
	"log/slog"
	"gopkg.in/yaml.v3"
)

type Periodic struct {
	modules.Module
	perioder perioder.Perioder[signaldbus.Message] `yaml:"-"`
	stop     context.CancelFunc                    `yaml:"-"`
}

func NewPeriodic(log *slog.Logger, cfgDir string) (*Periodic, error) {
	r := Periodic{
		Module:   modules.NewModule(log, cfgDir),
		perioder: perioder.NewPerioderImpl[signaldbus.Message](log.With()),
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
func (r *Periodic) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
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

func (r *Periodic) Add(add *addArgs, m signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
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
	var event perioder.ReocEvent[signaldbus.Message]
	if add.Until.IsZero() {
		event = perioder.NewReocEventImpl(add.Start, add.Every, add.Desc, m, func(time time.Time, event perioder.ReocEvent[signaldbus.Message]) {
			meta := event.Metadata()
			virtRcv(&meta)
		})
	} else {
		event = perioder.NewReocEventImplDeadline(add.Start, add.Every, add.Until, add.Desc, m, func(time time.Time, event perioder.ReocEvent[signaldbus.Message]) {
			meta := event.Metadata()
			virtRcv(&meta)
		})
	}
	r.perioder.Add(event)
	if _, err := signal.Respond(fmt.Sprintf("Added %v\n", event.String()), nil, &m, true); err != nil {
		r.Log.Error(fmt.Sprintf("error sending add success msg: %v", err))
	}
}

func (r *Periodic) Ls(ls *lsArgs, m signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	eventsAll := r.perioder.Events()
	events := make(map[uint]perioder.ReocEvent[signaldbus.Message])
	for k, v := range eventsAll {
		if v.Metadata().Sender == m.Sender {
			events[k] = v
		}
	}
	if _, err := signal.Respond(fmt.Sprintf("%v\n", events), nil, &m, true); err != nil {
		r.Log.Error(fmt.Sprintf("error sending ls output: %v", err))
	}
}

func (r *Periodic) Rm(rm *rmArgs, m signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
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

func (r *Periodic) Start(virtRcv func(*signaldbus.Message)) error {
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
		events := make(map[uint]perioder.ReocEventImplDeadline[signaldbus.Message])
		err = d.Decode(&events)
		if err != nil {
			// p.Log.Error(fmt.Sprintf("Error decoding to 'events.yaml': %v", err))
			return err
		}
		// add events
		for _, vIter := range events {
			v := vIter // force copy
			v.Foo = func(time time.Time, event perioder.ReocEvent[signaldbus.Message]) {
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

func (r *Periodic) Close(virtRcv func(*signaldbus.Message)) {
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
