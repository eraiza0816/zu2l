package commands

import (
	"fmt"
	"sort"
	"errors"
	"zutool/internal/api"
	"zutool/internal/presenter"

	"github.com/spf13/cobra"
)

// RunWeatherStatus executes the logic for the 'weather_status' command.
func RunWeatherStatus(client *api.Client, pres presenter.Presenter, cmd *cobra.Command, args []string) error {
	// 1. Parse arguments and flags
	cityCode := args[0]
	// TODO: Add validation for cityCode format (e.g., 3 digits)

	nFlag, _ := cmd.Flags().GetIntSlice("n")
	// Validate the -n flag values
	for _, n := range nFlag {
		if n < -1 || n > 2 {
			return fmt.Errorf("無効な日付オフセットです: %d (-1 から 2 の間で指定してください)", n)
		}
	}
	sort.Ints(nFlag) // Sort the days to display them chronologically

	// 2. Call the API client
	res, err := client.GetWeatherStatus(cityCode)
	if err != nil {
		// Handle potential API errors (e.g., city not found)
		var apiErr *api.APIError
		if errors.As(err, &apiErr) {
			// Return specific API errors directly
			return fmt.Errorf("気象状況の取得に失敗しました: %w", err)
		}
		// Wrap other potential errors
		return fmt.Errorf("気象状況の取得中に予期せぬエラーが発生しました: %w", err)
	}

	// 3. Iterate through requested days and call the presenter
	for _, n := range nFlag {
		var dayName string
		switch n {
		case -1:
			dayName = "yesterday"
		case 0:
			dayName = "today"
		case 1:
			dayName = "tomorrow"
		case 2:
			dayName = "dayaftertomorrow"
		default:
			// This case should not be reached due to earlier validation
			return fmt.Errorf("内部エラー: 無効な日付オフセット %d", n)
		}

		// Pass the relevant data slice and day info to the presenter
		err = pres.PresentWeatherStatus(res, n, dayName)
		if err != nil {
			// Log or print error for the specific day, but potentially continue with others?
			// For now, return the first error encountered during presentation.
			return fmt.Errorf("%s の結果表示に失敗しました: %w", dayName, err)
		}
	}

	return nil
}
