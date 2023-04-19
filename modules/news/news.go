package news

import (
	"fmt"
	"os"
	"path/filepath"
	"signalbot_go/internal/differ"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules"
	"signalbot_go/signaldbus"

	"github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

// news module. Should be instanciated with `NewNews`.
// data members are only global to be able to unmarshal them
type News struct {
	modules.Module
	fetcher      Fetcher                                 `yaml:"-"`
	LastBreaking differ.Differ[string, string, breaking] `yaml:"lastBreaking"` // stores last chat->user->sending
}

// instanciates a new News from a configuration file
// (cfgDir/news.yaml)
func NewNews(log *slog.Logger, cfgDir string) (*News, error) {
	r := News{
		Module:       modules.NewModule(log, cfgDir),
		fetcher:      Fetcher{},
		LastBreaking: make(differ.Differ[string, string, breaking]),
	}

	f, err := os.Open(filepath.Join(r.ConfigDir, "news.yaml"))
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

// can be used to validate the news struct
func (r *News) Validate() error {
	// validate the generic module first
	if err := r.Module.Validate(); err != nil {
		return err
	}
	return nil
}

// specifies the arguments when handling a request to this module
type Args struct {
	News     *struct{} `arg:"subcommand:news"`
	Breaking *struct {
		Diff bool `arg:"--diff" default:"false"`
	} `arg:"subcommand:breaking"`
	Quiet bool `arg:"-q,--quiet" default:"false"`
}

// Handle a message from the signaldbus. Parses the message, executes the query
// and responds to signal.
func (r *News) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
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

	// execute the query
	var resp string
	switch {
	case args.Breaking != nil:
		reader, err := r.fetcher.getBreakingReader()
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return
		}
		b, err := r.fetcher.getBreakingFromReader(reader)
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return
		}

		// respond
		if args.Breaking.Diff {
			resp = r.LastBreaking.DiffStore(m.Chat, m.Sender, b)
		} else {
			resp = b.String()
		}

	default: // args.News != nil:
		reader, err := r.fetcher.getNewsReader()
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return
		}
		b, err := r.fetcher.getNewsFromReader(reader)
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return
		}
		// respond
		resp = b.String()
	}

	if resp == "" {
		if args.Quiet {
			return
		} else {
			resp = "No (new) news"
		}
	}

	_, err = signal.Respond(resp, []string{}, m, true)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
}

// save config file in case something has changed (last with the differ might
// have changed)
func (r *News) Close(virtRcv func(*signaldbus.Message)) {
	r.Module.Close(virtRcv)

	f, err := os.Create(filepath.Join(r.ConfigDir, "news.yaml"))
	if err != nil {
		r.Log.Error(fmt.Sprintf("Error opening 'news.yaml': %v", err))
	}
	defer f.Close()

	e := yaml.NewEncoder(f)
	defer e.Close()
	err = e.Encode(r)
	if err != nil {
		r.Log.Error(fmt.Sprintf("Error endcoding to 'news.yaml': %v", err))
	}
}
