package presenter

import (
	"encoding/json"
	"fmt"
	"io" // Added io import
	"os"
	"time"
	"github.com/eraiza0816/zu2l/internal/models"
)

// JSONPresenter は Presenter インターフェースを実装し、データをJSON形式で出力します。
type JSONPresenter struct {
	// Writer は出力先を指定します。nil の場合は os.Stdout がデフォルトで使用されます。
	Writer io.Writer
}

// ensureWriter は設定された Writer を返します。nil の場合は os.Stdout を返します。
func (p *JSONPresenter) ensureWriter() io.Writer {
	if p.Writer == nil {
		return os.Stdout
	}
	return p.Writer
}

// marshalAndPrint はデータをインデント付きのJSONにマーシャリングし、ライターに出力するヘルパーメソッドです。
func (p *JSONPresenter) marshalAndPrint(data interface{}) error {
	// データをJSONバイトにマーシャリング (インデント付き)
	jsonBytes, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("データをJSONにマーシャリングできませんでした: %w", err)
	}
	// JSON文字列を ensureWriter で取得したライターに出力 (改行付き)
	_, err = fmt.Fprintln(p.ensureWriter(), string(jsonBytes))
	if err != nil {
		return fmt.Errorf("JSON出力の書き込みに失敗しました: %w", err)
	}
	return nil
}

// PresentPainStatus は痛み予報データをJSON形式で出力します。
func (p *JSONPresenter) PresentPainStatus(data models.GetPainStatusResponse) error {
	return p.marshalAndPrint(data)
}

// PresentWeatherPoint は地点検索結果データをJSON形式で出力します。
// kata および keyword パラメータはJSON出力では無視されます。
func (p *JSONPresenter) PresentWeatherPoint(data models.GetWeatherPointResponse, kata bool, keyword string) error {
	return p.marshalAndPrint(data)
}

// PresentWeatherStatus は気象状況データをJSON形式で出力します。
// dayOffset および dayName パラメータはJSON出力では無視されます。
func (p *JSONPresenter) PresentWeatherStatus(data models.GetWeatherStatusResponse, dayOffset int, dayName string) error {
	return p.marshalAndPrint(data)
}

// PresentOtenkiASP は Otenki ASP データをJSON形式で出力します。
// targetDates, cityName, cityCode パラメータはJSON出力では無視されます。
func (p *JSONPresenter) PresentOtenkiASP(data models.GetOtenkiASPResponse, targetDates []time.Time, cityName, cityCode string) error {
	return p.marshalAndPrint(data)
}

// コンパイル時チェック: JSONPresenter が Presenter インターフェースを実装していることを保証します。
var _ Presenter = (*JSONPresenter)(nil)
