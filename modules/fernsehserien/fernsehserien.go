package fernsehserien

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
	"signalbot_go/internal/differ"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules"
	"signalbot_go/signaldbus"
	"sort"
	"strings"

	"github.com/alexflint/go-arg"
	"log/slog"
	"gopkg.in/yaml.v3"
)

// fernsehserien module. Should be instanciated with `NewFernsehserien`.
// data members are only global to be able to unmarshal them
type Fernsehserien struct {
	modules.Module

	fetcher            Fetcher                                `yaml:"-"`
	Series             map[string]string                      `yaml:"series"`
	Aliases            map[string][]string                    `yaml:"aliases"`
	UnavailableSenders map[string]bool                        `yaml:"unavailableSenders"`
	Lasts              differ.Differ[string, string, sending] `yaml:"lasts"` // stores last chat->user->sending
}

// instanciates a new Fernsehserien from a configuration file
// (cfgDir/fernsehserien.yaml)
func NewFernsehserien(log *slog.Logger, cfgDir string) (*Fernsehserien, error) {
	r := Fernsehserien{
		Module: modules.NewModule(log, cfgDir),
		Lasts:  make(differ.Differ[string, string, sending]),
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

	if r.Aliases == nil {
		r.Aliases = make(map[string][]string)
	}
	if r.Series == nil {
		r.Series = make(map[string]string)
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

// validate a fernsehserien struct
func (r *Fernsehserien) Validate() error {
	// validate the generic module first
	if err := r.Module.Validate(); err != nil {
		return err
	}
	for alias, resolvedL := range r.Aliases {
		for _, resolved := range resolvedL {
			if _, ok := r.Series[resolved]; !ok {
				return fmt.Errorf("%s is an invalid series, alias %s is resolved to", resolved, alias)
			}
		}
	}
	return nil
}

// specifies the arguments when handling a request to this module
type Args struct {
	Which  string `arg:"positional"`
	Insert string `arg:"-i,--insert"`
	Quiet  bool   `arg:"-q,--quiet" default:"false"`
	Diff   bool   `arg:"--diff" default:"false"`
	Data   bool   `arg:"-d,--data" default:"true"`
}

// Handle a message from the signaldbus. Parses the message, executes the query
// and responds to signal.
func (r *Fernsehserien) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
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

	if args.Insert != "" {
		r.Series[args.Which] = args.Insert
		r.Aliases["all"] = append(r.Aliases["all"], args.Insert)
	}

	urls := make(map[string]string)
	chat := m.Chat
	if args.Which == "all" {
		urls = r.Series
		chat += "L" // different diffing for "all" command
	} else {
		resolvedL, ok := r.Aliases[args.Which]
		if !ok {
			errMsg := fmt.Sprintf("Error: %v is unknown", args.Which)
			r.Log.Error(errMsg)
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
			r.SendError(m, signal, builder.String())
			return
		}
		for _, re := range resolvedL {
			urls[re] = r.Series[re]
		}
	}

	// execute the query
	readers, err := r.fetcher.getReaders(urls)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
	defer func() {
		for _, r := range readers {
			r.Close()
		}
	}()
	items, err := r.fetcher.getFromReaders(readers, r.UnavailableSenders)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	// respond
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
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
	}
}

// save config file in case something has changed (module allows to add new
// series during runtime)
func (r *Fernsehserien) Close(virtRcv func(*signaldbus.Message)) {
	r.Module.Close(virtRcv)

	delete(r.Aliases, "all") // "all" alias is always a generated one
	f, err := os.Create(filepath.Join(r.ConfigDir, "fernsehserien.yaml"))
	if err != nil {
		r.Log.Error(fmt.Sprintf("Error opening 'fernsehserien.yaml': %v", err))
	}
	defer f.Close()

	e := yaml.NewEncoder(f)
	defer e.Close()
	err = e.Encode(r)
	if err != nil {
		r.Log.Error(fmt.Sprintf("Error endcoding to 'fernsehserien.yaml': %v", err))
	}
}
