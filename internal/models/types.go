package models

import (
	"fmt"
	"regexp"
	"time"
)

// --- Weather Point API Structures ---

// WeatherPoint は気象観測地点を表すエンティティです。
type WeatherPoint struct {
	CityCode string `json:"city_code"` // 5桁の都市コード
	NameKata string `json:"name_kata"` // 地点名の半角カタカナ
	Name     string `json:"name"`      // 地点名 (漢字など)
}

// Validate は WeatherPoint のフィールドが有効かどうかを検証します。
func (w *WeatherPoint) Validate() error {
	// CityCode は5桁の数字である必要があります。
	// 注: 元の正規表現 `^\\d{5}$` はGoでは `^\d{5}$` が正しいようです。
	if matched, _ := regexp.MatchString(`^\d{5}$`, w.CityCode); !matched {
		return fmt.Errorf("CityCode は5桁の数字である必要があります、取得値: %s", w.CityCode)
	}
	// NameKata は半角カタカナである必要があります。
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
	Result WeatherPoints `json:"result"` // ネストされた地点情報
}

// --- Pain Status API Structures ---

// GetPainStatus は痛み予報ステータスデータを表すエンティティです。
type GetPainStatus struct {
	AreaName    string  `json:"area_name"`    // 地域名
	TimeStart   string  `json:"time_start"`   // 開始時刻 (例: "00") - フォーマットが一定なら time.Time の使用を検討
	TimeEnd     string  `json:"time_end"`     // 終了時刻 (例: "03") - フォーマットが一定なら time.Time の使用を検討
	RateNormal  float64 `json:"rate_0"`       // 通常レベルの割合
	RateLittle  float64 `json:"rate_1"`       // やや注意レベルの割合 (APIドキュメントにはないがレスポンスに含まれる?)
	RatePainful float64 `json:"rate_2"`       // 注意レベルの割合 (JSONキーは rate_2)
	RateBad     float64 `json:"rate_3"`       // 警戒レベルの割合 (JSONキーは rate_3)
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
	PainnoterateStatus GetPainStatus `json:"painnoterate_status"` // 痛み予報ステータス
}

// --- Weather Status API Structures ---

// WeatherStatusByTime は特定の時刻における気象状況を表すエンティティです。
type WeatherStatusByTime struct {
	Time          string            `json:"time"`           // 時刻 (例: "0") - JSONに合わせるため int から string に変更
	Weather       WeatherEnum       `json:"weather"`        // 天気コード (WeatherEnum を使用)
	Temp          *string           `json:"temp"`           // 気温 (例: "15.3") - JSONに合わせるため *float64 から *string に変更 (nullの場合あり)
	Pressure      string            `json:"pressure"`       // 気圧 (例: "1007.5") - JSONに合わせるため float64 から string に変更
	PressureLevel PressureLevelEnum `json:"pressure_level"` // 気圧レベル (PressureLevelEnum を使用)
}

// Validate は WeatherStatusByTime のフィールドが有効かどうかを検証します。
// 注: Time の型が string に変更されたため、時刻の検証は削除されました。必要に応じてパース/検証を追加してください。
func (w *WeatherStatusByTime) Validate() error {
	// 必要であれば検証ロジックを追加 (例: Time が 0-23 の整数にパースできるかチェック)
	return nil
}

// GetWeatherStatusResponse は気象状況API (/getweatherstatus) のレスポンス構造体 (集約ルート) です。
type GetWeatherStatusResponse struct {
	PlaceName        string                `json:"place_name"`       // 地点名
	PlaceID          string                `json:"place_id"`         // 3桁の地点ID
	PrefecturesID    AreaEnum              `json:"prefectures_id"`   // 都道府県コード (AreaEnum を使用)
	DateTime         APIDateTime           `json:"dateTime"`         // 基準日時 (APIDateTime を使用)
	Yesterday        []WeatherStatusByTime `json:"yesterday"`        // 昨日の天気
	Today            []WeatherStatusByTime `json:"today"`            // 今日の天気
	Tomorrow         []WeatherStatusByTime `json:"tomorrow"`         // 明日の天気
	DayAfterTomorrow []WeatherStatusByTime `json:"dayaftertomorrow"` // 明後日の天気
}

