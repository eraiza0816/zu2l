package models

import "fmt"

// WeatherEmojiMap は天気コード (簡略化) を絵文字にマッピングします。
var WeatherEmojiMap = map[int]string{
	100: "☀", // 晴れ
	200: "☁", // くもり
	300: "☔", // 雨
	400: "🌨", // 雪
}

// ConfirmedOtenkiAspCityCodeMap は Otenki ASP で確認済みの都市コードとその名称をマッピングします。
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

// AreaEnum は都道府県コードを表す Enum (値オブジェクト) です。
type AreaEnum string

const (
	Hokkaido  AreaEnum = "01" // 北海道
	Aomori    AreaEnum = "02" // 青森
	Iwate     AreaEnum = "03" // 岩手
	Miyagi    AreaEnum = "04" // 宮城
	Akita     AreaEnum = "05" // 秋田
	Yamagata  AreaEnum = "06" // 山形
	Fukushima AreaEnum = "07" // 福島
	Ibaraki   AreaEnum = "08" // 茨城
	Tochigi   AreaEnum = "09" // 栃木
	Gunma     AreaEnum = "10" // 群馬
	Saitama   AreaEnum = "11" // 埼玉
	Chiba     AreaEnum = "12" // 千葉
	Tokyo     AreaEnum = "13" // 東京
	Kanagawa  AreaEnum = "14" // 神奈川
	Niigata   AreaEnum = "15" // 新潟
	Toyama    AreaEnum = "16" // 富山
	Ishikawa  AreaEnum = "17" // 石川
	Fukui     AreaEnum = "18" // 福井
	Yamanashi AreaEnum = "19" // 山梨
	Nagano    AreaEnum = "20" // 長野
	Gifu      AreaEnum = "21" // 岐阜
	Shizuoka  AreaEnum = "22" // 静岡
	Aichi     AreaEnum = "23" // 愛知
	Mie       AreaEnum = "24" // 三重
	Shiga     AreaEnum = "25" // 滋賀
	// Kyoto     AreaEnum = "26" // 京都 - API非対応？
	Osaka     AreaEnum = "27" // 大阪
	Hyogo     AreaEnum = "28" // 兵庫
	Nara      AreaEnum = "29" // 奈良
	Wakayama  AreaEnum = "30" // 和歌山
	Tottori   AreaEnum = "31" // 鳥取
	Shimane   AreaEnum = "32" // 島根
	Okayama   AreaEnum = "33" // 岡山
	Hiroshima AreaEnum = "34" // 広島
	Yamaguchi AreaEnum = "35" // 山口
	Tokushima AreaEnum = "36" // 徳島
	Kagawa    AreaEnum = "37" // 香川
	Ehime     AreaEnum = "38" // 愛媛
	Kochi     AreaEnum = "39" // 高知
	Fukuoka   AreaEnum = "40" // 福岡
	Saga      AreaEnum = "41" // 佐賀
	Nagasaki  AreaEnum = "42" // 長崎
	Kumamoto  AreaEnum = "43" // 熊本
	Oita      AreaEnum = "44" // 大分
	Miyazaki  AreaEnum = "45" // 宮崎
	Kagoshima AreaEnum = "46" // 鹿児島
	Okinawa   AreaEnum = "47" // 沖縄
)

// AreaCodeMap は都道府県名 (漢字およびカタカナ) を都道府県コードにマッピングします。
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
	// "京都": "26", "キョウト": "26", // 京都 - API非対応？
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

// PressureLevelEnum は気圧レベルコードを表す Enum (値オブジェクト) です。
type PressureLevelEnum string

const (
	Normal      PressureLevelEnum = "0" // 通常
	SlightAlert PressureLevelEnum = "2" // やや注意
	Caution     PressureLevelEnum = "3" // 注意
	Alert       PressureLevelEnum = "4" // 警戒
	SevereAlert PressureLevelEnum = "5" // 厳重警戒
)

// String は PressureLevelEnum の文字列表現を返します。
func (p PressureLevelEnum) String() string {
	switch p {
	case Normal:
		return "通常"
	case SlightAlert:
		return "やや注意"
	case Caution:
		return "注意"
	case Alert:
		return "警戒"
	case SevereAlert:
		return "厳重警戒"
	default:
		return fmt.Sprintf("不明な気圧レベル(%s)", string(p))
	}
}

// WeatherEnum は天気コードを表す Enum (値オブジェクト) です。
type WeatherEnum string

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

// String は WeatherEnum の文字列表現を返します。
func (w WeatherEnum) String() string {
	switch w {
	case Sunny:
		return "晴れ"
	case SunnyCloudy:
		return "晴れ時々くもり"
	case SunnySometimes:
		return "晴れ一時くもり"
	case SunnyRainy:
		return "晴れ時々雨"
	case SunnySnowy:
		return "晴れ時々雪"
	case CloudySunny:
		return "晴れのちくもり"
	case RainySunny:
		return "晴れのち雨"
	case SnowySunny:
		return "晴れのち雪"
	case Cloudy:
		return "くもり"
	case CloudySunny2:
		return "くもり時々晴れ"
	case CloudySometimes:
		return "くもり一時晴れ"
	case CloudyRainy:
		return "くもり時々雨"
	case CloudyRainyTemp:
		return "くもり一時雨"
	case CloudySnowy:
		return "くもり時々雪"
	case CloudySnowyTemp:
		return "くもり一時雪"
	case CloudyAfterSunny:
		return "くもりのち晴れ"
	case RainyCloudy:
		return "くもりのち雨"
	case SnowyCloudy:
		return "くもりのち雪"
	case Rain:
		return "雨"
	case RainySunny2:
		return "雨時々晴れ"
	case RainySometimes:
		return "雨一時晴れ"
	case RainyCloudy2:
		return "雨時々くもり"
	case RainyCloudyTemp:
		return "雨一時くもり"
	case RainySnowy:
		return "雨時々雪"
	case RainySnowyTemp:
		return "雨一時雪"
	case RainyAfterSunny:
		return "雨のち晴れ"
	case RainyAfterCloudy:
		return "雨のちくもり"
	case SnowyRainy2:
		return "雨のち雪"
	case Snow:
		return "雪"
	case SnowySunny2:
		return "雪時々晴れ"
	case SnowySometimes:
		return "雪一時晴れ"
	case SnowyCloudy2:
		return "雪時々くもり"
	case SnowyCloudyTemp:
		return "雪一時くもり"
	case SnowyRainy:
		return "雪時々雨"
	case SnowyRainyTemp:
		return "雪一時雨"
	case SnowyAfterSunny:
		return "雪のち晴れ"
	case RainySnowy3:
		return "雪のち雨"
	case SnowyAfterCloudy:
		return "雪のちくもり"
	default:
		return fmt.Sprintf("不明な天気(%s)", string(w))
	}
}
