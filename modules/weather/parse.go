package weather

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
	"math"
	"strconv"
	"strings"
	"time"
)

type openweatherWeatherCommon struct {
	Dt         int64   `json:"dt"`
	Humidity   float32 `json:"humidity"`
	Clouds     float32 `json:"clouds"`
	Uvi        float32 `json:"uvi"`
	Wind_speed float32 `json:"wind_speed"`
	WindDeg    float64 `json:"wind_deg"`
	Weather    []struct {
		Id          uint   `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		// Icon string `json:"icon"`
	}
}

func (o *openweatherWeatherCommon) String() string {
	// builder := strings.Builder{}
	return ""
}

type openweatherWeatherCurrHour struct {
	openweatherWeatherCommon
	Temp float32 `json:"temp"`
	Rain struct {
		OneHour float32 `json:"1h"`
	} `json:"rain"`
	Snow struct {
		OneHour float32 `json:"1h"`
	} `json:"snow"`
}

func (o *openweatherWeatherCurrHour) String() string {
	// builder := strings.Builder{}
	return ""
}

type openweatherWeatherDay struct {
	openweatherWeatherCommon
	Rain float32 `json:"rain"`
	Snow float32 `json:"snow"`
	Temp struct {
		Morn  float32 `json:"morn"`
		Day   float32 `json:"day"`
		Eve   float32 `json:"eve"`
		Night float32 `json:"night"`
		Min   float32 `json:"min"`
		Max   float32 `json:"max"`
	} `json:"temp"`
}

// func (o *openweatherWeatherDay) String() string {
// 	// builder := strings.Builder{}
// 	return ""
// }

type openweatherWeatherAlert struct {
	SenderName  string   `json:"sender_name"`
	Event       string   `json:"event"`
	Start       int64    `json:"start"`
	End         int64    `json:"end"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// func (o *openweatherWeatherAlert) String() string {
// 	// builder := strings.Builder{}
// 	return ""
// }

type openweatherResp struct {
	Lat            float32                      `json:"lat"`
	Lon            float32                      `json:"lon"`
	Timezone       string                       `json:"timezone"`
	TimezoneOffset int                          `json:"timezone_offset"`
	Current        openweatherWeatherCurrHour   `json:"current"`
	Hourly         []openweatherWeatherCurrHour `json:"hourly"`
	Daily          []openweatherWeatherDay      `json:"daily"`
	Alerts         []openweatherWeatherAlert    `json:"alerts"`
}

// TODO len(weather) > 1 => warn + write json to file

func (o *openweatherResp) String() string {
	builder := strings.Builder{}
	date := time.Unix(o.Current.Dt, 0)
	tz, err := time.LoadLocation(o.Timezone)
	if err == nil {
		date = date.In(tz)
	} else {
		// TODO log warning about invalid TZ, default to UTC
	}

	builder.WriteString(weatherCCs[o.Current.Weather[0].Id].icon)
	// builder.WriteString(date.Format(" Mon 02.01 15:04:05 "))
	builder.WriteString(date.Format(" Mon 02.01  "))
	builder.WriteString(strconv.FormatInt(int64(o.Current.Temp), 10))
	builder.WriteString("°C")
	builder.WriteRune('\n')

	builder.WriteString("Humidity: ")
	builder.WriteString(strconv.FormatFloat(float64(o.Current.Humidity), 'f', 0, 32))
	builder.WriteRune('%')
	builder.WriteRune(' ')
	if o.Current.Rain.OneHour != 0 {
		builder.WriteString(" R:")
		builder.WriteString(strconv.FormatFloat(float64(o.Current.Rain.OneHour), 'f', 1, 32))
		builder.WriteString("mm")
	}
	if o.Current.Snow.OneHour != 0 {
		builder.WriteString(" S:")
		builder.WriteString(strconv.FormatFloat(float64(o.Current.Snow.OneHour), 'f', 1, 32))
		builder.WriteString("mm")
	}
	builder.WriteRune('\n')

	builder.WriteString("Clouds: ")
	builder.WriteString(strconv.FormatFloat(float64(o.Current.Clouds), 'f', 0, 32))
	builder.WriteRune('%')
	builder.WriteRune('\n')

	builder.WriteString("UV:")
	builder.WriteString(strconv.FormatFloat(float64(o.Current.Uvi), 'f', 1, 32))
	builder.WriteString(" Wind: ")
	builder.WriteString(strconv.FormatFloat(float64(o.Current.Wind_speed), 'f', 0, 32))
	builder.WriteString("m/s ")
	builder.WriteString(wind[int(math.Round(o.Current.WindDeg/360*float64(len(wind))))])
	builder.WriteRune('\n')

	// daily (max 7)
	if len(o.Daily) > 0 {
		builder.WriteRune('\n')
		for i, d := range o.Daily {
			if i >= 7 {
				break
			}
			date := time.Unix(d.Dt, 0).In(tz)

			builder.WriteString(weatherCCs[d.Weather[0].Id].icon)
			builder.WriteString(date.Format(" Mon 02.01  "))
			builder.WriteString(strconv.FormatFloat(float64(d.Temp.Min), 'f', 0, 32))
			builder.WriteString("°C - ")
			builder.WriteString(strconv.FormatFloat(float64(d.Temp.Max), 'f', 0, 32))
			builder.WriteString("°C")
			if d.Rain != 0 {
				builder.WriteString(" R:")
				builder.WriteString(strconv.FormatFloat(float64(d.Rain), 'f', 1, 32))
				builder.WriteString("mm")
			}
			if d.Snow != 0 {
				builder.WriteString(" S:")
				builder.WriteString(strconv.FormatFloat(float64(d.Snow), 'f', 1, 32))
				builder.WriteString("mm")
			}
			builder.WriteRune('\n')
		}
	}

	// hourly (max 5)
	if len(o.Hourly) > 0 {
		builder.WriteRune('\n')
		for i, h := range o.Hourly {
			if i >= 5 {
				break
			}
			date := time.Unix(h.Dt, 0).In(tz)

			builder.WriteString(weatherCCs[h.Weather[0].Id].icon)
			builder.WriteString(date.Format(" 15:04 02.01  "))
			builder.WriteString(strconv.FormatFloat(float64(h.Temp), 'f', 0, 32))
			builder.WriteString("°C")
			if h.Rain.OneHour != 0 {
				builder.WriteString(" R:")
				builder.WriteString(strconv.FormatFloat(float64(h.Rain.OneHour), 'f', 1, 32))
				builder.WriteString("mm")
			}
			if h.Snow.OneHour != 0 {
				builder.WriteString(" S:")
				builder.WriteString(strconv.FormatFloat(float64(h.Snow.OneHour), 'f', 1, 32))
				builder.WriteString("mm")
			}
			builder.WriteRune('\n')
		}
	}

	if len(o.Alerts) > 0 {
		for _, a := range o.Alerts {
			builder.WriteRune('\n')
			sDate := time.Unix(a.Start, 0).In(tz)
			eDate := time.Unix(a.End, 0).In(tz)

			builder.WriteString("Wetterwarnung")
			builder.WriteString(sDate.Format(" (15:04 02.01 - "))
			builder.WriteString(eDate.Format("15:04 02.01):"))
			builder.WriteRune('\n')
			builder.WriteString(a.Description)
			builder.WriteString("\nQuelle: ")
			builder.WriteString(a.SenderName)
			builder.WriteRune('\n')
		}
	}

	builder.WriteRune('\n')
	builder.WriteString("Quelle: openweather")
	return builder.String()
}
