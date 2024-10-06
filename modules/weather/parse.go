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

type weatherHdr struct {
	Latitude float64
	Longitude float64
	Elevation float64
	Generationtime_ms float64
	Utc_offset_seconds int
	Timezone string
	Timezone_abbreviation string
}

type weatherHour struct {
	Time []string
	Weather_code []int

	Temperature_2m []float64
	Precipitation []float64
	Snowfall []float64
}
type weatherHourU struct {
	Time string
	Weather_code string

	Temperature_2m string
	Precipitation string
	Snowfall string
}

type weatherDaily struct {
	Time []string
	Weather_code []int

	Uv_index_max []float64
	Temperature_2m_max []float64
	Temperature_2m_min []float64
	Precipitation_sum []float64
	Snowfall_sum []float64
	Sunshine_duration []float64
	Wind_speed_10m_max []float64
	Wind_direction_10m_dominant []float64
}
type weatherDailyU struct {
	Time string
	Weather_code string

	Uv_index_max string
	Temperature_2m_max string
	Temperature_2m_min string
	Precipitation_sum string
	Snowfall_sum string
	Sunshine_duration string
	Wind_speed_10m_max string
	Wind_direction_10m_dominant string
}

type weatherCurr struct {
	Time string
	Weather_code int

	Temperature_2m float64
	Relative_humidity_2m float64
	Dew_point_2m float64
	Cloud_cover float64
	Wind_speed_10m float64
	Wind_direction_10m float64
	Wind_gusts_10m float64
	Precipitation float64
	Precipitation_probability float64
	Snowfall float64
}
type weatherCurrU struct {
	Time string
	Weather_code string

	Temperature_2m string
	Relative_humidity_2m string
	Dew_point_2m string
	Cloud_cover string
	Wind_speed_10m string
	Wind_direction_10m string
	Wind_gusts_10m string
	Precipitation string
	Precipitation_probability string
	Snowfall string
}

type weatherResp struct {
	weatherHdr
	Current         weatherCurr    `json:"current"`
	CurrentU        weatherCurrU   `json:"current_units"`
	Hourly          weatherHour    `json:"hourly"`
	HourlyU         weatherHourU   `json:"hourly_units"`
	Daily           weatherDaily   `json:"daily"`
	DailyU          weatherDailyU  `json:"daily_units"`
}

// TODO len(weather) > 1 => warn + write json to file

var weatherDateTimeFormat string = "2006-01-02T15:04"

