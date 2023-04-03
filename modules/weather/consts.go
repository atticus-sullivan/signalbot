package weather

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
	200: {text: "Thunderstorm with light rain", icon: icons["thunderstorm"]},
	201: {text: "Thunderstorm with rain", icon: icons["thunderstorm"]},
	202: {text: "Thunderstorm with heavy rain", icon: icons["thunderstorm"]},
	210: {text: "light thunderstorm", icon: icons["thunderstorm"]},
	211: {text: "thunderstorm", icon: icons["thunderstorm"]},
	212: {text: "heavy thunderstorm ", icon: icons["thunderstorm"]},
	221: {text: "ragged thunderstorm", icon: icons["thunderstorm"]},
	230: {text: "thunderstorm with light drizzle", icon: icons["thunderstorm"]},
	231: {text: "thunderstorm with drizzle", icon: icons["thunderstorm"]},
	232: {text: "thunderstorm with heavy drizzle", icon: icons["thunderstorm"]},

	300: {text: "light intensity drizzle", icon: icons["drizzle"]},
	301: {text: "drizzle", icon: icons["drizzle"]},
	302: {text: "heavy intensity drizzle", icon: icons["drizzle"]},
	310: {text: "light intensity drizzle rain", icon: icons["drizzle"]},
	311: {text: "drizzle rain", icon: icons["drizzle"]},
	312: {text: "heavy intensity drizzle rain", icon: icons["drizzle"]},
	313: {text: "shower rain and drizzle", icon: icons["drizzle"]},
	314: {text: "heavy shower rain and drizzle", icon: icons["drizzle"]},
	321: {text: "shower drizzle", icon: icons["drizzle"]},

	500: {text: "light rain", icon: icons["rain"]},
	501: {text: "moderate rain", icon: icons["rain"]},
	502: {text: "heavy intensity rain", icon: icons["rain"]},
	503: {text: "very heavy rain", icon: icons["rain"]},
	504: {text: "extreme rain", icon: icons["rain"]},
	511: {text: "freezing rain", icon: icons["snow"]},
	520: {text: "light intensity shower rain", icon: icons["showerRain"]},
	521: {text: "shower rain", icon: icons["showerRain"]},
	522: {text: "heavy intensity shower rain", icon: icons["showerRain"]},
	531: {text: "ragged shower rain", icon: icons["showerRain"]},

	600: {text: "light snow", icon: icons["snow"]},
	601: {text: "Snow", icon: icons["snow"]},
	602: {text: "Heavy snow", icon: icons["snow"]},
	611: {text: "Sleet", icon: icons["snow"]},
	612: {text: "Light shower sleet", icon: icons["snow"]},
	613: {text: "Shower sleet", icon: icons["snow"]},
	615: {text: "Light rain and snow", icon: icons["snow"]},
	616: {text: "Rain and snow", icon: icons["snow"]},
	620: {text: "Light shower snow", icon: icons["snow"]},
	621: {text: "Shower snow", icon: icons["snow"]},
	622: {text: "Heavy shower snow", icon: icons["snow"]},

	701: {text: "mist", icon: icons["fog"]},
	711: {text: "Smoke", icon: icons["fog"]},
	721: {text: "Haze", icon: icons["fog"]},
	731: {text: "sand/ dust whirls", icon: icons["fog"]},
	741: {text: "fog", icon: icons["fog"]},
	751: {text: "sand", icon: icons["fog"]},
	761: {text: "dust", icon: icons["fog"]},
	762: {text: "volcanic ash", icon: icons["fog"]},
	771: {text: "squalls", icon: icons["fog"]},
	781: {text: "Tornado", icon: icons["fog"]},

	800: {text: "clear sky", icon: icons["clear"]},

	801: {text: "few clouds: 11-25%", icon: icons["cloudsFew"]},
	802: {text: "scattered clouds: 25-50%", icon: icons["cloudsScat"]},
	803: {text: "broken clouds: 51-84%", icon: icons["cloudsBrok"]},
	804: {text: "overcast clouds: 85-100%", icon: icons["cloudsOverc"]},
}
