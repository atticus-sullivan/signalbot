package buechertreff

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/internal/signalsender"
	"signalbot_go/signaldbus"
	"sort"
	"strings"

	"github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)


type Buechertreff struct {
	log         *slog.Logger        `yaml:"-"`
	ConfigDir   string              `yaml:"-"`
	Series      map[string]string     `yaml:"series"`
}

func NewBuechertreff(log *slog.Logger, cfgDir string) (*Buechertreff, error) {
	r := Buechertreff{
		log:       log,
		ConfigDir: cfgDir,
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

func (b *Buechertreff) Validate() error {
	return nil
}

func (b *Buechertreff) sendError(m *signaldbus.Message, signal signalsender.SignalSender, reply string) {
	if _, err := signal.Respond(reply, nil, m, false); err != nil {
		b.log.Error(fmt.Sprintf("Error responding to %v", m))
	}
}

type Args struct {
	Which  string `arg:"positional"`
	Insert string `arg:"-i,--insert"`
}

func (b *Buechertreff) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	var args Args
	parser, err := arg.NewParser(arg.Config{}, &args)

	if err != nil {
		b.log.Error(fmt.Sprintf("buechertreff module: newParser -> %v", err))
		return
	}

	vargs, err := cmdsplit.Split(m.Message)
	if err != nil {
		errMsg := fmt.Sprintf("buechertreff module: Error on parsing message: %v", err)
		b.log.Error(errMsg)
		b.sendError(m, signal, errMsg)
		return
	}

	err = parser.Parse(vargs)

	if err != nil {
		switch err {
		case arg.ErrVersion:
			// not implemented
			errMsg := fmt.Sprintf("buechertreff module: Error: %v", "Version is not implemented")
			b.log.Error(errMsg)
			b.sendError(m, signal, errMsg)
			return
		case arg.ErrHelp:
			buf := new(bytes.Buffer)
			parser.WriteHelp(buf)

			if h, err := io.ReadAll(buf); err != nil {
				errMsg := fmt.Sprintf("Error: %v", err)
				b.log.Error(errMsg)
				b.sendError(m, signal, errMsg)
				return
			} else {
				errMsg := string(h)
				b.log.Info(fmt.Sprintf("buechertreff module: Error: %v", err))
				b.sendError(m, signal, errMsg)
				return
			}
		default:
			errMsg := fmt.Sprintf("Error: %v", err)
			b.log.Error(errMsg)
			b.sendError(m, signal, errMsg)
			return
		}
	}
	if args.Insert != "" {
		b.Series[args.Which] = args.Insert
	}

	url,ok := b.Series[args.Which]
	if !ok {
		errMsg := fmt.Sprintf("Error: %v is unknown", args.Which)
		b.log.Error(errMsg)
		builder := strings.Builder{}
		builder.WriteString(errMsg)
		builder.WriteRune('\n')
		builder.WriteString("Available series: ")
		sorted := make(sort.StringSlice, 0, len(b.Series))
		for k := range b.Series {
			sorted = append(sorted, k)
		}
		sorted.Sort()
		builder.WriteString(strings.Join(sorted, ", "))
		b.sendError(m, signal, builder.String())
		return
	}

	items,err := Get(url)
	if !ok {
		errMsg := fmt.Sprintf("Error: %v", err)
		b.log.Error(errMsg)
		b.sendError(m, signal, errMsg)
		return
	}

	_, err = signal.Respond(items.String(), []string{}, m, true)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		b.log.Error(errMsg)
		b.sendError(m, signal, errMsg)
	}
}

func (b *Buechertreff) Start(virtRcv func(*signaldbus.Message)) error {
	return nil
}

func (b *Buechertreff) Close(virtRcv func(*signaldbus.Message)) {
	f, err := os.Create(filepath.Join(b.ConfigDir, "buechertreff.yaml"))
	if err != nil {
		b.log.Error(fmt.Sprintf("buechertreff module: Error opening 'buechertreff.yaml': %v", err))
	}
	e := yaml.NewEncoder(f)
	err = e.Encode(b)
	if err != nil {
		b.log.Error(fmt.Sprintf("buechertreff module: Error endcoding to 'buechertreff.yaml': %v", err))
	}
}
