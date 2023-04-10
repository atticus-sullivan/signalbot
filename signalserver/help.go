package signalserver

import (
	"encoding/hex"
	"fmt"
	"signalbot_go/internal/signalsender"
	"signalbot_go/signaldbus"
	"strings"

	"golang.org/x/exp/slog"
)

type Help struct {
	log         *slog.Logger        `yaml:"-"`
	ConfigDir   string              `yaml:"-"`
	Handlers map[string]HandlerCfg `yaml:"-"`
	self              string `yaml:"-"`
}

func NewHelp(log *slog.Logger, cfgDir string, handlers map[string]HandlerCfg, self string) (*Help, error) {
	r := Help{
		log:       log,
		ConfigDir: cfgDir,
		Handlers: handlers,
		self: self,
	}

	// f, err := os.Open(filepath.Join(r.ConfigDir, "help.yaml"))
	// if err != nil {
	// 	return nil, err
	// }
	// defer f.Close()
	//
	// d := yaml.NewDecoder(f)
	// d.KnownFields(true)
	// err = d.Decode(&r)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// // validation
	// if err := r.Validate(); err != nil {
	// 	return nil, err
	// }
	//
	return &r, nil
}

func (r *Help) Validate() error {
	return nil
}

func (r *Help) sendError(m *signaldbus.Message, signal signalsender.SignalSender, reply string) {
	if _, err := signal.Respond(reply, nil, m); err != nil {
		r.log.Error(fmt.Sprintf("Error responding to %v", m))
	}
}

func (r *Help) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	var err error
	builder := strings.Builder{}
	for _,handler := range r.Handlers {
		chat := getChatId(m, r.self)
		if len(m.GroupId) > 0 {
			chat = hex.EncodeToString(m.GroupId)
		}
		if err := handler.Access.Check(m.Sender, chat); err != nil {
			continue
		}
		builder.WriteString(strings.Join(handler.Prefixes, ","))
		builder.WriteString("\n    ")
		builder.WriteString(handler.Help)
		builder.WriteRune('\n')
	}

	_, err = signal.Respond(builder.String(), []string{}, m)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.log.Error(errMsg)
		r.sendError(m, signal, errMsg)
	}
}

func (r *Help) Start(virtRcv func(*signaldbus.Message)) error {
	return nil
}

func (r *Help) Close(virtRcv func(*signaldbus.Message)) {
}