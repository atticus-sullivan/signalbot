package periodic

import (
	"bytes"
	"context"
	"fmt"
	"io"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/internal/perioder"
	"signalbot_go/internal/signalsender"
	"signalbot_go/signaldbus"
	"time"

	"github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
)

type Periodic struct {
	Log       *slog.Logger                          `yaml:"-"`
	ConfigDir string                                `yaml:"-"`
	perioder  perioder.Perioder[signaldbus.Message] `yaml:"-"`
	stop      context.CancelFunc                    `yaml:"-"`
}

func NewPeriodic(log *slog.Logger, cfgDir string) (*Periodic, error) {
	p := Periodic{
		Log:       log,
		ConfigDir: cfgDir,
		perioder:  perioder.NewPerioderImpl[signaldbus.Message](log.With()),
	}
	return &p, nil
}

func (p *Periodic) Validate() error {
	return nil
}

// shortcut for sending an error via signal. If this fails log error.
func (p *Periodic) sendError(m *signaldbus.Message, signal signalsender.SignalSender, reply string) {
	if _, err := signal.Respond(reply, nil, m); err != nil {
		p.Log.Error(fmt.Sprintf("Error responding to %v", m))
	}
}

type Args struct {
	Add *addArgs `arg:"subcommand:add"`
	Ls  *lsArgs  `arg:"subcommand:list"`
	Rm  *rmArgs  `arg:"subcommand:remove"`
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
func (p *Periodic) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	var args Args
	parser, err := arg.NewParser(arg.Config{}, &args)

	if err != nil {
		return
	}

	vargs, err := cmdsplit.Split(m.Message)
	if err != nil {
		errMsg := fmt.Sprintf("periodic module: Error on parsing message: %v", err)
		p.Log.Error(errMsg)
		p.sendError(m, signal, errMsg)
		return
	}

	err = parser.Parse(vargs)

	if err != nil {
		switch err {
		case arg.ErrVersion:
			// not implemented
			errMsg := fmt.Sprintf("periodic module: Error: %v", "Version is not implemented")
			p.Log.Error(errMsg)
			p.sendError(m, signal, errMsg)
			return
		case arg.ErrHelp:
			buf := new(bytes.Buffer)
			parser.WriteHelp(buf)

			if b, err := io.ReadAll(buf); err != nil {
				errMsg := fmt.Sprintf("Error: %v", err)
				p.Log.Error(errMsg)
				p.sendError(m, signal, errMsg)
				return
			} else {
				errMsg := fmt.Sprintf("Error: %v", string(b))
				p.Log.Info(fmt.Sprintf("periodic module: Error: %v", err))
				p.sendError(m, signal, errMsg)
				return
			}
		default:
			errMsg := fmt.Sprintf("Error: %v", err)
			p.Log.Error(errMsg)
			p.sendError(m, signal, errMsg)
			return
		}
	} else {
		switch {
		case args.Add != nil:
			if args.Add.Desc == "" {
				args.Add.Desc = args.Add.Msg
			}
			args.Add.Every += time.Duration(24*time.Hour) * time.Duration(args.Add.EveryD)
			args.Add.EveryD = 0
			if args.Add.Start.IsZero() {
				args.Add.Start = time.Now()
			}
			p.Add(args.Add, *m, signal, virtRcv)
		case args.Ls != nil:
			p.Ls(args.Ls, *m, signal, virtRcv)
		case args.Rm != nil:
			p.Rm(args.Rm, *m, signal, virtRcv)
		}
	}
}

func (p *Periodic) Start() error {
	var ctx context.Context
	ctx, p.stop = context.WithCancel(context.Background())
	p.perioder.Start(ctx)
	return nil
}

func (p *Periodic) Close() {
	p.stop()
}

func (p *Periodic) Add(add *addArgs, m signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	if add.Every == time.Duration(0) {
		errMsg := fmt.Sprintf("periodic module: Invalid duration: %v", add.Every)
		p.Log.Info(errMsg)
		p.sendError(&m, signal, errMsg)
		return
	}
	var err error
	m.Message, err = cmdsplit.Unescape(add.Msg)
	if err != nil {
		errMsg := fmt.Sprintf("periodic module: Error on unescaping message: %v", err)
		p.Log.Error(errMsg)
		p.sendError(&m, signal, errMsg)
		return
	}
	var event perioder.ReocEvent[signaldbus.Message]
	if add.Until.IsZero(){
		event = perioder.NewReocEventImpl(add.Start, add.Every, add.Desc, &m, func(time time.Time) {
			mCopy := m
			virtRcv(&mCopy)
		})
	} else {
		event = perioder.NewReocEventImplDeadline(add.Start, add.Every, add.Until, add.Desc, &m, func(time time.Time) {
			mCopy := m
			virtRcv(&mCopy)
		})
	}
	p.perioder.Add(event)
	if _, err := signal.Respond(fmt.Sprintf("Added %v\n", event.String()), nil, &m); err != nil {
		p.Log.Error(fmt.Sprintf("periodic module: error sending add success msg: %v", err))
	}
}

func (p *Periodic) Ls(ls *lsArgs, m signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	eventsAll := p.perioder.Events()
	events := make(map[uint]perioder.ReocEvent[signaldbus.Message])
	for k, v := range eventsAll {
		if v.Metadata().Sender == m.Sender {
			events[k] = v
		}
	}
	if _, err := signal.Respond(fmt.Sprintf("%v\n", events), nil, &m); err != nil {
		p.Log.Error(fmt.Sprintf("periodic module: error sending ls output: %v", err))
	}
}

func (p *Periodic) Rm(rm *rmArgs, m signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	event, ok := p.perioder.Events()[rm.Id]
	if !ok || event.Metadata().Sender != m.Sender {
		errMsg := fmt.Sprintf("periodic module: Error: Event with ID %d does not exist or you don't added this event", rm.Id)
		p.Log.Error(errMsg)
		p.sendError(&m, signal, errMsg)
		return
	}
	p.Log.Info(fmt.Sprintf("periodic module: canceling event with ID: %d (%s)", rm.Id, event.String()))
	event.Cancel()
	if _, err := signal.Respond(fmt.Sprintf("Removed %v\n", event.String()), nil, &m); err != nil {
		p.Log.Error(fmt.Sprintf("periodic module: error sending rm success msg: %v", err))
	}
}