// Validate は GetWeatherStatusResponse の PlaceID フィールドが有効かどうかを検証します。
func (g *GetWeatherStatusResponse) Validate() error {
	// PlaceID は3桁の数字である必要があります。
	// 注: 元の正規表現 `^\\d{3}$` はGoでは `^\d{3}$` が正しいようです。
	if matched, _ := regexp.MatchString(`^\d{3}$`, g.PlaceID); !matched {
		return fmt.Errorf("PlaceID は3桁の数字である必要があります、取得値: %s", g.PlaceID)
	}
	return nil
}

// --- Otenki ASP API Structures (Raw and Processed) ---

// RawHead は Otenki ASP API の生レスポンスの head 部分を表します。
type RawHead struct {
	ContentsID string      `json:"contentsId"` // コンテンツID
	Title      string      `json:"title"`      // タイトル
	DateTime   APIDateTime `json:"dateTime"`   // 基準日時
	Status     string      `json:"status"`     // ステータス
}

// RawProperty は Otenki ASP API の生レスポンス内の property 配列を表します。
// 構造が完全には定義/一貫していないため、interface{} を使用します。
type RawProperty struct {
	Property []interface{} `json:"property"` // プロパティの配列 (型は混在する可能性がある)
}

// RawRecord は Otenki ASP API の生レスポンス内の record 配列を表します。
type RawRecord struct {
	Record []RawProperty `json:"record"` // レコード (プロパティの配列) の配列
}

// RawElement は Otenki ASP API の生レスポンス内の element 配列を表します。
type RawElement struct {
	Element []RawRecord `json:"element"` // 要素 (レコードの配列) の配列
}

// RawBody は Otenki ASP API の生レスポンスの body 部分を表します。
type RawBody struct {
	Location RawElement `json:"location"` // 地点情報 (要素を含む)
}

// GetOtenkiASPRawResponse は Otenki ASP API からの生のレスポンス構造体全体を表します。
// パース処理の中間段階で使用されます。
type GetOtenkiASPRawResponse struct {
	Head RawHead `json:"head"` // ヘッダー部
	Body RawBody `json:"body"` // ボディ部
}

// Element は Otenki ASP レスポンスから処理・整形された要素を表すエンティティです。
type Element struct {
	ContentID string                 `json:"content_id"` // コンテンツID (例: "day_tenki")
	Title     string                 `json:"title"`      // タイトル (例: "天気")
	Records   map[time.Time]interface{} `json:"records"`    // 時刻をキーとし、対応する値を保持するマップ (値の型は柔軟性のため interface{})
}

// GetOtenkiASPResponse は Otenki ASP API の処理・整形後のレスポンス構造体 (集約ルート) です。
type GetOtenkiASPResponse struct {
	Status   string      `json:"status"`     // ステータス (RawHead からコピー)
	DateTime APIDateTime `json:"date_time"`  // 基準日時 (RawHead からコピー)
	Elements []Element   `json:"elements"`   // 処理済みの要素のリスト
}

// Validate は GetOtenkiASPResponse の検証を行います (現在は元の実装に従い何もしません)。
func (g *GetOtenkiASPResponse) Validate() error {
	return nil
}

// --- Common Structures ---

// ErrorResponse は汎用的な API エラーレスポンスを表します。
type ErrorResponse struct {
	ErrorCode    int    `json:"error_code"`    // エラーコード
	ErrorMessage string `json:"error_message"` // エラーメッセージ
}

// SetWeatherPointResponse は地点設定API (/setweatherpoint) のレスポンス構造体です。
type SetWeatherPointResponse struct {
	Response string `json:"response"` // レスポンス文字列 (通常 "ok")
}
