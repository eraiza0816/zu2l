package commands

import (
	"fmt"
	"strconv"
	"zutool/internal/api"
	"zutool/internal/models"
	"zutool/internal/presenter"

	"github.com/spf13/cobra"
)

// RunPainStatus executes the logic for the 'pain_status' command.
// It fetches pain status data using the API client and displays it using the presenter.
func RunPainStatus(client *api.Client, pres presenter.Presenter, cmd *cobra.Command, args []string) error {
	// 1. Parse and validate arguments/flags
	areaArg := args[0]
	var areaCode string

	// Check if the argument is a numeric code or a name
	if _, err := strconv.Atoi(areaArg); err == nil && len(areaArg) == 2 { // Basic validation for 2-digit code
		areaCode = areaArg
	} else {
		// Look up the code in the AreaCodeMap
		code, ok := models.AreaCodeMap[areaArg]
		if !ok {
			// Consider suggesting valid names/codes or listing them
			return fmt.Errorf("無効な地域コードまたは地域名です: %s", areaArg)
		}
		areaCode = code
	}

	// Get the optional set_weather_point flag
	setWeatherPointFlag, _ := cmd.Flags().GetString("set_weather_point")
	var setWeatherPoint *string
	if setWeatherPointFlag != "" {
		// TODO: Add validation for the city code format (e.g., 5 digits) if needed
		setWeatherPoint = &setWeatherPointFlag
	}

	// 2. Call the API client to get data
	res, err := client.GetPainStatus(areaCode, setWeatherPoint)
	if err != nil {
		// Return the error from the API client directly
		return fmt.Errorf("痛み予報の取得に失敗しました: %w", err)
	}

	// 3. Pass the data to the presenter
	err = pres.PresentPainStatus(res)
	if err != nil {
		return fmt.Errorf("結果の表示に失敗しました: %w", err)
	}

	return nil
}
