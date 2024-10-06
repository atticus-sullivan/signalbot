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

var icons map[string]string = map[string]string{
	"thunderstorm": "\U000026c8\ufe0f", // '⛈️', // 🌩️
	"drizzle":      "\U0001f327\ufe0f", // '🌧️',
	"rain":         "\U0001f326\ufe0f", // '🌦️',
	"snow":         "\U00002744\ufe0f", // '❄️', // 🌨️
	"showerRain":   "\U0001f327\ufe0f", // '🌧',
	"fog":          "\U0001f32b\ufe0f", // '🌫️',
	"clear":        "\U00002600\ufe0f", // '☀️',

	// all the same since no better icon found
	"cloudsFew":   "\U00002601\ufe0f", // '☁️',
	"cloudsScat":  "\U00002601\ufe0f", // '☁️',
	"cloudsBrok":  "\U00002601\ufe0f", // '☁️',
	"cloudsOverc": "\U00002601\ufe0f", // '☁️'
}

var langs map[string]bool = map[string]bool{
	"af":    true, // Afrikaans
	"al":    true, // Albanian
	"ar":    true, // Arabic
	"az":    true, // Azerbaijani
	"bg":    true, // Bulgarian
	"ca":    true, // Catalan
	"cz":    true, // Czech
	"da":    true, // Danish
	"de":    true, // German
	"el":    true, // Greek
	"en":    true, // English
	"eu":    true, // Basque
	"fa":    true, // Persian (Farsi)
	"fi":    true, // Finnish
	"fr":    true, // French
	"gl":    true, // Galician
	"he":    true, // Hebrew
	"hi":    true, // Hindi
	"hr":    true, // Croatian
	"hu":    true, // Hungarian
	"id":    true, // Indonesian
	"it":    true, // Italian
	"ja":    true, // Japanese
	"kr":    true, // Korean
	"la":    true, // Latvian
	"lt":    true, // Lithuanian
	"mk":    true, // Macedonian
	"no":    true, // Norwegian
	"nl":    true, // Dutch
	"pl":    true, // Polish
	"pt":    true, // Portuguese
	"pt_br": true, // Português Brasil
	"ro":    true, // Romanian
	"ru":    true, // Russian
	"sv":    true, //, se	Swedish
	"sk":    true, // Slovak
	"sl":    true, // Slovenian
	"sp":    true, //, es	Spanish
	"sr":    true, // Serbian
	"th":    true, // Thai
	"tr":    true, // Turkish
	"ua":    true, //, uk Ukrainian
	"vi":    true, // Vietnamese
	"zh_cn": true, // Chinese Simplified
	"zh_tw": true, // Chinese Traditional
	"zu":    true, // Zulu
}

var wind []string = []string{
	"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NW", "NNW", "N",
}

type weatherTextIcon struct {
	text string
	icon string
}

var weatherCCs map[uint]weatherTextIcon = map[uint]weatherTextIcon{
	0: {text: "clear sky", icon: icons["clear"]},

	1: {text: "few clouds: 11-25%", icon: icons["cloudsFew"]},
	2: {text: "scattered clouds: 25-50%", icon: icons["cloudsScat"]},
	3: {text: "broken clouds: 51-84%", icon: icons["cloudsBrok"]},

	45: {text: "fog", icon: icons["fog"]},
	48: {text: "rime fog", icon: icons["fog"]},

	51: {text: "light intensity drizzle rain", icon: icons["drizzle"]},
	53: {text: "drizzle rain", icon: icons["drizzle"]},
	55: {text: "heavy intensity drizzle", icon: icons["drizzle"]},

	56: {text: "light freezing drizzle", icon: icons["drizzle"]},
	57: {text: "freezing drizzle", icon: icons["drizzle"]},

	61: {text: "light rain", icon: icons["rain"]},
	63: {text: "rain", icon: icons["rain"]},
	65: {text: "heavy rain", icon: icons["rain"]},

	66: {text: "light freezing rain", icon: icons["snow"]},
	67: {text: "freezing rain", icon: icons["snow"]},

	71: {text: "light snow", icon: icons["snow"]},
	73: {text: "Snow", icon: icons["snow"]},
	75: {text: "Heavy snow", icon: icons["snow"]},

	77: {text: "snow grains", icon: icons["snow"]},

	80: {text: "light shower rain", icon: icons["showerRain"]},
	81: {text: "shower rain", icon: icons["showerRain"]},
	82: {text: "heavy shower rain", icon: icons["showerRain"]},

	85: {text: "light snow showers", icon: icons["snow"]},
	86: {text: "heavy snow showers", icon: icons["snow"]},

	95: {text: "thunderstorm", icon: icons["thunderstorm"]},
	96: {text: "thunderstorm with slight hail", icon: icons["thunderstorm"]},
	99: {text: "thunderstorm with heavy hail", icon: icons["thunderstorm"]},
}
