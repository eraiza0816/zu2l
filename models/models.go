package models

import (
	"fmt"
	"regexp"
	"strings" // Added for custom UnmarshalJSON
	"time"
)

type APIDateTime struct {
	time.Time
}

const apiDateTimeLayout = "2006-01-02 15"

// UnmarshalJSON implements the json.Unmarshaler interface for APIDateTime.
func (adt *APIDateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" {
		adt.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(apiDateTimeLayout, s)
	if err != nil {
		return fmt.Errorf("failed to parse APIDateTime %q: %w", s, err)
	}
	adt.Time = t
	return nil
}

// MarshalJSON implements the json.Marshaler interface for APIDateTime.
func (adt APIDateTime) MarshalJSON() ([]byte, error) {
	if adt.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + adt.Time.Format(apiDateTimeLayout) + `"`), nil
}


var WeatherEmojiMap = map[int]string{
	100: "☀",
	200: "☁",
	300: "☔",
	400: "🌨", // 雪
}

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

type WeatherPoint struct {
	CityCode string `json:"city_code"`
	NameKata string `json:"name_kata"`
	Name     string `json:"name"`
}

func (w *WeatherPoint) Validate() error {
	if matched, _ := regexp.MatchString(`^\\d{5}$`, w.CityCode); !matched {
		return fmt.Errorf("CityCode must be a 5-digit number")
	}

	if matched, _ := regexp.MatchString(`^[\\uff61-\\uff9f]+$`, w.NameKata); !matched {
		return fmt.Errorf("NameKata must be in half-width katakana")
	}
	return nil
}

type WeatherPoints struct {
	Root []WeatherPoint `json:"root"`
}

type GetWeatherPointResponse struct {
	Result WeatherPoints `json:"result"`
}

type AreaEnum string

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
	// "京都": "26", "キョウト": "26", // API未対応？
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

type GetPainStatus struct {
	AreaName      string        `json:"area_name"`
	TimeStart     string        `json:"time_start"`
	TimeEnd       string        `json:"time_end"`
	RateNormal    float64       `json:"rate_0"`
	RateLittle    float64       `json:"rate_1"`
	RatePainful   float64       `json:"rate_2"`
	RateBad       float64       `json:"rate_3"`
}

func (g *GetPainStatus) Validate() error {
	if g.RateNormal < 0 {
		return fmt.Errorf("RateNormal must be positive")
	}
	if g.RateLittle < 0 {
		return fmt.Errorf("RateLittle must be positive")
	}
	if g.RatePainful < 0 {
		return fmt.Errorf("RatePainful must be positive")
	}
	if g.RateBad < 0 {
		return fmt.Errorf("RateBad must be positive")
	}
	return nil
}

type GetPainStatusResponse struct {
	PainnoterateStatus GetPainStatus `json:"painnoterate_status"`
}

type PressureLevelEnum string

const (
	Normal      PressureLevelEnum = "0"
	SlightAlert PressureLevelEnum = "2"
	Caution     PressureLevelEnum = "3"
	Alert       PressureLevelEnum = "4"
	SevereAlert PressureLevelEnum = "5"
)

type WeatherEnum string

