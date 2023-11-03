package hugendubel

import (
	"fmt"
	"os"
	"path/filepath"
	"signalbot_go/internal/differ"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules"
	"signalbot_go/signaldbus"
	"sort"
	"strings"

	"github.com/alexflint/go-arg"
	"log/slog"
	"gopkg.in/yaml.v3"
)

// hugendubel module. Should be instanciated with `NewHugendubel`.
// data members are only global to be able to unmarshal them
type Hugendubel struct {
	modules.Module

	fetcher *Fetcher                             `yaml:"-"`
	Queries map[string]query                    `yaml:"queries"`
	Aliases map[string][]string                 `yaml:"aliases"`
	Lasts   differ.Differ[string, string, book] `yaml:"lasts"` // stores last chat->user->sending
	QuerySize uint `yaml:"querySize"`
}

// instanciates a new Hugendubel from a configuration file
// (cfgDir/hugendubel.yaml)
func NewHugendubel(log *slog.Logger, cfgDir string) (*Hugendubel, error) {
	r := Hugendubel{
		Module: modules.NewModule(log, cfgDir),
		Lasts:  make(differ.Differ[string, string, book]),
	}

	f, err := os.Open(filepath.Join(r.ConfigDir, "hugendubel.yaml"))
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

	if r.Aliases == nil {
		r.Aliases = make(map[string][]string)
	}
	if r.Queries == nil {
		r.Queries = make(map[string]query)
	}

	all := make([]string, 0, len(r.Queries))
	for s := range r.Queries {
		r.Aliases[s] = []string{s}
		all = append(all, s)
	}
	r.Aliases["all"] = all

	if r.QuerySize <= 20 {
		r.fetcher = NewFetcher(log, 20)
	} else {
		r.fetcher = NewFetcher(log, r.QuerySize)
	}

	// validation
	if err := r.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

// validate a hugendubel struct
func (r *Hugendubel) Validate() error {
	// validate the generic module first
	if err := r.Module.Validate(); err != nil {
		return err
	}
	for alias, resolvedL := range r.Aliases {
		for _, resolved := range resolvedL {
			if _, ok := r.Queries[resolved]; !ok {
				return fmt.Errorf("%s is an invalid series, alias %s is resolved to", resolved, alias)
			}
		}
	}
	return nil
}

// specifies the arguments when handling a request to this module
type Args struct {
	Which string `arg:"positional"`
	Quiet bool   `arg:"-q,--quiet" default:"false"`
	Diff  bool   `arg:"--diff" default:"false"`
}

// Handle a message from the signaldbus. Parses the message, executes the query
// and responds to signal.
func (r *Hugendubel) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	// parse the message
	var args Args
	parser, err := arg.NewParser(arg.Config{}, &args)
	if err != nil {
		r.Log.Error(fmt.Sprintf("newParser -> %v", err))
		return
	}

	if err := r.Module.Handle(m, signal, virtRcv, parser); err != nil {
		if err == arg.ErrHelp {
			return
		}
		errMsg := err.Error()
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	queries := make([]query, 0)
	chat := m.Chat
	if args.Which == "all" {
		for _, q := range r.Queries {
			queries = append(queries, q)
		}
		chat += "L" // different diffing for "all" command
	} else {
		resolvedL, ok := r.Aliases[args.Which]
		if !ok {
			errMsg := fmt.Sprintf("Error: %v is unknown", args.Which)
			r.Log.Error(errMsg)
			builder := strings.Builder{}
			builder.WriteString(errMsg)
			builder.WriteRune('\n')
			builder.WriteString("Available queries: ")
			sorted := make(sort.StringSlice, 0, len(r.Aliases))
			for k := range r.Aliases {
				sorted = append(sorted, k)
			}
			sorted.Sort()
			builder.WriteString(strings.Join(sorted, ", "))
			r.SendError(m, signal, builder.String())
			return
		}
		for _, re := range resolvedL {
			queries = append(queries, r.Queries[re])
		}
	}

	// execute the query
	items, err := r.fetcher.get(queries)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	// respond
	resp := strings.Builder{}
	if !args.Diff {
		resp.WriteString(items.String())
	} else {
		d := r.Lasts.DiffStore(chat, m.Sender, items)
		if d != "" {
			resp.WriteString("Diff:\n")
			resp.WriteString(d)
		}
	}
	respS := resp.String()

	if respS == "" {
		if args.Quiet {
			return
		}
		respS = "No data/changes"
	}

	_, err = signal.Respond(respS, []string{}, m, true)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
}

// save config file in case something has changed (module allows to add new
// series during runtime)
func (r *Hugendubel) Close(virtRcv func(*signaldbus.Message)) {
	r.Module.Close(virtRcv)

	delete(r.Aliases, "all") // "all" alias is always a generated one
	f, err := os.Create(filepath.Join(r.ConfigDir, "hugendubel.yaml"))
	if err != nil {
		r.Log.Error(fmt.Sprintf("Error opening 'hugendubel.yaml': %v", err))
	}
	defer f.Close()

	e := yaml.NewEncoder(f)
	defer e.Close()
	err = e.Encode(r)
	if err != nil {
		r.Log.Error(fmt.Sprintf("Error endcoding to 'hugendubel.yaml': %v", err))
	}
}
