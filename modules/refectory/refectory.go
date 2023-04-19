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

// Refectory module. Should be instanciated with `NewRefectory`.
// data members are only global to be able to unmarshal them
type Refectory struct {
	modules.Module
	fetcher     Fetcher             `yaml:"-"`
	Refectories map[string]uint     `yaml:"refectories"`
	Aliases     map[string][]string `yaml:"aliases"`
}

// instanciates a new Refectory from a configuration file
// (cfgDir/refectory.yaml)
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

	return &r, nil
}

// validates the refectory struct
func (r *Refectory) Validate() error {
	// validate the generic module first
	if err := r.Module.Validate(); err != nil {
		return err
	}
	for alias, resolvedL := range r.Aliases {
		for _, resolved := range resolvedL {
			if _, ok := r.Refectories[resolved]; !ok {
				return fmt.Errorf("%s is an invalid refectory, alias %s is resolved to", resolved, alias)
			}
		}
	}
	return nil
}

// specifies the arguments when handling a request to this module
type Args struct {
	Where string `arg:"positional"`
	When  int    `arg:"-d,--day" default:"0"`
	Quiet bool   `arg:"-q,--quiet" default:"false"`
}

// Handle a message from the signaldbus. Parses the message, executes the query
// and responds to signal.
func (r *Refectory) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	// parse the message
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
		return
	}

	for _, ref := range resolvedL {
		// execute the query
		reader, err := r.fetcher.getReader(r.Refectories[ref], date)
		if err != nil {
			var errMsg string
			if err == ErrNotOpenThatDay {
				errMsg = err.Error()
			} else {
				errMsg = fmt.Sprintf("Error: %v", err)
			}
			r.Log.Error(errMsg)
			if !args.Quiet || err != ErrNotOpenThatDay {
				r.SendError(m, signal, errMsg)
			}
			continue
		}
		defer reader.Close()
		menu, err := r.fetcher.getFromReader(reader)
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			continue
		}
		// respond
		menuS := fmt.Sprintf("%s on %s\n", ref, date.Format("2006-01-02")) + menu.String()
		_, err = signal.Respond(menuS, []string{}, m, true)
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			continue
		}
	}
}
