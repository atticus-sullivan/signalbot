package tv

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules/tv/internal/show"
	"signalbot_go/signaldbus"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

type Tv struct {
	log         *slog.Logger  `yaml:"-"`
	ConfigDir   string        `yaml:"-"`
	SenderOrder []string      `yaml:"senderOrder"`
	Location    string        `yaml:"location"`
	Timeout     time.Duration `yaml:"timeout"`
	loc         *time.Location
	fetcher     *Fetcher
}

func NewTv(log *slog.Logger, cfgDir string) (*Tv, error) {
	t := Tv{
		log:       log,
		ConfigDir: cfgDir,
	}

	f, err := os.Open(filepath.Join(t.ConfigDir, "tv.yaml"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.KnownFields(true)
	err = d.Decode(&t)
	if err != nil {
		return nil, err
	}

	t.loc, err = time.LoadLocation(t.Location)
	if err != nil {
		return nil, err
	}

	t.fetcher = NewFetcher(t.log.With(), t.loc, t.Timeout)

	// validation
	if err := t.Validate(); err != nil {
		return nil, err
	}

	return &t, nil
}

func (t *Tv) Validate() error {
	if t.Timeout == time.Duration(0) {
		return fmt.Errorf("Invalid timeout, cannot be 0")
	}
	return nil
}

func (t *Tv) sendError(m *signaldbus.Message, signal signalsender.SignalSender, reply string) {
	if _, err := signal.Respond(reply, nil, m); err != nil {
		t.log.Error(fmt.Sprintf("Error responding to %v", m))
	}
}

type Args struct {
	When string `arg:"positional"`
	Post uint   `arg:"-p,--post" default:"1"`
}

func (t *Tv) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	var args Args
	parser, err := arg.NewParser(arg.Config{}, &args)

	if err != nil {
		t.log.Error(fmt.Sprintf("tv module: newParser -> %v", err))
		return
	}

	vargs, err := cmdsplit.Split(m.Message)
	if err != nil {
		errMsg := fmt.Sprintf("tv module: Error on parsing message: %v", err)
		t.log.Error(errMsg)
		t.sendError(m, signal, errMsg)
		return
	}

	err = parser.Parse(vargs)

	if err != nil {
		switch err {
		case arg.ErrVersion:
			// not implemented
			errMsg := fmt.Sprintf("tv module: Error: %v", "Version is not implemented")
			t.log.Error(errMsg)
			t.sendError(m, signal, errMsg)
			return
		case arg.ErrHelp:
			buf := new(bytes.Buffer)
			parser.WriteHelp(buf)

			if b, err := io.ReadAll(buf); err != nil {
				errMsg := fmt.Sprintf("Error: %v", err)
				t.log.Error(errMsg)
				t.sendError(m, signal, errMsg)
				return
			} else {
				errMsg := string(b)
				t.log.Info(fmt.Sprintf("tv module: Error: %v", err))
				t.sendError(m, signal, errMsg)
				return
			}
		default:
			errMsg := fmt.Sprintf("Error: %v", err)
			t.log.Error(errMsg)
			t.sendError(m, signal, errMsg)
			return
		}
	} else {
		target := time.Now()

		switch args.When {
		case "prime":
			target = time.Date(target.Year(), target.Month(), target.Day(), 20, 15, 0, 0, t.loc)
		default:
			errMsg := fmt.Sprintf("tv module: Error: %v", "invalid 'when' parameter")
			t.log.Error(errMsg)
			t.sendError(m, signal, errMsg)
			return
		}

		out, err := t.format(target, args.Post)
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			t.log.Error(errMsg)
			t.sendError(m, signal, errMsg)
			return
		}

		_, err = signal.Respond(out, []string{}, m)
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			t.log.Error(errMsg)
			t.sendError(m, signal, errMsg)
			return
		}
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
	}
	return builder.String(), nil
}

func (t *Tv) Start(virtRcv func(*signaldbus.Message)) error {
	return nil
}

func (t *Tv) Close(virtRcv func(*signaldbus.Message)) {
}
