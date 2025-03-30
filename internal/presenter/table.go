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
	table.SetTablePadding("\t") // Use tabs to separate columns
	table.SetNoWhiteSpace(true)
	return table
}

// PresentPainStatus displays the pain status forecast as a table.
func (p *TablePresenter) PresentPainStatus(data models.GetPainStatusResponse) error {
	status := data.PainnoterateStatus
	titleA := fmt.Sprintf("今のみんなの体調は? <%s>", status.AreaName)
	titleB := fmt.Sprintf("(集計時間: %s時-%s時台)", status.TimeStart, status.TimeEnd)

	table := p.newTable()
	table.SetHeader([]string{fmt.Sprintf("%s\n%s", titleA, titleB)}) // Use SetHeader for the main title

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

// PresentOtenkiASP displays the Otenki ASP weather information as a table.
func (p *TablePresenter) PresentOtenkiASP(data models.GetOtenkiASPResponse, targetDates []time.Time, cityName, cityCode string) error {
	table := p.newTable()
	table.SetCaption(true, fmt.Sprintf("<%s|%s>の天気情報", cityName, cityCode))

	if len(data.Elements) == 0 {
		fmt.Fprintf(p.ensureWriter(), "表示する天気情報要素がありません。\n")
		return nil
	}

	// Sort elements for consistent column order (optional, based on desired order)
	// sort.Slice(data.Elements, func(i, j int) bool {
	// 	// Define sorting logic if needed, e.g., by ContentID or Title
	// 	return data.Elements[i].ContentID < data.Elements[j].ContentID
	// })

	headers := []string{"日付"}
	for _, element := range data.Elements {
		// Clean up title for header
		title := strings.Replace(element.Title, "(日別)", "", -1)
		headers = append(headers, title)
	}
	table.SetHeader(headers)

	// Sort targetDates to ensure chronological order in rows
	sort.Slice(targetDates, func(i, j int) bool {
		return targetDates[i].Before(targetDates[j])
	})

	if len(targetDates) == 0 {
		fmt.Fprintf(p.ensureWriter(), "表示対象の日付が指定されていません。\n")
		// Optionally, default to showing all available dates if none are specified
		// Or return an error/message as appropriate.
		return nil // Or return an error
	}


	for _, targetDate := range targetDates {
		row := []string{targetDate.Format("01/02")} // Format date as MM/DD

		for _, element := range data.Elements {
			value, ok := element.Records[targetDate]
			if !ok {
				row = append(row, "-") // Use "-" for missing data
				continue
			}

			// Format value based on its type
			switch v := value.(type) {
			case string:
				// Attempt to map weather codes to emojis if it's a weather element
				if strings.Contains(element.ContentID, "day_tenki") {
					weatherCodeInt, err := strconv.Atoi(v)
					if err == nil {
						simplifiedCode := (weatherCodeInt / 100) * 100
						if emoji, okEmoji := models.WeatherEmojiMap[simplifiedCode]; okEmoji {
							row = append(row, emoji) // Show emoji for weather
							continue // Skip default string formatting
						}
					}
				}
				row = append(row, v) // Default string representation
			case float64:
				// Format floats nicely (e.g., temperature, pressure)
				// Add units based on element type if possible/needed
				unit := ""
				if strings.Contains(element.ContentID, "temp") {
					unit = "℃"
				} else if strings.Contains(element.ContentID, "pre") {
					unit = "hPa" // Assuming hPa for pressure
				}
				row = append(row, fmt.Sprintf("%.1f%s", v, unit))
			case int: // Handle potential integer values
				row = append(row, strconv.Itoa(v))
			default:
				// Fallback for other types
				row = append(row, fmt.Sprintf("%v", v))
			}
		}
		table.Append(row)
	}

	table.Render()
	return nil
}

// Helper function min (already present in original main.go)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


// Compile-time check to ensure TablePresenter implements Presenter.
var _ Presenter = (*TablePresenter)(nil)
