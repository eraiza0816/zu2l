package models

// WeatherEmojiMap maps weather codes (simplified) to emojis.
var WeatherEmojiMap = map[int]string{
	100: "☀", // Sunny
	200: "☁", // Cloudy
	300: "☔", // Rainy
	400: "🌨", // Snowy
}

// ConfirmedOtenkiAspCityCodeMap maps confirmed city codes for Otenki ASP to their names.
var ConfirmedOtenkiAspCityCodeMap = map[string]string{
	"01101": "札幌",
	"04101": "仙台",
	"13101": "東京",
	"15103": "新潟",
	"17201": "金沢",
	"23106": "名古屋",
	"27128": "大阪",
	"34101": "広島",
	"39201": "高知",
	"40133": "福岡",
	"47201": "那覇",
}

// AreaEnum represents prefecture codes.
type AreaEnum string

// Constants for AreaEnum representing prefecture codes.
const (
	Hokkaido  AreaEnum = "01"
	Aomori    AreaEnum = "02"
	Iwate     AreaEnum = "03"
	Miyagi    AreaEnum = "04"
	Akita     AreaEnum = "05"
	Yamagata  AreaEnum = "06"
	Fukushima AreaEnum = "07"
	Ibaraki   AreaEnum = "08"
	Tochigi   AreaEnum = "09"
	Gunma     AreaEnum = "10"
	Saitama   AreaEnum = "11"
	Chiba     AreaEnum = "12"
	Tokyo     AreaEnum = "13"
	Kanagawa  AreaEnum = "14"
	Niigata   AreaEnum = "15"
	Toyama    AreaEnum = "16"
	Ishikawa  AreaEnum = "17"
	Fukui     AreaEnum = "18"
	Yamanashi AreaEnum = "19"
	Nagano    AreaEnum = "20"
	Gifu      AreaEnum = "21"
	Shizuoka  AreaEnum = "22"
	Aichi     AreaEnum = "23"
	Mie       AreaEnum = "24"
	Shiga     AreaEnum = "25"
	// Kyoto     AreaEnum = "26" // API not supported?
	Osaka     AreaEnum = "27"
	Hyogo     AreaEnum = "28"
	Nara      AreaEnum = "29"
	Wakayama  AreaEnum = "30"
	Tottori   AreaEnum = "31"
	Shimane   AreaEnum = "32"
	Okayama   AreaEnum = "33"
	Hiroshima AreaEnum = "34"
	Yamaguchi AreaEnum = "35"
	Tokushima AreaEnum = "36"
	Kagawa    AreaEnum = "37"
	Ehime     AreaEnum = "38"
	Kochi     AreaEnum = "39"
	Fukuoka   AreaEnum = "40"
	Saga      AreaEnum = "41"
	Nagasaki  AreaEnum = "42"
	Kumamoto  AreaEnum = "43"
	Oita      AreaEnum = "44"
	Miyazaki  AreaEnum = "45"
	Kagoshima AreaEnum = "46"
	Okinawa   AreaEnum = "47"
)

// AreaCodeMap maps prefecture names (Kanji and Katakana) to their codes.
var AreaCodeMap = map[string]string{
	"北海道": "01", "ホッカイドウ": "01",
	"青森": "02", "アオモリ": "02",
	"岩手": "03", "イワテ": "03",
	"宮城": "04", "ミヤギ": "04",
	"秋田": "05", "アキタ": "05",
	"山形": "06", "ヤマガタ": "06",
	"福島": "07", "フクシマ": "07",
	"茨城": "08", "イバラキ": "08",
	"栃木": "09", "トチギ": "09",
	"群馬": "10", "グンマ": "10",
	"埼玉": "11", "サイタマ": "11",
	"千葉": "12", "チバ": "12",
	"東京": "13", "トウキョウ": "13",
	"神奈川": "14", "カナガワ": "14",
	"新潟": "15", "ニイガタ": "15",
	"富山": "16", "トヤマ": "16",
	"石川": "17", "イシカワ": "17",
	"福井": "18", "フクイ": "18",
	"山梨": "19", "ヤマナシ": "19",
	"長野": "20", "ナガノ": "20",
	"岐阜": "21", "ギフ": "21",
	"静岡": "22", "シズオカ": "22",
	"愛知": "23", "アイチ": "23",
	"三重": "24", "ミエ": "24",
	"滋賀": "25", "シガ": "25",
	// "京都": "26", "キョウト": "26", // API not supported?
	"大阪": "27", "オオサカ": "27",
	"兵庫": "28", "ヒョウゴ": "28",
	"奈良": "29", "ナラ": "29",
	"和歌山": "30", "ワカヤマ": "30",
	"鳥取": "31", "トットリ": "31",
	"島根": "32", "シマネ": "32",
	"岡山": "33", "オカヤマ": "33",
	"広島": "34", "ヒロシマ": "34",
	"山口": "35", "ヤマグチ": "35",
	"徳島": "36", "トクシマ": "36",
	"香川": "37", "カガワ": "37",
	"愛媛": "38", "エヒメ": "38",
	"高知": "39", "コウチ": "39",
	"福岡": "40", "フクオカ": "40",
	"佐賀": "41", "サガ": "41",
	"長崎": "42", "ナガサキ": "42",
	"熊本": "43", "クマモト": "43",
	"大分": "44", "オオイタ": "44",
	"宮崎": "45", "ミヤザキ": "45",
	"鹿児島": "46", "カゴシマ": "46",
	"沖縄": "47", "オキナワ": "47",
}

