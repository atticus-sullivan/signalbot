package main

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
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"time"

	// "runtime"
	"signalbot_go/signalserver"
	// "strings"
	"syscall"

	"github.com/lmittmann/tint"
	"log/slog"
)

// outsourced configuration of logging. Call this to get a configured root logger
func logInit() *slog.Logger {
	// _, file, _, ok := runtime.Caller(0)
	// if !ok {
	// 	panic("failed to retrieve runtime information")
	// }
	// dir, _ := filepath.Split(file)
	// replace := func(groups []string, a slog.Attr) slog.Attr {
	// 	// Remove the directory from the source's filename.
	// 	if a.Key == slog.SourceKey {
	// 		a.Value = slog.StringValue(strings.TrimPrefix(a.Value.String(), dir))
	// 	}
	// 	return a
	// }
	// logger := slog.New(slog.HandlerOptions{AddSource: true, ReplaceAttr: replace, Level: slog.LevelInfo}.NewTextHandler(os.Stderr))
	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}))
	return logger
}

// get the directory for the configuration of this project
func getCfgDir() string {
	dir, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		usr, _ := user.Current()
		dir = filepath.Join(usr.HomeDir, ".config")
	}
	return filepath.Join(dir, "signalbot")
}

// get the directory for stored data of this project
func getDataDir() string {
	dir, ok := os.LookupEnv("XDG_DATA_HOME")
	if !ok {
		usr, _ := user.Current()
		dir = filepath.Join(usr.HomeDir, ".local", "share")
	}
	return filepath.Join(dir, "signalbot")
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	s, err := signalserver.NewSignalServer(logInit(), getCfgDir(), getDataDir())
	if err != nil {
		panic(err)
	}
	defer s.Close()
	if err := s.Start(); err != nil {
		panic(err)
	}
	// Block until a signal is received.
	<-c
}
