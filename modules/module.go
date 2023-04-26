package modules

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/internal/signalsender"
	"signalbot_go/signaldbus"

	"github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
)

type Module struct {
	Log       *slog.Logger `yaml:"-"`
	ConfigDir string       `yaml:"-"`
}

func NewModule(log *slog.Logger, cfgDir string) Module {
	r := Module{
		Log:       log,
		ConfigDir: cfgDir,
	}

	return r
}

func (r *Module) Validate() error {
	return nil
}

// shortcut for sending an error via signal. If this fails log error.
func (r *Module) SendError(m *signaldbus.Message, signal signalsender.SignalSender, reply string) {
	if _, err := signal.Respond(reply, nil, m, false); err != nil {
		r.Log.Error(fmt.Sprintf("Error responding to %v", m))
	}
}

// returns whether an error ocurred (error is already logged and sent to the user)
func (r *Module) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message), parser *arg.Parser) error {

	vargs, err := cmdsplit.Split(m.Message)
	if err != nil {
		errMsg := fmt.Sprintf("Error on parsing message: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return err
	}

	err = parser.Parse(vargs)

	if err != nil {
		switch err {
		case arg.ErrVersion:
			// not implemented
			err := errors.New("Version is not implemented")
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return err
		case arg.ErrHelp:
			buf := new(bytes.Buffer)
			parser.WriteHelp(buf)

			if b, err := io.ReadAll(buf); err != nil {
				errMsg := fmt.Sprintf("Error: %v", err)
				r.Log.Error(errMsg)
				r.SendError(m, signal, errMsg)
				return err
			} else {
				// b contains the help text
				errMsg := string(b)
				r.Log.Info(fmt.Sprintf("%v", arg.ErrHelp))
				r.SendError(m, signal, errMsg)
				return arg.ErrHelp
			}
		default:
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return err
		}
	}
	return nil
}

func (r *Module) Start(virtRcv func(*signaldbus.Message)) error {
	return nil
}

func (r *Module) Close(virtRcv func(*signaldbus.Message)) {
}
