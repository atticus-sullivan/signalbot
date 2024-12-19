package refectory

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
	"fmt"
	"os"
	"path/filepath"
	"signalbot_go/internal/signalsender"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/modules"
	"signalbot_go/signalcli"
	"sort"
	"strings"
	"time"

	"log/slog"

	"github.com/alexflint/go-arg"
	"gopkg.in/yaml.v3"
)

// Refectory module. Should be instanciated with `NewRefectory`.
// data members are only global to be able to unmarshal them
type Refectory[U FetcherReadCloser, T FetcherInter[U]] struct {
	modules.Module
	fetcher     T                   `yaml:"-"`
	Refectories map[string]uint     `yaml:"refectories"`
	Aliases     map[string][]string `yaml:"aliases"`
}

// instanciates a new Refectory from a configuration file
// (cfgDir/refectory.yaml)
func NewRefectory(log *slog.Logger, cfgDir string) (*Refectory[*fetcherAllReadCloser, *FetcherAll], error) {
	return newRefectoryWithFetcher(log, cfgDir, newFetcherAll())
}

func newRefectoryWithFetcher[U FetcherReadCloser, T FetcherInter[U]](log *slog.Logger, cfgDir string, fetcher T) (*Refectory[U,T], error) {
	r := Refectory[U,T]{
		Module: modules.NewModule(log, cfgDir),
		fetcher: fetcher,
	}
	r.fetcher.init(log)

	f, err := os.Open(filepath.Join(r.ConfigDir, "refectory.yaml"))
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

	for ref := range r.Refectories {
		r.Aliases[ref] = []string{ref}
	}


	// all aliases are lowercase
	for al, ref := range r.Aliases {
		delete(r.Aliases, al)
		r.Aliases[strings.ToLower(al)] = ref
	}

	// validation
	if err := r.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

// validates the refectory struct
func (r *Refectory[U,T]) Validate() error {
	// validate the generic module first
	if err := r.Module.Validate(); err != nil {
		return err
	}
	for alias, resolvedL := range r.Aliases {
		for _, resolved := range resolvedL {
			if _, ok := r.Refectories[resolved]; !ok {
				return fmt.Errorf("%s is an invalid refectory, alias %s is resolved to", resolved, alias)
			}
		}
	}
	return nil
}

// specifies the arguments when handling a request to this module
type Args struct {
	Where string `arg:"positional"`
	When  string `arg:"-d,--day" default:"0"`
	Quiet bool   `arg:"-q,--quiet" default:"false"`
}

// Handle a message from the signalcli. Parses the message, executes the query
// and responds to signal.
func (r *Refectory[U,T]) Handle(m *signalcli.Message, signal signalsender.SignalSender, virtRcv func(*signalcli.Message)) {
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

	days, err := cmdsplit.ParseNumberRange(args.When)
	if err != nil {
		errMsg := err.Error()
		r.Log.Warn(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
	for _, when := range days {
		date := time.Now().Add(time.Hour * 24 * time.Duration(when))
		date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

		args.Where = strings.ToLower(args.Where)

		resolvedL, ok := r.Aliases[args.Where]
		if !ok {
			errMsg := fmt.Sprintf("Error: %v is unknown", args.Where)
			r.Log.Error(errMsg)
			builder := strings.Builder{}
			builder.WriteString(errMsg)
			builder.WriteRune('\n')
			builder.WriteString("Available refectories: ")
			sorted := make(sort.StringSlice, 0, len(r.Aliases))
			for k := range r.Aliases {
				sorted = append(sorted, k)
			}
			sorted.Sort()
			builder.WriteString(strings.Join(sorted, ", "))
			r.SendError(m, signal, builder.String())
			continue
		}

		for _, ref := range resolvedL {
			// execute the query
			reader, err := r.fetcher.getReader(r.Refectories[ref], date)
			if err != nil {
				var errMsg string
				if err == ErrNotOpenThatDay {
					errMsg = err.Error()
				} else {
					errMsg = fmt.Sprintf("Error: %v", err)
				}
				r.Log.Error(errMsg)
				if !args.Quiet || err != ErrNotOpenThatDay {
					r.SendError(m, signal, errMsg)
				}
				continue
			}
			defer reader.Close()
			menu, err := r.fetcher.getFromReader(reader)
			if err != nil {
				errMsg := fmt.Sprintf("Error: %v", err)
				r.Log.Error(errMsg)
				r.SendError(m, signal, errMsg)
				continue
			}
			// respond
			menuS := fmt.Sprintf("%s on %s\n", ref, date.Format("Mon 2006-01-02")) + menu.String()
			_, err = signal.Respond(menuS, []string{}, m, true)
			if err != nil {
				errMsg := fmt.Sprintf("Error: %v", err)
				r.Log.Error(errMsg)
				r.SendError(m, signal, errMsg)
				continue
			}
		}
	}
}
