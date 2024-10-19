package tv

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
	"io"
	"signalbot_go/modules/tv/internal/show"
	"signalbot_go/modules/tv/scrapers"
	"time"

	"github.com/andybalholm/cascadia"
	"github.com/jellydator/ttlcache/v3"
	"log/slog"
)

// fetches stuff. Maybe some day this will have data members (e.g. if caching
// is implemented)
// Has to be instanciated via `NewFetcher`
type Fetcher struct {
	log            *slog.Logger
	loc            *time.Location
	timeout        time.Duration
	senderScrapers []subFetcher
	cache          *ttlcache.Cache[string, map[string][]show.Show]
}

var (
	cascZdf_zdf  cascadia.Matcher = cascadia.MustCompile(".timeline-ZDF")
	cascZdf_info cascadia.Matcher = cascadia.MustCompile(".timeline-ZDFinfo")
	cascZdf_neo  cascadia.Matcher = cascadia.MustCompile(".timeline-ZDFneo")
)

// instanciates a new Fetcher object
func NewFetcher(log *slog.Logger, loc *time.Location, timeout time.Duration) *Fetcher {
	f := &Fetcher{
		log:     log,
		loc:     loc,
		timeout: timeout,
		senderScrapers: []subFetcher{
			&scrapers.Vox{ScraperBase: scrapers.NewScraperBase(log, "vox", loc)},
			&scrapers.Rtl{ScraperBase: scrapers.NewScraperBase(log, "rtl", loc)},
			&scrapers.Rtl2{ScraperBase: scrapers.NewScraperBase(log, "rtl2", loc)},

			&scrapers.SatEins{ScraperBase: scrapers.NewScraperBase(log, "sat1", loc), Url: "https://www.sat1.de/tv-programm"},
			&scrapers.SatEins{ScraperBase: scrapers.NewScraperBase(log, "prosieben", loc), Url: "https://www.prosieben.de/tv-programm"},
			&scrapers.SatEins{ScraperBase: scrapers.NewScraperBase(log, "kabeleins", loc), Url: "https://www.kabeleins.de/tv-programm"},

			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "br", loc), Url: "https://programm-api.ard.de/program/api/program?mode=channel&channelIds=br"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "ndr", loc), Url: "https://programm-api.ard.de/program/api/program?mode=channel&channelIds=ndr"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "ard", loc), Url: "https://programm-api.ard.de/program/api/program?mode=channel&channelIds=daserste"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "arte", loc), Url: "https://programm-api.ard.de/program/api/program?mode=channel&channelIds=arte"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "alpha", loc), Url: "https://programm-api.ard.de/program/api/program?mode=channel&channelIds=alpha"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "one", loc), Url: "https://programm-api.ard.de/program/api/program?mode=channel&channelIds=one"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "wdr", loc), Url: "https://programm-api.ard.de/program/api/program?mode=channel&channelIds=wdr"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "swr", loc), Url: "https://programm-api.ard.de/program/api/program?mode=channel&channelIds=swr"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "3sat", loc), Url: "https://programm-api.ard.de/program/api/program?mode=channel&channelIds=3sat"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "hr", loc), Url: "https://programm-api.ard.de/program/api/program?mode=channel&channelIds=hr"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "phoenix", loc), Url: "https://programm-api.ard.de/program/api/program?mode=channel&channelIds=phoenix"},

			scrapers.NewZdf(scrapers.NewScraperBase(log, "zdf", loc), cascZdf_zdf),
			scrapers.NewZdf(scrapers.NewScraperBase(log, "zdfInfo", loc), cascZdf_info),
			scrapers.NewZdf(scrapers.NewScraperBase(log, "zdfNeo", loc), cascZdf_neo),
		},
		cache: ttlcache.New(ttlcache.WithTTL[string, map[string][]show.Show](30*time.Minute), ttlcache.WithDisableTouchOnHit[string, map[string][]show.Show]()),
	}
	return f
}

// subfetcher fetches the shows of one sender
type subFetcher interface {
	Get(time.Time) (io.ReadCloser, error)
	// GetFromFile() (io.ReadCloser, error)
	Parse(io.ReadCloser, chan<- show.Show, time.Time)
	Name() string
}

// queries all senders for shows and collects the result in a map (sender ->
// shows)
func (fetcher *Fetcher) Get() map[string][]show.Show {
	if v := fetcher.cache.Get("all"); v != nil && !v.IsExpired() {
		return v.Value()
	}

	channels := make([]chan show.Show, len(fetcher.senderScrapers))

	now := time.Now()
	for iS, fS := range fetcher.senderScrapers {
		// make copies of the loop variables before capturing them in the
		// goroutine
		i := iS
		f := fS
		channels[i] = make(chan show.Show)
		go func() {
			r, err := f.Get(now)
			if err != nil {
				// logging has to happen inside the Get function
				return
			}
			defer r.Close()
			f.Parse(r, channels[i], now)
		}()
	}

	res := make(map[string][]show.Show)
	for _, f := range fetcher.senderScrapers {
		res[f.Name()] = make([]show.Show, 0, 2)
	}

	var ele show.Show
	var ok bool
	var i uint
	timer := time.NewTimer(fetcher.timeout)
collect:
	for !finished(channels) {
		timer.Reset(fetcher.timeout)
		select {
		case <-timer.C:
			break collect
		case ele, ok = <-channels[0]:
			i = 0
		case ele, ok = <-channels[1]:
			i = 1
		case ele, ok = <-channels[2]:
			i = 2
		case ele, ok = <-channels[3]:
			i = 3
		case ele, ok = <-channels[4]:
			i = 4
		case ele, ok = <-channels[5]:
			i = 5
		case ele, ok = <-channels[6]:
			i = 6
		case ele, ok = <-channels[7]:
			i = 7
		case ele, ok = <-channels[8]:
			i = 8
		case ele, ok = <-channels[9]:
			i = 9
		case ele, ok = <-channels[10]:
			i = 10
		case ele, ok = <-channels[11]:
			i = 11
		case ele, ok = <-channels[12]:
			i = 12
		case ele, ok = <-channels[13]:
			i = 13
		case ele, ok = <-channels[14]:
			i = 14
		case ele, ok = <-channels[15]:
			i = 15
		case ele, ok = <-channels[16]:
			i = 16
		case ele, ok = <-channels[17]:
			i = 17
		case ele, ok = <-channels[18]:
			i = 18
		case ele, ok = <-channels[19]:
			i = 19
		}
		if ok {
			sender := fetcher.senderScrapers[i].Name()
			res[sender] = append(res[sender], ele)
		} else {
			channels[i] = nil // disable channel in select as send/rcv on nil block forever
		}
	}

	fetcher.cache.Set("all", res, ttlcache.DefaultTTL)
	return res
}

// check if there is still a channel from which can be read
func finished(s []chan show.Show) bool {
	for _, b := range s {
		if b != nil {
			return false
		}
	}
	return true
}
