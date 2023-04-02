package tv

import (
	"io"
	"signalbot_go/modules/tv/internal/show"
	"signalbot_go/modules/tv/scrapers"
	"time"

	"golang.org/x/exp/slog"
)

// could implement caching if necessary
type Fetcher struct {
	log            *slog.Logger
	loc            *time.Location
	timeout        time.Duration
	senderScrapers []subFetcher
}

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

			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "br", loc), Url: "https://programm.ard.de/TV/Programm/Sender?sender=-28107&datum=%s&hour=0&archiv=1"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "ndr", loc), Url: "https://programm.ard.de/TV/Programm/Sender?sender=-28226&datum=%s&hour=0&archiv=1"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "ard", loc), Url: "https://programm.ard.de/TV/Programm/Sender?sender=28106&datum=%s&hour=0&archiv=1"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "arte", loc), Url: "https://programm.ard.de/TV/Programm/Sender?sender=28724&datum=%s&hour=0&archiv=1"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "3sat", loc), Url: "https://programm.ard.de/TV/Programm/Sender?sender=28007&datum=%s&hour=0&archiv=1"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "hr", loc), Url: "https://programm.ard.de/TV/Programm/Sender?sender=28108&datum=%s&hour=0&archiv=1"},
			&scrapers.Ard2{ScraperBase: scrapers.NewScraperBase(log, "phoenix", loc), Url: "https://programm.ard.de/TV/Programm/Sender?sender=28725&datum=%s&hour=0&archiv=1"},

			scrapers.NewZdf(scrapers.NewScraperBase(log, "zdf", loc), ".timeline-ZDF"),
			scrapers.NewZdf(scrapers.NewScraperBase(log, "zdfInfo", loc), ".timeline-ZDFinfo"),
			scrapers.NewZdf(scrapers.NewScraperBase(log, "zdfNeo", loc), ".timeline-ZDFneo"),
		},
	}
	return f
}

type subFetcher interface {
	Get(time.Time) (io.ReadCloser, error)
	// GetFromFile() (io.ReadCloser, error)
	Parse(io.ReadCloser, chan<- show.Show, time.Time)
	Name() string
}

// Idea: add Url() function ->  build set of URLs, download and parse them to
// *html.Node -> sender parsing. This would reduce the amount of requests to
// the zdf page

func (fetcher *Fetcher) Get() map[string][]show.Show {
	channels := make([]chan show.Show, len(fetcher.senderScrapers))

	now := time.Now()
	for iS, fS := range fetcher.senderScrapers {
		// make copies of the loop variables before capturing them in the goroutine
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
		}
		if ok {
			sender := fetcher.senderScrapers[i].Name()
			res[sender] = append(res[sender], ele)
		} else {
			channels[i] = nil // disable channel in select as send/rcv on nil block forever
		}
	}
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
