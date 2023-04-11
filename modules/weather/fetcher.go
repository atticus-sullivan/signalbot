package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type Fetcher struct {
	ApiKey     string `yaml:"openweatherKey"`
	Lang       string `yaml:"lang"`
	Unitsystem string `yaml:"unitsystem"`
}

func (f *Fetcher) Validate() error {
	if v, ok := langs[f.Lang]; !ok || !v {
		return fmt.Errorf("Invalid lang selected. Was: %s", f.Lang)
	}
	if f.Unitsystem != "metric" && f.Unitsystem != "imperial" && f.Unitsystem != "standard" {
		return fmt.Errorf("Invalid unitsystem selected (has to be either 'metric', 'standard' or 'imperial'). Was: %s", f.Unitsystem)
	}
	if len(f.ApiKey) != 32 {
		return fmt.Errorf("Apikey should have 32 chars, but has %d", len(f.ApiKey))
	}
	return nil
}

const baseUrl string = "https://api.openweathermap.org/data/2.5/onecall?"

func (f *Fetcher) get(loc Position) (*openweatherResp, error) {
	r, err := f.getFromWeb(loc)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	ret, err := f.parse(r)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
func (f *Fetcher) getFromWeb(loc Position) (io.ReadCloser, error) {
	params := url.Values{
		"exclude": {"minutely"},
		"appid":   {f.ApiKey},
		"lat":     {strconv.FormatFloat(float64(loc.Lat), 'f', 4, 32)},
		"lon":     {strconv.FormatFloat(float64(loc.Lon), 'f', 4, 32)},
		"units":   {f.Unitsystem},
		"lang":    {f.Lang},
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

func (f *Fetcher) getFromFile(fn string) (io.Reader, error) {
	return os.Open(fn)
}

func (f *Fetcher) parse(r io.ReadCloser) (*openweatherResp, error) {
	d := json.NewDecoder(r)
	resp := &openweatherResp{}
	err := d.Decode(resp)
	if err != nil {
		panic(err)
	}
	return resp, nil
}
