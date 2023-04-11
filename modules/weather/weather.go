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
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

type Position struct {
	Lon float32 `yaml:"lon"`
	Lat float32 `yaml:"lat"`
}

type Weather struct {
	modules.Module
	Fetcher Fetcher `yaml:"fetcher"`

	MinuteLimit uint `yaml:"minuteLimit"`
	DayLimit    uint `yaml:"dayLimit"`
	MonthLimit  uint `yaml:"monthLimit"`

	Locations map[string]Position `yaml:"locations"`
}

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
	if err := r.Module.Validate(); err != nil {
		return nil, err
	}
	if err := r.Fetcher.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *Weather) Validate() error {
	return nil
}

type Args struct {
	Where string `arg:"positional"`
	// When  int    `arg:"-d,--day" default:"0"`
}

func (r *Weather) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
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

	resp, err := r.Fetcher.get(loc)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
	_, err = signal.Respond(resp.String(), []string{}, m, true)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
}

type calls struct {
	MonthDate  time.Time `yaml:"monthDate"`
	MonthCalls uint      `yaml:"monthCalls"`

	DayDate  time.Time `yaml:"dayDate"`
	DayCalls uint      `yaml:"dayCalls"`

	MinuteDate  time.Time `yaml:"minuteDate"`
	MinuteCalls uint      `yaml:"minuteCalls"`
}

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
