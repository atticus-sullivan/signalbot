package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules"
	"signalbot_go/signaldbus"

	"log/slog"
	"gopkg.in/yaml.v3"
)

// in addition to these errors there may be an fmt.Errorf error
var (
	ErrCommandEmpty error = errors.New("command cannot be empty")
	ErrNoRegFile    error = errors.New("Command is not a regular file")
	ErrNoExec       error = errors.New("Command is not executable")
)

// module to execute scripts on the server and respond with the output from the stdout.
// Create with NewCmd
type Cmd struct {
	modules.Module
	// store a command -> scriptname/-path mapping.
	// Might be replaced with argument parsing (e.g. https://pkg.go.dev/github.com/alexflint/go-arg)?
	Commands map[string]string `yaml:"commands"`
}

// create a new cmd instance from a configuration.
func NewCmd(log *slog.Logger, cfgDir string) (*Cmd, error) {
	r := Cmd{
		Module: modules.NewModule(log, cfgDir),
	}

	f, err := os.Open(filepath.Join(r.ConfigDir, "cmd.yaml"))
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

// check if stored values are valid (for now only the read configuration mapping)
func (r *Cmd) Validate() error {
	// validate the generic module first
	if err := r.Module.Validate(); err != nil {
		return err
	}
	for _, cmd := range r.Commands {
		if cmd == "" {
			return ErrCommandEmpty
		}
		var stat os.FileInfo
		stat, err := os.Stat(filepath.Join(r.ConfigDir, cmd))
		if err != nil {
			return fmt.Errorf("Error reading command. %v", err)
		}
		if !stat.Mode().IsRegular() {
			return ErrNoRegFile
		}
		if stat.Mode().Perm()&0111 != 0111 {
			return ErrNoExec
		}
	}
	return nil
}

// handle a signalmessage
func (r *Cmd) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	args, err := cmdsplit.Split(m.Message)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
	if len(args) < 1 {
		errMsg := "Error: too few arguments povided"
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
	if cmds, ok := r.Commands[args[0]]; ok {
		command := exec.Command(cmds, args[1:]...)
		command.Dir = r.ConfigDir

		out, err := command.StdoutPipe()
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return
		}

		if err := command.Start(); err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return
		}

		output, _ := io.ReadAll(out)

		if err := command.Wait(); err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return
		}

		r.Log.Info(fmt.Sprintf("Command returned successfully. Output:\n%s", output))
		_, err = signal.Respond(string(output), nil, m, true)
		if err != nil {
			r.Log.Error(fmt.Sprintf("Failed to send reply to %v", m))
		}

	} else {
		_, err := signal.Respond("Invalid command", nil, m, false)
		if err != nil {
			r.Log.Error(fmt.Sprintf("Failed to send reply to %v", m))
		}
	}
}
