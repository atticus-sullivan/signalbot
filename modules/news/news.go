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

type News struct {
	modules.Module
	fetcher      Fetcher                                     `yaml:"-"`
	LastBreaking differ.Differ[string, string, breakingResp] `yaml:"lastBreaking"` // stores last chat->user->sending
}

func NewNews(log *slog.Logger, cfgDir string) (*News, error) {
	r := News{
		Module:       modules.NewModule(log, cfgDir),
		fetcher:      Fetcher{},
		LastBreaking: make(differ.Differ[string, string, breakingResp]),
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
	if err := r.Module.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *News) Validate() error {
	return nil
}

type Args struct {
	News     *struct{} `arg:"subcommand:news"`
	Breaking *struct {
		Diff bool `arg:"--diff" default:"false"`
	} `arg:"subcommand:breaking"`
	Quiet bool `arg:"-q,--quiet" default:"false"`
}

func (r *News) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
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

	var resp string
	switch {
	case args.Breaking != nil:
		b, err := r.fetcher.getBreaking()
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return
		}

		if args.Breaking.Diff {
			resp = r.LastBreaking.DiffStore(m.Chat, m.Sender, b)
		} else {
			resp = b.String()
		}

	default: // args.News != nil:
		b, err := r.fetcher.getNews()
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return
		}
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
