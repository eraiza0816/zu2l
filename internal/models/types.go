package models

import (
	"fmt"
	"regexp"
	"time"
)

// --- Weather Point API Structures ---

// WeatherPoint は気象観測地点を表すエンティティです。
type WeatherPoint struct {
	CityCode string `json:"city_code"`
	NameKata string `json:"name_kata"`
	Name     string `json:"name"`
}

// Validate は WeatherPoint のフィールドが有効かどうかを検証します。
func (w *WeatherPoint) Validate() error {
	// 注: 元の正規表現 `^\\d{5}$` はGoでは `^\d{5}$` が正しいようです。
	if matched, _ := regexp.MatchString(`^\d{5}$`, w.CityCode); !matched {
		return fmt.Errorf("CityCode は5桁の数字である必要があります、取得値: %s", w.CityCode)
	}
	// 注: 元の正規表現 `^[\\uff61-\\uff9f]+$` はGoでは `^[\uff61-\uff9f]+$` が正しいようです。
	if matched, _ := regexp.MatchString(`^[\uff61-\uff9f]+$`, w.NameKata); !matched {
		return fmt.Errorf("NameKata は半角カタカナである必要があります、取得値: %s", w.NameKata)
	}
	return nil
}

// WeatherPoints は WeatherPoint のコレクションです。
// GetWeatherPointResponse 内のネストされた "root" 配列に対応します。
type WeatherPoints struct {
	Root []WeatherPoint `json:"root"`
}

// GetWeatherPointResponse は地点検索API (/getweatherpoint) のレスポンス構造体 (集約ルート) です。
type GetWeatherPointResponse struct {
	Result WeatherPoints `json:"result"`
}

// --- Pain Status API Structures ---

// GetPainStatus は痛み予報ステータスデータを表すエンティティです。
type GetPainStatus struct {
	AreaName    string  `json:"area_name"`
	TimeStart   string  `json:"time_start"`   // 例: "00" - フォーマットが一定なら time.Time の使用を検討
	TimeEnd     string  `json:"time_end"`     // 例: "03" - フォーマットが一定なら time.Time の使用を検討
	RateNormal  float64 `json:"rate_0"`
	RateLittle  float64 `json:"rate_1"`       // APIドキュメントにはないがレスポンスに含まれる?
	RatePainful float64 `json:"rate_2"`       // JSONキーは rate_2
	RateBad     float64 `json:"rate_3"`       // JSONキーは rate_3
}

// Validate は GetPainStatus の割合フィールドが有効 (非負) かどうかを検証します。
func (g *GetPainStatus) Validate() error {
	if g.RateNormal < 0 {
		return fmt.Errorf("RateNormal は非負である必要があります、取得値: %f", g.RateNormal)
	}
	if g.RateLittle < 0 {
		return fmt.Errorf("RateLittle は非負である必要があります、取得値: %f", g.RateLittle)
	}
	if g.RatePainful < 0 {
		return fmt.Errorf("RatePainful は非負である必要があります、取得値: %f", g.RatePainful)
	}
	if g.RateBad < 0 {
		return fmt.Errorf("RateBad は非負である必要があります、取得値: %f", g.RateBad)
	}
	return nil
}

// GetPainStatusResponse は痛み予報API (/getpainstatus) のレスポンス構造体 (集約ルート) です。
type GetPainStatusResponse struct {
	PainnoterateStatus GetPainStatus `json:"painnoterate_status"`
}

// --- Weather Status API Structures ---

// WeatherStatusByTime は特定の時刻における気象状況を表すエンティティです。
type WeatherStatusByTime struct {
	Time          string            `json:"time"`           // 例: "0" - JSONに合わせるため int から string に変更
	Weather       WeatherEnum       `json:"weather"`
	Temp          *string           `json:"temp"`           // 例: "15.3" - JSONに合わせるため *float64 から *string に変更 (nullの場合あり)
	Pressure      string            `json:"pressure"`       // 例: "1007.5" - JSONに合わせるため float64 から string に変更
	PressureLevel PressureLevelEnum `json:"pressure_level"`
}