// PressureLevelEnum represents pressure level codes.
type PressureLevelEnum string

// Constants for PressureLevelEnum.
const (
	Normal      PressureLevelEnum = "0" // 通常
	SlightAlert PressureLevelEnum = "2" // やや注意
	Caution     PressureLevelEnum = "3" // 注意
	Alert       PressureLevelEnum = "4" // 警戒
	SevereAlert PressureLevelEnum = "5" // 厳重警戒
)

// WeatherEnum represents weather condition codes.
type WeatherEnum string

// Constants for WeatherEnum.
const (
	Sunny           WeatherEnum = "100" // 晴れ
	SunnyCloudy     WeatherEnum = "101" // 晴れ時々くもり
	SunnySometimes  WeatherEnum = "102" // 晴れ一時くもり
	SunnyRainy      WeatherEnum = "103" // 晴れ時々雨
	SunnySnowy      WeatherEnum = "105" // 晴れ時々雪
	CloudySunny     WeatherEnum = "111" // 晴れのちくもり
	RainySunny      WeatherEnum = "114" // 晴れのち雨
	SnowySunny      WeatherEnum = "116" // 晴れのち雪

	Cloudy          WeatherEnum = "200" // くもり
	CloudySunny2    WeatherEnum = "201" // くもり時々晴れ
	CloudySometimes WeatherEnum = "202" // くもり一時晴れ
	CloudyRainy     WeatherEnum = "203" // くもり時々雨
	CloudyRainyTemp WeatherEnum = "204" // くもり一時雨
	CloudySnowy     WeatherEnum = "205" // くもり時々雪
	CloudySnowyTemp WeatherEnum = "206" // くもり一時雪
	CloudyAfterSunny WeatherEnum = "211" // くもりのち晴れ
	RainyCloudy     WeatherEnum = "214" // くもりのち雨
	SnowyCloudy     WeatherEnum = "216" // くもりのち雪

	Rain            WeatherEnum = "300" // 雨
	RainySunny2     WeatherEnum = "301" // 雨時々晴れ
	RainySometimes  WeatherEnum = "302" // 雨一時晴れ
	RainyCloudy2    WeatherEnum = "303" // 雨時々くもり
	RainyCloudyTemp WeatherEnum = "304" // 雨一時くもり
	RainySnowy      WeatherEnum = "305" // 雨時々雪
	RainySnowyTemp  WeatherEnum = "306" // 雨一時雪
	RainyAfterSunny WeatherEnum = "311" // 雨のち晴れ
	RainyAfterCloudy WeatherEnum = "313" // 雨のちくもり
	SnowyRainy2     WeatherEnum = "315" // 雨のち雪

	Snow            WeatherEnum = "400" // 雪
	SnowySunny2     WeatherEnum = "401" // 雪時々晴れ
	SnowySometimes  WeatherEnum = "402" // 雪一時晴れ
	SnowyCloudy2    WeatherEnum = "403" // 雪時々くもり
	SnowyCloudyTemp WeatherEnum = "404" // 雪一時くもり
	SnowyRainy      WeatherEnum = "405" // 雪時々雨
	SnowyRainyTemp  WeatherEnum = "406" // 雪一時雨
	SnowyAfterSunny WeatherEnum = "411" // 雪のち晴れ
	RainySnowy3     WeatherEnum = "414" // 雪のち雨
	SnowyAfterCloudy WeatherEnum = "416" // 雪のちくもり
)
