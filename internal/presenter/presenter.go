package presenter

import (
	"time"
	"github.com/eraiza0816/zu2l/internal/models"
)

// Presenter は API 結果を表示するためのインターフェースを定義します。
// 実装は異なる出力形式（例: テーブル、JSON）を扱います。
type Presenter interface {
	// PresentPainStatus は痛み予報ステータスを表示します。
	PresentPainStatus(data models.GetPainStatusResponse) error

	// PresentWeatherPoint は地点検索結果を表示します。
	PresentWeatherPoint(data models.GetWeatherPointResponse, kata bool, keyword string) error

	// PresentWeatherStatus は詳細な気象状況を表示します。
	PresentWeatherStatus(data models.GetWeatherStatusResponse, dayOffset int, dayName string) error

	// PresentOtenkiASP は Otenki ASP の気象情報を表示します。
	PresentOtenkiASP(data models.GetOtenkiASPResponse, targetDates []time.Time, cityName, cityCode string) error
}
