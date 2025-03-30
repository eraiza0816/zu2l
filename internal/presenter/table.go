package presenter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"zutool/internal/models"

	"github.com/olekukonko/tablewriter"
)

// TablePresenter implements the Presenter interface to output data in table format.
type TablePresenter struct {
	// Writer specifies the output destination. Defaults to os.Stdout if nil.
	Writer io.Writer
}

// ensureWriter returns the configured writer or os.Stdout if nil.
func (p *TablePresenter) ensureWriter() io.Writer {
	if p.Writer == nil {
		return os.Stdout
	}
	return p.Writer
}

// newTable creates and configures a new tablewriter.Table instance.
// This is a helper function based on the original main.go logic.
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

// PresentPainStatus displays the pain status forecast as a table.
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
		// Ensure rate is non-negative before repeating emoji
		repeatCount := 0
		if rate >= 0 {
			repeatCount = int(rate / 2) // Simple scaling for emoji count
		}
		emojiRepeat := strings.Repeat(emoji, repeatCount)
		table.Append([]string{fmt.Sprintf("%s %.0f%%", emojiRepeat, rate)})
	}

	var footerParts []string
	for i, emoji := range sicknessEmojis {
		footerParts = append(footerParts, fmt.Sprintf("%s･･･%s", emoji, sicknessLabels[i]))
	}
	footerText := fmt.Sprintf("[%s]", strings.Join(footerParts, ", "))
	table.SetFooter([]string{footerText}) // Use SetFooter for the legend

	table.Render()
	return nil
}

// PresentWeatherPoint displays the weather point search results as a table.
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

// renderWeatherStatusSubTable renders a 12-hour segment of the weather status.
// It's a helper function based on the original main.go logic.
func (p *TablePresenter) renderWeatherStatusSubTable(
	dayData []models.WeatherStatusByTime,
	startHour int,
	prevPressure float64,
	title string,
) float64 {
	table := p.newTable()
	if title != "" {
		table.SetCaption(true, title)
	}

	var headers []string
	dataLen := len(dayData)
	if dataLen == 0 {
		fmt.Fprintf(p.ensureWriter(), "データがありません (%d時台)\n", startHour)
		return prevPressure // Return previous pressure if no data
	}

	// Adjust loop to handle potentially less than 12 hours of data
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

		// Use simplified weather code (first digit) for emoji mapping
		weatherCodeInt, err := strconv.Atoi(string(byTime.Weather))
		emoji := "?"
		if err == nil {
			simplifiedCode := (weatherCodeInt / 100) * 100
			if e, ok := models.WeatherEmojiMap[simplifiedCode]; ok {
				emoji = e
			}
		}
		weathers[i] = emoji

		// Handle temperature display (convert string to float)
		if byTime.Temp != nil {
			tempFloat, err := strconv.ParseFloat(*byTime.Temp, 64)
			if err != nil {
				// Handle conversion error
				temps[i] = "?℃" // Indicate error
				fmt.Fprintf(os.Stderr, "警告: 温度 '%s' の数値変換に失敗しました: %v\n", *byTime.Temp, err)
			} else {
				temps[i] = fmt.Sprintf("%.1f℃", tempFloat)
			}
		} else {
			temps[i] = "-℃" // Indicate missing temperature data
		}

		// Convert pressure string to float64 for comparison and display
		pressureFloat, err := strconv.ParseFloat(byTime.Pressure, 64)
		if err != nil {
			// Handle conversion error, e.g., log it or default value
			// For now, default to 0 and skip arrow logic if conversion fails
			pressureFloat = 0 // Or consider using lastPressure?
			fmt.Fprintf(os.Stderr, "警告: 気圧 '%s' の数値変換に失敗しました: %v\n", byTime.Pressure, err)
		}

		var arrow string
		// Handle initial case or conversion error
		if (lastPressure == 0 && i == 0 && startHour == 0) || err != nil {
			arrow = "→" // Or ""
		} else if pressureFloat > lastPressure {
			arrow = "↗"
		} else if pressureFloat < lastPressure {
			arrow = "↘"
		} else {
			arrow = "→"
		}
		// Use the converted float for display
		pressures[i] = fmt.Sprintf("%s\n%.1f", arrow, pressureFloat)
		// Update lastPressure with the converted float
		lastPressure = pressureFloat

		// Map PressureLevelEnum to a more descriptive representation if needed, or keep as is
		pressureLevels[i] = string(byTime.PressureLevel) // Keeping the code for now
	}

	table.Append(weathers)
	table.Append(temps)
	table.Append(pressures)
	table.Append(pressureLevels)
	table.Render()

	return lastPressure
}

