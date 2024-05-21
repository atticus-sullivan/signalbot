package modules

// signalbot
// Copyright (C) 2024  Lukas Heindl
// 
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/internal/signalsender"
	"signalbot_go/signalcli"

	"github.com/alexflint/go-arg"
	"log/slog"
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
func (r *Module) SendError(m *signalcli.Message, signal signalsender.SignalSender, reply string) {
	if _, err := signal.Respond(reply, nil, m, false); err != nil {
		r.Log.Error(fmt.Sprintf("Error responding to %v", m))
	}
}

// returns whether an error ocurred (error is already logged and sent to the user)
func (r *Module) Handle(m *signalcli.Message, signal signalsender.SignalSender, virtRcv func(*signalcli.Message), parser *arg.Parser) error {

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

func (r *Module) Start(virtRcv func(*signalcli.Message)) error {
	return nil
}

func (r *Module) Close(virtRcv func(*signalcli.Message)) {
}
