package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules"
	"signalbot_go/signaldbus"

	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

// module to execute scripts on the server and respond with the output from the stdout.
// Create with NewCmd
// TODO what about concurrency
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
	if err := r.Module.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

// check if stored values are valid (for now only the read configuration mapping)
func (r *Cmd) Validate() error {
	for _, cmd := range r.Commands {
		if cmd == "" {
			return fmt.Errorf("command cannot be empty")
		}
		var stat os.FileInfo
		stat, err := os.Stat(filepath.Join(r.ConfigDir, cmd))
		if err != nil {
			return fmt.Errorf("Error reading command. %v", err)
		}
		if !stat.Mode().IsRegular() {
			return fmt.Errorf("Command is not a regular file")
		}
		if stat.Mode().Perm()&0111 != 0111 {
			return fmt.Errorf("Command is not executable, %d (file: %s)", stat.Mode().Perm(), filepath.Join(r.ConfigDir, cmd))
		}
	}
	return nil
}

// handle a signalmessage
func (c *Cmd) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	if cmds, ok := c.Commands[m.Message]; ok {
		command := exec.Command(cmds)
		command.Dir = c.ConfigDir

		out, err := command.StdoutPipe()
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			c.Log.Error(errMsg)
			c.SendError(m, signal, errMsg)
			return
		}

		if err := command.Start(); err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			c.Log.Error(errMsg)
			c.SendError(m, signal, errMsg)
			return
		}

		output, _ := io.ReadAll(out)

		if err := command.Wait(); err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			c.Log.Error(errMsg)
			c.SendError(m, signal, errMsg)
			return
		}

		c.Log.Info(fmt.Sprintf("Command returned successfully. Output:\n%s", output))
		_, err = signal.Respond(string(output), nil, m, true)
		if err != nil {
			c.Log.Error(fmt.Sprintf("Failed to send reply to %v", m))
		}

	} else {
		_, err := signal.Respond("Invalid command", nil, m, false)
		if err != nil {
			c.Log.Error(fmt.Sprintf("Failed to send reply to %v", m))
		}
	}
}