// PresentWeatherStatus displays the detailed weather status as tables (12h segments).
func (p *TablePresenter) PresentWeatherStatus(data models.GetWeatherStatusResponse, dayOffset int, dayName string) error {
	var dayData []models.WeatherStatusByTime
	switch dayName { // Use dayName passed from command logic
	case "yesterday":
		dayData = data.Yesterday
	case "today":
		dayData = data.Today
	case "tomorrow":
		dayData = data.Tomorrow
	case "dayaftertomorrow":
		dayData = data.DayAfterTomorrow
	default:
		return fmt.Errorf("invalid dayName provided: %s", dayName)
	}

	if len(dayData) == 0 {
		fmt.Fprintf(p.ensureWriter(), "%s のデータがありません。\n", dayName)
		return nil
	}
	if len(dayData) < 24 {
		fmt.Fprintf(p.ensureWriter(), "警告: %s のデータが24時間分ありません (%d時間分)。利用可能なデータを表示します。\n", dayName, len(dayData))
	}

	// Calculate the date string based on the API response DateTime and the offset
	displayDate := data.DateTime.AddDate(0, 0, dayOffset).Format("2006-01-02")
	title := fmt.Sprintf("<%s|%s>の気圧予報\n%s = %s",
		data.PlaceName, data.PlaceID, dayName, displayDate)

	// Render first 12 hours (0-11)
	prevPressure := p.renderWeatherStatusSubTable(dayData[0:min(12, len(dayData))], 0, 0, title)

	// Render next 12 hours (12-23) if data exists
	if len(dayData) > 12 {
		p.renderWeatherStatusSubTable(dayData[12:min(24, len(dayData))], 12, prevPressure, "")
	}

	return nil
}

// PresentOtenkiASP displays the Otenki ASP weather information in the original multi-column format
// with a manually printed header row based on the user's ideal format.
func (p *TablePresenter) PresentOtenkiASP(data models.GetOtenkiASPResponse, targetDates []time.Time, cityName, cityCode string) error {
	table := p.newTable() // Keep table setup
	table.SetCaption(true, fmt.Sprintf("<%s|%s>の天気情報", cityName, cityCode))

	if len(data.Elements) == 0 {
		fmt.Fprintf(p.ensureWriter(), "表示する天気情報要素がありません。\n")
		return nil
	}

	// --- Data processing logic from the successful multi-column state ---
	// Sort elements (optional, keep commented)
	// sort.Slice(data.Elements, func(i, j int) bool { ... })

	// Sort targetDates
	sort.Slice(targetDates, func(i, j int) bool {
		return targetDates[i].Before(targetDates[j])
	})

	if len(targetDates) == 0 {
		fmt.Fprintf(p.ensureWriter(), "表示対象の日付が指定されていません。\n")
		return nil
	}

	idealHeaderString := "日付\t天気\t降水確率\t最高気温\t最低気温\t最大風速\t最大風速時風向\t気圧予報レベル\t最小湿度"
	fmt.Fprintln(p.ensureWriter(), idealHeaderString)

	for _, targetDate := range targetDates {
		row := []string{targetDate.Format("01/02")} // First column is the date for the row

		for _, element := range data.Elements { // Loop through elements (these become the columns after the date)
			value, ok := element.Records[targetDate]
			valueStr := "-" // Default value if missing
			if ok {
				// Formatting logic based on parser providing correct types (string or float64)
				switch v := value.(type) {
				case string:
					// Attempt emoji mapping for weather code
					if element.ContentID == "day_tenki" { // Check specific ContentID
						weatherCodeInt, err := strconv.Atoi(v)
						if err == nil {
							simplifiedCode := (weatherCodeInt / 100) * 100
							if emoji, okEmoji := models.WeatherEmojiMap[simplifiedCode]; okEmoji {
								valueStr = emoji // Use emoji
							} else {
								valueStr = v // Use code if no emoji
							}
						} else {
							valueStr = v // Use original string if not int code (e.g., weather text)
						}
					} else {
						valueStr = v // Use string directly for other elements (dir, level)
					}
				case float64:
					// Format floats (pop, temps, wind_v, humidity)
					// Format pop, humidity, level, dir as integer if they have no decimal part
					if element.ContentID == "day_pre" || element.ContentID == "low_humidity" || element.ContentID == "zutu_level_day" || element.ContentID == "day_wind_d" {
						if v == float64(int(v)) {
							valueStr = strconv.Itoa(int(v))
						} else {
							valueStr = fmt.Sprintf("%.1f", v) // Keep decimal if present
						}
					} else {
						// Format temps, wind_v to one decimal place
						valueStr = fmt.Sprintf("%.1f", v)
					}
				default:
					// Fallback for unexpected types from parser
					fmt.Fprintf(os.Stderr, "警告: ContentID '%s' の予期しない値の型: %T, 値: %v\n", element.ContentID, v, v)
					valueStr = fmt.Sprintf("%v", v)
				}
			}
			row = append(row, valueStr)
		}
		table.Append(row)
	}

	// Render the table (without its own header)
	table.Render()
	return nil
}

// Helper function min (if not already present or imported)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


// Compile-time check to ensure TablePresenter implements Presenter.
var _ Presenter = (*TablePresenter)(nil)
