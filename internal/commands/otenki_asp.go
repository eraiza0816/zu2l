package commands

import (
	"fmt"
	"sort"
	"time"
	github.com/eraiza0816/zu2l/api" // Update import path
	"github.com/eraiza0816/zu2l/internal/models"
	"github.com/eraiza0816/zu2l/internal/presenter"

	"github.com/spf13/cobra"
)

// RunOtenkiAsp executes the logic for the 'otenki_asp' command.
func RunOtenkiAsp(client *api.Client, pres presenter.Presenter, cmd *cobra.Command, args []string) error {
	// 1. Parse arguments and flags
	cityArg := args[0]
	nFlag, _ := cmd.Flags().GetIntSlice("n")

	// Validate city code/name
	var cityCode string
	var cityName string
	var found bool
	if name, ok := models.ConfirmedOtenkiAspCityCodeMap[cityArg]; ok {
		cityCode = cityArg
		cityName = name
		found = true
	} else {
		for code, name := range models.ConfirmedOtenkiAspCityCodeMap {
			if name == cityArg {
				cityCode = code
				cityName = name
				found = true
				break
			}
		}
	}
	if !found {
		// Provide a helpful error message listing supported values
		supportedValues := make([]string, 0, len(models.ConfirmedOtenkiAspCityCodeMap)*2)
		for code, name := range models.ConfirmedOtenkiAspCityCodeMap {
			supportedValues = append(supportedValues, code, name)
		}
		sort.Strings(supportedValues)
		return fmt.Errorf("無効な都市コードまたは都市名です: '%s' (サポートされている値: %v)", cityArg, supportedValues)
	}

	// Validate the -n flag (day numbers)
	for _, n := range nFlag {
		if n < 0 || n > 6 { // Otenki ASP supports 0-6 days
			return fmt.Errorf("無効な日付オフセットです: %d (0 から 6 の間で指定してください)", n)
		}
	}
	sort.Ints(nFlag) // Sort requested days

	// 2. Call the API client
	res, err := client.GetOtenkiASP(cityCode)
	if err != nil {
		// Handle potential API errors
		return fmt.Errorf("Otenki ASP データの取得に失敗しました: %w", err)
	}

	// 3. Filter target dates based on flags and available data
	var availableDates []time.Time
	if len(res.Elements) > 0 && len(res.Elements[0].Records) > 0 {
		// Assume all elements have the same date keys, get from the first element
		for t := range res.Elements[0].Records {
			availableDates = append(availableDates, t)
		}
		sort.Slice(availableDates, func(i, j int) bool {
			return availableDates[i].Before(availableDates[j])
		})
	} else {
		// Handle case where no data/records are available
		fmt.Println("利用可能な日付データがありません。") // Inform user
		return nil // Or return an error? Decided to just print and exit cleanly.
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

	// If no specific days were requested via -n, or if the requested days resulted in an empty list
	// (e.g., API returned fewer days than requested), consider defaulting or erroring.
	// Current behavior: If targetDates is empty, the presenter will likely show nothing or an empty table.
	// Let's keep this behavior for now.

	// 4. Pass data to the presenter
	err = pres.PresentOtenkiASP(res, targetDates, cityName, cityCode)
	if err != nil {
		return fmt.Errorf("結果の表示に失敗しました: %w", err)
	}

	return nil
}
