package weather

import (
	"fmt"
	"os"
	"path/filepath"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules"
	"signalbot_go/signaldbus"
	"time"

	"github.com/alexflint/go-arg"
	"log/slog"
	"gopkg.in/yaml.v3"
)

type Position struct {
	Lon float32 `yaml:"lon"`
	Lat float32 `yaml:"lat"`
}

// Weather module. Should be instanciated with `NewWeather`.
// data members are only global to be able to unmarshal them
type Weather struct {
	modules.Module
	Fetcher Fetcher `yaml:"fetcher"`

	MinuteLimit uint `yaml:"minuteLimit"`
	DayLimit    uint `yaml:"dayLimit"`
	MonthLimit  uint `yaml:"monthLimit"`

	Locations map[string]Position `yaml:"locations"`
}

// instanciates a new Weather from a configuration file
// (cfgDir/weather.yaml)
func NewWeather(log *slog.Logger, cfgDir string) (*Weather, error) {
	r := Weather{
		Module:  modules.NewModule(log, cfgDir),
		Fetcher: Fetcher{},
	}

	f, err := os.Open(filepath.Join(r.ConfigDir, "weather.yaml"))
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

// validates the weather struct
func (r *Weather) Validate() error {
	// validate the generic module first
	if err := r.Module.Validate(); err != nil {
		return err
	}
	// validate the fetcher
	if err := r.Fetcher.Validate(); err != nil {
		return err
	}
	return nil
}

// specifies the arguments when handling a request to this module
type Args struct {
	Where string `arg:"positional"`
	// When  int    `arg:"-d,--day" default:"0"`
}

// Handle a message from the signaldbus. Parses the message, executes the query
// and responds to signal.
func (r *Weather) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
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

	// check quota
	if fine, err := r.incQuota(); err != nil {
		errMsg := fmt.Sprintf("Error checking quota. %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	} else if !fine {
		errMsg := "Quota exceeded."
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	loc, ok := r.Locations[args.Where]
	if !ok {
		errMsg := fmt.Sprintf("location %v is unknown", args.Where)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	// execute the query
	reader, err := r.Fetcher.getReader(loc)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
	defer reader.Close()
	resp, err := r.Fetcher.getFromReader(reader)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	// respond
	_, err = signal.Respond(resp.String(), []string{}, m, true)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
}

// track amount of calls made in the current minute, day and month
type calls struct {
	MinuteDate  time.Time `yaml:"minuteDate"`
	MinuteCalls uint      `yaml:"minuteCalls"`

	DayDate  time.Time `yaml:"dayDate"`
	DayCalls uint      `yaml:"dayCalls"`

	MonthDate  time.Time `yaml:"monthDate"`
	MonthCalls uint      `yaml:"monthCalls"`
}

// check and increase the quota of the current minute, day and month
func (w *Weather) incQuota() (bool, error) {
	fn := filepath.Join(w.ConfigDir, "openweather.calls")

	c, err := func() (*calls, error) {
		f, err := os.Open(fn)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		d := yaml.NewDecoder(f)
		c := &calls{}
		err = d.Decode(c)
		if err != nil {
			return nil, err
		}
		return c, nil
	}()
	if err != nil {
		return false, err
	}

	// increment
	nowMi := time.Now().Truncate(time.Minute)
	nowDa := nowMi.Truncate(24 * time.Hour)
	nowMo := time.Date(nowMi.Year(), nowMi.Month(), 1, 0, 0, 0, 0, time.UTC)

	if nowMo != c.MonthDate {
		c.MonthDate = nowMo
		c.MonthCalls = 1
		c.DayDate = nowDa
		c.DayCalls = 1
		c.MinuteDate = nowMi
		c.MinuteCalls = 1

	} else if nowDa != c.DayDate {
		c.MonthCalls += 1

		c.DayDate = nowDa
		c.DayCalls = 1
		c.MinuteDate = nowMi
		c.MinuteCalls = 1

	} else if nowMi != c.MinuteDate {
		c.MonthCalls += 1
		c.DayCalls += 1

		c.MinuteDate = nowMi
		c.MinuteCalls = 1

	} else {
		c.MonthCalls += 1
		c.DayCalls += 1
		c.MinuteCalls += 1
	}

	// check calls
	ret := c.MonthCalls < w.MonthLimit && c.DayCalls < w.DayLimit && c.MinuteCalls < w.MinuteLimit

	// write back
	f, err := os.Create(fn)
	if err != nil {
		return false, err
	}
	defer f.Close()
	e := yaml.NewEncoder(f)
	defer e.Close()
	err = e.Encode(c)
	if err != nil {
		return false, err
	}

	return ret, nil
}
