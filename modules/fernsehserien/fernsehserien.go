package fernsehserien

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/internal/differ"
	"signalbot_go/internal/signalsender"
	"signalbot_go/signaldbus"
	"sort"
	"strings"

	"github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

type Fernsehserien struct {
	log         *slog.Logger        `yaml:"-"`
	ConfigDir   string              `yaml:"-"`
	Series      map[string]string   `yaml:"series"`
	Aliases     map[string][]string `yaml:"aliases"`
	UnavailableSenders map[string]bool `yaml:"unavailableSenders"`
	Lasts differ.Differ[string,string,sending] `yaml:"lasts"` // stores last chat->user->sending
	getChat func(m *signaldbus.Message) string
}

func NewFernsehserien(log *slog.Logger, cfgDir string, getChat func(m *signaldbus.Message) string) (*Fernsehserien, error) {
	r := Fernsehserien{
		log:       log,
		ConfigDir: cfgDir,
		getChat: getChat,
		Lasts: make(differ.Differ[string,string,sending]),
	}

	f, err := os.Open(filepath.Join(r.ConfigDir, "fernsehserien.yaml"))
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

	all := make([]string, 0, len(r.Series))
	for s := range r.Series {
		r.Aliases[s] = []string{s}
		all = append(all, s)
	}
	r.Aliases["all"] = all

	// validation
	if err := r.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *Fernsehserien) Validate() error {
	for alias, resolvedL := range r.Aliases {
		for _, resolved := range resolvedL {
			if _, ok := r.Series[resolved]; !ok {
				return fmt.Errorf("%s is an invalid series, alias %s is resolved to", resolved, alias)
			}
		}
	}
	return nil
}

func (r *Fernsehserien) sendError(m *signaldbus.Message, signal signalsender.SignalSender, reply string) {
	if _, err := signal.Respond(reply, nil, m, false); err != nil {
		r.log.Error(fmt.Sprintf("Error responding to %v", m))
	}
}

type Args struct {
	Which  string `arg:"positional"`
	Insert string `arg:"-i,--insert"`
	Quiet  bool    `arg:"-q,--quiet" default:"false"`
	Diff   bool    `arg:"--diff" default:"false"`
	Data   bool    `arg:"-d,--data" default:"true"`
}

func (r *Fernsehserien) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	var args Args
	parser, err := arg.NewParser(arg.Config{}, &args)

	if err != nil {
		r.log.Error(fmt.Sprintf("newParser -> %v", err))
		return
	}

	vargs, err := cmdsplit.Split(m.Message)
	if err != nil {
		errMsg := fmt.Sprintf("Error on parsing message: %v", err)
		r.log.Error(errMsg)
		r.sendError(m, signal, errMsg)
		return
	}

	err = parser.Parse(vargs)

	if err != nil {
		switch err {
		case arg.ErrVersion:
			// not implemented
			errMsg := fmt.Sprintf("Error: %v", "Version is not implemented")
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
				r.log.Info(fmt.Sprintf("Error: %v", err))
				r.sendError(m, signal, errMsg)
				return
			}
		default:
			errMsg := fmt.Sprintf("Error: %v", err)
			r.log.Error(errMsg)
			r.sendError(m, signal, errMsg)
			return
		}
	}
	if args.Insert != "" {
		r.Series[args.Which] = args.Insert
		r.Aliases["all"] = append(r.Aliases["all"], args.Insert)
	}

	urls := make(map[string]string)
	chat := r.getChat(m)
	if args.Which == "all" {
		urls = r.Series
		chat += "L" // different diffing for "all" command
	} else {
		resolvedL, ok := r.Aliases[args.Which]
		if !ok {
			errMsg := fmt.Sprintf("Error: %v is unknown", args.Which)
			r.log.Error(errMsg)
			builder := strings.Builder{}
			builder.WriteString(errMsg)
			builder.WriteRune('\n')
			builder.WriteString("Available series: ")
			sorted := make(sort.StringSlice, 0, len(r.Aliases))
			for k := range r.Aliases {
				sorted = append(sorted, k)
			}
			sorted.Sort()
			builder.WriteString(strings.Join(sorted, ", "))
			r.sendError(m, signal, builder.String())
			return
		}
		for _,re := range resolvedL {
			urls[re] = r.Series[re]
		}
	}

	items,err := Get(urls, r.UnavailableSenders)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.log.Error(errMsg)
		r.sendError(m, signal, errMsg)
		return
	}

	
	resp := strings.Builder{}
	if args.Data {
		resp.WriteString(items.String())
	}
	if args.Data && args.Diff {
		resp.WriteRune('\n')
	}
	if args.Diff {
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
		r.log.Error(errMsg)
		r.sendError(m, signal, errMsg)
	}
}

func (r *Fernsehserien) Start(virtRcv func(*signaldbus.Message)) error {
	return nil
}

func (r *Fernsehserien) Close(virtRcv func(*signaldbus.Message)) {
	delete(r.Aliases, "all") // "all" alias is always a generated one
	f, err := os.Create(filepath.Join(r.ConfigDir, "fernsehserien.yaml"))
	if err != nil {
		r.log.Error(fmt.Sprintf("Error opening 'buechertreff.yaml': %v", err))
	}
	e := yaml.NewEncoder(f)
	err = e.Encode(r)
	if err != nil {
		r.log.Error(fmt.Sprintf("Error endcoding to 'buechertreff.yaml': %v", err))
	}
}
