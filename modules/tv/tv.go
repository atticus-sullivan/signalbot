package tv

import (
	"fmt"
	"os"
	"path/filepath"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules"
	"signalbot_go/modules/tv/internal/show"
	"signalbot_go/signaldbus"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

type Tv struct {
	modules.Module
	SenderOrder []string      `yaml:"senderOrder"`
	Location    string        `yaml:"location"`
	Timeout     time.Duration `yaml:"timeout"`
	loc         *time.Location
	fetcher     *Fetcher
}

func NewTv(log *slog.Logger, cfgDir string) (*Tv, error) {
	r := Tv{
		Module: modules.NewModule(log, cfgDir),
	}

	f, err := os.Open(filepath.Join(r.ConfigDir, "tv.yaml"))
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

	r.loc, err = time.LoadLocation(r.Location)
	if err != nil {
		return nil, err
	}

	r.fetcher = NewFetcher(r.Log, r.loc, r.Timeout)

	// validation
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if err := r.Module.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (t *Tv) Validate() error {
	if t.Timeout == time.Duration(0) {
		return fmt.Errorf("Invalid timeout, cannot be 0")
	}
	return nil
}

type Args struct {
	When string `arg:"positional"`
	Post uint   `arg:"-p,--post" default:"1"`
}

func (r *Tv) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
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

	target := time.Now()

	switch args.When {
	case "prime":
		target = time.Date(target.Year(), target.Month(), target.Day(), 20, 15, 0, 0, r.loc)
	default:
		errMsg := "invalid 'when' parameter"
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	out, err := r.format(target, args.Post)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	_, err = signal.Respond(out, []string{}, m, true)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
}

func (t *Tv) format(target time.Time, postOrig uint) (string, error) {
	g := t.fetcher.Get()

	builder := strings.Builder{}

	first := true
	for _, sender := range t.SenderOrder {
		shows, ok := g[sender]
		if !ok {
			continue
		}
		if !first {
			builder.WriteRune('\n')
		} else {
			first = false
		}
		builder.WriteString(sender)
		builder.WriteRune('\n')
		var last *show.Show = nil
		post := postOrig
		for _, s := range shows {
			if post == 0 {
				break
			}
			if s.Time.Compare(target) == +1 {
				builder.WriteString(last.String())
				builder.WriteRune('\n')
				post--
			}
			sNew := s
			last = &sNew
		}
		if post != 0 {
			builder.WriteString(last.String())
			builder.WriteRune('\n')
		}
	}
	return builder.String(), nil
}
