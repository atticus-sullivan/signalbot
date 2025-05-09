package spotify

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
	"signalbot_go/signalcli"
	"sort"
	"strings"

	"github.com/alexflint/go-arg"
	"gopkg.in/yaml.v3"
	"log/slog"
)

// spotify module. Should be instanciated with `NewSpotify`.
// data members are only global to be able to unmarshal them
type Spotify struct {
	modules.Module

	fetcher   *Fetcher                             `yaml:"-"`
	Queries   map[string]string                    `yaml:"queries"`
	Aliases   map[string][]string                  `yaml:"aliases"`
	Lasts     differ.Differ[string, string, album] `yaml:"lasts"` // stores last chat->user->sending
	QuerySize uint                                 `yaml:"querySize"`
	ClientId string                                 `yaml:"clientId"`
	ClientSecret string                                 `yaml:"clientSecret"`
}

// instanciates a new Spotify from a configuration file
// (cfgDir/spotify.yaml)
func NewSpotify(log *slog.Logger, cfgDir string) (*Spotify, error) {
	r := Spotify{
		Module: modules.NewModule(log, cfgDir),
		Lasts:  make(differ.Differ[string, string, album]),
	}

	f, err := os.Open(filepath.Join(r.ConfigDir, "spotify.yaml"))
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
	if r.Queries == nil {
		r.Queries = make(map[string]string)
	}

	all := make([]string, 0, len(r.Queries))
	for s := range r.Queries {
		r.Aliases[s] = []string{s}
		all = append(all, s)
	}
	r.Aliases["all"] = all

	if r.QuerySize <= 20 {
		r.fetcher = NewFetcher(log, 20, r.ClientId, r.ClientSecret)
	} else {
		r.fetcher = NewFetcher(log, r.QuerySize, r.ClientId, r.ClientSecret)
	}

	// validation
	if err := r.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

// validate a spotify struct
func (r *Spotify) Validate() error {
	// validate the generic module first
	if err := r.Module.Validate(); err != nil {
		return err
	}
	for alias, resolvedL := range r.Aliases {
		for _, resolved := range resolvedL {
			if _, ok := r.Queries[resolved]; !ok {
				return fmt.Errorf("%s is an invalid series, alias %s is resolved to", resolved, alias)
			}
		}
	}
	return nil
}

// specifies the arguments when handling a request to this module
type Args struct {
	Which string `arg:"positional"`
	Quiet bool   `arg:"-q,--quiet" default:"false"`
	Diff  bool   `arg:"--diff" default:"false"`
}

// Handle a message from the signalcli. Parses the message, executes the query
// and responds to signal.
func (r *Spotify) Handle(m *signalcli.Message, signal signalsender.SignalSender, virtRcv func(*signalcli.Message)) {
	// parse the message
	var args Args
	parser, err := arg.NewParser(arg.Config{}, &args)
	if err != nil {
		r.Log.Error(fmt.Sprintf("newParser -> %v", err))
		return
	}

	if err := r.Module.Handle(m, signal, virtRcv, parser); err != nil {
		if err == arg.ErrHelp {
			return
		}
		errMsg := err.Error()
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	queries := make([]string, 0)
	chat := m.Chat
	if args.Which == "all" {
		for _, q := range r.Queries {
			queries = append(queries, q)
		}
		chat += "L" // different diffing for "all" command
	} else {
		resolvedL, ok := r.Aliases[args.Which]
		if !ok {
			errMsg := fmt.Sprintf("Error: %v is unknown", args.Which)
			r.Log.Error(errMsg)
			builder := strings.Builder{}
			builder.WriteString(errMsg)
			builder.WriteRune('\n')
			builder.WriteString("Available queries: ")
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
			queries = append(queries, r.Queries[re])
		}
	}

	// execute the query
	items, err := r.fetcher.get(queries)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	// respond
	resp := strings.Builder{}
	if !args.Diff {
		resp.WriteString(items.String())
	} else {
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
		return
	}
}

// save config file in case something has changed (module allows to add new
// series during runtime)
func (r *Spotify) Close(virtRcv func(*signalcli.Message)) {
	r.Module.Close(virtRcv)

	delete(r.Aliases, "all") // "all" alias is always a generated one
	f, err := os.Create(filepath.Join(r.ConfigDir, "spotify.yaml"))
	if err != nil {
		r.Log.Error(fmt.Sprintf("Error opening 'spotify.yaml': %v", err))
	}
	defer f.Close()

	e := yaml.NewEncoder(f)
	defer e.Close()
	err = e.Encode(r)
	if err != nil {
		r.Log.Error(fmt.Sprintf("Error endcoding to 'spotify.yaml': %v", err))
	}
}
