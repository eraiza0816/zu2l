package presenter

import (
	"time" // time.Time を使うためインポート
	"github.com/eraiza0816/zu2l/internal/models"
)

// Presenter は API 結果を表示するためのインターフェースを定義します。
// 実装は異なる出力形式（例: テーブル、JSON）を扱います。
type Presenter interface {
	// PresentPainStatus は痛み予報ステータスを表示します。
	PresentPainStatus(data models.GetPainStatusResponse) error

	// PresentWeatherPoint は地点検索結果を表示します。
	// kata: カタカナ名を含めるかどうか。
	// keyword: 使用された検索キーワード。
	PresentWeatherPoint(data models.GetWeatherPointResponse, kata bool, keyword string) error

	// PresentWeatherStatus は詳細な気象状況を表示します。
	// dayOffset: 今日からのオフセット (-1:昨日, 0:今日, 1:明日, 2:明後日)。
	// dayName: 日付の名前 ("yesterday", "today" など)。
	PresentWeatherStatus(data models.GetWeatherStatusResponse, dayOffset int, dayName string) error

	// PresentOtenkiASP は Otenki ASP の気象情報を表示します。
	// targetDates: 表示対象の特定の日付のスライス。
	// cityName: 都市名。
	// cityCode: 都市コード。
	PresentOtenkiASP(data models.GetOtenkiASPResponse, targetDates []time.Time, cityName, cityCode string) error
}
