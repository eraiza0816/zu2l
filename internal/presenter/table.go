package presenter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"
	"github.com/eraiza0816/zu2l/internal/models"

	"github.com/olekukonko/tablewriter"
)

type TablePresenter struct {
	Writer io.Writer
}

func (p *TablePresenter) ensureWriter() io.Writer {
	if p.Writer == nil {
		return os.Stdout
	}
	return p.Writer
}

func (p *TablePresenter) newTable() *tablewriter.Table {
	writer := p.ensureWriter()
	table := tablewriter.NewWriter(writer)
	return table
}

func (p *TablePresenter) PresentPainStatus(data models.GetPainStatusResponse) error {
	status := data.PainnoterateStatus

	table := p.newTable()

	sicknessLabels := []string{"普通", "少し痛い", "痛い", "かなり痛い"}
	rates := []float64{status.RateNormal, status.RateLittle, status.RatePainful, status.RateBad}

	for i, label := range sicknessLabels {
		rate := rates[i]

		table.Append([]string{fmt.Sprintf("%s: %.0f%%", label, rate)})
	}

	table.Render()
	return nil
}

func (p *TablePresenter) PresentWeatherPoint(data models.GetWeatherPointResponse, kata bool, keyword string) error {
	table := p.newTable()
	headers := []string{"地域コード", "地域名"}
	if kata {
		headers = append(headers, "地域カナ")
	}

	if len(data.Result.Root) == 0 {
		fmt.Fprintf(p.ensureWriter(), "「%s」に一致する地域が見つかりませんでした。\n", keyword)
		return nil
	}

	for _, weatherPoint := range data.Result.Root {
		row := []string{weatherPoint.CityCode, weatherPoint.Name}
		if kata {
			row = append(row, weatherPoint.NameKata)
		}
		table.Append(row)
	}
	table.Render()
	return nil
}

// renderWeatherStatusSubTable は気象状況の12時間分のセグメントを描画します。
// prevPressure は前の12時間セグメントの最後の気圧を保持し、矢印表示に使用します。
// 最後の気圧を返します。
func (p *TablePresenter) renderWeatherStatusSubTable(
	dayData []models.WeatherStatusByTime, // 表示する日の時間別データ
	startHour int,                        // このセグメントの開始時間 (0 または 12)
	prevPressure float64,                 // 前のセグメントの最後の気圧 (矢印表示用)
	title string,                         // テーブルのタイトル (最初のセグメントのみ)
) float64 {
	table := p.newTable()
	if title != "" {
	}

	var headers []string
	dataLen := len(dayData)
	if dataLen == 0 {
		fmt.Fprintf(p.ensureWriter(), "データがありません (%d時台)\n", startHour)
		return prevPressure
	}

	numHours := min(12, dataLen)
	for i := 0; i < numHours; i++ {
		headers = append(headers, strconv.Itoa(startHour+i))
	}

	weathers := make([]string, numHours)
	temps := make([]string, numHours)
	pressures := make([]string, numHours)
	pressureLevels := make([]string, numHours)

	lastPressure := prevPressure
	for i := 0; i < numHours; i++ {
		byTime := dayData[i]

		weatherCodeInt, err := strconv.Atoi(string(byTime.Weather))
		emoji := "?"
		if err == nil {
			simplifiedCode := (weatherCodeInt / 100) * 100
			if e, ok := models.WeatherEmojiMap[simplifiedCode]; ok {
				emoji = e
			}
		}
		weathers[i] = emoji

		if byTime.Temp != nil {
			tempFloat, err := strconv.ParseFloat(*byTime.Temp, 64)
			if err != nil {
				temps[i] = "?℃"
				fmt.Fprintf(os.Stderr, "警告: 温度 '%s' の数値変換に失敗しました: %v\n", *byTime.Temp, err)
			} else {
				temps[i] = fmt.Sprintf("%.1f℃", tempFloat)
			}
		} else {
			temps[i] = "-℃"
		}

		pressureFloat, err := strconv.ParseFloat(byTime.Pressure, 64)
		if err != nil {
			pressureFloat = 0
			fmt.Fprintf(os.Stderr, "警告: 気圧 '%s' の数値変換に失敗しました: %v\n", byTime.Pressure, err)
		}

		var arrow string
		if (lastPressure == 0 && i == 0 && startHour == 0) || err != nil {
			arrow = "→"
		} else if pressureFloat > lastPressure {
			arrow = "↗"
		} else if pressureFloat < lastPressure {
			arrow = "↘"
		} else {
			arrow = "→"
		}
		pressures[i] = fmt.Sprintf("%s\n%.1f", arrow, pressureFloat)
		lastPressure = pressureFloat

		pressureLevels[i] = string(byTime.PressureLevel)
	}

	table.Append(weathers)
	table.Append(temps)
	table.Append(pressures)
	table.Append(pressureLevels)
	table.Render()

	return lastPressure
}