func (o *weatherResp) String() string {
	builder := strings.Builder{}
	date, err := time.Parse(weatherDateTimeFormat, o.Current.Time)
	if err != nil {
		// TODO log warning about invalid timestamp
	}
	tz, err := time.LoadLocation(o.Timezone)
	if err == nil {
		date = date.In(tz)
	} else {
		// TODO log warning about invalid TZ, default to UTC
	}

	builder.WriteString(weatherCCs[uint(o.Current.Weather_code)].icon)
	// builder.WriteString(date.Format(" Mon 02.01 15:04:05 "))
	builder.WriteString(date.Format(" Mon 02.01  "))
	builder.WriteString(strconv.FormatInt(int64(o.Current.Temperature_2m), 10))
	builder.WriteString(o.CurrentU.Temperature_2m)
	builder.WriteRune('\n')

	builder.WriteString("Humidity: ")
	builder.WriteString(strconv.FormatFloat(float64(o.Current.Relative_humidity_2m), 'f', 0, 32))
	builder.WriteRune('%')
	builder.WriteRune(' ')
	// TODO split into rain and shower?
	if o.Current.Precipitation != 0 {
		builder.WriteString(" P:")
		builder.WriteString(strconv.FormatFloat(float64(o.Current.Precipitation), 'f', 1, 32))
		builder.WriteString(o.CurrentU.Precipitation)
	}
	if o.Current.Snowfall != 0 {
		builder.WriteString(" S:")
		builder.WriteString(strconv.FormatFloat(float64(o.Current.Snowfall), 'f', 1, 32))
		builder.WriteString(o.CurrentU.Snowfall)
	}
	builder.WriteRune('\n')

	builder.WriteString("Clouds: ")
	builder.WriteString(strconv.FormatFloat(float64(o.Current.Cloud_cover), 'f', 0, 32))
	builder.WriteString(o.CurrentU.Cloud_cover)
	builder.WriteRune('\n')

	builder.WriteString("Wind: ")
	builder.WriteString(strconv.FormatFloat(float64(o.Current.Wind_speed_10m), 'f', 0, 32))
	builder.WriteString(o.CurrentU.Wind_speed_10m)
	builder.WriteString(" ")
	builder.WriteString(wind[int(math.Round(o.Current.Wind_direction_10m/360*float64(len(wind))))])
	builder.WriteRune('\n')

	// daily (max 7)
	builder.WriteRune('\n')
	for i := 0; i < 7; i++ {
		date, err := time.Parse(weatherDateTimeFormat, o.Daily.Time[i])
		if err != nil {
			// TODO log warning about invalid timestamp
		}
		date = date.In(tz)

		builder.WriteString(weatherCCs[uint(o.Daily.Weather_code[i])].icon)
		builder.WriteString(date.Format(" Mon 02.01  "))
		builder.WriteString(strconv.FormatFloat(float64(o.Daily.Temperature_2m_min[i]), 'f', 0, 32))
		builder.WriteString(o.DailyU.Temperature_2m_min)
		builder.WriteString(" - ")
		builder.WriteString(strconv.FormatFloat(float64(o.Daily.Temperature_2m_max[i]), 'f', 0, 32))
		builder.WriteString(o.DailyU.Temperature_2m_max)
		if len(o.Daily.Precipitation_sum) > i {
			if o.Daily.Precipitation_sum[i] != 0 {
				builder.WriteString(" R:")
				builder.WriteString(strconv.FormatFloat(o.Daily.Precipitation_sum[i], 'f', 1, 32))
				builder.WriteString(o.DailyU.Precipitation_sum)
			}
		}
		if len(o.Daily.Snowfall_sum) > i {
			if o.Daily.Snowfall_sum[i] != 0 {
				builder.WriteString(" S:")
				builder.WriteString(strconv.FormatFloat(o.Daily.Snowfall_sum[i], 'f', 1, 32))
				builder.WriteString(o.DailyU.Snowfall_sum)
			}
		}
		builder.WriteRune('\n')
	}

	// hourly (max 5)
	builder.WriteRune('\n')
	for i := 0; i < 5; i++ {
		date, err := time.Parse(weatherDateTimeFormat, o.Hourly.Time[i])
		if err != nil {
			// TODO log warning about invalid timestamp
		}
		date = date.In(tz)

		builder.WriteString(weatherCCs[uint(o.Hourly.Weather_code[i])].icon)
		builder.WriteString(date.Format(" 15:04 02.01  "))
		builder.WriteString(strconv.FormatFloat(float64(o.Hourly.Temperature_2m[i]), 'f', 0, 32))
		builder.WriteString(o.HourlyU.Temperature_2m)
		if o.Hourly.Precipitation[i] != 0 {
			builder.WriteString(" R:")
			builder.WriteString(strconv.FormatFloat(o.Hourly.Precipitation[i], 'f', 1, 32))
			builder.WriteString(o.HourlyU.Precipitation)
		}
		if o.Hourly.Snowfall[i] != 0 {
			builder.WriteString(" S:")
			builder.WriteString(strconv.FormatFloat(o.Hourly.Snowfall[i], 'f', 1, 32))
			builder.WriteString(o.HourlyU.Snowfall)
		}
		builder.WriteRune('\n')
	}

	builder.WriteRune('\n')
	builder.WriteString("Quelle: open-meteo.com")
	return builder.String()
}
