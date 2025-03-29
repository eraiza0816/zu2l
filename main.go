package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"zutool/api"
	"zutool/models"
)

var (
	jsonFlag bool
)

func newTable(writer *os.File) *tablewriter.Table {
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

func main() {
	rootCmd := &cobra.Command{
		Use:   "zutool",
		Short: "Get info from zutool <https://zutool.jp/>",
		Long:  "A command-line tool to fetch weather and pain forecast information from zutool.jp.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	painStatusCommand := &cobra.Command{
		Use:     "pain_status [area_code]",
		Aliases: []string{"ps"},
		Short:   "Get pain status forecast by prefecture",
		Long:    "Fetches and displays the pain status forecast for the specified prefecture code.",
		Args:    cobra.ExactArgs(1),
		RunE:    runPainStatus,
	}
	painStatusCommand.Flags().StringP("set_weather_point", "s", "", "Set weather point code (e.g., '13113') to get area-specific forecast")
	rootCmd.AddCommand(painStatusCommand)

	weatherPointCommand := &cobra.Command{
		Use:     "weather_point [keyword]",
		Aliases: []string{"wp"},
		Short:   "Search for weather points (locations)",
		Long:    "Searches for weather point locations based on the provided keyword (e.g., city name).",
		Args:    cobra.ExactArgs(1),
		RunE:    runWeatherPoint,
	}
	weatherPointCommand.Flags().BoolP("kata", "k", false, "Include katakana name in the output table")
	rootCmd.AddCommand(weatherPointCommand)

	weatherStatusCommand := &cobra.Command{
		Use:     "weather_status [city_code]",
		Aliases: []string{"ws"},
		Short:   "Get detailed weather status by city",
		Long:    "Fetches and displays detailed weather status (temperature, pressure, etc.) for the specified city code.",
		Args:    cobra.ExactArgs(1),
		RunE:    runWeatherStatus,
	}
	weatherStatusCommand.Flags().IntSliceP("n", "n", []int{0}, "Specify day number(s) to show (-1 to 2)")
	rootCmd.AddCommand(weatherStatusCommand)

	otenkiAspCommand := &cobra.Command{
		Use:     "otenki_asp [city_code]",
		Aliases: []string{"oa"},
		Short:   "Get weather information from Otenki ASP",
		Long:    "Fetches various weather forecast elements (weather, temp, wind, etc.) from the Otenki ASP service for specific major city codes.",
		Args:    cobra.ExactArgs(1),
		RunE:    runOtenkiAsp,
	}
	otenkiAspCommand.Flags().IntSliceP("n", "n", []int{0, 1, 2, 3, 4, 5, 6}, "Specify forecast day number(s) to show (0 to 6)")
	rootCmd.AddCommand(otenkiAspCommand)

	rootCmd.PersistentFlags().BoolVarP(&jsonFlag, "json", "j", false, "Output results in JSON format")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runPainStatus(cmd *cobra.Command, args []string) error {
	areaArg := args[0]
	var areaCode string

	if _, err := strconv.Atoi(areaArg); err == nil {
		// TODO: Add validation to ensure it's a valid 2-digit code if needed
		areaCode = areaArg
	} else {
		code, ok := models.AreaCodeMap[areaArg]
		if !ok {
			return fmt.Errorf("無効な地域コードまたは地域名です: %s", areaArg)
		}
		areaCode = code
	}

	setWeatherPointFlag, _ := cmd.Flags().GetString("set_weather_point")
	var setWeatherPoint *string
	if setWeatherPointFlag != "" {
		setWeatherPoint = &setWeatherPointFlag
	}

	res, err := api.GetPainStatus(areaCode, setWeatherPoint)
	if err != nil {
		return err
	}

	if jsonFlag {
		jsonBytes, _ := json.MarshalIndent(res, "", "    ")
		fmt.Println(string(jsonBytes))
		return nil
	}

	status := res.PainnoterateStatus
	titleA := fmt.Sprintf("今のみんなの体調は? <%s>", status.AreaName)
	titleB := fmt.Sprintf("(集計時間: %s時-%s時台)", status.TimeStart, status.TimeEnd)

	table := newTable(os.Stdout)
	table.SetHeader([]string{fmt.Sprintf("%s\n%s", titleA, titleB)})

	sicknessEmojis := []string{"😃", "😐", "😞", "🤯"}
	sicknessLabels := []string{"普通", "少し痛い", "痛い", "かなり痛い"}
	rates := []float64{status.RateNormal, status.RateLittle, status.RatePainful, status.RateBad}

	for i, emoji := range sicknessEmojis {
		rate := rates[i]
		emojiRepeat := strings.Repeat(emoji, int(rate/2))
		table.Append([]string{fmt.Sprintf("%s %.0f%%", emojiRepeat, rate)})
	}

	var footerParts []string
	for i, emoji := range sicknessEmojis {
		footerParts = append(footerParts, fmt.Sprintf("%s･･･%s", emoji, sicknessLabels[i]))
	}
	footerText := fmt.Sprintf("[%s]", strings.Join(footerParts, ", "))
	footerSlice := []string{footerText}
	table.SetFooter(footerSlice)

	table.Render()
	return nil
}

func runWeatherPoint(cmd *cobra.Command, args []string) error {
	keyword := args[0]
	kataFlag, _ := cmd.Flags().GetBool("kata")

	res, err := api.GetWeatherPoint(keyword)
	if err != nil {
		return err
	}

	if jsonFlag {
		jsonBytes, _ := json.MarshalIndent(res, "", "    ")
		fmt.Println(string(jsonBytes))
		return nil
	}

	table := newTable(os.Stdout)
	table.SetHeader([]string{"地域コード", "地域名"})
	if kataFlag {
		table.SetHeader([]string{"地域コード", "地域名", "地域カナ"})
	}
	table.SetCaption(true, fmt.Sprintf("「%s」の検索結果", keyword))

	for _, weatherPoint := range res.Result.Root {
		row := []string{weatherPoint.CityCode, weatherPoint.Name}
		if kataFlag {
			row = append(row, weatherPoint.NameKata)
		}
		table.Append(row)
	}
	table.Render()
	return nil
}

func renderWeatherStatusTable(
	writer *os.File,
	dayData []models.WeatherStatusByTime,
	startHour int,
	prevPressure float64,
	title string,
) float64 {
	table := newTable(writer)
	if title != "" {
		table.SetCaption(true, title)
	}

	var headers []string
	for i := startHour; i < startHour+12; i++ {
		headers = append(headers, strconv.Itoa(i))
	}
	table.SetHeader(headers)

	weathers := make([]string, 12)
	temps := make([]string, 12)
	pressures := make([]string, 12)
	pressureLevels := make([]string, 12)

	lastPressure := prevPressure
	for i, byTime := range dayData {
		idx := i

		weatherCode, err := strconv.Atoi(string(byTime.Weather))
		if err == nil {
			weathers[idx] = models.WeatherEmojiMap[weatherCode/100*100]
		} else {
			weathers[idx] = "?"
		}

		if byTime.Temp != nil {
			temps[idx] = fmt.Sprintf("%.1f℃", *byTime.Temp)
		} else {
			temps[idx] = "-℃"
		}

		pressure := byTime.Pressure
		var arrow string
		if lastPressure == 0 {
			arrow = "→"
		} else if pressure > lastPressure {
			arrow = "↗"
		} else if pressure < lastPressure {
			arrow = "↘"
		} else {
			arrow = "→"
		}
		pressures[idx] = fmt.Sprintf("%s\n%.1f", arrow, pressure)
		lastPressure = pressure

		// TODO: Map PressureLevelEnum value to a descriptive name if needed
		pressureLevels[idx] = string(byTime.PressureLevel)
	}

	table.Append(weathers)
	table.Append(temps)
	table.Append(pressures)
	table.Append(pressureLevels)
	table.Render()

	return lastPressure
}

func runWeatherStatus(cmd *cobra.Command, args []string) error {
	cityCode := args[0]
	nFlag, _ := cmd.Flags().GetIntSlice("n")

	for _, n := range nFlag {
		if n < -1 || n > 2 {
			return fmt.Errorf("invalid choice for -n: %d (must be between -1 and 2)", n)
		}
	}

	res, err := api.GetWeatherStatus(cityCode)
	if err != nil {
		return err
	}

	if jsonFlag {
		jsonBytes, _ := json.MarshalIndent(res, "", "    ")
		fmt.Println(string(jsonBytes))
		return nil
	}

	sort.Ints(nFlag)

	for _, n := range nFlag {
		var dayData []models.WeatherStatusByTime
		var dayName string
		var dayOffset int

		switch n {
		case -1:
			dayData = res.Yesterday
			dayName = "yesterday"
			dayOffset = -1
		case 0:
			dayData = res.Today
			dayName = "today"
			dayOffset = 0
		case 1:
			dayData = res.Tomorrow
			dayName = "tomorrow"
			dayOffset = 1
		case 2:
			dayData = res.DayAfterTomorrow
			dayName = "dayaftertomorrow"
			dayOffset = 2
		}

		if len(dayData) < 24 {
			fmt.Fprintf(os.Stderr, "Warning: Unexpected data length for %s (%d hours), expected 24. Rendering available data.\n", dayName, len(dayData))
		}

		title := fmt.Sprintf("<%s|%s>の気圧予報\n%s = %s",
			res.PlaceName, cityCode, dayName,
			res.DateTime.AddDate(0, 0, dayOffset).Format("2006-01-02"))

		prevPressure := renderWeatherStatusTable(os.Stdout, dayData[0:min(12, len(dayData))], 0, 0, title)

		if len(dayData) > 12 {
			renderWeatherStatusTable(os.Stdout, dayData[12:min(24, len(dayData))], 12, prevPressure, "")
		}
	}
	return nil
}

func runOtenkiAsp(cmd *cobra.Command, args []string) error {
	cityCode := args[0]
	nFlag, _ := cmd.Flags().GetIntSlice("n")

	if _, ok := models.ConfirmedOtenkiAspCityCodeMap[cityCode]; !ok {
		return fmt.Errorf("invalid choice for city_code: '%s' (supported codes: %v)", cityCode, getMapKeys(models.ConfirmedOtenkiAspCityCodeMap))
	}
	for _, n := range nFlag {
		if n < 0 || n > 6 {
			return fmt.Errorf("invalid choice for -n: %d (must be between 0 and 6)", n)
		}
	}

	res, err := api.GetOtenkiASP(cityCode)
	if err != nil {
		return err
	}

	if jsonFlag {
		jsonBytes, _ := json.MarshalIndent(res, "", "    ")
		fmt.Println(string(jsonBytes))
		return nil
	}

	table := newTable(os.Stdout)

	cityName := models.ConfirmedOtenkiAspCityCodeMap[cityCode]
	table.SetCaption(true, fmt.Sprintf("<%s|%s>の天気情報", cityName, cityCode))

	headers := []string{"日付"}
	elementOrder := make([]models.Element, len(res.Elements))
	copy(elementOrder, res.Elements)
	for _, element := range elementOrder {
		title := strings.Replace(element.Title, "(日別)", "", -1)
		headers = append(headers, title)
	}
	table.SetHeader(headers)

	var availableDates []time.Time
	if len(elementOrder) > 0 {
		for t := range elementOrder[0].Records {
			availableDates = append(availableDates, t)
		}
		sort.Slice(availableDates, func(i, j int) bool {
			return availableDates[i].Before(availableDates[j])
		})
	}

	targetDateIndices := make(map[int]bool)
	for _, n := range nFlag {
		targetDateIndices[n] = true
	}
	var targetDates []time.Time
	for i, date := range availableDates {
		if targetDateIndices[i] {
			targetDates = append(targetDates, date)
		}
	}

	for _, targetDate := range targetDates {
		row := []string{targetDate.Format("01/02")}

		for _, element := range elementOrder {
			value, ok := element.Records[targetDate]
			if !ok {
				row = append(row, "-")
				continue
			}

			switch v := value.(type) {
			case string:
				// TODO: Potentially map WeatherEnum code to name/emoji here if desired.
				row = append(row, v)
			case float64:
				row = append(row, fmt.Sprintf("%.1f", v))
			default:
				row = append(row, fmt.Sprintf("%v", v))
			}
		}
		table.Append(row)
	}

	table.Render()
	return nil
}

func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