// PresentWeatherStatus は詳細な気象状況をテーブル形式 (12時間ごとのセグメント) で表示します。
func (p *TablePresenter) PresentWeatherStatus(data models.GetWeatherStatusResponse, dayOffset int, dayName string) error {
	var dayData []models.WeatherStatusByTime
	switch dayName {
	case "yesterday":
		dayData = data.Yesterday
	case "today":
		dayData = data.Today
	case "tomorrow":
		dayData = data.Tomorrow
	case "dayaftertomorrow":
		dayData = data.DayAfterTomorrow
	default:
		return fmt.Errorf("無効な dayName が提供されました: %s", dayName)
	}

	if len(dayData) == 0 {
		fmt.Fprintf(p.ensureWriter(), "%s のデータがありません。\n", dayName)
		return nil
	}
	// データが24時間分ない場合に警告を表示
	if len(dayData) < 24 {
		fmt.Fprintf(p.ensureWriter(), "警告: %s のデータが24時間分ありません (%d時間分)。利用可能なデータを表示します。\n", dayName, len(dayData))
	}

	displayDate := data.DateTime.AddDate(0, 0, dayOffset).Format("2006-01-02")
	title := fmt.Sprintf("<%s|%s>の気圧予報\n%s = %s",
		data.PlaceName, data.PlaceID, dayName, displayDate)

	prevPressure := p.renderWeatherStatusSubTable(dayData[0:min(12, len(dayData))], 0, 0, title)

	if len(dayData) > 12 {
		p.renderWeatherStatusSubTable(dayData[12:min(24, len(dayData))], 12, prevPressure, "")
	}

	return nil
}

func (p *TablePresenter) PresentOtenkiASP(data models.GetOtenkiASPResponse, targetDates []time.Time, cityName, cityCode string) error {
	table := p.newTable()

	if len(data.Elements) == 0 {
		fmt.Fprintf(p.ensureWriter(), "表示する天気情報要素がありません。\n")
		return nil
	}

	sort.Slice(targetDates, func(i, j int) bool {
		return targetDates[i].Before(targetDates[j])
	})

	if len(targetDates) == 0 {
		fmt.Fprintf(p.ensureWriter(), "表示対象の日付が指定されていません。\n")
		return nil
	}

	// tablewriter のヘッダーではなく、理想的なヘッダー文字列を手動で出力
	idealHeaderString := "日付\t天気\t降水確率\t最高気温\t最低気温\t最大風速\t最大風速時風向\t気圧予報レベル\t最小湿度"
	fmt.Fprintln(p.ensureWriter(), idealHeaderString)

	for _, targetDate := range targetDates {
		row := []string{targetDate.Format("01/02")}

		for _, element := range data.Elements {
			value, ok := element.Records[targetDate]
			valueStr := "-"
			if ok {
				switch v := value.(type) {
				case string:
					if element.ContentID == "day_tenki" {
						weatherCodeInt, err := strconv.Atoi(v)
						if err == nil {
							simplifiedCode := (weatherCodeInt / 100) * 100
							if emoji, okEmoji := models.WeatherEmojiMap[simplifiedCode]; okEmoji {
								valueStr = emoji
							} else {
								valueStr = v
							}
						} else {
							valueStr = v
						}
					} else {
						valueStr = v
					}
				case float64:
					if element.ContentID == "day_pre" || element.ContentID == "low_humidity" || element.ContentID == "zutu_level_day" || element.ContentID == "day_wind_d" {
						if v == float64(int(v)) {
							valueStr = strconv.Itoa(int(v))
						} else {
							valueStr = fmt.Sprintf("%.1f", v)
						}
					} else {
						valueStr = fmt.Sprintf("%.1f", v)
					}
				default:
					fmt.Fprintf(os.Stderr, "警告: ContentID '%s' の予期しない値の型: %T, 値: %v\n", element.ContentID, v, v)
					valueStr = fmt.Sprintf("%v", v)
				}
			}
			row = append(row, valueStr)
		}
		table.Append(row)
	}

	table.Render()
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


var _ Presenter = (*TablePresenter)(nil)