// Validate は WeatherStatusByTime のフィールドが有効かどうかを検証します。
// 注: Time の型が string に変更されたため、時刻の検証は削除されました。必要に応じてパース/検証を追加してください。
func (w *WeatherStatusByTime) Validate() error {
	// 必要であれば検証ロジックを追加 (例: Time が 0-23 の整数にパースできるかチェック)
	return nil
}

// GetWeatherStatusResponse は気象状況API (/getweatherstatus) のレスポンス構造体 (集約ルート) です。
type GetWeatherStatusResponse struct {
	PlaceName        string                `json:"place_name"`
	PlaceID          string                `json:"place_id"`
	PrefecturesID    AreaEnum              `json:"prefectures_id"`
	DateTime         APIDateTime           `json:"dateTime"`
	Yesterday        []WeatherStatusByTime `json:"yesterday"`
	Today            []WeatherStatusByTime `json:"today"`
	Tomorrow         []WeatherStatusByTime `json:"tomorrow"`
	DayAfterTomorrow []WeatherStatusByTime `json:"dayaftertomorrow"`
}

// Validate は GetWeatherStatusResponse の PlaceID フィールドが有効かどうかを検証します。
func (g *GetWeatherStatusResponse) Validate() error {
	// 注: 元の正規表現 `^\\d{3}$` はGoでは `^\d{3}$` が正しいようです。
	if matched, _ := regexp.MatchString(`^\d{3}$`, g.PlaceID); !matched {
		return fmt.Errorf("PlaceID は3桁の数字である必要があります、取得値: %s", g.PlaceID)
	}
	return nil
}

// --- Otenki ASP API Structures (Raw and Processed) ---

// RawHead は Otenki ASP API の生レスポンスの head 部分を表します。
type RawHead struct {
	ContentsID string      `json:"contentsId"`
	Title      string      `json:"title"`
	DateTime   APIDateTime `json:"dateTime"`
	Status     string      `json:"status"`
}

// RawProperty は Otenki ASP API の生レスポンス内の property 配列を表します。
// 構造が完全には定義/一貫していないため、interface{} を使用します。
type RawProperty struct {
	Property []interface{} `json:"property"` // 型は混在する可能性がある
}

// RawRecord は Otenki ASP API の生レスポンス内の record 配列を表します。
type RawRecord struct {
	Record []RawProperty `json:"record"`
}

// RawElement は Otenki ASP API の生レスポンス内の element 配列を表します。
type RawElement struct {
	Element []RawRecord `json:"element"`
}

// RawBody は Otenki ASP API の生レスポンスの body 部分を表します。
type RawBody struct {
	Location RawElement `json:"location"`
}

// GetOtenkiASPRawResponse は Otenki ASP API からの生のレスポンス構造体全体を表します。
// パース処理の中間段階で使用されます。
type GetOtenkiASPRawResponse struct {
	Head RawHead `json:"head"`
	Body RawBody `json:"body"`
}

// Element は Otenki ASP レスポンスから処理・整形された要素を表すエンティティです。
type Element struct {
	ContentID string                 `json:"content_id"` // 例: "day_tenki"
	Title     string                 `json:"title"`      // 例: "天気"
	Records   map[time.Time]interface{} `json:"records"`    // 時刻をキーとし、対応する値を保持するマップ (値の型は柔軟性のため interface{})
}

// GetOtenkiASPResponse は Otenki ASP API の処理・整形後のレスポンス構造体 (集約ルート) です。
type GetOtenkiASPResponse struct {
	Status   string      `json:"status"`
	DateTime APIDateTime `json:"date_time"`
	Elements []Element   `json:"elements"`
}

// Validate は GetOtenkiASPResponse の検証を行います。
func (g *GetOtenkiASPResponse) Validate() error {
	// 現在は特に検証しない
	return nil
}

// --- Common Structures ---

// ErrorResponse は汎用的な API エラーレスポンスを表します。
type ErrorResponse struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

// SetWeatherPointResponse は地点設定API (/setweatherpoint) のレスポンス構造体です。
type SetWeatherPointResponse struct {
	Response string `json:"response"` // 通常 "ok"
}
