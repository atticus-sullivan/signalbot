package refectory

import (
	"fmt"
	"os"
	"path/filepath"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules"
	"signalbot_go/signaldbus"
	"time"

	"github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

type Refectory struct {
	modules.Module
	fetcher     Fetcher             `yaml:"-"`
	Refectories map[string]uint     `yaml:"refectories"`
	Aliases     map[string][]string `yaml:"aliases"`
}

func NewRefectory(log *slog.Logger, cfgDir string) (*Refectory, error) {
	r := Refectory{
		Module: modules.NewModule(log, cfgDir),
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
	if err := r.Module.Validate(); err != nil {
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

type Args struct {
	Where string `arg:"positional"`
	When  int    `arg:"-d,--day" default:"0"`
	Quiet bool   `arg:"-q,--quiet" default:"false"`
}

func (r *Refectory) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
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

	date := time.Now().Add(time.Hour * 24 * time.Duration(args.When))

	resolvedL, ok := r.Aliases[args.Where]
	if !ok {
		errMsg := fmt.Sprintf("Error: %v is unknown", args.Where)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
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
			r.Log.Error(errMsg)
			if !args.Quiet || err != NotOpenThatDay {
				r.SendError(m, signal, errMsg)
			}
			continue
		}
		_, err = signal.Respond(menuS, []string{}, m, true)
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
		}
	}
}
