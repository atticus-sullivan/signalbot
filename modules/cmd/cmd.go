package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"signalbot_go/internal/signalsender"
	"signalbot_go/signaldbus"

	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

// module to execute scripts on the server and respond with the output from the stdout.
// Create with NewCmd
// TODO what about concurrency
type Cmd struct {
	Log *slog.Logger `yaml:"-"`
	// the configuration will be read from here. In addition, this will be the
	// working directory for any script being executed.
	ConfigDir string `yaml:"-"`
	// store a command -> scriptname/-path mapping.
	// Might be replaced with argument parsing (e.g. https://pkg.go.dev/github.com/alexflint/go-arg)?
	Commands map[string]string `yaml:"commands"`
}

// create a new cmd instance from a configuration.
func NewCmd(log *slog.Logger, cfgDir string) (*Cmd, error) {
	c := Cmd{
		Log:       log,
		ConfigDir: cfgDir,
	}

	f, err := os.Open(filepath.Join(c.ConfigDir, "cmd.yaml"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.KnownFields(true)
	err = d.Decode(&c)
	if err != nil {
		return nil, err
	}

	// validation
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

// check if stored values are valid (for now only the read configuration mapping)
func (c *Cmd) Validate() error {
	for _, cmd := range c.Commands {
		if cmd == "" {
			return fmt.Errorf("command cannot be empty")
		}
		var stat os.FileInfo
		stat, err := os.Stat(filepath.Join(c.ConfigDir, cmd))
		if err != nil {
			return fmt.Errorf("Error reading command. %v", err)
		}
		if !stat.Mode().IsRegular() {
			return fmt.Errorf("Command is not a regular file")
		}
		if stat.Mode().Perm()&0111 != 0111 {
			return fmt.Errorf("Command is not executable, %d (file: %s)", stat.Mode().Perm(), filepath.Join(c.ConfigDir, cmd))
		}
	}
	return nil
}

// shortcut for sending an error via signal. If this fails log error.
func (c *Cmd) sendError(m *signaldbus.Message, signal signalsender.SignalSender, reply string) {
	if _, err := signal.Respond(reply, nil, m); err != nil {
		c.Log.Error(fmt.Sprintf("Error responding to %v", m))
	}
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
			c.sendError(m, signal, errMsg)
			return
		}

		if err := command.Start(); err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			c.Log.Error(errMsg)
			c.sendError(m, signal, errMsg)
			return
		}

		output, _ := io.ReadAll(out)

		if err := command.Wait(); err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			c.Log.Error(errMsg)
			c.sendError(m, signal, errMsg)
			return
		}

		c.Log.Info(fmt.Sprintf("Command returned successfully. Output:\n%s", output))
		_, err = signal.Respond(string(output), nil, m)
		if err != nil {
			c.Log.Error(fmt.Sprintf("Failed to send reply to %v", m))
		}

	} else {
		_, err := signal.Respond("Invalid command", nil, m)
		if err != nil {
			c.Log.Error(fmt.Sprintf("Failed to send reply to %v", m))
		}
	}
}

func (c *Cmd) Start() error {
	return nil
}

func (c *Cmd) Close() {
}
