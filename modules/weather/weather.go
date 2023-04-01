package weather

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/internal/signalsender"
	"signalbot_go/signaldbus"
	"strconv"
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
	log       *slog.Logger `yaml:"-"`
	ConfigDir string       `yaml:"-"`

	ApiKey      string `yaml:"openweatherKey"`
	Lang        string `yaml:"lang"`
	Unitsystem  string `yaml:"unitsystem"`
	MinuteLimit uint   `yaml:"minuteLimit"`
	DayLimit    uint   `yaml:"dayLimit"`
	MonthLimit  uint   `yaml:"monthLimit"`

	Locations map[string]Position `yaml:"locations"`
}

func NewWeather(log *slog.Logger, cfgDir string) (*Weather, error) {
	w := Weather{
		log:       log,
		ConfigDir: cfgDir,
	}

	f, err := os.Open(filepath.Join(w.ConfigDir, "weather.yaml"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.KnownFields(true)
	err = d.Decode(&w)
	if err != nil {
		return nil, err
	}

	// validation
	if err := w.Validate(); err != nil {
		return nil, err
	}

	return &w, nil
}

func (w *Weather) Validate() error {
	if v, ok := langs[w.Lang]; !ok || !v {
		return fmt.Errorf("Invalid lang selected. Was: %s", w.Lang)
	}
	if w.Unitsystem != "metric" && w.Unitsystem != "imperial" && w.Unitsystem != "standard" {
		return fmt.Errorf("Invalid unitsystem selected (has to be either 'metric', 'standard' or 'imperial'). Was: %s", w.Unitsystem)
	}
	if len(w.ApiKey) != 32 {
		return fmt.Errorf("Apikey should have 32 chars, but has %d", len(w.ApiKey))
	}
	return nil
}

func (w *Weather) sendError(m *signaldbus.Message, signal signalsender.SignalSender, reply string) {
	if _, err := signal.Respond(reply, nil, m); err != nil {
		w.log.Error(fmt.Sprintf("Error responding to %v", m))
	}
}

type Args struct {
	Where string `arg:"positional"`
	// When  int    `arg:"-d,--day" default:"0"`
}

func (w *Weather) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	var args Args
	parser, err := arg.NewParser(arg.Config{}, &args)

	if err != nil {
		w.log.Error(fmt.Sprintf("weather module: newParser -> %v", err))
		return
	}

	vargs, err := cmdsplit.Split(m.Message)
	if err != nil {
		errMsg := fmt.Sprintf("weather module: Error on parsing message: %v", err)
		w.log.Error(errMsg)
		w.sendError(m, signal, errMsg)
		return
	}

	err = parser.Parse(vargs)

	if err != nil {
		switch err {
		case arg.ErrVersion:
			// not implemented
			errMsg := fmt.Sprintf("weather module: Error: %v", "Version is not implemented")
			w.log.Error(errMsg)
			w.sendError(m, signal, errMsg)
			return
		case arg.ErrHelp:
			buf := new(bytes.Buffer)
			parser.WriteHelp(buf)

			if b, err := io.ReadAll(buf); err != nil {
				errMsg := fmt.Sprintf("openweather Error: %v", err)
				w.log.Error(errMsg)
				w.sendError(m, signal, errMsg)
				return
			} else {
				errMsg := string(b)
				w.log.Info(fmt.Sprintf("weather module: Error: %v", err))
				w.sendError(m, signal, errMsg)
				return
			}
		default:
			errMsg := fmt.Sprintf("openweather Error: %v", err)
			w.log.Error(errMsg)
			w.sendError(m, signal, errMsg)
			return
		}
	} else {
		// check quota
		if fine, err := w.incQuota(); err != nil {
			errMsg := fmt.Sprintf("openweather Error checking quota. %v", err)
			w.log.Error(errMsg)
			w.sendError(m, signal, errMsg)
			return
		} else if !fine {
			errMsg := "openweather Error: Quota exceeded."
			w.log.Error(errMsg)
			w.sendError(m, signal, errMsg)
			return
		}

		loc, ok := w.Locations[args.Where]
		if !ok {
			errMsg := fmt.Sprintf("openweather Error: location %v is unknown", args.Where)
			w.log.Error(errMsg)
			w.sendError(m, signal, errMsg)
			return
		}

		r, err := w.get(loc)
		if err != nil {
			errMsg := fmt.Sprintf("openweather Error: %v", err)
			w.log.Error(errMsg)
			w.sendError(m, signal, errMsg)
			return
		}
		defer r.Close()
		resp, err := w.parse(r)
		if err != nil {
			errMsg := fmt.Sprintf("openweather Error: %v", err)
			w.log.Error(errMsg)
			w.sendError(m, signal, errMsg)
			return
		}
		_, err = signal.Respond(resp, []string{}, m)
		if err != nil {
			errMsg := fmt.Sprintf("openweather Error: %v", err)
			w.log.Error(errMsg)
			w.sendError(m, signal, errMsg)
			return
		}
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

const baseUrl string = "https://api.openweathermap.org/data/2.5/onecall?"

func (w *Weather) get(loc Position) (io.ReadCloser, error) {
	params := url.Values{
		"exclude": {"minutely"},
		"appid":   {w.ApiKey},
		"lat":     {strconv.FormatFloat(float64(loc.Lat), 'f', 4, 32)},
		"lon":     {strconv.FormatFloat(float64(loc.Lon), 'f', 4, 32)},
		"units":   {w.Unitsystem},
		"lang":    {w.Lang},
	}
	resp, err := http.Get(baseUrl + params.Encode())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getting weather information failed with statusCode %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func (w *Weather) getFromFile() (io.ReadCloser, error) {
	return os.Open(filepath.Join(w.ConfigDir, "example.json"))
}

func (w *Weather) parse(r io.ReadCloser) (string, error) {
	d := json.NewDecoder(r)
	resp := &openweatherResp{}
	err := d.Decode(resp)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%v", resp), nil
}

func (w *Weather) Start(virtRcv func(*signaldbus.Message)) error {
	return nil
}

func (w *Weather) Close(virtRcv func(*signaldbus.Message)) {
}
