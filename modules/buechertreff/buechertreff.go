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

type Buechertreff struct {
	modules.Module
	Series  map[string]string `yaml:"series"`
	fetcher Fetcher           `yaml:"-"`
}

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
	if err := r.Module.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *Buechertreff) Validate() error {
	return nil
}

type Args struct {
	Which  string `arg:"positional"`
	Insert string `arg:"-i,--insert"`
}

func (r *Buechertreff) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
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

	items, err := r.fetcher.get(url)
	if !ok {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	_, err = signal.Respond(items.String(), []string{}, m, true)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
	}
}

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
