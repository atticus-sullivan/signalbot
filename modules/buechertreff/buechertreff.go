package buechertreff

import (
	"fmt"
	"os"
	"path/filepath"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules"
	"signalbot_go/signaldbus"
	"sort"
	"strings"

	"github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

// buechertreff module. Should be instanciated with `NewBuechertreff`.
// data members are only global to be able to unmarshal them
type Buechertreff struct {
	modules.Module
	Series  map[string]string `yaml:"series"`
	fetcher Fetcher           `yaml:"-"`
}

// instanciates a new Buechertreff from a configuration file
// (cfgDir/buechertreff.yaml)
func NewBuechertreff(log *slog.Logger, cfgDir string) (*Buechertreff, error) {
	r := Buechertreff{
		Module:  modules.NewModule(log, cfgDir),
		fetcher: Fetcher{},
	}

	f, err := os.Open(filepath.Join(r.ConfigDir, "buechertreff.yaml"))
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

	// validation
	if err := r.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

// validates the buechertreff struct
func (r *Buechertreff) Validate() error {
	// validate the generic module first
	if err := r.Module.Validate(); err != nil {
		return err
	}
	return nil
}

// specifies the arguments when handling a request to this module
type Args struct {
	Which  string `arg:"positional"`
	Insert string `arg:"-i,--insert"`
}

// Handle a message from the signaldbus. Parses the message, executes the query
// and responds to signal.
func (r *Buechertreff) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
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

	if args.Insert != "" {
		r.Series[args.Which] = args.Insert
	}

	url, ok := r.Series[args.Which]
	if !ok {
		errMsg := fmt.Sprintf("Error: %v is unknown", args.Which)
		r.Log.Error(errMsg)
		builder := strings.Builder{}
		builder.WriteString(errMsg)
		builder.WriteRune('\n')
		builder.WriteString("Available series: ")
		sorted := make(sort.StringSlice, 0, len(r.Series))
		for k := range r.Series {
			sorted = append(sorted, k)
		}
		sorted.Sort()
		builder.WriteString(strings.Join(sorted, ", "))
		r.SendError(m, signal, builder.String())
		return
	}

	// execute the query
	reader, err := r.fetcher.getReader(url)
	if !ok {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
	defer reader.Close()
	items, err := r.fetcher.getFromReader(reader)
	if !ok {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	// respond
	_, err = signal.Respond(items.String(), []string{}, m, true)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
}

// save config file in case something has changed (module allows to add new
// series during runtime)
func (r *Buechertreff) Close(virtRcv func(*signaldbus.Message)) {
	r.Module.Close(virtRcv)

	f, err := os.Create(filepath.Join(r.ConfigDir, "buechertreff.yaml"))
	if err != nil {
		r.Log.Error(fmt.Sprintf("Error opening 'buechertreff.yaml': %v", err))
	}
	defer f.Close()

	e := yaml.NewEncoder(f)
	defer e.Close()
	err = e.Encode(r)
	if err != nil {
		r.Log.Error(fmt.Sprintf("Error endcoding to 'buechertreff.yaml': %v", err))
	}
}