const (
	Sunny          WeatherEnum = "100"
	SunnyCloudy    WeatherEnum = "101"
	SunnySometimes WeatherEnum = "102"
	SunnyRainy     WeatherEnum = "103"
	SunnySnowy     WeatherEnum = "105"
	CloudySunny    WeatherEnum = "111"
	RainySunny     WeatherEnum = "114"
	SnowySunny     WeatherEnum = "116"

	Cloudy         WeatherEnum = "200"
	CloudySunny2   WeatherEnum = "201" // くもり時々晴れ
	CloudySometimes WeatherEnum = "202" // くもり一時晴れ
	CloudyRainy    WeatherEnum = "203" // くもり時々雨
	CloudyRainyTemp WeatherEnum = "204" // くもり一時雨
	CloudySnowy    WeatherEnum = "205" // くもり時々雪
	CloudySnowyTemp WeatherEnum = "206" // くもり一時雪
	CloudyAfterSunny WeatherEnum = "211" // くもりのち晴れ
	RainyCloudy    WeatherEnum = "214" // くもりのち雨
	SnowyCloudy    WeatherEnum = "216" // くもりのち雪

	Rain           WeatherEnum = "300"
	RainySunny2    WeatherEnum = "301" // 雨時々晴れ
	RainySometimes WeatherEnum = "302" // 雨一時晴れ
	RainyCloudy2   WeatherEnum = "303" // 雨時々くもり
	RainyCloudyTemp WeatherEnum = "304" // 雨一時くもり
	RainySnowy     WeatherEnum = "305" // 雨時々雪
	RainySnowyTemp WeatherEnum = "306" // 雨一時雪
	RainyAfterSunny WeatherEnum = "311" // 雨のち晴れ
	RainyAfterCloudy WeatherEnum = "313" // 雨のちくもり
	SnowyRainy2    WeatherEnum = "315" // 雨のち雪

	Snow           WeatherEnum = "400"
	SnowySunny2    WeatherEnum = "401" // 雪時々晴れ
	SnowySometimes WeatherEnum = "402" // 雪一時晴れ
	SnowyCloudy2   WeatherEnum = "403" // 雪時々くもり
	SnowyCloudyTemp WeatherEnum = "404" // 雪一時くもり
	SnowyRainy     WeatherEnum = "405" // 雪時々雨
	SnowyRainyTemp WeatherEnum = "406" // 雪一時雨
	SnowyAfterSunny WeatherEnum = "411" // 雪のち晴れ
	RainySnowy3    WeatherEnum = "414" // 雪のち雨
	SnowyAfterCloudy WeatherEnum = "416"
)

type WeatherStatusByTime struct {
	Time          string        `json:"time"` // Changed from int to string
	Weather       WeatherEnum   `json:"weather"`
	Temp          *float64      `json:"temp"`
	Pressure      float64       `json:"pressure"`
	PressureLevel PressureLevelEnum `json:"pressure_level"`
}

// Validate method removed as Time is now a string and API is assumed to provide valid "0"-"23"

type GetWeatherStatusResponse struct {
	PlaceName      string              `json:"place_name"`
	PlaceID        string              `json:"place_id"`
	PrefecturesID  AreaEnum            `json:"prefectures_id"`
	DateTime       APIDateTime         `json:"dateTime"` // Changed type
	Yesterday      []WeatherStatusByTime `json:"yesterday"`
	Today          []WeatherStatusByTime `json:"today"`
	Tomorrow       []WeatherStatusByTime `json:"tomorrow"`
	DayAfterTomorrow []WeatherStatusByTime `json:"dayaftertomorrow"`
}

func (g *GetWeatherStatusResponse) Validate() error {
	if matched, _ := regexp.MatchString(`^\\d{3}$`, g.PlaceID); !matched {
		return fmt.Errorf("PlaceID must be a 3-digit number")
	}
	return nil
}

type RawHead struct {
	ContentsID string      `json:"contentsId"`
	Title      string      `json:"title"`
	DateTime   APIDateTime `json:"dateTime"` // Changed type
	Status     string      `json:"status"`
}

type RawProperty struct {
	Property []interface{} `json:"property"`
}

type RawRecord struct {
	Record []RawProperty `json:"record"`
}

type RawElement struct {
	Element []RawRecord `json:"element"`
}

type RawBody struct {
	Location RawElement `json:"location"`
}

type GetOtenkiASPRawResponse struct {
	Head RawHead `json:"head"`
	Body RawBody `json:"body"`
}

type Element struct {
	ContentID string                 `json:"content_id"`
	Title     string                 `json:"title"`
	Records   map[time.Time]interface{} `json:"records"`
}

type GetOtenkiASPResponse struct {
	Status   string      `json:"status"`
	DateTime APIDateTime `json:"date_time"` // Changed type
	Elements []Element `json:"elements"`
}

func (g *GetOtenkiASPResponse) Validate() error {
	return nil
}

type ErrorResponse struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

type SetWeatherPointResponse struct {
	Response string `json:"response"`
}
