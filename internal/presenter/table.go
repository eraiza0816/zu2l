package presenter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"github.com/eraiza0816/zu2l/internal/models"

	"github.com/olekukonko/tablewriter"
)

// TablePresenter は Presenter インターフェースを実装し、データをテーブル形式で出力します。
type TablePresenter struct {
	Writer io.Writer // 出力先 (nil の場合は os.Stdout)
}

func (p *TablePresenter) ensureWriter() io.Writer {
	if p.Writer == nil {
		return os.Stdout
	}
	return p.Writer
}

// newTable は新しい tablewriter.Table インスタンスを作成し、設定します。
func (p *TablePresenter) newTable() *tablewriter.Table {
	writer := p.ensureWriter()
	table := tablewriter.NewWriter(writer)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	return table
}

// PresentPainStatus は痛み予報ステータスをテーブル形式で表示します。
func (p *TablePresenter) PresentPainStatus(data models.GetPainStatusResponse) error {
	status := data.PainnoterateStatus
	titleA := fmt.Sprintf("今のみんなの体調は? <%s>", status.AreaName)
	titleB := fmt.Sprintf("(集計時間: %s時-%s時台)", status.TimeStart, status.TimeEnd)

	table := p.newTable()
	table.SetHeader([]string{fmt.Sprintf("%s\n%s", titleA, titleB)})

	sicknessEmojis := []string{"😃", "😐", "😞", "🤯"}
	sicknessLabels := []string{"普通", "少し痛い", "痛い", "かなり痛い"}
	rates := []float64{status.RateNormal, status.RateLittle, status.RatePainful, status.RateBad}

	for i, emoji := range sicknessEmojis {
		rate := rates[i]
		repeatCount := 0
		if rate >= 0 {
			// 絵文字の繰り返し回数を割合に基づいて単純にスケーリング (rate / 2)
			repeatCount = int(rate / 2)
		}
		emojiRepeat := strings.Repeat(emoji, repeatCount)
		table.Append([]string{fmt.Sprintf("%s %.0f%%", emojiRepeat, rate)})
	}

	var footerParts []string
	for i, emoji := range sicknessEmojis {
		footerParts = append(footerParts, fmt.Sprintf("%s･･･%s", emoji, sicknessLabels[i]))
	}
	footerText := fmt.Sprintf("[%s]", strings.Join(footerParts, ", "))
	table.SetFooter([]string{footerText}) // 凡例のために SetFooter を使用

	table.Render()
	return nil
}

// PresentWeatherPoint は地点検索結果をテーブル形式で表示します。
func (p *TablePresenter) PresentWeatherPoint(data models.GetWeatherPointResponse, kata bool, keyword string) error {
	table := p.newTable()
	headers := []string{"地域コード", "地域名"}
	if kata {
		headers = append(headers, "地域カナ")
	}
	table.SetHeader(headers)
	table.SetCaption(true, fmt.Sprintf("「%s」の検索結果", keyword))

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
) float64 { // このセグメントの最後の気圧を返す
	table := p.newTable()
	if title != "" {
		table.SetCaption(true, title)
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
	table.SetHeader(headers)

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
			pressureFloat = 0 // 変換エラー時は 0 とする
			fmt.Fprintf(os.Stderr, "警告: 気圧 '%s' の数値変換に失敗しました: %v\n", byTime.Pressure, err)
		}

		// 気圧の変化を示す矢印
		var arrow string
		if (lastPressure == 0 && i == 0 && startHour == 0) || err != nil { // 初回 or 変換エラー
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

// PresentOtenkiASP は Otenki ASP の気象情報を元の複数列形式で表示します。
// ユーザーの理想的なフォーマットに基づいて手動でヘッダー行を出力します。
func (p *TablePresenter) PresentOtenkiASP(data models.GetOtenkiASPResponse, targetDates []time.Time, cityName, cityCode string) error {
	table := p.newTable()
	table.SetCaption(true, fmt.Sprintf("<%s|%s>の天気情報", cityName, cityCode))

	if len(data.Elements) == 0 {
		fmt.Fprintf(p.ensureWriter(), "表示する天気情報要素がありません。\n")
		return nil
	}

	// sort.Slice(data.Elements, func(i, j int) bool { ... }) // 要素ソートはオプション

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
			valueStr := "-" // Default value if missing
			if ok {
				// パーサーが正しい型 (string または float64) を提供することを前提としたフォーマットロジック
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


// コンパイル時チェック: TablePresenter が Presenter インターフェースを実装していることを保証します。
var _ Presenter = (*TablePresenter)(nil)
