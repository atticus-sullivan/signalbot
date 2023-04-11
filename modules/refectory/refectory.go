package refectory

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/internal/signalsender"
	"signalbot_go/signaldbus"
	"time"

	"github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

type Refectory struct {
	log         *slog.Logger        `yaml:"-"`
	fetcher     Fetcher             `yaml:"-"`
	ConfigDir   string              `yaml:"-"`
	Refectories map[string]uint     `yaml:"refectories"`
	Aliases     map[string][]string `yaml:"aliases"`
}

func NewRefectory(log *slog.Logger, cfgDir string) (*Refectory, error) {
	r := Refectory{
		log:       log,
		ConfigDir: cfgDir,
		fetcher: Fetcher{
			log: log,
		},
	}

	f, err := os.Open(filepath.Join(r.ConfigDir, "refectory.yaml"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.KnownFields(true)
	err = d.Decode(&r)
	if err != nil {
		return nil, err
	}

	for ref := range r.Refectories {
		r.Aliases[ref] = []string{ref}
	}

	// validation
	if err := r.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *Refectory) Validate() error {
	for alias, resolvedL := range r.Aliases {
		for _, resolved := range resolvedL {
			if _, ok := r.Refectories[resolved]; !ok {
				return fmt.Errorf("%s is an invalid refectory, alias %s is resolved to", resolved, alias)
			}
		}
	}
	return nil
}

func (r *Refectory) sendError(m *signaldbus.Message, signal signalsender.SignalSender, reply string) {
	if _, err := signal.Respond(reply, nil, m, false); err != nil {
		r.log.Error(fmt.Sprintf("Error responding to %v", m))
	}
}

type Args struct {
	Where string `arg:"positional"`
	When  int    `arg:"-d,--day" default:"0"`
	Quiet  bool    `arg:"-q,--quiet" default:"false"`
}

func (r *Refectory) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	var args Args
	parser, err := arg.NewParser(arg.Config{}, &args)

	if err != nil {
		r.log.Error(fmt.Sprintf("periodic module: newParser -> %v", err))
		return
	}

	vargs, err := cmdsplit.Split(m.Message)
	if err != nil {
		errMsg := fmt.Sprintf("periodic module: Error on parsing message: %v", err)
		r.log.Error(errMsg)
		r.sendError(m, signal, errMsg)
		return
	}

	err = parser.Parse(vargs)

	if err != nil {
		switch err {
		case arg.ErrVersion:
			// not implemented
			errMsg := fmt.Sprintf("periodic module: Error: %v", "Version is not implemented")
			r.log.Error(errMsg)
			r.sendError(m, signal, errMsg)
			return
		case arg.ErrHelp:
			buf := new(bytes.Buffer)
			parser.WriteHelp(buf)

			if b, err := io.ReadAll(buf); err != nil {
				errMsg := fmt.Sprintf("Error: %v", err)
				r.log.Error(errMsg)
				r.sendError(m, signal, errMsg)
				return
			} else {
				errMsg := string(b)
				r.log.Info(fmt.Sprintf("periodic module: Error: %v", err))
				r.sendError(m, signal, errMsg)
				return
			}
		default:
			errMsg := fmt.Sprintf("Error: %v", err)
			r.log.Error(errMsg)
			r.sendError(m, signal, errMsg)
			return
		}
	} else {
		date := time.Now().Add(time.Hour * 24 * time.Duration(args.When))

		resolvedL, ok := r.Aliases[args.Where]
		if !ok {
			errMsg := fmt.Sprintf("Error: %v is unknown", args.Where)
			r.log.Error(errMsg)
			r.sendError(m, signal, errMsg)
		}

		for _, ref := range resolvedL {
			menuS, err := r.fetcher.getMenuString(ref, r.Refectories[ref], date)
			if err != nil {
				var errMsg string
				if err == NotOpenThatDay {
					errMsg = err.Error()
				} else {
					errMsg = fmt.Sprintf("Error: %v", err)
				}
				r.log.Error(errMsg)
				if !args.Quiet || err != NotOpenThatDay {
					r.sendError(m, signal, errMsg)
				}
				continue
			}
			_, err = signal.Respond(menuS, []string{}, m, true)
			if err != nil {
				errMsg := fmt.Sprintf("Error: %v", err)
				r.log.Error(errMsg)
				r.sendError(m, signal, errMsg)
			}
		}
	}
}

func (r *Refectory) Start(virtRcv func(*signaldbus.Message)) error {
	return nil
}

func (r *Refectory) Close(virtRcv func(*signaldbus.Message)) {
}
